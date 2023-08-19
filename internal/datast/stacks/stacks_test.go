package stacks_test

import (
	"testing"

	"github.com/hnakamur/mjingo/internal/datast/stacks"
)

func TestStacks(t *testing.T) {
	var stk []string
	stacks.Push(&stk, "hello")
	if got, want := len(stk), 1; got != want {
		t.Errorf("slice length mismatch, got=%d, want=%d", got, want)
	}
	if got, want := stk[0], "hello"; got != want {
		t.Errorf("element mismatch, got=%s, want=%s", got, want)
	}

	v, ok := stacks.Peek(stk)
	if got, want := len(stk), 1; got != want {
		t.Errorf("slice length mismatch, got=%d, want=%d", got, want)
	}
	if got, want := v, "hello"; got != want {
		t.Errorf("element mismatch, got=%s, want=%s", got, want)
	}
	if got, want := ok, true; got != want {
		t.Errorf("ok mismatch, got=%v, want=%v", got, want)
	}

	v, ok = stacks.TryPop(&stk)
	if got, want := len(stk), 0; got != want {
		t.Errorf("slice length mismatch, got=%d, want=%d", got, want)
	}
	if got, want := v, "hello"; got != want {
		t.Errorf("element mismatch, got=%s, want=%s", got, want)
	}
	if got, want := ok, true; got != want {
		t.Errorf("ok mismatch, got=%v, want=%v", got, want)
	}
}
