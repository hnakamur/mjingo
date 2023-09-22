package mjingo

import (
	"fmt"
	"io"
	"strings"
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
	t.Run("indent0", func(t *testing.T) {
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
			t.Errorf("result mismatch,\n got=%q,\nwant=%q", got, want)
		}
	})
	t.Run("indent1", func(t *testing.T) {
		b := new(strings.Builder)
		a := newPadAdapter(b, true)
		s := newDebugStruct("Error").field("kind", UnknownFunction.debugString()).
			field("detail", "missing_function is unknown").
			field("name", "bad_basic_block.txt").
			field("line", 3)
		fmt.Fprintf(a, "%s", s)
		if got, want := b.String(), `    Error { kind: UnknownFunction, detail: missing_function is unknown, name: bad_basic_block.txt, line: 3 }`; got != want {
			t.Errorf("result mismatch,\n got=%q,\nwant=%q", got, want)
		}

		b = new(strings.Builder)
		a = newPadAdapter(b, true)
		fmt.Fprintf(a, "%q", s)
		if got, want := b.String(), "    Error {\n"+
			"        kind: UnknownFunction,\n"+
			"        detail: missing_function is unknown,\n"+
			"        name: bad_basic_block.txt,\n"+
			"        line: 3,\n"+
			"    }"; got != want {
			t.Errorf("result mismatch,\n got=%q,\nwant=%q", got, want)
		}
	})
}

func TestDebugList(t *testing.T) {
	t.Run("indent1", func(t *testing.T) {
		t.Run("s", func(t *testing.T) {
			b := new(strings.Builder)
			a := newPadAdapter(b, true)
			l := newDebugList([]any{1, 2})
			fmt.Fprintf(a, "%s", l)
			if got, want := b.String(), "    [1, 2]"; got != want {
				t.Errorf("result mismatch,\n got=%q,\nwant=%q", got, want)
			}
		})
		t.Run("q", func(t *testing.T) {
			b := new(strings.Builder)
			a := newPadAdapter(b, true)
			l := newDebugList([]any{1, 2})
			fmt.Fprintf(a, "%q", l)
			if got, want := b.String(), "    [\n"+
				"        1,\n"+
				"        2,\n"+
				"    ]"; got != want {
				t.Errorf("result mismatch,\n got=%q,\nwant=%q", got, want)
			}
		})
	})
}

func TestPadAdapter(t *testing.T) {
	var b strings.Builder
	a := newPadAdapter(&b, false)
	io.WriteString(a, "foo\nbar\nbaz")
	if got, want := b.String(), "foo\n    bar\n    baz"; got != want {
		t.Errorf("result mismatch, got=%q, want=%q", got, want)
	}
}
