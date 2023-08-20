package vm

import (
	"fmt"
	"io"
	"slices"

	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/stacks"
	"github.com/hnakamur/mjingo/value"
)

type virtualMachine struct {
	env *Environment
}

func newVirtualMachine(env *Environment) *virtualMachine {
	return &virtualMachine{env: env}
}

func (m *virtualMachine) eval(instructions compiler.Instructions, root value.Value, blocks map[string]compiler.Instructions, out *Output, escape compiler.AutoEscape) (option.Option[value.Value], error) {
	state := State{
		env:          m.env,
		instructions: instructions,
		ctx:          *newContext(*newFrame(root)),
		autoEscape:   escape,
	}
	return m.evalState(&state, out)
}

func (m *virtualMachine) evalState(state *State, out *Output) (option.Option[value.Value], error) {
	var stack []value.Value
	return m.evalImpl(state, out, &stack, 0)
}

// type autoEscapeStack = stack[autoEscape]

func (m *virtualMachine) evalImpl(state *State, out *Output, stack *[]value.Value, pc uint) (option.Option[value.Value], error) {
	initialAutoEscape := state.autoEscape
	undefinedBehavior := state.undefinedBehavior()
	autoEscapeStack := []compiler.AutoEscape{}
	nextRecursionJump := option.None[recursionJump]()
	loadedFilters := [compiler.MaxLocals]option.Option[FilterFunc]{}
	loadedTests := [compiler.MaxLocals]option.Option[TestFunc]{}

	for pc < uint(len(state.instructions.Instructions())) {
		var a, b value.Value

		inst := state.instructions.Instructions()[pc]
		// log.Printf("evalImpl pc=%d, instr=%s %+v", pc, instr.kind, instr)
		switch inst := inst.(type) {
		case compiler.EmitRawInstruction:
			if _, err := io.WriteString(out, inst.Val); err != nil {
				return option.None[value.Value](), err
			}
		case compiler.EmitInstruction:
			v := stacks.Pop(stack)
			if err := m.env.format(v, state, out); err != nil {
				return option.None[value.Value](), err
			}
		case compiler.StoreLocalInstruction:
			state.ctx.store(inst.Name, stacks.Pop(stack))
		case compiler.LookupInstruction:
			var v value.Value
			if val := state.lookup(inst.Name); option.IsSome(val) {
				v = option.Unwrap(val)
			} else {
				v = value.Undefined
			}
			stacks.Push(stack, v)
		case compiler.GetAttrInstruction:
			a = stacks.Pop(stack)
			// This is a common enough operation that it's interesting to consider a fast
			// path here.  This is slightly faster than the regular attr lookup because we
			// do not need to pass down the error object for the more common success case.
			// Only when we cannot look up something, we start to consider the undefined
			// special case.
			if v := a.GetAttrFast(inst.Name); option.IsSome(v) {
				if v, err := assertValid(option.Unwrap(v), pc, state); err != nil {
					return option.None[value.Value](), err
				} else {
					stacks.Push(stack, v)
				}
			} else {
				if v, err := undefinedBehavior.HandleUndefined(a.IsUndefined()); err != nil {
					return option.None[value.Value](), processErr(err, pc, state)
				} else {
					stacks.Push(stack, v)
				}
			}
		case compiler.GetItemInstruction:
			a = stacks.Pop(stack)
			b = stacks.Pop(stack)
			if v := b.GetItemOpt(a); option.IsSome(v) {
				if v, err := assertValid(option.Unwrap(v), pc, state); err != nil {
					return option.None[value.Value](), err
				} else {
					stacks.Push(stack, v)
				}
			} else {
				if v, err := undefinedBehavior.HandleUndefined(b.IsUndefined()); err != nil {
					return option.None[value.Value](), processErr(err, pc, state)
				} else {
					stacks.Push(stack, v)
				}
			}
		case compiler.SliceInstruction:
			step := stacks.Pop(stack)
			stop := stacks.Pop(stack)
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if a.IsUndefined() && undefinedBehavior == compiler.UndefinedBehaviorStrict {
				return option.None[value.Value](), processErr(internal.NewError(internal.UndefinedError, ""), pc, state)
			}
			if s, err := value.Slice(a, b, stop, step); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				stacks.Push(stack, s)
			}
		case compiler.LoadConstInstruction:
			stacks.Push(stack, inst.Val)
		case compiler.BuildMapInstruction:
			m := value.NewIndexMapWithCapacity(inst.PairCount)
			for i := uint(0); i < inst.PairCount; i++ {
				val := stacks.Pop(stack)
				key := stacks.Pop(stack)
				m.Set(value.KeyRefFromValue(key), val)
			}
			stacks.Push(stack, value.FromIndexMap(m))
		case compiler.BuildListInstruction:
			v := make([]value.Value, 0, untrustedSizeHint(inst.Count))
			for i := uint(0); i < inst.Count; i++ {
				v = append(v, stacks.Pop(stack))
			}
			slices.Reverse(v)
			stacks.Push(stack, value.FromSlice(v))
		case compiler.UnpackListInstruction:
			if err := m.unpackList(stack, inst.Count); err != nil {
				return option.None[value.Value](), err
			}
		case compiler.ListAppendInstruction:
			a = stacks.Pop(stack)
			// this intentionally only works with actual sequences
			if v, ok := stacks.Pop(stack).(value.SeqValue); ok {
				v.Append(a)
				stacks.Push[[]value.Value, value.Value](stack, v)
			} else {
				err := internal.NewError(internal.InvalidOperation, "cannot append to non-list")
				return option.None[value.Value](), processErr(err, pc, state)
			}
		case compiler.AddInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := value.Add(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case compiler.SubInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := value.Sub(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case compiler.MulInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := value.Mul(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case compiler.DivInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := value.Div(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case compiler.IntDivInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := value.IntDiv(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case compiler.EqInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, value.FromBool(value.Equal(a, b)))
		case compiler.NeInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, value.FromBool(!value.Equal(a, b)))
		case compiler.GtInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, value.FromBool(value.Cmp(a, b) > 0))
		case compiler.GteInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, value.FromBool(value.Cmp(a, b) >= 0))
		case compiler.LtInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, value.FromBool(value.Cmp(a, b) < 0))
		case compiler.LteInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			stacks.Push(stack, value.FromBool(value.Cmp(a, b) <= 0))
		case compiler.PowInstruction:
			b = stacks.Pop(stack)
			a = stacks.Pop(stack)
			if v, err := value.Pow(a, b); err != nil {
				return option.None[value.Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case compiler.NotInstruction:
			a = stacks.Pop(stack)
			stacks.Push(stack, value.FromBool(!a.IsTrue()))
		case compiler.StringConcatInstruction:
			a = stacks.Pop(stack)
			b = stacks.Pop(stack)
			v := value.StringConcat(b, a)
			stacks.Push(stack, v)
		case compiler.InInstruction:
			a = stacks.Pop(stack)
			b = stacks.Pop(stack)
			// the in-operator can fail if the value is undefined and
			// we are in strict mode.
			if err := state.undefinedBehavior().AssertIterable(a); err != nil {
				return option.None[value.Value](), err
			}
			rv, err := value.Contains(a, b)
			if err != nil {
				return option.None[value.Value](), err
			}
			stacks.Push(stack, rv)
		case compiler.NegInstruction:
			a = stacks.Pop(stack)
			if v, err := value.Neg(a); err != nil {
				return option.None[value.Value](), err
			} else {
				stacks.Push(stack, v)
			}
		case compiler.PopFrameInstruction:
			if optLoopCtx := state.ctx.popFrame().currentLoop; option.IsSome(optLoopCtx) {
				loopCtx := option.Unwrap(optLoopCtx)
				if option.IsSome(loopCtx.currentRecursionJump) {
					recurJump := option.Unwrap(loopCtx.currentRecursionJump)
					loopCtx.currentRecursionJump = option.None[recursionJump]()
					pc = recurJump.target
					if recurJump.endCapture {
						// TODO: implement
						// stacks.Push(stack, )
					}
					continue
				}
			}
		case compiler.JumpInstruction:
			pc = inst.JumpTarget
			continue
		case compiler.JumpIfFalseInstruction:
			a = stacks.Pop(stack)
			if !a.IsTrue() {
				pc = inst.JumpTarget
				continue
			}
		case compiler.PushAutoEscapeInstruction:
			a = stacks.Pop(stack)
			stacks.Push(&autoEscapeStack, state.autoEscape)
			if escape, err := m.deriveAutoEscape(a, initialAutoEscape); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				state.autoEscape = escape
			}
		case compiler.PopAutoEscapeInstruction:
			if autoEscape, ok := stacks.TryPop(&autoEscapeStack); ok {
				state.autoEscape = autoEscape
			} else {
				panic("unreachable")
			}
		case compiler.BeginCaptureInstruction:
			out.beginCapture(inst.Mode)
		case compiler.EndCaptureInstruction:
			stacks.Push(stack, out.endCapture(state.autoEscape))
		case compiler.ApplyFilterInstruction:
			f := func() option.Option[FilterFunc] { return state.env.getFilter(inst.Name) }
			var tf FilterFunc
			if optVal := getOrLookupLocal(loadedFilters[:], inst.LocalID, f); option.IsSome(optVal) {
				tf = option.Unwrap(optVal)
			} else {
				err := internal.NewError(internal.UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[value.Value](), processErr(err, pc, state)
			}
			args := stacks.SliceTop(*stack, inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				stacks.DropTop(stack, inst.ArgCount)
				stacks.Push(stack, rv)
			}
		case compiler.PerformTestInstruction:
			f := func() option.Option[TestFunc] { return state.env.getTest(inst.Name) }
			var tf TestFunc
			if optVal := getOrLookupLocal(loadedTests[:], inst.LocalID, f); option.IsSome(optVal) {
				tf = option.Unwrap(optVal)
			} else {
				err := internal.NewError(internal.UnknownTest, fmt.Sprintf("test %s is unknown", inst.Name))
				return option.None[value.Value](), processErr(err, pc, state)
			}
			args := stacks.SliceTop(*stack, inst.ArgCount)
			if rv, err := tf(state, args); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			} else {
				stacks.DropTop(stack, inst.ArgCount)
				stacks.Push(stack, value.FromBool(rv))
			}
		case compiler.DupTopInstruction:
			if val, ok := stacks.Peek(*stack); ok {
				stacks.Push(stack, val.Clone())
			} else {
				panic("stack must not be empty")
			}
		case compiler.DiscardTopInstruction:
			stacks.Pop(stack)
		case compiler.PushLoopInstruction:
			a = stacks.Pop(stack)
			if err := m.pushLoop(state, a, inst.Flags, pc, nextRecursionJump); err != nil {
				return option.None[value.Value](), processErr(err, pc, state)
			}
		case compiler.IterateInstruction:
			var l *loopState
			if mayLoopState := state.ctx.currentLoop(); option.IsSome(mayLoopState) {
				l = option.Unwrap(mayLoopState)
			} else {
				panic("no currentLoop")
			}
			l.object.idx++
			next := option.None[value.Value]()
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
					return option.None[value.Value](), err
				} else {
					stacks.Push(stack, v)
				}
			} else {
				pc = inst.JumpTarget
				continue
			}
			//  Instruction::PushDidNotIterate
		default:
			panic(fmt.Sprintf("not implemented for instruction %s", inst.Typ()))
		}
		pc++
	}
	if v, ok := stacks.TryPop(stack); ok {
		return option.Some(v), nil
	}
	return option.None[value.Value](), nil
}

func untrustedSizeHint(val uint) uint {
	return min(val, 1024)
}

func (m *virtualMachine) deriveAutoEscape(val value.Value, initialAutoEscape compiler.AutoEscape) (compiler.AutoEscape, error) {
	strVal := val.AsStr()
	if option.IsSome(strVal) {
		switch option.Unwrap(strVal) {
		case "html":
			return compiler.AutoEscapeHTML{}, nil
		case "json":
			return compiler.AutoEscapeJSON{}, nil
		case "none":
			return compiler.AutoEscapeNone{}, nil
		}
	} else if v, ok := val.(value.BoolValue); ok && v.B {
		if _, ok := initialAutoEscape.(compiler.AutoEscapeNone); ok {
			return compiler.AutoEscapeHTML{}, nil
		}
		return initialAutoEscape, nil
	}
	return nil, internal.NewError(internal.InvalidOperation, "invalid value to autoescape tag")
}

func (m *virtualMachine) pushLoop(state *State, iterable value.Value,
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
		object: loop{
			idx:         uint(0),
			len:         l,
			depth:       depth,
			valueTriple: optValueTriple{option.None[value.Value](), option.None[value.Value](), it.Next()},
		},
		iterator: it,
	})
	return state.ctx.pushFrame(*f)
}

func (m *virtualMachine) unpackList(stack *[]value.Value, count uint) error {
	top := stacks.Pop(stack)
	var seq value.SeqObject
	if optSeq := top.AsSeq(); option.IsSome(optSeq) {
		seq = option.Unwrap(optSeq)
	} else {
		return internal.NewError(internal.CannotUnpack, "not a sequence")
	}
	if seq.ItemCount() != count {
		return internal.NewError(internal.CannotUnpack,
			fmt.Sprintf("sequence of wrong length (expected %d, got %d)", count, seq.ItemCount()))
	}
	for i := uint(0); i < count; i++ {
		item := option.Unwrap(seq.GetItem(i))
		stacks.Push(stack, item)
	}
	return nil
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

func assertValid(v value.Value, pc uint, st *State) (value.Value, error) {
	if vInvalid, ok := v.(value.InvalidValue); ok {
		detail := vInvalid.Detail
		err := internal.NewError(internal.BadSerialization, detail)
		processErr(err, pc, st)
		return nil, err
	}
	return v, nil
}

func processErr(err error, pc uint, st *State) error {
	er, ok := err.(*internal.Error)
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
