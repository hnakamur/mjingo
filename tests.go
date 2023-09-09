package mjingo

import (
	"reflect"
	"strings"
)

type BoxedTest = func(State, []Value) (bool, error)

func boxedTestFromFunc(fn any) BoxedTest {
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

	numIn := fnType.NumIn()
	if numIn < 1 && numIn > 3 {
		panic("only functions with argument count between 1 and 3 are supported")
	}
	checkFuncArgTypes(fnType)

	fnVal := reflect.ValueOf(fn)
	return func(state State, values []Value) (bool, error) {
		reflectVals := make([]reflect.Value, 0, numIn)
		inOffset := 0
		if fnType.In(0) == reflectType[State]() {
			reflectVals = append(reflectVals, reflect.ValueOf(state))
			inOffset++
		}
		wantValuesLen := numIn - inOffset
		if len(values) < wantValuesLen {
			return false, newError(MissingArgument, "")
		}
		if len(values) > wantValuesLen {
			return false, newError(TooManyArguments, "")
		}
		for i, val := range values {
			goVal, err := goValueFromValue(val, fnType.In(i+inOffset))
			if err != nil {
				return false, err
			}
			reflectVals = append(reflectVals, reflect.ValueOf(goVal))
		}
		ret := fnVal.Call(reflectVals)
		switch len(ret) {
		case 1:
			return ret[0].Interface().(bool), nil
		case 2:
			return ret[0].Interface().(bool), ret[1].Interface().(error)
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
	boolVal, ok := val.(boolValue)
	return ok && boolVal.B
}

func isFalse(val Value) bool {
	boolVal, ok := val.(boolValue)
	return ok && !boolVal.B
}

func isFilter(state State, name string) bool {
	return state.Env().getFilter(name).IsSome()
}

func isTest(state State, name string) bool {
	return state.Env().getTest(name).IsSome()
}
