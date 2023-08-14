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
		var a value
		// var b value

		instr := state.instructions.instructions[pc]
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
				panic("not implemented")
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
				stack.push(v.data)
			} else {
				v, err := undefinedBehavior.handleUndefined(a.isUndefined())
				if err != nil {
					processErr(err, pc, state)
					return option[value]{}, err
				}
				stack.push(v)
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

func processErr(err error, pc uint, st *virtualMachineState) {
	er, ok := err.(*Error)
	if !ok {
		return
	}
	// only attach line information if the error does not have line info yet.
	if !er.line().valid {
		if spn := st.instructions.getSpan(pc); spn.valid {
			er.setFilenameAndSpan(st.instructions.name, spn.data)
		} else if lineno := st.instructions.getLine(pc); lineno.valid {
			er.setFilenameAndLine(st.instructions.name, lineno.data)
		}
	}
}
