package mjingo

import (
	"reflect"
	"strings"
)

type BoxedTest = func(*State, []Value) (bool, error)

// 1 argument functions

func BoxTestFromFixedArity1ArgNoErrFunc[A JustOneArgTypes](f func(A) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		var a A
		var err error
		values, err = convertArgToGoVar(state, values, &a)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a), nil
	}
}

func BoxTestFromFixedArity1ArgWithErrFunc[A JustOneArgTypes](f func(A) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		var a A
		var err error
		values, err = convertArgToGoVar(state, values, &a)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a)
	}
}

func BoxTestFromVariadic1ArgNoErrFunc[A VariadicLastArgElemTypes](f func(...A) bool) BoxedTest {
	return func(*State, []Value) (bool, error) {
		var a []A
		// TODO implement
		return f(a...), nil
	}
}

func BoxTestFromVariadic1ArgWithErrFunc[A VariadicLastArgElemTypes](f func(...A) (bool, error)) BoxedTest {
	return func(*State, []Value) (bool, error) {
		var a []A
		// TODO implement
		return f(a...)
	}
}

// 2 argument functions

func BoxTestFromFixedArity2ArgNoErrFunc[A FirstArgTypes, B FixedArityLastArgTypes](f func(A, B) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		var a A
		var b B
		var err error
		values, err = convertArgToGoVar(state, values, &a)
		if err != nil {
			return false, err
		}
		values, err = convertArgToGoVar(state, values, &b)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b), nil
	}
}

func BoxTestFromFixedArity2ArgWithErrFunc[A FirstArgTypes, B FixedArityLastArgTypes](f func(A, B) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		var a A
		var b B
		var err error
		values, err = convertArgToGoVar(state, values, &a)
		if err != nil {
			return false, err
		}
		values, err = convertArgToGoVar(state, values, &b)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b)
	}
}

func BoxTestFromVariadic2ArgNoErrFunc[A FirstArgTypes, B VariadicLastArgElemTypes](f func(A, ...B) bool) BoxedTest {
	return func(*State, []Value) (bool, error) {
		var a A
		var b []B
		// TODO implement
		return f(a, b...), nil
	}
}

func BoxTestFromVariadic2ArgWithErrFunc[A FirstArgTypes, B VariadicLastArgElemTypes](f func(A, ...B) (bool, error)) BoxedTest {
	return func(*State, []Value) (bool, error) {
		var a A
		var b []B
		// TODO implement
		return f(a, b...)
	}
}

// 3 argument functions

func BoxTestFromFixedArity3ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C FixedArityLastArgTypes](f func(A, B, C) bool) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		var a A
		var b B
		var c C
		var err error
		values, err = convertArgToGoVar(state, values, &a)
		if err != nil {
			return false, err
		}
		values, err = convertArgToGoVar(state, values, &b)
		if err != nil {
			return false, err
		}
		values, err = convertArgToGoVar(state, values, &c)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c), nil
	}
}

func BoxTestFromFixedArity3ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C FixedArityLastArgTypes](f func(A, B, C) (bool, error)) BoxedTest {
	return func(state *State, values []Value) (bool, error) {
		var a A
		var b B
		var c C
		var err error
		values, err = convertArgToGoVar(state, values, &a)
		if err != nil {
			return false, err
		}
		values, err = convertArgToGoVar(state, values, &b)
		if err != nil {
			return false, err
		}
		values, err = convertArgToGoVar(state, values, &c)
		if err != nil {
			return false, err
		}
		if len(values) > 0 {
			return false, NewError(TooManyArguments, "")
		}
		return f(a, b, c)
	}
}

func BoxTestFromVariadic3ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C VariadicLastArgElemTypes](f func(A, B, ...C) bool) BoxedTest {
	return func(*State, []Value) (bool, error) {
		var a A
		var b B
		var c []C
		// TODO implement
		return f(a, b, c...), nil
	}
}

func BoxTestFromVariadic3ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C VariadicLastArgElemTypes](f func(A, B, ...C) (bool, error)) BoxedTest {
	return func(*State, []Value) (bool, error) {
		var a A
		var b B
		var c []C
		// TODO implement
		return f(a, b, c...)
	}
}

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

	argTypes := buildArgTypesOfFunc(fn)
	if err := checkArgTypes(argTypes); err != nil {
		panic(err.Error())
	}
	fnVal := reflect.ValueOf(fn)
	return func(state *State, values []Value) (bool, error) {
		goVals, err := argsToGoValuesReflect(state, values, argTypes)
		if err != nil {
			return false, err
		}
		reflectVals := make([]reflect.Value, len(goVals))
		for i, goVal := range goVals {
			if fnType.IsVariadic() && i == fnType.NumIn()-1 {
				reflectVals[i] = reflect.ValueOf(goVal).Convert(sliceTypeForRestTypeReflect(argTypes[i]))
			} else {
				reflectVals[i] = reflect.ValueOf(goVal)
			}
		}
		var retVals []reflect.Value
		if fnType.IsVariadic() {
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
func isNumber(val Value) bool { return val.kind() == valueKindNumber }

// Checks if this value is a string.
//
// ```jinja
// {{ "42" is string }} -> true
// {{ 42 is string }} -> false
// ```
func isString(val Value) bool { return val.kind() == valueKindString }

// Checks if this value is a sequence
//
// ```jinja
// {{ [1, 2, 3] is sequence }} -> true
// {{ 42 is sequence }} -> false
// ```
func isSequence(val Value) bool { return val.kind() == valueKindSeq }

// Checks if this value is a mapping
//
// ```jinja
// {{ {"foo": "bar"} is mapping }} -> true
// {{ [1, 2, 3] is mapping }} -> false
// ```
func isMapping(val Value) bool { return val.kind() == valueKindMap }

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
