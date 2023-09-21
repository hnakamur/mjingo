package mjingo

import (
	"fmt"
	"testing"
)

func TestFormatCenterAligned(t *testing.T) {
	if got, want := formatCenterAligned(" foo.txt ", '-', 79),
		`----------------------------------- foo.txt -----------------------------------`; got != want {
		t.Errorf("result mismatch, got=%s, want=%s", got, want)
	}
	if got, want := formatCenterAligned(" fo.txt ", '-', 79),
		`----------------------------------- fo.txt ------------------------------------`; got != want {
		t.Errorf("result mismatch, got=%s, want=%s", got, want)
	}
}

func TestDebugStruct(t *testing.T) {
	s := newDebugStruct("Error").field("kind", UnknownFunction.debugString()).
		field("detail", "missing_function is unknown").
		field("name", "bad_basic_block.txt").
		field("line", 3)
	if got, want := fmt.Sprintf("%s", s), `Error { kind: UnknownFunction, detail: missing_function is unknown, name: bad_basic_block.txt, line: 3 }`; got != want {
		t.Errorf("result mismatch, got=%s, want=%s", got, want)
	}
	if got, want := fmt.Sprintf("%q", s), "Error {\n"+
		"    kind: UnknownFunction,\n"+
		"    detail: missing_function is unknown,\n"+
		"    name: bad_basic_block.txt,\n"+
		"    line: 3,\n"+
		"}"; got != want {
		t.Errorf("result mismatch, got=%s, want=%s", got, want)
	}
}
