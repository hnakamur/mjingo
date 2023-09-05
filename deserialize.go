package mjingo

import (
	"fmt"
	"reflect"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

func checkFuncArgTypes(fnType reflect.Type) (optionCount int) {
	numIn := fnType.NumIn()
	for i := 0; i < numIn; i++ {
		supported, optional := checkFuncArgType(fnType.In(i), i)
		if !supported {
			panic(fmt.Sprintf("unsupported type for argument %d of function", i))
		}
		if optionCount > 0 && !optional {
			panic(fmt.Sprintf("non optional type of argument %d after optional type of argument in function", i))
		}
		if optional {
			optionCount++
		}
	}
	return
}

func checkFuncArgType(argType reflect.Type, argPos int) (supported, optional bool) {
	switch argType {
	case typeFromPtr((**vmState)(nil)):
		return argPos == 0, false
	case typeFromPtr((*Value)(nil)), typeFromPtr((*string)(nil)), typeFromPtr((*uint)(nil)),
		typeFromPtr((*int32)(nil)), typeFromPtr((*bool)(nil)):
		return true, false
	case typeFromPtr((*option.Option[Value])(nil)), typeFromPtr((*option.Option[string])(nil)),
		typeFromPtr((*option.Option[int32])(nil)), typeFromPtr((*option.Option[bool])(nil)):
		return true, true
	}
	return false, false
}

func goValueFromValue(val Value, destType reflect.Type) (any, error) {
	switch destType {
	case typeFromPtr((*Value)(nil)):
		return val, nil
	case typeFromPtr((*string)(nil)):
		return stringFromValue(option.Some(val))
	case typeFromPtr((*uint)(nil)):
		return uintTryFromValue(val)
	case typeFromPtr((*int32)(nil)):
		return i32TryFromValue(val)
	case typeFromPtr((*bool)(nil)):
		return boolTryFromValue(val)
	case typeFromPtr((*option.Option[Value])(nil)):
		if val == nil {
			return option.None[Value](), nil
		}
		return option.Some(val), nil
	case typeFromPtr((*option.Option[string])(nil)):
		return goOptValueFromValue(val, func(val Value) (string, error) {
			return stringFromValue(option.Some(val))
		})
	case typeFromPtr((*option.Option[int32])(nil)):
		return goOptValueFromValue(val, i32TryFromValue)
	case typeFromPtr((*option.Option[bool])(nil)):
		return goOptValueFromValue(val, boolTryFromValue)
	}
	panic("unsupported destination type")
}

func goOptValueFromValue[T any](val Value, fn func(val Value) (T, error)) (any, error) {
	if val == nil {
		return option.None[T](), nil
	}
	return mapResultOK(option.Some[T])(fn(val))
}

func mapResultOK[T any, U any](fn func(ok T) U) func(ok T, err error) (U, error) {
	return func(ok T, err error) (U, error) {
		if err != nil {
			var zero U
			return zero, err
		}
		return fn(ok), nil
	}
}
