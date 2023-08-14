package mjingo

import (
	"fmt"
	"io"
	"log"
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
	for pc < uint(len(state.instructions.instructions)) {
		instr := state.instructions.instructions[pc]
		log.Printf("instr=%v", instr)
		switch instr.kind {
		case instructionKindEmitRaw:
			val := instr.data.(emitRawInstructionData)
			if _, err := io.WriteString(out, val); err != nil {
				return option[value]{}, err
			}
		case instructionKindEmit:
			v := stack.pop()
			log.Printf("value=%+v", v)
			if _, err := fmt.Fprintf(out, "%s", v.data); err != nil {
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
		default:
			panic(fmt.Sprintf("not implemented for instruction %s", instr.kind))
		}
		pc++
	}
	return stack.tryPop(), nil
}
