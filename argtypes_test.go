package mjingo

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/hnakamur/mjingo/option"
)

func TestArgsTo1GoValue(t *testing.T) {
	intEqual := func(a, b int) bool { return a == b }

	t.Run("State", func(t *testing.T) {
		st, err := ArgsTo1GoValue[State](State(nil), nil)
		if err != nil {
			t.Errorf("err mismatch, got=%v, want=%v", err, nil)
		}
		if got, want := st, State(nil); got != want {
			t.Errorf("state mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
		}
	})
	t.Run("int", func(t *testing.T) {
		n, err := ArgsTo1GoValue[int](State(nil), []Value{valueFromI64(3)})
		if err != nil {
			t.Errorf("err mismatch, got=%v, want=%v", err, nil)
		}
		if got, want := n, 3; got != want {
			t.Errorf("n mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
		}
	})
	t.Run("Option[int]", func(t *testing.T) {
		t.Run("None", func(t *testing.T) {
			n, err := ArgsTo1GoValue[option.Option[int]](State(nil), nil)
			if err != nil {
				t.Errorf("err mismatch, got=%v, want=%v", err, nil)
			}
			if got, want := n, option.None[int](); !got.Equal(want, intEqual) {
				t.Errorf("n mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
			}
		})
		t.Run("Some", func(t *testing.T) {
			n, err := ArgsTo1GoValue[option.Option[int]](State(nil), []Value{valueFromI64(3)})
			if err != nil {
				t.Errorf("err mismatch, got=%v, want=%v", err, nil)
			}
			if got, want := n, option.Some(3); !got.Equal(want, intEqual) {
				t.Errorf("n mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
			}
		})
	})
	t.Run("[]int", func(t *testing.T) {
		slice, err := ArgsTo1GoValue[[]int](State(nil), []Value{valueFromI64(3), valueFromI64(4)})
		if err != nil {
			t.Errorf("err mismatch, got=%v, want=%v", err, nil)
		}
		if got, want := slice, []int{3, 4}; !slices.Equal(got, want) {
			t.Errorf("ret mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
		}
	})
	t.Run("Rest[string]", func(t *testing.T) {
		rest, err := ArgsTo1GoValue[Rest[string]](State(nil),
			[]Value{valueFromString("ab"), valueFromString("cd")})
		if err != nil {
			t.Errorf("err mismatch, got=%v, want=%v", err, nil)
		}
		if got, want := rest, Rest[string]([]string{"ab", "cd"}); !slices.Equal(got, want) {
			t.Errorf("ret mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
		}
	})
	t.Run("Kwargs", func(t *testing.T) {
		t.Run("WithValue", func(t *testing.T) {
			kwargs, err := ArgsTo1GoValue[Kwargs](State(nil),
				[]Value{valueFromKwargs(newKwargs(*valueMapFromEntries([]valueMapEntry{
					{Key: keyRefFromString("a"), Value: valueFromI64(3)},
				})))})
			if err != nil {
				t.Errorf("err mismatch, got=%v, want=%v", err, nil)
			}
			if got, want := kwargs.PeekValue("a"), option.Some(valueFromI64(3)); got.Compare(want, valueCmp) != 0 {
				t.Errorf("ret mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
			}
		})
		t.Run("NoValue", func(t *testing.T) {
			kwargs, err := ArgsTo1GoValue[Kwargs](State(nil), nil)
			if err != nil {
				t.Errorf("err mismatch, got=%v, want=%v", err, nil)
			}
			if got, want := kwargs.PeekValue("a"), option.None[Value](); got.Compare(want, valueCmp) != 0 {
				t.Errorf("ret mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
			}
		})
	})
}

func TestArgsTo2GoValues(t *testing.T) {
	t.Run("StateInt", func(t *testing.T) {
		st, n, err := ArgsTo2GoValues[State, int](State(nil), []Value{valueFromI64(3)})
		if err != nil {
			t.Errorf("err mismatch, got=%v, want=%v", err, nil)
		}
		if got, want := st, State(nil); got != want {
			t.Errorf("ret#0 mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
		}
		if got, want := n, 3; got != want {
			t.Errorf("ret#1 mismatch, got=%v (%T), want=%v (%T)", got, got, want, want)
		}
	})
	t.Run("MissingArgument", func(t *testing.T) {
		_, _, err := ArgsTo2GoValues[State, int](State(nil), nil)
		if err != nil {
			var merr *Error
			if errors.As(err, &merr) && merr.Type() == MissingArgument {
				return
			}
		}
		t.Errorf("err mismatch, got=%v, want=%v", err, NewError(MissingArgument, ""))
	})
	t.Run("TooManyArguments", func(t *testing.T) {
		_, _, err := ArgsTo2GoValues[State, int](State(nil),
			[]Value{valueFromI64(3), valueFromI64(4)})
		if err != nil {
			var merr *Error
			if errors.As(err, &merr) && merr.Type() == TooManyArguments {
				return
			}
		}
		t.Errorf("err mismatch, got=%v, want=%v", err, NewError(TooManyArguments, ""))
	})
}

func TestCheckArgTypes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testCases := [][]reflect.Type{
			{reflectType[State](), reflectType[int]()},
			{reflectType[State](), reflectType[int](), reflectType[option.Option[int]]()},
			{reflectType[State](), reflectType[int](), reflectType[option.Option[int]](),
				reflectType[Rest[int]]()},
			{reflectType[State](), reflectType[int](), reflectType[option.Option[int]](),
				reflectType[[]string]()},
			{reflectType[State](), reflectType[int](), reflectType[option.Option[int]](),
				reflectType[option.Option[string]](), reflectType[Kwargs]()},
		}
		for i, tc := range testCases {
			err := checkArgTypes(tc)
			if err != nil {
				t.Errorf("must not get error but got error, i=%d, err=%v", i, err)
			}
		}
	})
	t.Run("error", func(t *testing.T) {
		testCases := []struct {
			argTypes []reflect.Type
			detail   string
		}{
			{
				argTypes: []reflect.Type{reflectType[int](), reflectType[State]()},
				detail:   "argument of State type must be the first argument",
			},
			{
				argTypes: []reflect.Type{reflectType[option.Option[int]](), reflectType[string]()},
				detail:   "argument of non-optional type cannot be after argument of optional type",
			},
			{
				argTypes: []reflect.Type{reflectType[[]int](), reflectType[string]()},
				detail:   "argument of slice type must be the last argument",
			},
			{
				argTypes: []reflect.Type{reflectType[Rest[int]](), reflectType[string]()},
				detail:   "argument of Rest type must be the last argument",
			},
			{
				argTypes: []reflect.Type{reflectType[Kwargs](), reflectType[string]()},
				detail:   "argument of Kwargs type must be the last argument",
			},
			{
				argTypes: []reflect.Type{reflectType[string](), reflectType[complex64]()},
				detail:   "argument type complex64 is unsupported",
			},
		}
		for i, tc := range testCases {
			err := checkArgTypes(tc.argTypes)
			if err != nil {
				var merr *Error
				if errors.As(err, &merr) {
					if got, want := merr.Type(), InvalidOperation; got != want {
						t.Errorf("error Type() mismatch, i=%d, got=%v, want=%v", i, got, want)
					}
					if got, want := merr.detail, tc.detail; got != want {
						t.Errorf("error detail mismatch, i=%d, got=%v, want=%v", i, got, want)
					}
				} else {
					t.Errorf("error Go type mismatch, i=%d, got=%T, want=%T", i, err, (*Error)(nil))
				}
			} else {
				t.Errorf("must get error but no error, i=%d", i)
			}
		}
	})
}
