package internal

import (
	"fmt"
	"io"
	"slices"

	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/internal/datast/stacks"
)

// the cost of a single include against the stack limit.
const includeRecursionConst = 10

// the cost of a single macro call against the stack limit.
const macroRecursionConst = 5

type virtualMachine struct {
	env *Environment
}

func newVirtualMachine(env *Environment) *virtualMachine {
	return &virtualMachine{env: env}
}

func (m *virtualMachine) eval(instructions Instructions, root Value, blocks map[string]Instructions, out *Output, escape AutoEscape) (option.Option[Value], error) {
	state := State{
		env:          m.env,
		instructions: instructions,
		ctx:          *newContext(*newFrame(root)),
		autoEscape:   escape,
	}
	return m.evalState(&state, out)
}

func (m *virtualMachine) evalMacro(insts Instructions, pc uint, closure Value,
	caller option.Option[Value], out *Output, state *State, args []Value) (option.Option[Value], error) {
	ctx := newContext(*newFrame(closure))
	if option.IsSome(caller) {
		ctx.store("caller", option.Unwrap(caller))
	}
	if err := ctx.incrDepth(state.ctx.depth() + macroRecursionConst); err != nil {
		return option.None[Value](), err
	}

	return m.evalImpl(&State{
		env:          m.env,
		ctx:          *ctx,
		currentBlock: option.None[string](),
		autoEscape:   state.autoEscape,
		instructions: insts,
		blocks:       make(map[string]blockStack),
		macros:       state.macros, // TODO: clone
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

loop:
	for pc < uint(len(state.instructions.Instructions())) {
		var a, b Value

		inst := state.instructions.Instructions()[pc]
		// log.Printf("evalImpl pc=%d, instr=%s %+v", pc, instr.kind, instr)
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
			if val := state.lookup(inst.Name); option.IsSome(val) {
				v = option.Unwrap(val)
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
			if v := a.GetAttrFast(inst.Name); option.IsSome(v) {
				if v, err := assertValid(option.Unwrap(v), pc, state); err != nil {
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
			if v := b.GetItemOpt(a); option.IsSome(v) {
				if v, err := assertValid(option.Unwrap(v), pc, state); err != nil {
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
		case PopFrameInstruction:
			if optLoopCtx := state.ctx.popFrame().currentLoop; option.IsSome(optLoopCtx) {
				loopCtx := option.Unwrap(optLoopCtx)
				if option.IsSome(loopCtx.currentRecursionJump) {
					recurJump := option.Unwrap(loopCtx.currentRecursionJump)
					loopCtx.currentRecursionJump = option.None[recursionJump]()
					pc = recurJump.target
					if recurJump.endCapture {
						stacks.Push(stack, out.endCapture(state.autoEscape))
					}
					continue
				}
			}
		case PushLoopInstruction:
			a = stacks.Pop(stack)
			if err := m.pushLoop(state, a, inst.Flags, pc, nextRecursionJump); err != nil {
				return option.None[Value](), processErr(err, pc, state)
			}
		case IterateInstruction:
			var l *loopState
			if mayLoopState := state.ctx.currentLoop(); option.IsSome(mayLoopState) {
				l = option.Unwrap(mayLoopState)
			} else {
				panic("no currentLoop")
			}
			l.object.idx++
			next := option.None[Value]()
			triple := &l.object.valueTriple
			triple[0] = triple[1]
			triple[1] = triple[2]
			triple[2] = l.iterator.Next()
			if option.IsSome(triple[1]) {
				next = option.Some(option.Unwrap(triple[1]).Clone())
			}
			if option.IsSome(next) {
				item := option.Unwrap(next)
				if v, err := assertValid(item, pc, state); err != nil {
					return option.None[Value](), err
				} else {
					stacks.Push(stack, v)
				}
			} else {
				pc = inst.JumpTarget
				continue
			}
		case JumpInstruction:
			pc = inst.JumpTarget
			continue
		case JumpIfFalseInstruction:
			a = stacks.Pop(stack)
			if !a.IsTrue() {
				pc = inst.JumpTarget
				continue
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
			if optVal := getOrLookupLocal(loadedFilters[:], inst.LocalID, f); option.IsSome(optVal) {
				tf = option.Unwrap(optVal)
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
			if optVal := getOrLookupLocal(loadedTests[:], inst.LocalID, f); option.IsSome(optVal) {
				tf = option.Unwrap(optVal)
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
		case DupTopInstruction:
			if val, ok := stacks.Peek(*stack); ok {
				stacks.Push(stack, val.Clone())
			} else {
				panic("stack must not be empty")
			}
		case DiscardTopInstruction:
			stacks.Pop(stack)
		case BuildMacroInstruction:
			m.buildMacro(stack, state, inst.Offset, inst.Name, inst.Flags)
		case ReturnInstruction:
			break loop
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

func untrustedSizeHint(val uint) uint {
	return min(val, 1024)
}

func (m *virtualMachine) deriveAutoEscape(val Value, initialAutoEscape AutoEscape) (AutoEscape, error) {
	strVal := val.AsStr()
	if option.IsSome(strVal) {
		switch option.Unwrap(strVal) {
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
	if optLoopState := state.ctx.currentLoop(); option.IsSome(optLoopState) {
		loopState := option.Unwrap(optLoopState)
		if option.IsSome(loopState.recurseJumpTarget) {
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
		object: loop{
			idx:         uint(0),
			len:         l,
			depth:       depth,
			valueTriple: optValueTriple{option.None[Value](), option.None[Value](), it.Next()},
		},
		iterator: it,
	})
	return state.ctx.pushFrame(*f)
}

func (m *virtualMachine) unpackList(stack *[]Value, count uint) error {
	top := stacks.Pop(stack)
	var seq SeqObject
	if optSeq := top.AsSeq(); option.IsSome(optSeq) {
		seq = option.Unwrap(optSeq)
	} else {
		return NewError(CannotUnpack, "not a sequence")
	}
	if seq.ItemCount() != count {
		return NewError(CannotUnpack,
			fmt.Sprintf("sequence of wrong length (expected %d, got %d)", count, seq.ItemCount()))
	}
	for i := uint(0); i < count; i++ {
		item := option.Unwrap(seq.GetItem(i))
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
	} else if optVal := tryGetItem(vec, localID); option.IsSome(optVal) {
		return optVal
	} else {
		optVal := f()
		if option.IsNone(optVal) {
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
	if option.IsNone[uint](er.Line()) {
		if spn := st.instructions.GetSpan(pc); option.IsSome(spn) {
			er.SetFilenameAndSpan(st.instructions.Name(), option.Unwrap(spn))
		} else if lineno := st.instructions.GetLine(pc); option.IsSome(lineno) {
			er.SetFilenameAndLine(st.instructions.Name(), option.Unwrap(lineno))
		}
	}
	return er
}
