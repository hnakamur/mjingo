package mjingo

import (
	"fmt"
	"io"
	"slices"
)

type virtualMachine struct {
	env *Environment
}

func newVirtualMachine(env *Environment) *virtualMachine {
	return &virtualMachine{env: env}
}

func (m *virtualMachine) eval(instructions instructions, root value, blocks map[string]instructions, out *Output, escape autoEscape) (option[value], error) {
	state := virtualMachineState{
		env:          m.env,
		instructions: instructions,
		ctx:          *newContext(*newFrame(root)),
		autoEscape:   escape,
	}
	return m.evalState(&state, out)
}

func (m *virtualMachine) evalState(state *virtualMachineState, out *Output) (option[value], error) {
	return m.evalImpl(state, out, newVMStack(), 0)
}

type autoEscapeStack = stack[autoEscape]

func (m *virtualMachine) evalImpl(state *virtualMachineState, out *Output, stack vmStack, pc uint) (option[value], error) {
	initialAutoEscape := state.autoEscape
	undefinedBehavior := state.undefinedBehavior()
	autoEscapeStack := autoEscapeStack{}
	nextRecursionJump := option[recursionJump]{}
	loadedTests := [maxLocals]option[TestFunc]{}

	for pc < uint(len(state.instructions.instructions)) {
		var a, b value

		inst := state.instructions.instructions[pc]
		// log.Printf("evalImpl pc=%d, instr=%s %+v", pc, instr.kind, instr)
		switch inst := inst.(type) {
		case emitRawInstruction:
			if _, err := io.WriteString(out, inst.val); err != nil {
				return option[value]{}, err
			}
		case emitInstruction:
			v := stack.pop()
			if err := m.env.format(v, state, out); err != nil {
				return option[value]{}, err
			}
		case storeLocalInstruction:
			state.ctx.store(inst.name, stack.pop())
		case lookupInstruction:
			var v value
			if val := state.lookup(inst.name); val.valid {
				v = val.data
			} else {
				v = valueUndefined
			}
			stack.push(v)
		case getAttrInstruction:
			a = stack.pop()
			// This is a common enough operation that it's interesting to consider a fast
			// path here.  This is slightly faster than the regular attr lookup because we
			// do not need to pass down the error object for the more common success case.
			// Only when we cannot look up something, we start to consider the undefined
			// special case.
			if v := a.getAttrFast(inst.name); v.valid {
				if v, err := assertValid(v.data, pc, state); err != nil {
					return option[value]{}, err
				} else {
					stack.push(v)
				}
			} else {
				if v, err := undefinedBehavior.handleUndefined(a.isUndefined()); err != nil {
					return option[value]{}, processErr(err, pc, state)
				} else {
					stack.push(v)
				}
			}
		case getItemInstruction:
			a = stack.pop()
			b = stack.pop()
			if v := b.getItemOpt(a); v.valid {
				if v, err := assertValid(v.data, pc, state); err != nil {
					return option[value]{}, err
				} else {
					stack.push(v)
				}
			} else {
				if v, err := undefinedBehavior.handleUndefined(b.isUndefined()); err != nil {
					return option[value]{}, processErr(err, pc, state)
				} else {
					stack.push(v)
				}
			}
		case sliceInstruction:
			step := stack.pop()
			stop := stack.pop()
			b = stack.pop()
			a = stack.pop()
			if a.isUndefined() && undefinedBehavior == UndefinedBehaviorStrict {
				return option[value]{}, processErr(&Error{typ: UndefinedError}, pc, state)
			}
			if s, err := opsSlice(a, b, stop, step); err != nil {
				return option[value]{}, processErr(err, pc, state)
			} else {
				stack.push(s)
			}
		case loadConstInstruction:
			stack.push(inst.val)
		case buildMapInstruction:
			m := newValueIndexMapWithCapacity(inst.pairCount)
			for i := uint(0); i < inst.pairCount; i++ {
				val := stack.pop()
				key := stack.pop()
				m.Store(keyRefFromValue(key), val)
			}
			stack.push(valueFromValueIndexMap(m))
		case buildListInstruction:
			v := make([]value, 0, untrustedSizeHint(inst.count))
			for i := uint(0); i < inst.count; i++ {
				v = append(v, stack.pop())
			}
			slices.Reverse(v)
			stack.push(seqValue{items: v})
		case unpackListInstruction:
			if err := m.unpackList(&stack, inst.count); err != nil {
				return option[value]{}, err
			}
		case listAppendInstruction:
			a = stack.pop()
			// this intentionally only works with actual sequences
			if v, ok := stack.pop().(seqValue); ok {
				v.items = append(v.items, a)
				stack.push(v)
			} else {
				err := newError(InvalidOperation, "cannot append to non-list")
				return option[value]{}, processErr(err, pc, state)
			}
		case addInstruction:
			b = stack.pop()
			a = stack.pop()
			if v, err := opsAdd(a, b); err != nil {
				return option[value]{}, err
			} else {
				stack.push(v)
			}
		case subInstruction:
			b = stack.pop()
			a = stack.pop()
			if v, err := opsSub(a, b); err != nil {
				return option[value]{}, err
			} else {
				stack.push(v)
			}
		case powInstruction:
			b = stack.pop()
			a = stack.pop()
			if v, err := opsPow(a, b); err != nil {
				return option[value]{}, err
			} else {
				stack.push(v)
			}
		case stringConcatInstruction:
			b = stack.pop()
			a = stack.pop()
			v := opsStringConcat(a, b)
			stack.push(v)
		case negInstruction:
			a = stack.pop()
			if v, err := opsNeg(a); err != nil {
				return option[value]{}, err
			} else {
				stack.push(v)
			}
		case popFrameInstruction:
			if optLoopCtx := state.ctx.popFrame().currentLoop; optLoopCtx.valid {
				loopCtx := optLoopCtx.data
				if loopCtx.currentRecursionJump.valid {
					recurJump := loopCtx.currentRecursionJump.data
					loopCtx.currentRecursionJump = option[recursionJump]{}
					pc = recurJump.target
					if recurJump.endCapture {
						// TODO: implement
						// stack.push()
					}
					continue
				}
			}
		case jumpInstruction:
			pc = inst.jumpTarget
			continue
		case jumpIfFalseInstruction:
			a = stack.pop()
			if !a.isTrue() {
				pc = inst.jumpTarget
				continue
			}
		case pushAutoEscapeInstruction:
			a = stack.pop()
			autoEscapeStack.push(state.autoEscape)
			if escape, err := m.deriveAutoEscape(a, initialAutoEscape); err != nil {
				return option[value]{}, processErr(err, pc, state)
			} else {
				state.autoEscape = escape
			}
		case popAutoEscapeInstruction:
			if mayAutoEsc := autoEscapeStack.pop(); mayAutoEsc != nil {
				state.autoEscape = *mayAutoEsc
			} else {
				panic("unreachable")
			}
		case beginCaptureInstruction:
			out.beginCapture(inst.mode)
		case endCaptureInstruction:
			stack.push(out.endCapture(state.autoEscape))
		case performTestInstruction:
			f := func() option[TestFunc] { return state.env.getTest(inst.name) }
			var tf TestFunc
			if optVal := getOrLookupLocal(loadedTests[:], inst.localId, f); optVal.valid {
				tf = optVal.data
			} else {
				err := newError(UnknownTest, fmt.Sprintf("test %s is unknown", inst.name))
				return option[value]{}, processErr(err, pc, state)
			}
			args := stack.sliceTop(inst.argCount)
			if rv, err := tf(state, args); err != nil {
				return option[value]{}, processErr(err, pc, state)
			} else {
				stack.dropTop(inst.argCount)
				stack.push(boolValue{b: rv})
			}
		case dupTopInstruction:
			if val := stack.peek(); val == nil {
				panic("stack must not be empty")
			} else {
				stack.push((*val).clone())
			}
		case discardTopInstruction:
			stack.pop()
		case pushLoopInstruction:
			a = stack.pop()
			if err := m.pushLoop(state, a, inst.flags, pc, nextRecursionJump); err != nil {
				return option[value]{}, processErr(err, pc, state)
			}
		case iterateInstruction:
			var l *loopState
			if mayLoopState := state.ctx.currentLoop(); mayLoopState.valid {
				l = mayLoopState.data
			} else {
				panic("no currentLoop")
			}
			l.object.idx++
			next := option[value]{}
			triple := &l.object.valueTriple
			triple[0] = triple[1]
			triple[1] = triple[2]
			triple[2] = l.iterator.next()
			if triple[1].valid {
				next = option[value]{valid: true, data: triple[1].data.clone()}
			}
			if next.valid {
				item := next.data
				if v, err := assertValid(item, pc, state); err != nil {
					return option[value]{}, err
				} else {
					stack.push(v)
				}
			} else {
				pc = inst.jumpTarget
				continue
			}
			//  Instruction::PushDidNotIterate
		default:
			panic(fmt.Sprintf("not implemented for instruction %s", inst.typ()))
		}
		pc++
	}
	return stack.tryPop(), nil
}

func (m *virtualMachine) deriveAutoEscape(val value, initialAutoEscape autoEscape) (autoEscape, error) {
	strVal := val.asStr()
	if strVal.valid {
		switch strVal.data {
		case "html":
			return autoEscapeHTML{}, nil
		case "json":
			return autoEscapeJSON{}, nil
		case "none":
			return autoEscapeNone{}, nil
		}
	} else if v, ok := val.(boolValue); ok && v.b {
		if _, ok := initialAutoEscape.(autoEscapeNone); ok {
			return autoEscapeHTML{}, nil
		}
		return initialAutoEscape, nil
	}
	return nil, newError(InvalidOperation, "invalid value to autoescape tag")
}

func (m *virtualMachine) pushLoop(state *virtualMachineState, iterable value,
	flags uint8, pc uint, currentRecursionJump option[recursionJump]) error {
	it, err := state.undefinedBehavior().tryIter(iterable)
	if err != nil {
		return err
	}
	l := it.len
	depth := uint(0)
	if optLoopState := state.ctx.currentLoop(); optLoopState.valid {
		loopState := optLoopState.data
		if loopState.recurseJumpTarget.valid {
			depth = loopState.object.depth + 1
		}
	}
	recursive := (flags & loopFlagRecursive) != 0
	withLoopVar := (flags & loopFlagWithLoopVar) != 0
	recurseJumpTarget := option[uint]{}
	if recursive {
		recurseJumpTarget = option[uint]{valid: true, data: pc}
	}
	f := newFrameDefault()
	f.currentLoop = option[loopState]{
		valid: true,
		data: loopState{
			withLoopVar:          withLoopVar,
			recurseJumpTarget:    recurseJumpTarget,
			currentRecursionJump: currentRecursionJump,
			object: loop{
				idx:         uint(0),
				len:         l,
				depth:       depth,
				valueTriple: optValueTriple{option[value]{}, option[value]{}, it.next()},
			},
			iterator: it,
		},
	}
	if err := state.ctx.pushFrame(*f); err != nil {
		return err
	}
	return nil
}

func (m *virtualMachine) unpackList(stack *vmStack, count uint) error {
	top := stack.pop()
	var seq seqObject
	if optSeq := top.asSeq(); optSeq.valid {
		seq = optSeq.data
	} else {
		return newError(CannotUnpack, "not a sequence")
	}
	if seq.itemCount() != count {
		return newError(CannotUnpack,
			fmt.Sprintf("sequence of wrong length (expected %d, got %d)", count, seq.itemCount()))
	}
	for i := uint(0); i < count; i++ {
		item := seq.getItem(i).data
		stack.push(item)
	}
	return nil
}

func getOrLookupLocal[T any](vec []option[T], localId uint8, f func() option[T]) option[T] {
	tryGetItem := func(vec []option[T], localId uint8) option[T] {
		if localId < uint8(len(vec)) {
			return vec[localId]
		}
		return option[T]{}
	}

	if localId == ^(uint8)(0) {
		return f()
	} else if optVal := tryGetItem(vec, localId); optVal.valid {
		return optVal
	} else {
		optVal := f()
		if !optVal.valid {
			return option[T]{}
		}
		vec[localId] = optVal
		return optVal
	}
}

func assertValid(v value, pc uint, st *virtualMachineState) (value, error) {
	if vInvalid, ok := v.(invalidValue); ok {
		detail := vInvalid.detail
		err := &Error{
			typ:    BadSerialization,
			detail: option[string]{valid: true, data: detail},
		}
		processErr(err, pc, st)
		return nil, err
	}
	return v, nil
}

func processErr(err error, pc uint, st *virtualMachineState) error {
	er, ok := err.(*Error)
	if !ok {
		return err
	}
	// only attach line information if the error does not have line info yet.
	if !er.line().valid {
		if spn := st.instructions.getSpan(pc); spn.valid {
			er.setFilenameAndSpan(st.instructions.name, spn.data)
		} else if lineno := st.instructions.getLine(pc); lineno.valid {
			er.setFilenameAndLine(st.instructions.name, lineno.data)
		}
	}
	return er
}
