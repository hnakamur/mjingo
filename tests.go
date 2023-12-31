package mjingo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hnakamur/mjingo/internal/rustfmt"
)

// BoxedTest is the type of a boxed test.
//
// A boxed test can be registered as test functions to the environment via
// [Environment.AddTest].
//
// A boxed test can be created by wrapping a "unboxed" test using one of BoxedTextFrom* function.
// Or a user can define a boxed test directly.
// In that case, [ConvertArgToGoValue] and [ConvertVariadicArgsToGoValue] can be used to convert a Value
// to Go's data type.
type BoxedTest = func(*State, []Value) (bool, error)

// 1 argument functions

// BoxedTestFromFixedArity1ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity1ArgNoErrFunc[A JustOneArgTypes](f func(A) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a), nil
	}
}

// BoxedTestFromFixedArity1ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity1ArgWithErrFunc[A JustOneArgTypes](f func(A) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a)
	}
}

// BoxedTestFromVariadic1ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic1ArgNoErrFunc[A VariadicArgElemTypes](f func(...A) bool) BoxedTest {
	return func(_state *State, values []Value) (bool, error) {
		a, err := ConvertVariadicArgsToGoValue[[]A, A](values)
		if err != nil {
			return false, err
		}
		return f(a...), nil
	}
}

// BoxedTestFromVariadic1ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic1ArgWithErrFunc[A VariadicArgElemTypes](f func(...A) (bool, error)) BoxedTest {
	return func(_state *State, values []Value) (bool, error) {
		a, err := ConvertVariadicArgsToGoValue[[]A, A](values)
		if err != nil {
			return false, err
		}
		return f(a...)
	}
}

// 2 argument functions

// BoxedTestFromFixedArity2ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity2ArgNoErrFunc[A FirstArgTypes, B FixedArityLastArgTypes](f func(A, B) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b), nil
	}
}

// BoxedTestFromFixedArity2ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity2ArgWithErrFunc[A FirstArgTypes, B FixedArityLastArgTypes](f func(A, B) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b)
	}
}

// BoxedTestFromVariadic2ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic2ArgNoErrFunc[A FirstArgTypes, B VariadicArgElemTypes](f func(A, ...B) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, err := ConvertVariadicArgsToGoValue[[]B, B](values)
		if err != nil {
			return false, err
		}
		return f(a, b...), nil
	}
}

// BoxedTestFromVariadic2ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic2ArgWithErrFunc[A FirstArgTypes, B VariadicArgElemTypes](f func(A, ...B) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, err := ConvertVariadicArgsToGoValue[[]B, B](values)
		if err != nil {
			return false, err
		}
		return f(a, b...)
	}
}

// 3 argument functions

// BoxedTestFromFixedArity3ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity3ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C FixedArityLastArgTypes](f func(A, B, C) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c), nil
	}
}

// BoxedTestFromFixedArity3ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity3ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C FixedArityLastArgTypes](f func(A, B, C) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c)
	}
}

// BoxedTestFromVariadic3ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic3ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C VariadicArgElemTypes](f func(A, B, ...C) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, err := ConvertVariadicArgsToGoValue[[]C, C](values)
		if err != nil {
			return false, err
		}
		return f(a, b, c...), nil
	}
}

// BoxedTestFromVariadic3ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic3ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C VariadicArgElemTypes](f func(A, B, ...C) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, err := ConvertVariadicArgsToGoValue[[]C, C](values)
		if err != nil {
			return false, err
		}
		return f(a, b, c...)
	}
}

// 4 argument functions

// BoxedTestFromFixedArity4ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity4ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D FixedArityLastArgTypes](f func(A, B, C, D) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c, d), nil
	}
}

// BoxedTestFromFixedArity4ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity4ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D FixedArityLastArgTypes](f func(A, B, C, D) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c, d)
	}
}

// BoxedTestFromVariadic4ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic4ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D VariadicArgElemTypes](f func(A, B, C, ...D) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, err := ConvertVariadicArgsToGoValue[[]D, D](values)
		if err != nil {
			return false, err
		}
		return f(a, b, c, d...), nil
	}
}

// BoxedTestFromVariadic4ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic4ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D VariadicArgElemTypes](f func(A, B, C, ...D) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, err := ConvertVariadicArgsToGoValue[[]D, D](values)
		if err != nil {
			return false, err
		}
		return f(a, b, c, d...)
	}
}

// 5 argument functions

// BoxedTestFromFixedArity5ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity5ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E FixedArityLastArgTypes](f func(A, B, C, D, E) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return false, err
		}
		e, values, err := ConvertArgToGoValue[E](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c, d, e), nil
	}
}

// BoxedTestFromFixedArity5ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromFixedArity5ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E FixedArityLastArgTypes](f func(A, B, C, D, E) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return false, err
		}
		e, values, err := ConvertArgToGoValue[E](state, values)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c, d, e)
	}
}

// BoxedTestFromVariadic5ArgNoErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic5ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E VariadicArgElemTypes](f func(A, B, C, D, ...E) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return false, err
		}
		e, err := ConvertVariadicArgsToGoValue[[]E, E](values)
		if err != nil {
			return false, err
		}
		return f(a, b, c, d, e...), nil
	}
}

// BoxedTestFromVariadic5ArgWithErrFunc creates a boxed function which wraps f.
func BoxedTestFromVariadic5ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E VariadicArgElemTypes](f func(A, B, C, D, ...E) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return false, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return false, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return false, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return false, err
		}
		e, err := ConvertVariadicArgsToGoValue[[]E, E](values)
		if err != nil {
			return false, err
		}
		return f(a, b, c, d, e...)
	}
}

// BoxedTestFromFuncReflect creates a boxed test which wraps f using Go's reflect package.
//
// This may be slower so caller should prefer generic BoxedTestFrom* functions.
func BoxedTestFromFuncReflect(fn any) BoxedTest {
	if bt, ok := fn.(BoxedTest); ok {
		return bt
	}

	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("argument must be a function")
	}

	numOut := fnType.NumOut()
	if numOut != 1 && numOut != 2 {
		panic("return value count must be 1 or 2")
	}
	assertType[bool](fnType.Out(0), "type of first return value must be bool")
	if numOut == 2 {
		assertType[error](fnType.Out(1), "type of seond return value must be error")
	}

	variadic := fnType.IsVariadic()
	argTypes := buildArgTypesOfFunc(fn)
	if err := checkArgTypes(argTypes, variadic); err != nil {
		panic(err.Error())
	}
	fnVal := reflect.ValueOf(fn)
	return func(state *State, values []Value) (bool, error) {
		goVals, err := argsToGoValuesReflect(state, values, argTypes, variadic)
		if err != nil {
			return false, err
		}
		reflectVals := make([]reflect.Value, len(goVals))
		for i, goVal := range goVals {
			reflectVals[i] = reflect.ValueOf(goVal)
		}
		var retVals []reflect.Value
		if variadic {
			retVals = fnVal.CallSlice(reflectVals)
		} else {
			retVals = fnVal.Call(reflectVals)
		}
		switch len(retVals) {
		case 1:
			return retVals[0].Interface().(bool), nil
		case 2:
			retVal0 := retVals[0].Interface().(bool)
			retVal1 := retVals[1].Interface()
			if retVal1 != nil {
				return retVal0, retVal1.(error)
			}
			return retVal0, nil
		}
		panic("unreachable")
	}
}

func isUndefined(val Value) bool {
	return val.isUndefined()
}

func isDefined(val Value) bool {
	return !val.isUndefined()
}

func isNone(val Value) bool {
	return val.isNone()
}

func isSafe(val Value) bool {
	return val.isSafe()
}

// Checks if a value is odd.
//
// ```jinja
// {{ 41 is odd }} -> true
// ```
func isOdd(val Value) bool {
	x, err := val.tryToI64()
	if err != nil {
		return false
	}
	return x%2 != 0
}

// Checks if a value is even.
//
// ```jinja
// {{ 42 is even }} -> true
// ```
func isEven(val Value) bool {
	x, err := val.tryToI64()
	if err != nil {
		return false
	}
	return x%2 == 0
}

// Checks if this value is a number.
//
// ```jinja
// {{ 42 is number }} -> true
// {{ "42" is number }} -> false
// ```
func isNumber(val Value) bool { return val.Kind() == ValueKindNumber }

// Checks if this value is a string.
//
// ```jinja
// {{ "42" is string }} -> true
// {{ 42 is string }} -> false
// ```
func isString(val Value) bool { return val.Kind() == ValueKindString }

// Checks if this value is a sequence
//
// ```jinja
// {{ [1, 2, 3] is sequence }} -> true
// {{ 42 is sequence }} -> false
// ```
func isSequence(val Value) bool { return val.Kind() == ValueKindSeq }

// Checks if this value is a mapping
//
// ```jinja
// {{ {"foo": "bar"} is mapping }} -> true
// {{ [1, 2, 3] is mapping }} -> false
// ```
func isMapping(val Value) bool { return val.Kind() == ValueKindMap }

// Checks if the value is starting with a string.
//
// ```jinja
// {{ "foobar" is startingwith("foo") }} -> true
// {{ "foobar" is startingwith("bar") }} -> false
// ```
func isStartingWith(v, other string) bool { return strings.HasPrefix(v, other) }

// Checks if the value is ending with a string.
//
// ```jinja
// {{ "foobar" is endingwith("bar") }} -> true
// {{ "foobar" is endingwith("foo") }} -> false
// ```
func isEndingWith(v, other string) bool { return strings.HasSuffix(v, other) }

func isEq(val, other Value) bool { return valueEqual(val, other) }
func isNe(val, other Value) bool { return !valueEqual(val, other) }
func isLt(val, other Value) bool { return valueCmp(val, other) < 0 }
func isLe(val, other Value) bool { return valueCmp(val, other) <= 0 }
func isGt(val, other Value) bool { return valueCmp(val, other) > 0 }
func isGe(val, other Value) bool { return valueCmp(val, other) >= 0 }

func isIn(val, other Value) bool {
	b, err := opContains(other, val)
	if err != nil {
		return false
	}
	return b.isTrue()
}

func isTrue(val Value) bool {
	boolVal, ok := val.data.(boolValue)
	return ok && boolVal.B
}

func isFalse(val Value) bool {
	boolVal, ok := val.data.(boolValue)
	return ok && !boolVal.B
}

func isFilter(state *State, name string) bool {
	return state.Env().getFilter(name).IsSome()
}

func isTest(state *State, name string) bool {
	return state.Env().getTest(name).IsSome()
}

type testObject struct {
	name string
	test BoxedTest
}

var _ = (Object)(testObject{})
var _ = (Caller)(testObject{})
var _ = (rustfmt.Formatter)(testObject{})

func newTestObject(name string, test BoxedTest) testObject {
	return testObject{name: name, test: test}
}

func (testObject) Kind() ObjectKind { return ObjectKindPlain }

func (to testObject) Call(state *State, args []Value) (Value, error) {
	rv, err := to.test(state, args)
	if err != nil {
		return Value{}, err
	}
	return valueFromBool(rv), nil
}

// SupportRustFormat implements rustfmt.Formatter.
func (testObject) SupportsCustomVerb(verb rune) bool {
	return verb == rustfmt.DebugVerb || verb == rustfmt.DisplayVerb
}

// Format implements rustfmt.Formatter.
func (to testObject) Format(f fmt.State, verb rune) {
	switch verb {
	case rustfmt.DisplayVerb, rustfmt.DebugVerb:
		fmt.Fprintf(f, "<test %s>", to.name)
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods testObject
		type testObject hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), testObject(to))
	}
}
