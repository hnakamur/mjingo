package mjingo

import (
	"reflect"
	"testing"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

func TestFilterReflect(t *testing.T) {
	t.Run("boolFilter", func(t *testing.T) {
		ty := reflect.TypeOf(boolFilter)
		if got, want := ty.Kind(), reflect.Func; got != want {
			t.Errorf("type kind mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.NumIn(), 1; got != want {
			t.Errorf("input parameter count mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.In(0), reflect.TypeOf((*Value)(nil)).Elem(); got != want {
			t.Errorf("input parameter #0 type mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.NumOut(), 1; got != want {
			t.Errorf("output parameter count mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.Out(0), reflect.TypeOf((*bool)(nil)).Elem(); got != want {
			t.Errorf("out parameter #0 type mismatch, got=%v, want=%v", got, want)
		}
	})
	t.Run("listFilter", func(t *testing.T) {
		ty := reflect.TypeOf(listFilter)
		if got, want := ty.Kind(), reflect.Func; got != want {
			t.Errorf("type kind mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.NumIn(), 2; got != want {
			t.Errorf("input parameter count mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.In(0), reflect.TypeOf((*vmState)(nil)); got != want {
			t.Errorf("input parameter #0 type mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.In(1), reflect.TypeOf((*Value)(nil)).Elem(); got != want {
			t.Errorf("input parameter #1 type mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.NumOut(), 2; got != want {
			t.Errorf("output parameter count mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.Out(0), reflect.TypeOf((*Value)(nil)).Elem(); got != want {
			t.Errorf("out parameter #0 type mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.Out(1), reflect.TypeOf((*error)(nil)).Elem(); got != want {
			t.Errorf("out parameter #1 type mismatch, got=%v, want=%v", got, want)
		}
	})
	t.Run("round", func(t *testing.T) {
		ty := reflect.TypeOf(round)
		if got, want := ty.Kind(), reflect.Func; got != want {
			t.Errorf("type kind mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.NumIn(), 2; got != want {
			t.Errorf("input parameter count mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.In(0), reflect.TypeOf((*Value)(nil)).Elem(); got != want {
			t.Errorf("input parameter #0 type mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.In(1), reflect.TypeOf((*option.Option[int32])(nil)).Elem(); got != want {
			t.Errorf("input parameter #1 type mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.NumOut(), 2; got != want {
			t.Errorf("output parameter count mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.Out(0), reflect.TypeOf((*Value)(nil)).Elem(); got != want {
			t.Errorf("out parameter #0 type mismatch, got=%v, want=%v", got, want)
		}
		if got, want := ty.Out(1), reflect.TypeOf((*error)(nil)).Elem(); got != want {
			t.Errorf("out parameter #1 type mismatch, got=%v, want=%v", got, want)
		}
	})
}

func TestCallFiltersWithReflect(t *testing.T) {
	v := reflect.ValueOf(lower)
	arg := reflect.ValueOf("Hello")
	out := v.Call([]reflect.Value{arg})
	if got, want := len(out), 1; got != want {
		t.Errorf("out parameter count mismatch, got=%v, want=%v", got, want)
	}
	if got, want := out[0].Interface(), "hello"; got != want {
		t.Errorf("out parameter #1 mismatch, got=%v, want=%v", got, want)
	}
}

func TestReflectCallSlice(t *testing.T) {
	sum := func(s string, values ...int) int {
		var ret int
		for _, v := range values {
			ret += v
		}
		return ret
	}

	v := reflect.ValueOf(sum)
	s := reflect.ValueOf("Hello")
	values := reflect.ValueOf([]int{1, 2, 3})
	out := v.CallSlice([]reflect.Value{s, values})
	if got, want := len(out), 1; got != want {
		t.Errorf("out parameter count mismatch, got=%v, want=%v", got, want)
	}
	if got, want := out[0].Interface(), 6; got != want {
		t.Errorf("out parameter #1 mismatch, got=%v, want=%v", got, want)
	}
}

func TestReflectCallVaradic(t *testing.T) {
	sum := func(s string, values ...int) int {
		var ret int
		for _, v := range values {
			ret += v
		}
		return ret
	}

	v := reflect.ValueOf(sum)
	s := reflect.ValueOf("Hello")
	v1 := reflect.ValueOf(int(1))
	v2 := reflect.ValueOf(int(2))
	v3 := reflect.ValueOf(int(3))
	out := v.Call([]reflect.Value{s, v1, v2, v3})
	if got, want := len(out), 1; got != want {
		t.Errorf("out parameter count mismatch, got=%v, want=%v", got, want)
	}
	if got, want := out[0].Interface(), 6; got != want {
		t.Errorf("out parameter #1 mismatch, got=%v, want=%v", got, want)
	}
}

func TestVaradicFunc(t *testing.T) {
	// func mapFilter(state *vmState, val Value, args ...Value) ([]Value, error) {
	ty := reflect.TypeOf(mapFilter)
	if got, want := ty.NumIn(), 3; got != want {
		t.Errorf("in parameter count mismatch, got=%d, want=%d", got, want)
	}
	arg2Ty := ty.In(2)
	assertType[[]Value](arg2Ty, "arg 2 must be []Value")
	assertType[Value](arg2Ty.Elem(), "arg 2 elem must be Value")
	if got, want := ty.IsVariadic(), true; got != want {
		t.Errorf("arg 2 variadic mismatch, got=%v, want=%v", got, want)
	}
}
