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
