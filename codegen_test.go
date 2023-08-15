package mjingo

import "testing"

func TestModifyInstructions(t *testing.T) {
	insts := []instruction{
		jumpIfFalseInstruction{jumpTarget: 0},
	}

	// This is not good since inst is a copy.
	// if inst, ok := insts[0].(jumpIfFalseInst); ok {
	// 	inst.jumpTarget = 1
	// }

	if _, ok := insts[0].(jumpIfFalseInstruction); ok {
		insts[0] = jumpIfFalseInstruction{jumpTarget: 1}
	}

	if got, want := insts[0].(jumpIfFalseInstruction).jumpTarget, uint(1); got != want {
		t.Errorf("jumpTarget mismatch, got=%d, want=%d", got, want)
	}
}
