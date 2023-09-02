package vm

import (
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	stackpkg "github.com/hnakamur/mjingo/internal/datast/stack"
	"github.com/hnakamur/mjingo/internal/value"
)

// the cost of a single include against the stack limit.
const includeRecursionConst = 10

// the cost of a single macro call against the stack limit.
const macroRecursionConst = 5

func prepareBlocks(blocks map[string]compiler.Instructions) map[string]*blockStack {
	rv := make(map[string]*blockStack, len(blocks))
	for name, insts := range blocks {
		rv[name] = &blockStack{instrs: []compiler.Instructions{insts}}
	}
	return rv
}

type virtualMachine struct {
	env *Environment
}

func newVirtualMachine(env *Environment) *virtualMachine {
	return &virtualMachine{env: env}
}

func (m *virtualMachine) eval(instructions compiler.Instructions, root value.Value, blocks map[string]compiler.Instructions, out *Output, escape AutoEscape) (option.Option[value.Value], error) {
	state := State{
		env:          m.env,
		ctx:          *newContext(*newFrame(root)),
		autoEscape:   escape,
		instructions: instructions,
		blocks:       prepareBlocks(blocks),
	}
	return m.evalState(&state, out)
}

func (m *virtualMachine) evalMacro(insts compiler.Instructions, pc uint, closure value.Value,
	caller option.Option[value.Value], out *Output, state *State, args []value.Value) (option.Option[value.Value], error) {
	ctx := newContext(*newFrame(closure))
	if caller.IsSome() {
		ctx.store("caller", caller.Unwrap())
	}
	if err := ctx.incrDepth(state.ctx.depth() + macroRecursionConst); err != nil {
		return option.None[value.Value](), err
	}

	stack := stackpkg.Stack[value.Value](args)
	return m.evalImpl(&State{
		env:             m.env,
		ctx:             *ctx,
		currentBlock:    option.None[string](),
		autoEscape:      state.autoEscape,
		instructions:    insts,
		blocks:          make(map[string]*blockStack),
		loadedTemplates: *hashset.NewStrHashSet(),
		macros:          state.macros, // TODO: clone
	}, out, &stack, pc)
}

func (m *virtualMachine) evalState(state *State, out *Output) (option.Option[value.Value], error) {
	var stack stackpkg.Stack[value.Value]
	return m.evalImpl(state, out, &stack, 0)
}

func (m *virtualMachine) evalImpl(state *State, out *Output, stack *stackpkg.Stack[value.Value], pc uint) (option.Option[value.Value], error) {
	initialAutoEscape := state.autoEscape
	undefinedBehavior := state.undefinedBehavior()
	var autoEscapeStack stackpkg.Stack[AutoEscape]
	nextRecursionJump := option.None[recursionJump]()
	loadedFilters := [compiler.MaxLocals]option.Option[FilterFunc]{}
	loadedTests := [compiler.MaxLocals]option.Option[TestFunc]{}

	// If we are extending we are holding the instructions of the target parent
	// template here.  This is used to detect multiple extends and the evaluation
	// uses these instructions when it makes it to the end of the instructions.
	parentInstructions := option.None[compiler.Instructions]()

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
		var inst compiler.Instruction
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
				parentInstructions = option.None[compiler.Instructions]()
			} else {
				break loop
			}
			out.endCapture(AutoEscapeNone{})
			pc = 0
			continue
		}
		// log.Printf("evalImpl pc=%d, instr=%s %+v", pc, inst.Typ(), inst)

		var a, b value.Value

		switch inst := inst.(type) {
		case compiler.EmitRawInstruction:
			if _, err := io.WriteString(out, inst.Val); err != nil {
				return option.None[value.Value](), err
			}
		case compiler.EmitInstruction:
			v := stack.Pop()
			if err := m.env.format(v, state, out); err != nil {
				return option.None[value.Value](), err
			}
		case compiler.StoreLocalInstruction:
			state.ctx.store(inst.Name, stack.Pop())
		case compiler.LookupInstruction:
			var v value.Value
			if val := state.lookup(inst.Name); val.IsSome() {
				v = val.Unwrap()
			} else {
				v = value.Undefined
			}
			stack.Push(v)
		case compiler.GetAttrInstruction:
			a = stack.Pop()
			// This is a common enough operation that it's interesting to consider a fast
			// path here.  This is slightly faster than the regular attr lookup because we
			// do not need to pass down the error object for the more common success case.
			// Only when we cannot look up something, we start to consider the undefined
			// special case.
			if v := a.GetAttrFast(inst.Name); v.IsSome() {
				if v, err := assertValid(v.Unwrap(), pc, state); err != nil {
					return option.None[value.Value](), err
				} else {
					stack.Push(v)
				}
			} else {
				if v, err := undefinedBehavior.HandleUndefined(a.IsUndefined()); err != nil {
					return option.None[value.Value](), processErr(err, pc, state)
				} else {
					stack.Push(v)
				}
			}
		case compiler.GetItemInstruction:
			a = stack.Pop()
			b = stack.Pop()
			if v := b.GetItemOpt(a); v.IsSome() {
				if v, err := assertValid(v.Unwrap(), pc, state); err != nil {
					return option.None[value.Value](), err
				} else {
					stack.Push(v)
				}
			} else {
				if v, err := undefinedBehavior.HandleUndefined(b.IsUndefined()); err != nil {
					return option.None[value.Value](), processErr(err, pc, state)
				} else {
					stack.Push(v)
				}
			}
		case compiler.SliceInstruction:
			step := stack.Pop()
			stop := stack.Pop()
			b = stack.Pop()
			a = stack.Pop()
			if a.IsUndefined() && undefinedBehavior == UndefinedBehaviorStrict {
				return option.None[value.Value](), processErr(common.NewError(common.UndefinedError, ""), pc, state)
			}
			if s, err := value.Slice(a, b, stop, step); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				stack.Push(s)
			}
		case compiler.LoadConstInstruction:
			stack.Push(inst.Val)
		case compiler.BuildMapInstruction:
			m := value.ValueMapWithCapacity(inst.PairCount)
			for i := uint(0); i < inst.PairCount; i++ {
				val := stack.Pop()
				key := stack.Pop()
				m.Set(value.KeyRefFromValue(key), val)
			}
			stack.Push(value.ValueFromIndexMap(m))
		case compiler.BuildKwargsInstruction:
			m := value.ValueMapWithCapacity(inst.PairCount)
			for i := uint(0); i < inst.PairCount; i++ {
				val := stack.Pop()
				key := stack.Pop()
				m.Set(value.KeyRefFromValue(key), val)
			}
			stack.Push(value.MapValue{Map: m, Type: value.MapTypeKwargs})
		case compiler.BuildListInstruction:
			v := make([]value.Value, 0, untrustedSizeHint(inst.Count))
			for i := uint(0); i < inst.Count; i++ {
				v = append(v, stack.Pop())
			}
			slices.Reverse(v)
			stack.Push(value.ValueFromSlice(v))
		case compiler.UnpackListInstruction:
			if err := m.unpackList(stack, inst.Count); err != nil {
				return option.None[value.Value](), err
			}
		case compiler.ListAppendInstruction:
			a = stack.Pop()
			// this intentionally only works with actual sequences
			if v, ok := stack.Pop().(value.SeqValue); ok {
				v.Append(a)
				stack.Push(v)
			} else {
				err := common.NewError(common.InvalidOperation, "cannot append to non-list")
				return option.None[value.Value](), processErr(err, pc, state)
			}
		case compiler.AddInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := value.Add(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.SubInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := value.Sub(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.MulInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := value.Mul(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.DivInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := value.Div(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.IntDivInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := value.IntDiv(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.RemInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := value.Rem(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.PowInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := value.Pow(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.EqInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(value.ValueFromBool(value.Equal(a, b)))
		case compiler.NeInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(value.ValueFromBool(!value.Equal(a, b)))
		case compiler.GtInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(value.ValueFromBool(value.Cmp(a, b) > 0))
		case compiler.GteInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(value.ValueFromBool(value.Cmp(a, b) >= 0))
		case compiler.LtInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(value.ValueFromBool(value.Cmp(a, b) < 0))
		case compiler.LteInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(value.ValueFromBool(value.Cmp(a, b) <= 0))
		case compiler.NotInstruction:
			a = stack.Pop()
			stack.Push(value.ValueFromBool(!a.IsTrue()))
		case compiler.StringConcatInstruction:
			a = stack.Pop()
			b = stack.Pop()
			v := value.StringConcat(b, a)
			stack.Push(v)
		case compiler.InInstruction:
			a = stack.Pop()
			b = stack.Pop()
			// the in-operator can fail if the value is undefined and
			// we are in strict mode.
			if err := state.undefinedBehavior().AssertIterable(a); err != nil {
				return option.None[value.Value](), err
			}
			rv, err := value.Contains(a, b)
			if err != nil {
				return option.None[value.Value](), err
			}
			stack.Push(rv)
		case compiler.NegInstruction:
			a = stack.Pop()
			if v, err := value.Neg(a); err != nil {
				return option.None[value.Value](), err
			} else {
				stack.Push(v)
			}
		case compiler.PushWithInstruction:
			if err := state.ctx.pushFrame(*newFrameDefault()); err != nil {
				return option.None[value.Value](), err
			}
		case compiler.PopFrameInstruction:
			if optLoopCtx := state.ctx.popFrame().currentLoop; optLoopCtx.IsSome() {
				loopCtx := optLoopCtx.Unwrap()
				if loopCtx.currentRecursionJump.IsSome() {
					recurJump := loopCtx.currentRecursionJump.Unwrap()
					loopCtx.currentRecursionJump = option.None[recursionJump]()
					pc = recurJump.target
					if recurJump.endCapture {
						stack.Push(out.endCapture(state.autoEscape))
					}
					continue
				}
			}
		case compiler.IsUndefinedInstruction:
			a = stack.Pop()
			stack.Push(value.ValueFromBool(a.IsUndefined()))
		case compiler.PushLoopInstruction:
			a = stack.Pop()
			if err := m.pushLoop(state, a, inst.Flags, pc, nextRecursionJump); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			}
		case compiler.IterateInstruction:
			var l *loopState
			if mayLoopState := state.ctx.currentLoop(); mayLoopState.IsSome() {
				l = mayLoopState.Unwrap()
			} else {
				panic("no currentLoop")
			}
			l.object.idx++
			next := option.None[value.Value]()
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
					return option.None[value.Value](), err
				} else {
					stack.Push(v)
				}
			} else {
				pc = inst.JumpTarget
				continue
			}
		case compiler.PushDidNotIterateInstruction:
			l := state.ctx.currentLoop().Unwrap()
			stack.Push(value.ValueFromBool(l.object.idx == 0))
		case compiler.JumpInstruction:
			pc = inst.JumpTarget
			continue
		case compiler.JumpIfFalseInstruction:
			a = stack.Pop()
			if !a.IsTrue() {
				pc = inst.JumpTarget
				continue
			}
		case compiler.JumpIfFalseOrPopInstruction:
			if a, ok := stack.Peek(); ok {
				if a.IsTrue() {
					stack.Pop()
				} else {
					pc = inst.JumpTarget
					continue
				}
			} else {
				panic("unreachable")
			}
		case compiler.JumpIfTrueOrPopInstruction:
			if a, ok := stack.Peek(); ok {
				if a.IsTrue() {
					pc = inst.JumpTarget
					continue
				} else {
					stack.Pop()
				}
			} else {
				panic("unreachable")
			}
		case compiler.CallBlockInstruction:
			if parentInstructions.IsNone() && !out.isDiscarding() {
				m.callBlock(inst.Name, state, out)
			}
		case compiler.PushAutoEscapeInstruction:
			a = stack.Pop()
			autoEscapeStack.Push(state.autoEscape)
			if escape, err := m.deriveAutoEscape(a, initialAutoEscape); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				state.autoEscape = escape
			}
		case compiler.PopAutoEscapeInstruction:
			if autoEscape, ok := autoEscapeStack.TryPop(); ok {
				state.autoEscape = autoEscape
			} else {
				panic("unreachable")
			}
		case compiler.BeginCaptureInstruction:
			out.beginCapture(inst.Mode)
		case compiler.EndCaptureInstruction:
			stack.Push(out.endCapture(state.autoEscape))
		case compiler.ApplyFilterInstruction:
			f := func() option.Option[FilterFunc] { return state.env.getFilter(inst.Name) }
			var tf FilterFunc
			if optVal := getOrLookupLocal(loadedFilters[:], inst.LocalID, f); optVal.IsSome() {
				tf = optVal.Unwrap()
			} else {
				err := common.NewError(common.UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[value.Value](), processErr(err, pc, state)
			}
			args := stack.SliceTop(inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				stack.DropTop(inst.ArgCount)
				stack.Push(rv)
			}
		case compiler.PerformTestInstruction:
			f := func() option.Option[TestFunc] { return state.env.getTest(inst.Name) }
			var tf TestFunc
			if optVal := getOrLookupLocal(loadedTests[:], inst.LocalID, f); optVal.IsSome() {
				tf = optVal.Unwrap()
			} else {
				err := common.NewError(common.UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[value.Value](), processErr(err, pc, state)
			}
			args := stack.SliceTop(inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				stack.DropTop(inst.ArgCount)
				stack.Push(value.ValueFromBool(rv))
			}
		case compiler.CallFunctionInstruction:
			if inst.Name == "super" {
				// super is a special function reserved for super-ing into blocks.
				if inst.ArgCount != 0 {
					err := common.NewError(common.InvalidOperation, "super() takes no arguments")
					return option.None[value.Value](), processErr(err, pc, state)
				}
				val, err := m.performSuper(state, out, true)
				if err != nil {
					return option.None[value.Value](), processErr(err, pc, state)
				}
				stack.Push(val)
			} else if inst.Name == "loop" {
				// loop is a special name which when called recurses the current loop.
				if inst.ArgCount != 1 {
					err := common.NewError(common.InvalidOperation,
						fmt.Sprintf("loop() takes one argument, got %d", inst.ArgCount))
					return option.None[value.Value](), processErr(err, pc, state)
				}
				// leave the one argument on the stack for the recursion
				if err := recurseLoop(true); err != nil {
					return option.None[value.Value](), err
				}
				continue
			} else if optFunc := state.lookup(inst.Name); optFunc.IsSome() {
				f := optFunc.Unwrap()
				args := stack.SliceTop(inst.ArgCount)
				a, err := Call(f, state, args)
				if err != nil {
					return option.None[value.Value](), err
				}
				stack.DropTop(inst.ArgCount)
				stack.Push(a)
			} else {
				err := common.NewError(common.UnknownFunction, fmt.Sprintf("%s is unknown", inst.Name))
				return option.None[value.Value](), processErr(err, pc, state)
			}
		case compiler.CallMethodInstruction:
			args := stack.SliceTop(inst.ArgCount)
			a, err := CallMethod(args[0], state, inst.Name, args[1:])
			if err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			}
			stack.DropTop(inst.ArgCount)
			stack.Push(a)
		case compiler.CallObjectInstruction:
			panic("not implemented for CallObjectInstruction")
		case compiler.DupTopInstruction:
			if val, ok := stack.Peek(); ok {
				stack.Push(val.Clone())
			} else {
				panic("stack must not be empty")
			}
		case compiler.DiscardTopInstruction:
			stack.Pop()
		case compiler.FastSuperInstruction:
			if _, err := m.performSuper(state, out, false); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			}
		case compiler.FastRecurseInstruction:
			if err := recurseLoop(false); err != nil {
				return option.None[value.Value](), err
			}
			continue
		case compiler.LoadBlocksInstruction:
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
			// However for the common case compiler.this is convenient because it
			// lets you put some imports there and for as long as you do not
			// create name clashes this works fine.
			a = stack.Pop()
			if parentInstructions.IsSome() {
				err := common.NewError(common.InvalidOperation, "tried to extend a second time in a template")
				return option.None[value.Value](), processErr(err, pc, state)
			}
			insts, err := m.loadBlocks(a, state)
			if err != nil {
				return option.None[value.Value](), err
			}
			parentInstructions = option.Some(insts)
			out.beginCapture(compiler.CaptureModeDiscard)
		case compiler.IncludeInstruction:
			a = stack.Pop()
			if err := m.performInclude(a, state, out, inst.IgnoreMissing); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			}
		case compiler.ExportLocalsInstruction:
			locals := state.ctx.currentLocals()
			module := value.ValueMapWithCapacity(uint(len(*locals)))
			for key, val := range *locals {
				module.Set(value.KeyRefFromValue(value.ValueFromString(key)), val.Clone())
			}
			stack.Push(value.ValueFromIndexMap(module))
		case compiler.BuildMacroInstruction:
			m.buildMacro(stack, state, inst.Offset, inst.Name, inst.Flags)
		case compiler.ReturnInstruction:
			break loop
		case compiler.EncloseInstruction:
			state.ctx.enclose(state.env, inst.Name)
		case compiler.GetClosureInstruction:
			closure := state.ctx.closure()
			stack.Push(value.ValueFromObject(&closure))
		default:
			panic("unreachable")
		}
		pc++
	}
	if v, ok := stack.TryPop(); ok {
		return option.Some(v), nil
	}
	return option.None[value.Value](), nil
}

func (m *virtualMachine) performInclude(name value.Value, state *State, out *Output, ignoreMissing bool) error {
	var choices value.SeqObject
	if optChoices := name.AsSeq(); optChoices.IsSome() {
		choices = optChoices.Unwrap()
	} else {
		choices = value.NewSliceSeqObject([]value.Value{name})
	}

	var templatesTried stackpkg.Stack[value.Value]
	l := choices.ItemCount()
	for i := uint(0); i < l; i++ {
		choice := choices.GetItem(i).Unwrap()
		optName := choice.AsStr()
		if optName.IsNone() {
			return common.NewError(common.InvalidOperation, "template name was not a string")
		}
		tmpl, err := m.env.GetTemplate(optName.Unwrap())
		if err != nil {
			var er *common.Error
			if errors.As(err, &er) && er.Type() == common.TemplateNotFound {
				templatesTried.Push(choice)
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
			return common.NewError(common.BadInclude, fmt.Sprintf("error in \"%s\"", tmpl.name())).WithSource(err)
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
		return common.NewError(common.TemplateNotFound, detail)
	}
	return nil
}

func (m *virtualMachine) performSuper(state *State, out *Output, capture bool) (value.Value, error) {
	if state.currentBlock.IsNone() {
		return nil, common.NewError(common.InvalidOperation, "cannot super outside of block")
	}
	name := state.currentBlock.Unwrap()

	blockStack := state.blocks[name]
	if !blockStack.push() {
		return nil, common.NewError(common.InvalidOperation, "no parent block exists")
	}

	if capture {
		out.beginCapture(compiler.CaptureModeCapture)
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
		return nil, common.NewError(common.EvalBlock, "error in super block").WithSource(err)
	}
	if capture {
		return out.endCapture(state.autoEscape), nil
	}
	return value.Undefined, nil
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
		return 0, common.NewError(common.InvalidOperation, "cannot recurse outside of recursive loop")
	}
	return 0, common.NewError(common.InvalidOperation, "cannot recurse outside of loop")
}

func (m *virtualMachine) loadBlocks(name value.Value, state *State) (compiler.Instructions, error) {
	optName := name.AsStr()
	if optName.IsNone() {
		return compiler.Instructions{}, common.NewError(common.InvalidOperation, "template name was not a string")
	}
	strName := optName.Unwrap()
	if state.loadedTemplates.Contains(strName) {
		return compiler.Instructions{}, common.NewError(common.InvalidOperation,
			fmt.Sprintf("cycle in template inheritance. %s was referenced more than once", name))
	}
	tmpl, err := m.env.GetTemplate(strName)
	if err != nil {
		return compiler.Instructions{}, err
	}
	newInsts, newBlocks, err := tmpl.instructionsAndBlocks()
	if err != nil {
		return compiler.Instructions{}, err
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

func (m *virtualMachine) callBlock(name string, state *State, out *Output) (option.Option[value.Value], error) {
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
	return option.None[value.Value](), common.NewError(common.UnknownBlock, fmt.Sprintf("block '%s' not found", name))
}

func (m *virtualMachine) deriveAutoEscape(val value.Value, initialAutoEscape AutoEscape) (AutoEscape, error) {
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
	} else if v, ok := val.(value.BoolValue); ok && v.B {
		if _, ok := initialAutoEscape.(AutoEscapeNone); ok {
			return AutoEscapeHTML{}, nil
		}
		return initialAutoEscape, nil
	}
	return nil, common.NewError(common.InvalidOperation, "invalid value to autoescape tag")
}

func (m *virtualMachine) pushLoop(state *State, iterable value.Value,
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
	recursive := (flags & compiler.LoopFlagRecursive) != 0
	withLoopVar := (flags & compiler.LoopFlagWithLoopVar) != 0
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
			valueTriple: [3]option.Option[value.Value]{option.None[value.Value](), option.None[value.Value](), it.Next()},
		},
		iterator: it,
	})
	return state.ctx.pushFrame(*f)
}

func (m *virtualMachine) unpackList(stack *stackpkg.Stack[value.Value], count uint) error {
	top := stack.Pop()
	var seq value.SeqObject
	if optSeq := top.AsSeq(); optSeq.IsSome() {
		seq = optSeq.Unwrap()
	} else {
		return common.NewError(common.CannotUnpack, "not a sequence")
	}
	if seq.ItemCount() != count {
		return common.NewError(common.CannotUnpack,
			fmt.Sprintf("sequence of wrong length (expected %d, got %d)", count, seq.ItemCount()))
	}
	for i := count - 1; ; i-- {
		item := seq.GetItem(i).Unwrap()
		stack.Push(item)
		if i == 0 {
			break
		}
	}
	return nil
}

func (m *virtualMachine) buildMacro(stack *stackpkg.Stack[value.Value], state *State, offset uint, name string, flags uint8) {
	var argSpec []string
	if args, ok := stack.Pop().(value.SeqValue); ok {
		argSpec = slicex.Map(args.Items, func(arg value.Value) string {
			if strVal, ok := arg.(value.StringValue); ok {
				return strVal.Str
			}
			panic("unreachable")
		})
	} else {
		panic("unreachable")
	}
	closure := stack.Pop()
	macroRefID := uint(len(state.macros))
	state.macros.Push(tuple2[compiler.Instructions, uint]{a: state.instructions, b: offset})
	macro := &Macro{
		data: MacroData{
			name:            name,
			argSpec:         argSpec,
			macroRefID:      macroRefID,
			closure:         closure,
			callerReference: flags&compiler.MacroCaller != 0,
		},
	}
	stack.Push(value.ValueFromObject(macro))
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

func assertValid(v value.Value, pc uint, st *State) (value.Value, error) {
	if vInvalid, ok := v.(value.InvalidValue); ok {
		detail := vInvalid.Detail
		err := common.NewError(common.BadSerialization, detail)
		processErr(err, pc, st)
		return nil, err
	}
	return v, nil
}

func processErr(err error, pc uint, st *State) error {
	er, ok := err.(*common.Error)
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
