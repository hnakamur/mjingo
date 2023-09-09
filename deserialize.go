package mjingo

import (
	"fmt"
	"reflect"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

func checkFuncArgTypes(fnType reflect.Type) (optionCount int) {
	numIn := fnType.NumIn()
	varadic := fnType.IsVariadic()
	for i := 0; i < numIn; i++ {
		argType := fnType.In(i)
		if (varadic && i == numIn-1) || argType.Kind() == reflect.Slice {
			argType = argType.Elem()
		}
		supported, optional := checkFuncArgType(argType, i)
		if !supported {
			panic(fmt.Sprintf("unsupported type for argument %d of function", i))
		}
		if optionCount > 0 && !optional && !(varadic && i == numIn-1) {
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
	case reflectType[*vmState]():
		return argPos == 0, false
	case reflectType[Value](), reflectType[string](), reflectType[uint](),
		reflectType[uint32](), reflectType[int32](), reflectType[bool]():
		return true, false
	case reflectType[option.Option[Value]](), reflectType[option.Option[string]](),
		reflectType[option.Option[int32]](), reflectType[option.Option[uint32]](),
		reflectType[option.Option[bool]](),
		reflectType[kwArgs]():
		return true, true
	}
	return false, false
}

func goValueFromValue(val Value, destType reflect.Type) (any, error) {
	switch destType {
	case reflectType[Value]():
		return val, nil
	case reflectType[bool]():
		return boolTryFromValue(val)
	case reflectType[int8]():
		return i8TryFromValue(val)
	case reflectType[int16]():
		return i16TryFromValue(val)
	case reflectType[int32]():
		return i32TryFromValue(val)
	case reflectType[int64]():
		return i64TryFromValue(val)
	case reflectType[int]():
		return intTryFromValue(val)
	case reflectType[uint8]():
		return u8TryFromValue(val)
	case reflectType[uint16]():
		return u16TryFromValue(val)
	case reflectType[uint32]():
		return u32TryFromValue(val)
	case reflectType[uint64]():
		return u64TryFromValue(val)
	case reflectType[uint]():
		return uintTryFromValue(val)
	case reflectType[string]():
		return stringFromValue(option.Some(val))
	case reflectType[kwArgs]():
		return kwArgsTryFromValue(val)
	case reflectType[option.Option[Value]]():
		return goOptValueFromValue(val, func(val Value) (Value, error) { return val, nil })
	case reflectType[option.Option[bool]]():
		return goOptValueFromValue(val, boolTryFromValue)
	case reflectType[option.Option[int8]]():
		return goOptValueFromValue(val, i8TryFromValue)
	case reflectType[option.Option[int16]]():
		return goOptValueFromValue(val, i16TryFromValue)
	case reflectType[option.Option[int32]]():
		return goOptValueFromValue(val, i32TryFromValue)
	case reflectType[option.Option[int64]]():
		return goOptValueFromValue(val, i64TryFromValue)
	case reflectType[option.Option[int]]():
		return goOptValueFromValue(val, intTryFromValue)
	case reflectType[option.Option[uint]]():
		return goOptValueFromValue(val, uintTryFromValue)
	case reflectType[option.Option[uint8]]():
		return goOptValueFromValue(val, u8TryFromValue)
	case reflectType[option.Option[uint16]]():
		return goOptValueFromValue(val, u16TryFromValue)
	case reflectType[option.Option[uint32]]():
		return goOptValueFromValue(val, u32TryFromValue)
	case reflectType[option.Option[uint64]]():
		return goOptValueFromValue(val, u64TryFromValue)
	case reflectType[option.Option[string]]():
		return goOptValueFromValue(val, func(val Value) (string, error) {
			return stringFromValue(option.Some(val))
		})
	case reflectType[[]Value]():
		return valueSliceTryFromValue(val)
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
