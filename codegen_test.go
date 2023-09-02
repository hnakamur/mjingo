package mjingo

import "testing"

func TestModifyInstructions(t *testing.T) {
	insts := []instruction{
		jumpIfFalseInstruction{JumpTarget: 0},
	}

	// This is not good since inst is a copy.
	// if inst, ok := insts[0].(jumpIfFalseInst); ok {
	// 	inst.jumpTarget = 1
	// }

	if _, ok := insts[0].(jumpIfFalseInstruction); ok {
		insts[0] = jumpIfFalseInstruction{JumpTarget: 1}
	}

	if got, want := insts[0].(jumpIfFalseInstruction).JumpTarget, uint(1); got != want {
		t.Errorf("jumpTarget mismatch, got=%d, want=%d", got, want)
	}
}
