package mjingo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodegen(t *testing.T) {
	fileExists := func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}

	inputFilenames := mustGlob(t, []string{"tests", "inputs"}, []string{"*.txt", "*.html"})
	for _, inputFilename := range inputFilenames {
		inputFileBasename := filepath.Base(inputFilename)
		t.Run(inputFileBasename, func(t *testing.T) {
			inputContent := mustReadFile(t, inputFilename)
			keepTrailingNewline := false
			ct, err := newCompiledTemplate(inputFileBasename, inputContent, defaultSyntaxConfig, keepTrailingNewline)
			codegenSnapPath := filepath.Join("tests", "inputs", inputFileBasename+".codegen.snap")
			if err != nil {
				if fileExists(codegenSnapPath) {
					t.Errorf("should not get error, but got: %v", err)
				}
				return
			}
			testVerifyInstsAndBlocksWithSnapshot(t, ct.instructions, ct.blocks,
				codegenSnapPath)
		})
	}
}

func TestCodegenFragments(t *testing.T) {
	inputFilenames := mustGlob(t, []string{"tests", "fragment-inputs"}, []string{"*.txt", "*.html"})
	for _, inputFilename := range inputFilenames {
		inputFileBasename := filepath.Base(inputFilename)
		t.Run(inputFileBasename, func(t *testing.T) {
			inputContent := mustReadFile(t, inputFilename)
			keepTrailingNewline := false
			ct, err := newCompiledTemplate(inputFileBasename, inputContent, defaultSyntaxConfig, keepTrailingNewline)
			if err != nil {
				t.Fatal(err)
			}
			testVerifyInstsAndBlocksWithSnapshot(t, ct.instructions, ct.blocks,
				filepath.Join("tests", "fragment-inputs", inputFileBasename+".codegen.snap"))
		})
	}
}

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
