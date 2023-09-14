package mjingo

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/hnakamur/mjingo/option"
)

func TestCheckArgTypes(t *testing.T) {
	t.Run("fixedAry", func(t *testing.T) {
		testCases := [][]reflect.Type{
			{reflectType[*State](), reflectType[int]()},
			{reflectType[*State](), reflectType[int](), reflectType[option.Option[int]]()},
			{reflectType[*State](), reflectType[int](), reflectType[option.Option[int]](),
				reflectType[option.Option[string]](), reflectType[Kwargs]()},
			{reflectType[[]int](), reflectType[string]()},
		}
		for i, tc := range testCases {
			err := checkArgTypes(tc, false)
			if err != nil {
				t.Errorf("must not get error but got error, i=%d, err=%v", i, err)
			}
		}
	})
	t.Run("variadic", func(t *testing.T) {
		testCases := [][]reflect.Type{
			{reflectType[*State](), reflectType[int](), reflectType[option.Option[int]](),
				reflectType[[]int]()},
		}
		for i, tc := range testCases {
			err := checkArgTypes(tc, true)
			if err != nil {
				t.Errorf("must not get error but got error, i=%d, err=%v", i, err)
			}
		}
	})
	t.Run("error", func(t *testing.T) {
		testCases := []struct {
			argTypes []reflect.Type
			variadic bool
			detail   string
		}{
			{
				argTypes: []reflect.Type{reflectType[int](), reflectType[*State]()},
				variadic: false,
				detail:   "argument of State type must be the first argument",
			},
			{
				argTypes: []reflect.Type{reflectType[option.Option[int]](), reflectType[string]()},
				variadic: false,
				detail:   "argument of non-optional type cannot be after argument of optional type",
			},
			{
				argTypes: []reflect.Type{
					reflectType[*State](), reflectType[int](), reflectType[option.Option[int]](),
					reflectType[[]string]()},
				variadic: false,
				detail:   "argument of non-optional type cannot be after argument of optional type",
			},
			{
				argTypes: []reflect.Type{reflectType[Kwargs](), reflectType[string]()},
				variadic: false,
				detail:   "argument of Kwargs type must be the last argument",
			},
			{
				argTypes: []reflect.Type{reflectType[string](), reflectType[complex64]()},
				variadic: false,
				detail:   "argument type complex64 is unsupported",
			},
		}
		for i, tc := range testCases {
			err := checkArgTypes(tc.argTypes, tc.variadic)
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

func TestSliceOfSameType(t *testing.T) {
	f := func(ptr any) {
		switch p := ptr.(type) {
		case *[]string:
			*p = append(*p, "ab")
			*p = append(*p, "cd")
		}
	}

	var a []string
	f(&a)
	if got, want := a, []string{"ab", "cd"}; !slices.Equal(got, want) {
		t.Errorf("result mismatch, got=%v, want=%v", got, want)
	}
}

func TestDestPtrs(t *testing.T) {
	var a int
	var b string
	var c []float64
	destPtrs := []any{&a, &b, &c}
	for i, destPtr := range destPtrs {
		var want string
		switch i {
		case 0:
			want = "*int"
		case 1:
			want = "*string"
		case 2:
			want = "*[]float64"
		}
		got := fmt.Sprintf("%T", destPtr)
		if got != want {
			t.Errorf("pointer type mismatch, i=%d, got=%s, want=%s", i, got, want)
		}
	}
}

func testGenericFuncCallerNoErr[A FirstArgTypes, B FixedArityLastArgTypes, R RetValTypes](a A, b B, f func(A, B) R) R {
	return f(a, b)
}

func testGenericFuncCallerWithErr[A FirstArgTypes, B FixedArityLastArgTypes, R RetValTypes](f func(A, B) (R, error)) (R, error) {
	var a A
	var b B
	return f(a, b)
}

func testGenericFuncCallerVariadicNoErr[A FirstArgTypes, B FixedArityLastArgTypes, R RetValTypes](f func(A, ...B) R) R {
	var a A
	var b []B
	return f(a, b...)
}

func testGenericFuncCallerVariadicWithErr[A FirstArgTypes, B FixedArityLastArgTypes, R RetValTypes](f func(A, ...B) (R, error)) (R, error) {
	var a A
	var b []B
	return f(a, b...)
}

func TestCallGenericFunc(t *testing.T) {
	t.Run("case1", func(t *testing.T) {
		got := testGenericFuncCallerNoErr[bool, string, string](false, "", func(a bool, b string) string {
			return fmt.Sprintf("%T %T", a, b)
		})
		const want = "bool string"
		if got != want {
			t.Errorf("result mismatch, got=%s, want=%s", got, want)
		}
	})
	t.Run("byteRetVal", func(t *testing.T) {
		got := testGenericFuncCallerNoErr[bool, string, byte](false, "", func(_a bool, _b string) byte {
			return 'a'
		})
		const want = 'a'
		if got != want {
			t.Errorf("result mismatch, got=%v, want=%v", got, want)
		}
	})
	t.Run("byteSlice", func(t *testing.T) {
		got := testGenericFuncCallerNoErr[[]byte, bool]([]byte("foo"), false, func(a []byte, _b bool) string {
			return string(a)
		})
		const want = "foo"
		if got != want {
			t.Errorf("result mismatch, got=%v, want=%v", got, want)
		}
	})
}
