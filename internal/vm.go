package internal

import (
	"errors"
	"fmt"
	"io"
	"log"
	"slices"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/internal/datast/stacks"
)

// the cost of a single include against the stack limit.
const includeRecursionConst = 10

// the cost of a single macro call against the stack limit.
const macroRecursionConst = 5

func prepareBlocks(blocks map[string]Instructions) map[string]*blockStack {
	rv := make(map[string]*blockStack, len(blocks))
	for name, insts := range blocks {
		rv[name] = &blockStack{instrs: []Instructions{insts}}
	}
	return rv
}

type virtualMachine struct {
	env *Environment
}

func newVirtualMachine(env *Environment) *virtualMachine {
	return &virtualMachine{env: env}
}

func (m *virtualMachine) eval(instructions Instructions, root Value, blocks map[string]Instructions, out *Output, escape AutoEscape) (option.Option[Value], error) {
	state := State{
		env:          m.env,
		ctx:          *newContext(*newFrame(root)),
		autoEscape:   escape,
		instructions: instructions,
		blocks:       prepareBlocks(blocks),
	}
	return m.evalState(&state, out)
}

func (m *virtualMachine) evalMacro(insts Instructions, pc uint, closure Value,
	caller option.Option[Value], out *Output, state *State, args []Value) (option.Option[Value], error) {
	ctx := newContext(*newFrame(closure))
	if caller.IsSome() {
		ctx.store("caller", caller.Unwrap())
	}
	if err := ctx.incrDepth(state.ctx.depth() + macroRecursionConst); err != nil {
		return option.None[Value](), err
	}

	return m.evalImpl(&State{
		env:             m.env,
		ctx:             *ctx,
		currentBlock:    option.None[string](),
		autoEscape:      state.autoEscape,
		instructions:    insts,
		blocks:          make(map[string]*blockStack),
		loadedTemplates: *hashset.NewStrHashSet(),
		macros:          state.macros, // TODO: clone
	}, out, &args, pc)
}

func (m *virtualMachine) evalState(state *State, out *Output) (option.Option[Value], error) {
	var stack []Value
	return m.evalImpl(state, out, &stack, 0)
}

func (m *virtualMachine) evalImpl(state *State, out *Output, stack *[]Value, pc uint) (option.Option[Value], error) {
	initialAutoEscape := state.autoEscape
	undefinedBehavior := state.undefinedBehavior()
	autoEscapeStack := []AutoEscape{}
	nextRecursionJump := option.None[recursionJump]()
	loadedFilters := [MaxLocals]option.Option[FilterFunc]{}
	loadedTests := [MaxLocals]option.Option[TestFunc]{}

	// If we are extending we are holding the instructions of the target parent
	// template here.  This is used to detect multiple extends and the evaluation
	// uses these instructions when it makes it to the end of the instructions.
	parentInstructions := option.None[Instructions]()

	recurseLoop := func(capture bool) error {
		jumpTarget, err := m.prepareLoopRecursion(state)
		// log.Printf("recurseLoop capture=%v, jumpTarget=%d, err=%v", capture, jumpTarget, err)
		if err != nil {
			return processErr(err, pc, state)
		}
		// the way this works is that we remember the next instruction
		// as loop exit jump target.  Whenever a loop is pushed, it
		// memorizes the value in `next_loop_iteration_jump` to jump
		// to.
		nextRecursionJump = option.Some(recursionJump{
			target:     pc + 1,
			endCapture: capture,
		})
		pc = jumpTarget
		return nil
	}

loop:
	for {
		var inst Instruction
		if pc < uint(len(state.instructions.Instructions())) {
			inst = state.instructions.Instructions()[pc]
		} else {
			// when an extends statement appears in a template, when we hit the
			// last instruction we need to check if parent instructions were
			// stashed away (which means we found an extends tag which invoked
			// `LoadBlocks`).  If we do find instructions, we reset back to 0
			// from the new instructions.
			if parentInstructions.IsSome() {
				state.instructions = parentInstructions.Unwrap()
				parentInstructions = option.None[Instructions]()
			} else {
				break loop
			}
			out.endCapture(AutoEscapeNone{})
			pc = 0
			continue
		}
		// log.Printf("evalImpl pc=%d, instr=%s %+v", pc, inst.Typ(), inst)

		var a, b Value

		switch inst := inst.(type) {
		case EmitRawInstruction:
			if _, err := io.WriteString(out, inst.Val); err != nil {
				return option.None[Value](), err
			}
		case EmitInstruction:
			v := stacks.Pop(stack)
			if err := m.env.format(v, state, out); err != nil {
				return option.None[Value](), err
			}
		case StoreLocalInstruction:
			state.ctx.store(inst.Name, stacks.Pop(stack))
		case LookupInstruction:
			var v Value
			if val := state.lookup(inst.Name); val.IsSome() {
				v = val.Unwrap()
			} else {
				v = Undefined
			}
			stacks.Push(stack, v)
		case GetAttrInstruction:
			a = stacks.Pop(stack)
			// This is a common enough operation that it's interesting to consider a fast
			// path here.  This is slightly faster than the regular attr lookup because we
			// do not need to pass down the error object for the more common success case.
			// Only when we cannot look up something, we start to consider the undefined
			// special case.
			if v := a.GetAttrFast(inst.Name); v.IsSome() {
				if v, err := assertValid(v.Unwrap(), pc, state); err != nil {
					return option.None[Value](), err
				} else {
					stacks.Push(stack, v)
				}
			} else {
				if v, err := undefinedBehavior.HandleUndefined(a.IsUndefined()); err != nil {
					return option.None[Value](), processErr(err, pc, state)
				} else {
					stacks.Push(stack, v)
				}
			}
		case GetItemInstruction:
			a = stacks.Pop(stack)
			b = stacks.Pop(stack)
			if v := b.GetItemOpt(a); v.IsSome() {
				if v, err := assertValid(v.Unwrap(), pc, state); err != nil {
					return option.None[Value](), err
				} else {
					stacks.Push(stack, v)
				}
			} else {
				if v, err := undefinedBehavior.HandleUndefined(b.IsUndefined()); err != nil {
					return option.None[Value](), processErr(err, pc, state)
				} else {
					stacks.Push(stack, v)
				}
			}
		case SliceInstruction:
			step := stacks.Pop(stack)
			stop := stacks.Pop(stack)
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if a.IsUndefined() && undefinedBehavior == UndefinedBehaviorStrict {
				return option.None[Value](), processErr(NewError(UndefinedError, ""), pc, state)
			}
			if s, err := Slice(a, b, stop, step); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				stacks.Push(stack, s)
			}
		case LoadConstInstruction:
			stacks.Push(stack, inst.Val)
		case BuildMapInstruction:
			m := NewIndexMapWithCapacity(inst.PairCount)
			for i := uint(0); i < inst.PairCount; i++ {
				val := stacks.Pop(stack)
				key := stacks.Pop(stack)
				m.Set(KeyRefFromValue(key), val)
			}
			stacks.Push(stack, ValueFromIndexMap(m))
		case BuildKwargsInstruction:
			m := NewIndexMapWithCapacity(inst.PairCount)
			for i := uint(0); i < inst.PairCount; i++ {
				val := stacks.Pop(stack)
				key := stacks.Pop(stack)
				m.Set(KeyRefFromValue(key), val)
			}
			stacks.Push[[]Value, Value](stack, mapValue{m: m, mapTyp: mapTypeKwargs})
		case BuildListInstruction:
			v := make([]Value, 0, untrustedSizeHint(inst.Count))
			for i := uint(0); i < inst.Count; i++ {
				v = append(v, stacks.Pop(stack))
			}
			slices.Reverse(v)
			stacks.Push(stack, ValueFromSlice(v))
		case UnpackListInstruction:
			if err := m.unpackList(stack, inst.Count); err != nil {
				return option.None[Value](), err
			}
		case ListAppendInstruction:
			a = stacks.Pop(stack)
			// this intentionally only works with actual sequences
			if v, ok := stacks.Pop(stack).(SeqValue); ok {
				v.Append(a)
				stacks.Push[[]Value, Value](stack, v)
			} else {
				err := NewError(InvalidOperation, "cannot append to non-list")
				return option.None[Value](), processErr(err, pc, state)
			}
		case AddInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := Add(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case SubInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := Sub(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case MulInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := Mul(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case DivInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := Div(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case IntDivInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := IntDiv(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case RemInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := Rem(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case PowInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := Pow(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case EqInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(Equal(a, b)))
		case NeInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(!Equal(a, b)))
		case GtInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(Cmp(a, b) > 0))
		case GteInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(Cmp(a, b) >= 0))
		case LtInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(Cmp(a, b) < 0))
		case LteInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(Cmp(a, b) <= 0))
		case NotInstruction:
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(!a.IsTrue()))
		case StringConcatInstruction:
			a = stacks.Pop(stack)
			b = stacks.Pop(stack)
			v := StringConcat(b, a)
			stacks.Push(stack, v)
		case InInstruction:
			a = stacks.Pop(stack)
			b = stacks.Pop(stack)
			// the in-operator can fail if the value is undefined and
			// we are in strict mode.
			if err := state.undefinedBehavior().AssertIterable(a); err != nil {
				return option.None[Value](), err
			}
			rv, err := Contains(a, b)
			if err != nil {
				return option.None[Value](), err
			}
			stacks.Push(stack, rv)
		case NegInstruction:
			a = stacks.Pop(stack)
			if v, err := Neg(a); err != nil {
				return option.None[Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case PushWithInstruction:
			if err := state.ctx.pushFrame(*newFrameDefault()); err != nil {
				return option.None[Value](), err
			}
		case PopFrameInstruction:
			if optLoopCtx := state.ctx.popFrame().currentLoop; optLoopCtx.IsSome() {
				loopCtx := optLoopCtx.Unwrap()
				if loopCtx.currentRecursionJump.IsSome() {
					recurJump := loopCtx.currentRecursionJump.Unwrap()
					loopCtx.currentRecursionJump = option.None[recursionJump]()
					pc = recurJump.target
					if recurJump.endCapture {
						stacks.Push(stack, out.endCapture(state.autoEscape))
					}
					continue
				}
			}
		case IsUndefinedInstruction:
			a = stacks.Pop(stack)
			stacks.Push(stack, ValueFromBool(a.IsUndefined()))
		case PushLoopInstruction:
			a = stacks.Pop(stack)
			if err := m.pushLoop(state, a, inst.Flags, pc, nextRecursionJump); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
		case IterateInstruction:
			var l *loopState
			if mayLoopState := state.ctx.currentLoop(); mayLoopState.IsSome() {
				l = mayLoopState.Unwrap()
			} else {
				panic("no currentLoop")
			}
			l.object.idx++
			next := option.None[Value]()
			triple := &l.object.valueTriple
			triple[0] = triple[1]
			triple[1] = triple[2]
			triple[2] = l.iterator.Next()
			if triple[1].IsSome() {
				next = option.Some(triple[1].Unwrap().Clone())
			}
			if next.IsSome() {
				item := next.Unwrap()
				if v, err := assertValid(item, pc, state); err != nil {
					return option.None[Value](), err
				} else {
					stacks.Push(stack, v)
				}
			} else {
				pc = inst.JumpTarget
				continue
			}
		case PushDidNotIterateInstruction:
			l := state.ctx.currentLoop().Unwrap()
			stacks.Push(stack, ValueFromBool(l.object.idx == 0))
		case JumpInstruction:
			pc = inst.JumpTarget
			continue
		case JumpIfFalseInstruction:
			a = stacks.Pop(stack)
			if !a.IsTrue() {
				pc = inst.JumpTarget
				continue
			}
		case JumpIfFalseOrPopInstruction:
			if a, ok := stacks.Peek(*stack); ok {
				if a.IsTrue() {
					stacks.Pop(stack)
				} else {
					pc = inst.JumpTarget
					continue
				}
			} else {
				panic("unreachable")
			}
		case JumpIfTrueOrPopInstruction:
			if a, ok := stacks.Peek(*stack); ok {
				if a.IsTrue() {
					pc = inst.JumpTarget
					continue
				} else {
					stacks.Pop(stack)
				}
			} else {
				panic("unreachable")
			}
		case CallBlockInstruction:
			if parentInstructions.IsNone() && !out.isDiscarding() {
				m.callBlock(inst.Name, state, out)
			}
		case PushAutoEscapeInstruction:
			a = stacks.Pop(stack)
			stacks.Push(&autoEscapeStack, state.autoEscape)
			if escape, err := m.deriveAutoEscape(a, initialAutoEscape); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				state.autoEscape = escape
			}
		case PopAutoEscapeInstruction:
			if autoEscape, ok := stacks.TryPop(&autoEscapeStack); ok {
				state.autoEscape = autoEscape
			} else {
				panic("unreachable")
			}
		case BeginCaptureInstruction:
			out.beginCapture(inst.Mode)
		case EndCaptureInstruction:
			stacks.Push(stack, out.endCapture(state.autoEscape))
		case ApplyFilterInstruction:
			f := func() option.Option[FilterFunc] { return state.env.getFilter(inst.Name) }
			var tf FilterFunc
			if optVal := getOrLookupLocal(loadedFilters[:], inst.LocalID, f); optVal.IsSome() {
				tf = optVal.Unwrap()
			} else {
				err := NewError(UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[Value](), processErr(err, pc, state)
			}
			args := stacks.SliceTop(*stack, inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				stacks.DropTop(stack, inst.ArgCount)
				stacks.Push(stack, rv)
			}
		case PerformTestInstruction:
			f := func() option.Option[TestFunc] { return state.env.getTest(inst.Name) }
			var tf TestFunc
			if optVal := getOrLookupLocal(loadedTests[:], inst.LocalID, f); optVal.IsSome() {
				tf = optVal.Unwrap()
			} else {
				err := NewError(UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[Value](), processErr(err, pc, state)
			}
			args := stacks.SliceTop(*stack, inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				stacks.DropTop(stack, inst.ArgCount)
				stacks.Push(stack, ValueFromBool(rv))
			}
		case CallFunctionInstruction:
			if inst.Name == "super" {
				if inst.ArgCount != 0 {
					err := NewError(InvalidOperation, "super() takes no arguments")
					return option.None[Value](), processErr(err, pc, state)
				}
				panic("not implemented")
				// stacks.Push(stack)
			} else if inst.Name == "loop" {
				if inst.ArgCount != 1 {
					err := NewError(InvalidOperation,
						fmt.Sprintf("loop() takes one argument, got %d", inst.ArgCount))
					return option.None[Value](), processErr(err, pc, state)
				}
				// leave the one argument on the stack for the recursion
				if err := recurseLoop(true); err != nil {
					return option.None[Value](), err
				}
				continue
			} else if optFunc := state.lookup(inst.Name); optFunc.IsSome() {
				f := optFunc.Unwrap()
				args := stacks.SliceTop(*stack, inst.ArgCount)
				a, err := f.Call(state, args)
				if err != nil {
					return option.None[Value](), err
				}
				stacks.DropTop(stack, inst.ArgCount)
				stacks.Push(stack, a)
			} else {
				err := NewError(UnknownFunction, fmt.Sprintf("%s is unknown", inst.Name))
				return option.None[Value](), processErr(err, pc, state)
			}
		case CallMethodInstruction:
			args := stacks.SliceTop(*stack, inst.ArgCount)
			log.Printf("stack[len(stack)-1]=%p, args[0]=%p", &(*stack)[len(*stack)-1], &args[0])
			a, err := args[0].CallMethod(state, inst.Name, args[1:])
			if err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
			stacks.DropTop(stack, inst.ArgCount)
			stacks.Push(stack, a)
		case CallObjectInstruction:
			panic("not implemented for CallObjectInstruction")
		case DupTopInstruction:
			if val, ok := stacks.Peek(*stack); ok {
				stacks.Push(stack, val.Clone())
			} else {
				panic("stack must not be empty")
			}
		case DiscardTopInstruction:
			stacks.Pop(stack)
		case FastSuperInstruction:
			if _, err := m.performSuper(state, out, false); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
		case FastRecurseInstruction:
			if err := recurseLoop(false); err != nil {
				return option.None[Value](), err
			}
			continue
		case LoadBlocksInstruction:
			// Explanation on the behavior of `LoadBlocks` and rendering of
			// inherited templates:
			//
			// MiniJinja inherits the behavior from Jinja2 where extending
			// loads the blocks (`LoadBlocks`) and the rest of the template
			// keeps executing but with output disabled, only at the end the
			// parent template is then invoked.  This has the effect that
			// you can still set variables or declare macros and that they
			// become visible in the blocks.
			//
			// This behavior has a few downsides.  First of all what happens
			// in the parent template overrides what happens in the child.
			// For instance if you declare a macro named `foo` after `{%
			// extends %}` and then a variable with that named is also set
			// in the parent template, then you won't be able to call that
			// macro in the body.
			//
			// The reason for this is that blocks unlike macros do not have
			// closures in Jinja2/MiniJinja.
			//
			// However for the common case this is convenient because it
			// lets you put some imports there and for as long as you do not
			// create name clashes this works fine.
			a = stacks.Pop(stack)
			if parentInstructions.IsSome() {
				err := NewError(InvalidOperation, "tried to extend a second time in a template")
				return option.None[Value](), processErr(err, pc, state)
			}
			insts, err := m.loadBlocks(a, state)
			if err != nil {
				return option.None[Value](), err
			}
			parentInstructions = option.Some(insts)
			out.beginCapture(CaptureModeDiscard)
		case IncludeInstruction:
			a = stacks.Pop(stack)
			if err := m.performInclude(a, state, out, inst.IgnoreMissing); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
		case ExportLocalsInstruction:
			locals := state.ctx.currentLocals()
			module := NewIndexMapWithCapacity(uint(len(*locals)))
			for key, value := range *locals {
				module.Set(KeyRefFromString(key), value.Clone())
				// TODO: Use KeyRefFromValue instead of KeyRefFromString
				// module.Set(KeyRefFromValue(ValueFromString(key)), value.Clone())
			}
			stacks.Push(stack, ValueFromIndexMap(module))
		case BuildMacroInstruction:
			m.buildMacro(stack, state, inst.Offset, inst.Name, inst.Flags)
		case ReturnInstruction:
			break loop
		case EncloseInstruction:
			state.ctx.enclose(state.env, inst.Name)
		case GetClosureInstruction:
			closure := state.ctx.closure()
			stacks.Push(stack, ValueFromObject(&closure))
		default:
			panic(fmt.Sprintf("not implemented for instruction %s", inst.Typ()))
		}
		pc++
	}
	if v, ok := stacks.TryPop(stack); ok {
		return option.Some(v), nil
	}
	return option.None[Value](), nil
}

func (m *virtualMachine) performInclude(name Value, state *State, out *Output, ignoreMissing bool) error {
	var choices SeqObject
	if optChoices := name.AsSeq(); optChoices.IsSome() {
		choices = optChoices.Unwrap()
	} else {
		choices = newSliceSeqObject([]Value{name})
	}

	var templatesTried []Value
	l := choices.ItemCount()
	for i := uint(0); i < l; i++ {
		choice := choices.GetItem(i).Unwrap()
		optName := choice.AsStr()
		if optName.IsNone() {
			return NewError(InvalidOperation, "template name was not a string")
		}
		tmpl, err := m.env.GetTemplate(optName.Unwrap())
		if err != nil {
			var er *Error
			if errors.As(err, &er) && er.typ == TemplateNotFound {
				stacks.Push(&templatesTried, choice)
			} else {
				return err
			}
			continue
		}

		newInsts, newBlocks, err := tmpl.instructionsAndBlocks()
		if err != nil {
			return err
		}
		oldEscape := state.autoEscape
		state.autoEscape = tmpl.initialAutoEscape
		oldInsts := state.instructions
		state.instructions = newInsts
		oldBlocks := state.blocks
		state.blocks = prepareBlocks(newBlocks)
		oldClosure := state.ctx.takeClosure()
		if err := state.ctx.incrDepth(includeRecursionConst); err != nil {
			return err
		}
		_, err = m.evalState(state, out)
		state.ctx.resetClosure(oldClosure)
		state.ctx.decrDepth(includeRecursionConst)
		state.autoEscape = oldEscape
		state.instructions = oldInsts
		state.blocks = oldBlocks
		if err != nil {
			return NewError(BadInclude, fmt.Sprintf("error in \"%s\"", tmpl.name())).WithSource(err)
		}
		return nil
	}

	if len(templatesTried) != 0 && !ignoreMissing {
		var detail string
		if len(templatesTried) == 1 {
			detail = fmt.Sprintf("tried to include non-existing template %v", templatesTried[0])
		} else {
			detail = fmt.Sprintf("tried to include one of multiple templates, none of which existed %s", templatesTried)
		}
		return NewError(TemplateNotFound, detail)
	}
	return nil
}

func (m *virtualMachine) performSuper(state *State, out *Output, capture bool) (Value, error) {
	if state.currentBlock.IsNone() {
		return nil, NewError(InvalidOperation, "cannot super outside of block")
	}
	name := state.currentBlock.Unwrap()

	blockStack := state.blocks[name]
	if !blockStack.push() {
		return nil, NewError(InvalidOperation, "no parent block exists")
	}

	if capture {
		out.beginCapture(CaptureModeCapture)
	}

	oldInsts := state.instructions
	state.instructions = blockStack.instructions()
	if err := state.ctx.pushFrame(*newFrameDefault()); err != nil {
		return nil, err
	}
	_, err := m.evalState(state, out)
	state.ctx.popFrame()
	state.instructions = oldInsts
	state.blocks[name].pop()
	if err != nil {
		return nil, NewError(EvalBlock, "error in super block").WithSource(err)
	}
	if capture {
		return out.endCapture(state.autoEscape), nil
	}
	return Undefined, nil
}

func untrustedSizeHint(val uint) uint {
	return min(val, 1024)
}

func (m *virtualMachine) prepareLoopRecursion(state *State) (uint, error) {
	if optLoopState := state.ctx.currentLoop(); optLoopState.IsSome() {
		loopCtx := optLoopState.Unwrap()
		if loopCtx.recurseJumpTarget.IsSome() {
			return loopCtx.recurseJumpTarget.Unwrap(), nil
		}
		return 0, NewError(InvalidOperation, "cannot recurse outside of recursive loop")
	}
	return 0, NewError(InvalidOperation, "cannot recurse outside of loop")
}

func (m *virtualMachine) loadBlocks(name Value, state *State) (Instructions, error) {
	optName := name.AsStr()
	if optName.IsNone() {
		return Instructions{}, NewError(InvalidOperation, "template name was not a string")
	}
	strName := optName.Unwrap()
	if state.loadedTemplates.Contains(strName) {
		return Instructions{}, NewError(InvalidOperation,
			fmt.Sprintf("cycle in template inheritance. %s was referenced more than once", name))
	}
	tmpl, err := m.env.GetTemplate(strName)
	if err != nil {
		return Instructions{}, err
	}
	newInsts, newBlocks, err := tmpl.instructionsAndBlocks()
	if err != nil {
		return Instructions{}, err
	}
	for strName, insts := range newBlocks {
		if _, ok := state.blocks[strName]; ok {
			state.blocks[strName].appendInstructions(insts)
		} else {
			state.blocks[strName] = newBlockStack(insts)
		}

	}
	return newInsts, nil
}

func (m *virtualMachine) callBlock(name string, state *State, out *Output) (option.Option[Value], error) {
	if blockStack, ok := state.blocks[name]; ok {
		oldBlock := state.currentBlock
		state.currentBlock = option.Some(name)
		oldInsts := state.instructions
		state.instructions = blockStack.instructions()
		state.ctx.pushFrame(*newFrameDefault())
		rv, err := m.evalState(state, out)
		state.ctx.popFrame()
		state.instructions = oldInsts
		state.currentBlock = oldBlock
		return rv, err
	}
	return option.None[Value](), NewError(UnknownBlock, fmt.Sprintf("block '%s' not found", name))
}

func (m *virtualMachine) deriveAutoEscape(val Value, initialAutoEscape AutoEscape) (AutoEscape, error) {
	strVal := val.AsStr()
	if strVal.IsSome() {
		switch strVal.Unwrap() {
		case "html":
			return AutoEscapeHTML{}, nil
		case "json":
			return AutoEscapeJSON{}, nil
		case "none":
			return AutoEscapeNone{}, nil
		}
	} else if v, ok := val.(BoolValue); ok && v.B {
		if _, ok := initialAutoEscape.(AutoEscapeNone); ok {
			return AutoEscapeHTML{}, nil
		}
		return initialAutoEscape, nil
	}
	return nil, NewError(InvalidOperation, "invalid value to autoescape tag")
}

func (m *virtualMachine) pushLoop(state *State, iterable Value,
	flags uint8, pc uint, currentRecursionJump option.Option[recursionJump]) error {
	it, err := state.undefinedBehavior().TryIter(iterable)
	if err != nil {
		return err
	}
	l := it.Len()
	depth := uint(0)
	if optLoopState := state.ctx.currentLoop(); optLoopState.IsSome() {
		loopState := optLoopState.Unwrap()
		if loopState.recurseJumpTarget.IsSome() {
			depth = loopState.object.depth + 1
		}
	}
	recursive := (flags & LoopFlagRecursive) != 0
	withLoopVar := (flags & LoopFlagWithLoopVar) != 0
	recurseJumpTarget := option.None[uint]()
	if recursive {
		recurseJumpTarget = option.Some(pc)
	}
	f := newFrameDefault()
	f.currentLoop = option.Some(loopState{
		withLoopVar:          withLoopVar,
		recurseJumpTarget:    recurseJumpTarget,
		currentRecursionJump: currentRecursionJump,
		object: LoopObject{
			idx:         ^uint(0),
			len:         l,
			depth:       depth,
			valueTriple: [3]option.Option[Value]{option.None[Value](), option.None[Value](), it.Next()},
		},
		iterator: it,
	})
	log.Printf("virtualMachine.pushLoop, loop=%p", &f.currentLoop.AsPtr().object)
	err = state.ctx.pushFrame(*f)
	log.Printf("virtualMachine.pushLoop, pushed loop=%p",
		&state.ctx.stack[len(state.ctx.stack)-1].currentLoop.AsPtr().object)
	return err
}

func (m *virtualMachine) unpackList(stack *[]Value, count uint) error {
	top := stacks.Pop(stack)
	var seq SeqObject
	if optSeq := top.AsSeq(); optSeq.IsSome() {
		seq = optSeq.Unwrap()
	} else {
		return NewError(CannotUnpack, "not a sequence")
	}
	if seq.ItemCount() != count {
		return NewError(CannotUnpack,
			fmt.Sprintf("sequence of wrong length (expected %d, got %d)", count, seq.ItemCount()))
	}
	for i := uint(0); i < count; i++ {
		item := seq.GetItem(i).Unwrap()
		stacks.Push(stack, item)
	}
	return nil
}

func (m *virtualMachine) buildMacro(stack *[]Value, state *State, offset uint, name string, flags uint8) {
	var argSpec []string
	if args, ok := stacks.Pop(stack).(SeqValue); ok {
		argSpec = slicex.Map(args.items, func(arg Value) string {
			if strVal, ok := arg.(stringValue); ok {
				return strVal.str
			}
			panic("unreachable")
		})
	} else {
		panic("unreachable")
	}
	closure := stacks.Pop(stack)
	macroRefID := uint(len(state.macros))
	stacks.Push(&state.macros, tuple2[Instructions, uint]{a: state.instructions, b: offset})
	macro := &Macro{
		data: MacroData{
			name:            name,
			argSpec:         argSpec,
			macroRefID:      macroRefID,
			closure:         closure,
			callerReference: flags&macroCaller != 0,
		},
	}
	stacks.Push(stack, ValueFromObject(macro))
}

func getOrLookupLocal[T any](vec []option.Option[T], localID uint8, f func() option.Option[T]) option.Option[T] {
	tryGetItem := func(vec []option.Option[T], localId uint8) option.Option[T] {
		if localId < uint8(len(vec)) {
			return vec[localId]
		}
		return option.None[T]()
	}

	if localID == ^(uint8)(0) {
		return f()
	} else if optVal := tryGetItem(vec, localID); optVal.IsSome() {
		return optVal
	} else {
		optVal := f()
		if optVal.IsNone() {
			return option.None[T]()
		}
		vec[localID] = optVal
		return optVal
	}
}

func assertValid(v Value, pc uint, st *State) (Value, error) {
	if vInvalid, ok := v.(InvalidValue); ok {
		detail := vInvalid.Detail
		err := NewError(BadSerialization, detail)
		processErr(err, pc, st)
		return nil, err
	}
	return v, nil
}

func processErr(err error, pc uint, st *State) error {
	er, ok := err.(*Error)
	if !ok {
		return err
	}
	// only attach line information if the error does not have line info yet.
	if er.Line().IsNone() {
		if spn := st.instructions.GetSpan(pc); spn.IsSome() {
			er.SetFilenameAndSpan(st.instructions.Name(), spn.Unwrap())
		} else if lineno := st.instructions.GetLine(pc); lineno.IsSome() {
			er.SetFilenameAndLine(st.instructions.Name(), lineno.Unwrap())
		}
	}
	return er
}
