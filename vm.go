package mjingo

import (
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	stackpkg "github.com/hnakamur/mjingo/internal/datast/stack"
	"github.com/hnakamur/mjingo/option"
)

// the cost of a single include against the stack limit.
const includeRecursionConst = 10

// the cost of a single macro call against the stack limit.
const macroRecursionConst = 5

func prepareBlocks(blocks map[string]instructions) map[string]*blockStack {
	rv := make(map[string]*blockStack, len(blocks))
	for name, insts := range blocks {
		rv[name] = &blockStack{instrs: []instructions{insts}}
	}
	return rv
}

type virtualMachine struct {
	env *Environment
}

func newVirtualMachine(env *Environment) *virtualMachine {
	return &virtualMachine{env: env}
}

func (m *virtualMachine) eval(insts instructions, root Value, blocks map[string]instructions, out *output, escape AutoEscape) (option.Option[Value], error) {
	state := State{
		env:          m.env,
		ctx:          *newContext(*newFrame(root)),
		autoEscape:   escape,
		instructions: insts,
		blocks:       prepareBlocks(blocks),
	}
	return m.evalState(&state, out)
}

func (m *virtualMachine) evalMacro(insts instructions, pc uint, closure Value,
	caller option.Option[Value], out *output, state *State, args []Value) (option.Option[Value], error) {
	ctx := newContext(*newFrame(closure))
	if caller.IsSome() {
		ctx.store("caller", caller.Unwrap())
	}
	if err := ctx.incrDepth(state.ctx.depth() + macroRecursionConst); err != nil {
		return option.None[Value](), err
	}

	stack := stackpkg.Stack[Value](args)
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

func (m *virtualMachine) evalState(state *State, out *output) (option.Option[Value], error) {
	var stack stackpkg.Stack[Value]
	return m.evalImpl(state, out, &stack, 0)
}

func (m *virtualMachine) evalImpl(state *State, out *output, stack *stackpkg.Stack[Value], pc uint) (option.Option[Value], error) {
	initialAutoEscape := state.autoEscape
	undefinedBehavior := state.UndefinedBehavior()
	var autoEscapeStack stackpkg.Stack[AutoEscape]
	nextRecursionJump := option.None[recursionJump]()
	loadedFilters := [maxLocals]option.Option[BoxedFilter]{}
	loadedTests := [maxLocals]option.Option[BoxedTest]{}

	// If we are extending we are holding the instructions of the target parent
	// template here.  This is used to detect multiple extends and the evaluation
	// uses these instructions when it makes it to the end of the instructions.
	parentInstructions := option.None[instructions]()

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
		var inst instruction
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
				parentInstructions = option.None[instructions]()
			} else {
				break loop
			}
			out.endCapture(autoEscapeNone{})
			pc = 0
			continue
		}
		// log.Printf("evalImpl pc=%d, instr=%s %+v", pc, inst.Typ(), inst)

		var a, b Value

		switch inst := inst.(type) {
		case emitRawInstruction:
			if _, err := io.WriteString(out, inst.Val); err != nil {
				return option.None[Value](), err
			}
		case emitInstruction:
			v := stack.Pop()
			if err := m.env.format(v, state, out); err != nil {
				return option.None[Value](), err
			}
		case storeLocalInstruction:
			state.ctx.store(inst.Name, stack.Pop())
		case lookupInstruction:
			var v Value
			if val := state.lookup(inst.Name); val.IsSome() {
				v = val.Unwrap()
			} else {
				v = Undefined
			}
			stack.Push(v)
		case getAttrInstruction:
			a = stack.Pop()
			// This is a common enough operation that it's interesting to consider a fast
			// path here.  This is slightly faster than the regular attr lookup because we
			// do not need to pass down the error object for the more common success case.
			// Only when we cannot look up something, we start to consider the undefined
			// special case.
			if v := a.getAttrFast(inst.Name); v.IsSome() {
				if v, err := assertValid(v.Unwrap(), pc, state); err != nil {
					return option.None[Value](), err
				} else {
					stack.Push(v)
				}
			} else {
				if v, err := undefinedBehavior.handleUndefined(a.isUndefined()); err != nil {
					return option.None[Value](), processErr(err, pc, state)
				} else {
					stack.Push(v)
				}
			}
		case getItemInstruction:
			a = stack.Pop()
			b = stack.Pop()
			if v := b.getItemOpt(a); v.IsSome() {
				if v, err := assertValid(v.Unwrap(), pc, state); err != nil {
					return option.None[Value](), err
				} else {
					stack.Push(v)
				}
			} else {
				if v, err := undefinedBehavior.handleUndefined(b.isUndefined()); err != nil {
					return option.None[Value](), processErr(err, pc, state)
				} else {
					stack.Push(v)
				}
			}
		case sliceInstruction:
			step := stack.Pop()
			stop := stack.Pop()
			b = stack.Pop()
			a = stack.Pop()
			if a.isUndefined() && undefinedBehavior == UndefinedBehaviorStrict {
				return option.None[Value](), processErr(NewError(UndefinedError, ""), pc, state)
			}
			if s, err := opSlice(a, b, stop, step); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				stack.Push(s)
			}
		case loadConstInstruction:
			stack.Push(inst.Val)
		case buildMapInstruction:
			m := valueMapWithCapacity(inst.PairCount)
			for i := uint(0); i < inst.PairCount; i++ {
				val := stack.Pop()
				key := stack.Pop()
				m.Set(keyRefFromValue(key), val)
			}
			stack.Push(valueFromIndexMap(m))
		case buildKwargsInstruction:
			m := valueMapWithCapacity(inst.PairCount)
			for i := uint(0); i < inst.PairCount; i++ {
				val := stack.Pop()
				key := stack.Pop()
				m.Set(keyRefFromValue(key), val)
			}
			stack.Push(valueFromKwargs(newKwargs(*m)))
		case buildListInstruction:
			v := make([]Value, 0, untrustedSizeHint(inst.Count))
			for i := uint(0); i < inst.Count; i++ {
				v = append(v, stack.Pop())
			}
			slices.Reverse(v)
			stack.Push(valueFromSlice(v))
		case unpackListInstruction:
			if err := m.unpackList(stack, inst.Count); err != nil {
				return option.None[Value](), err
			}
		case listAppendInstruction:
			a = stack.Pop()
			// this intentionally only works with actual sequences
			v := stack.Pop()
			if valData, ok := v.data.(seqValue); ok {
				valData.Append(a)
				stack.Push(v)
			} else {
				err := NewError(InvalidOperation, "cannot append to non-list")
				return option.None[Value](), processErr(err, pc, state)
			}
		case addInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := opAdd(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case subInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := opSub(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case mulInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := opMul(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case divInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := opDiv(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case intDivInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := opIntDiv(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case remInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := opRem(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case powInstruction:
			b = stack.Pop()
			a = stack.Pop()
			if v, err := opPow(a, b); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case eqInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(valueFromBool(valueEqual(a, b)))
		case neInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(valueFromBool(!valueEqual(a, b)))
		case gtInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(valueFromBool(valueCmp(a, b) > 0))
		case gteInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(valueFromBool(valueCmp(a, b) >= 0))
		case ltInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(valueFromBool(valueCmp(a, b) < 0))
		case lteInstruction:
			b = stack.Pop()
			a = stack.Pop()
			stack.Push(valueFromBool(valueCmp(a, b) <= 0))
		case notInstruction:
			a = stack.Pop()
			stack.Push(valueFromBool(!a.isTrue()))
		case stringConcatInstruction:
			a = stack.Pop()
			b = stack.Pop()
			v := opStringConcat(b, a)
			stack.Push(v)
		case inInstruction:
			a = stack.Pop()
			b = stack.Pop()
			// the in-operator can fail if the value is undefined and
			// we are in strict mode.
			if err := state.UndefinedBehavior().assertIterable(a); err != nil {
				return option.None[Value](), err
			}
			rv, err := opContains(a, b)
			if err != nil {
				return option.None[Value](), err
			}
			stack.Push(rv)
		case negInstruction:
			a = stack.Pop()
			if v, err := opNeg(a); err != nil {
				return option.None[Value](), err
			} else {
				stack.Push(v)
			}
		case pushWithInstruction:
			if err := state.ctx.pushFrame(*newFrameDefault()); err != nil {
				return option.None[Value](), err
			}
		case popFrameInstruction:
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
		case isUndefinedInstruction:
			a = stack.Pop()
			stack.Push(valueFromBool(a.isUndefined()))
		case pushLoopInstruction:
			a = stack.Pop()
			if err := m.pushLoop(state, a, inst.Flags, pc, nextRecursionJump); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
		case iterateInstruction:
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
				next = option.Some(triple[1].Unwrap().clone())
			}
			if next.IsSome() {
				item := next.Unwrap()
				if v, err := assertValid(item, pc, state); err != nil {
					return option.None[Value](), err
				} else {
					stack.Push(v)
				}
			} else {
				pc = inst.JumpTarget
				continue
			}
		case pushDidNotIterateInstruction:
			l := state.ctx.currentLoop().Unwrap()
			stack.Push(valueFromBool(l.object.idx == 0))
		case jumpInstruction:
			pc = inst.JumpTarget
			continue
		case jumpIfFalseInstruction:
			a = stack.Pop()
			if !a.isTrue() {
				pc = inst.JumpTarget
				continue
			}
		case jumpIfFalseOrPopInstruction:
			if a, ok := stack.Peek(); ok {
				if a.isTrue() {
					stack.Pop()
				} else {
					pc = inst.JumpTarget
					continue
				}
			} else {
				panic("unreachable")
			}
		case jumpIfTrueOrPopInstruction:
			if a, ok := stack.Peek(); ok {
				if a.isTrue() {
					pc = inst.JumpTarget
					continue
				} else {
					stack.Pop()
				}
			} else {
				panic("unreachable")
			}
		case callBlockInstruction:
			if parentInstructions.IsNone() && !out.isDiscarding() {
				m.callBlock(inst.Name, state, out)
			}
		case pushAutoEscapeInstruction:
			a = stack.Pop()
			autoEscapeStack.Push(state.autoEscape)
			if escape, err := m.deriveAutoEscape(a, initialAutoEscape); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				state.autoEscape = escape
			}
		case popAutoEscapeInstruction:
			if autoEscape, ok := autoEscapeStack.TryPop(); ok {
				state.autoEscape = autoEscape
			} else {
				panic("unreachable")
			}
		case beginCaptureInstruction:
			out.beginCapture(inst.Mode)
		case endCaptureInstruction:
			stack.Push(out.endCapture(state.autoEscape))
		case applyFilterInstruction:
			f := func() option.Option[BoxedFilter] { return state.env.getFilter(inst.Name) }
			var tf BoxedFilter
			if optVal := getOrLookupLocal(loadedFilters[:], inst.LocalID, f); optVal.IsSome() {
				tf = optVal.Unwrap()
			} else {
				err := NewError(UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[Value](), processErr(err, pc, state)
			}
			args := stack.SliceTop(inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				stack.DropTop(inst.ArgCount)
				stack.Push(rv)
			}
		case performTestInstruction:
			f := func() option.Option[BoxedTest] { return state.env.getTest(inst.Name) }
			var tf BoxedTest
			if optVal := getOrLookupLocal(loadedTests[:], inst.LocalID, f); optVal.IsSome() {
				tf = optVal.Unwrap()
			} else {
				err := NewError(UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[Value](), processErr(err, pc, state)
			}
			args := stack.SliceTop(inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			} else {
				stack.DropTop(inst.ArgCount)
				stack.Push(valueFromBool(rv))
			}
		case callFunctionInstruction:
			if inst.Name == "super" {
				// super is a special function reserved for super-ing into blocks.
				if inst.ArgCount != 0 {
					err := NewError(InvalidOperation, "super() takes no arguments")
					return option.None[Value](), processErr(err, pc, state)
				}
				val, err := m.performSuper(state, out, true)
				if err != nil {
					return option.None[Value](), processErr(err, pc, state)
				}
				stack.Push(val)
			} else if inst.Name == "loop" {
				// loop is a special name which when called recurses the current loop.
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
				args := stack.SliceTop(inst.ArgCount)
				a, err := valueCall(f, state, args)
				if err != nil {
					return option.None[Value](), err
				}
				stack.DropTop(inst.ArgCount)
				stack.Push(a)
			} else {
				err := NewError(UnknownFunction, fmt.Sprintf("%s is unknown", inst.Name))
				return option.None[Value](), processErr(err, pc, state)
			}
		case callMethodInstruction:
			args := stack.SliceTop(inst.ArgCount)
			a, err := callMethod(args[0], state, inst.Name, args[1:])
			if err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
			stack.DropTop(inst.ArgCount)
			stack.Push(a)
		case callObjectInstruction:
			panic("not implemented for CallObjectInstruction")
		case dupTopInstruction:
			if val, ok := stack.Peek(); ok {
				stack.Push(val.clone())
			} else {
				panic("stack must not be empty")
			}
		case discardTopInstruction:
			stack.Pop()
		case fastSuperInstruction:
			if _, err := m.performSuper(state, out, false); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
		case fastRecurseInstruction:
			if err := recurseLoop(false); err != nil {
				return option.None[Value](), err
			}
			continue
		case loadBlocksInstruction:
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
			a = stack.Pop()
			if parentInstructions.IsSome() {
				err := NewError(InvalidOperation, "tried to extend a second time in a template")
				return option.None[Value](), processErr(err, pc, state)
			}
			insts, err := m.loadBlocks(a, state)
			if err != nil {
				return option.None[Value](), err
			}
			parentInstructions = option.Some(insts)
			out.beginCapture(captureModeDiscard)
		case includeInstruction:
			a = stack.Pop()
			if err := m.performInclude(a, state, out, inst.IgnoreMissing); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
		case exportLocalsInstruction:
			locals := state.ctx.currentLocals()
			module := valueMapWithCapacity(uint(len(*locals)))
			for key, val := range *locals {
				module.Set(keyRefFromValue(valueFromString(key)), val.clone())
			}
			stack.Push(valueFromIndexMap(module))
		case buildMacroInstruction:
			m.buildMacro(stack, state, inst.Offset, inst.Name, inst.Flags)
		case returnInstruction:
			break loop
		case encloseInstruction:
			state.ctx.enclose(state.env, inst.Name)
		case getClosureInstruction:
			closure := state.ctx.closure()
			stack.Push(valueFromObject(&closure))
		default:
			panic("unreachable")
		}
		pc++
	}
	if v, ok := stack.TryPop(); ok {
		return option.Some(v), nil
	}
	return option.None[Value](), nil
}

func (m *virtualMachine) performInclude(name Value, state *State, out *output, ignoreMissing bool) error {
	var choices SeqObject
	if optChoices := name.asSeq(); optChoices.IsSome() {
		choices = optChoices.Unwrap()
	} else {
		choices = newSliceSeqObject([]Value{name})
	}

	var templatesTried stackpkg.Stack[Value]
	l := choices.ItemCount()
	for i := uint(0); i < l; i++ {
		choice := choices.GetItem(i).Unwrap()
		var name string
		if !valueAsOptionString(choice).UnwrapTo(&name) {
			return NewError(InvalidOperation, "template name was not a string")
		}
		tmpl, err := m.env.GetTemplate(name)
		if err != nil {
			var er *Error
			if errors.As(err, &er) && er.Type() == TemplateNotFound {
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
			return NewError(BadInclude, fmt.Sprintf("error in \"%s\"", tmpl.name())).withSource(err)
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

func (m *virtualMachine) performSuper(state *State, out *output, capture bool) (Value, error) {
	if state.currentBlock.IsNone() {
		return Value{}, NewError(InvalidOperation, "cannot super outside of block")
	}
	name := state.currentBlock.Unwrap()

	blockStack := state.blocks[name]
	if !blockStack.push() {
		return Value{}, NewError(InvalidOperation, "no parent block exists")
	}

	if capture {
		out.beginCapture(captureModeCapture)
	}

	oldInsts := state.instructions
	state.instructions = blockStack.instructions()
	if err := state.ctx.pushFrame(*newFrameDefault()); err != nil {
		return Value{}, err
	}
	_, err := m.evalState(state, out)
	state.ctx.popFrame()
	state.instructions = oldInsts
	state.blocks[name].pop()
	if err != nil {
		return Value{}, NewError(EvalBlock, "error in super block").withSource(err)
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

func (m *virtualMachine) loadBlocks(name Value, state *State) (instructions, error) {
	var strName string
	if !valueAsOptionString(name).UnwrapTo(&strName) {
		return instructions{}, NewError(InvalidOperation, "template name was not a string")
	}
	if state.loadedTemplates.Contains(strName) {
		return instructions{}, NewError(InvalidOperation,
			fmt.Sprintf("cycle in template inheritance. %s was referenced more than once", name))
	}
	tmpl, err := m.env.GetTemplate(strName)
	if err != nil {
		return instructions{}, err
	}
	newInsts, newBlocks, err := tmpl.instructionsAndBlocks()
	if err != nil {
		return instructions{}, err
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

func (m *virtualMachine) callBlock(name string, state *State, out *output) (option.Option[Value], error) {
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
	var strVal string
	if valueAsOptionString(val).UnwrapTo(&strVal) {
		switch strVal {
		case "html":
			return autoEscapeHTML{}, nil
		case "json":
			return autoEscapeJSON{}, nil
		case "none":
			return autoEscapeNone{}, nil
		}
	} else if v, ok := val.data.(boolValue); ok && v.B {
		if _, ok := initialAutoEscape.(autoEscapeNone); ok {
			return autoEscapeHTML{}, nil
		}
		return initialAutoEscape, nil
	}
	return nil, NewError(InvalidOperation, "invalid value to autoescape tag")
}

func (m *virtualMachine) pushLoop(state *State, iterable Value,
	flags uint8, pc uint, currentRecursionJump option.Option[recursionJump]) error {
	it, err := state.UndefinedBehavior().tryIter(iterable)
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
	recursive := (flags & loopFlagRecursive) != 0
	withLoopVar := (flags & loopFlagWithLoopVar) != 0
	recurseJumpTarget := option.None[uint]()
	if recursive {
		recurseJumpTarget = option.Some(pc)
	}
	f := newFrameDefault()
	f.currentLoop = option.Some(loopState{
		withLoopVar:          withLoopVar,
		recurseJumpTarget:    recurseJumpTarget,
		currentRecursionJump: currentRecursionJump,
		object: loopObject{
			idx:         ^uint(0),
			len:         l,
			depth:       depth,
			valueTriple: [3]option.Option[Value]{option.None[Value](), option.None[Value](), it.Next()},
		},
		iterator: it,
	})
	return state.ctx.pushFrame(*f)
}

func (m *virtualMachine) unpackList(stack *stackpkg.Stack[Value], count uint) error {
	top := stack.Pop()
	var seq SeqObject
	if optSeq := top.asSeq(); optSeq.IsSome() {
		seq = optSeq.Unwrap()
	} else {
		return NewError(CannotUnpack, "not a sequence")
	}
	if seq.ItemCount() != count {
		return NewError(CannotUnpack,
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

func (m *virtualMachine) buildMacro(stack *stackpkg.Stack[Value], state *State, offset uint, name string, flags uint8) {
	var argSpec []string
	if args, ok := stack.Pop().data.(seqValue); ok {
		argSpec = slicex.Map(args.Items, func(arg Value) string {
			if strVal, ok := arg.data.(stringValue); ok {
				return strVal.Str
			}
			panic("unreachable")
		})
	} else {
		panic("unreachable")
	}
	closure := stack.Pop()
	macroRefID := uint(len(state.macros))
	state.macros.Push(macroStackElem{insts: state.instructions, offset: offset})
	macro := &macro{
		data: macroData{
			name:            name,
			argSpec:         argSpec,
			macroRefID:      macroRefID,
			closure:         closure,
			callerReference: flags&macroCaller != 0,
		},
	}
	stack.Push(valueFromObject(macro))
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
	if vInvalid, ok := v.data.(invalidValue); ok {
		detail := vInvalid.Detail
		err := NewError(BadSerialization, detail)
		processErr(err, pc, st)
		return Value{}, err
	}
	return v, nil
}

func processErr(err error, pc uint, st *State) error {
	er, ok := err.(*Error)
	if !ok {
		return err
	}
	// only attach line information if the error does not have line info yet.
	if er.line().IsNone() {
		if spn := st.instructions.GetSpan(pc); spn.IsSome() {
			er.setFilenameAndSpan(st.instructions.Name(), spn.Unwrap())
		} else if lineno := st.instructions.GetLine(pc); lineno.IsSome() {
			er.setFilenameAndLine(st.instructions.Name(), lineno.Unwrap())
		}
	}
	return er
}
