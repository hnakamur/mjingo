package mjingo

import (
	"fmt"
	"io"
)

type virtualMachine struct {
	env *Environment
}

func newVirtualMachine(env *Environment) *virtualMachine {
	return &virtualMachine{env: env}
}

func (m *virtualMachine) eval(instructions instructions, root value, blocks map[string]instructions, out io.Writer) (option[value], error) {
	state := virtualMachineState{
		env:          m.env,
		instructions: instructions,
		ctx:          *newContext(*newFrame(root)),
	}
	return m.evalState(&state, out)
}

func (m *virtualMachine) evalState(state *virtualMachineState, out io.Writer) (option[value], error) {
	return m.evalImpl(state, out, newVMStack(), 0)
}

func (m *virtualMachine) evalImpl(state *virtualMachineState, out io.Writer, stack vmStack, pc uint) (option[value], error) {
	undefinedBehavior := state.undefinedBehavior()

	for pc < uint(len(state.instructions.instructions)) {
		var a, b value

		instr := state.instructions.instructions[pc]
		// log.Printf("evalImpl pc=%d, instr=%s %+v", pc, instr.kind, instr)
		switch instr.kind {
		case instructionKindEmitRaw:
			val := instr.data.(emitRawInstructionData)
			if _, err := io.WriteString(out, val); err != nil {
				return option[value]{}, err
			}
		case instructionKindEmit:
			v := stack.pop()
			if _, err := fmt.Fprintf(out, "%v", v.data); err != nil {
				return option[value]{}, err
			}
		case instructionKindLookup:
			name := instr.data.(lookupInstructionData)
			var v value
			if val := state.lookup(name); val.valid {
				v = val.data
			} else {
				v = valueUndefined
			}
			stack.push(v)
		case instructionKindGetAttr:
			name := instr.data.(getAttrInstructionData)
			a = stack.pop()
			// This is a common enough operation that it's interesting to consider a fast
			// path here.  This is slightly faster than the regular attr lookup because we
			// do not need to pass down the error object for the more common success case.
			// Only when we cannot look up something, we start to consider the undefined
			// special case.
			if v := a.getAttrFast(name); v.valid {
				if v2, err := assertValid(v.data, pc, state); err != nil {
					return option[value]{}, err
				} else {
					stack.push(v2)
				}
			} else {
				if v, err := undefinedBehavior.handleUndefined(a.isUndefined()); err != nil {
					return option[value]{}, processErr(err, pc, state)
				} else {
					stack.push(v)
				}
			}
		case instructionKindGetItem:
			a = stack.pop()
			b = stack.pop()
			if v := b.getItemOpt(a); v.valid {
				if v2, err := assertValid(v.data, pc, state); err != nil {
					return option[value]{}, err
				} else {
					stack.push(v2)
				}
			} else {
				if v, err := undefinedBehavior.handleUndefined(b.isUndefined()); err != nil {
					return option[value]{}, processErr(err, pc, state)
				} else {
					stack.push(v)
				}
			}
		case instructionKindSlice:
			step := stack.pop()
			stop := stack.pop()
			b = stack.pop()
			a = stack.pop()
			if a.isUndefined() && undefinedBehavior == UndefinedBehaviorStrict {
				return option[value]{}, processErr(&Error{kind: UndefinedError}, pc, state)
			}
			if s, err := opsSlice(a, b, stop, step); err != nil {
				return option[value]{}, processErr(err, pc, state)
			} else {
				stack.push(s)
			}
		case instructionKindLoadConst:
			v := instr.data.(loadConstInstructionData)
			stack.push(v)
		default:
			panic(fmt.Sprintf("not implemented for instruction %s", instr.kind))
		}
		pc++
	}
	return stack.tryPop(), nil
}

func assertValid(v value, pc uint, st *virtualMachineState) (value, error) {
	if v.kind == valueKindInvalid {
		detail := v.data.(invalidValueData)
		err := &Error{
			kind:   BadSerialization,
			detail: option[string]{valid: true, data: detail},
		}
		processErr(err, pc, st)
		return value{}, err
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
