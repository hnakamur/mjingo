package mjingo

import (
	"fmt"
	"reflect"

	"github.com/hnakamur/mjingo/option"
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
	case reflectType[State]():
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

func ValueTryToGoValue[T any](val Value) (T, error) {
	goVal, err := valueTryToGoValueReflect(val, reflectType[T]())
	if err != nil {
		var zero T
		return zero, err
	}
	return goVal.(T), nil
}

func valueTryToGoValueReflect(val Value, destType reflect.Type) (any, error) {
	switch destType {
	case reflectType[Value]():
		return val, nil
	case reflectType[bool]():
		return valueTryToGoBool(val)
	case reflectType[int8]():
		return valueTryToGoInt8(val)
	case reflectType[int16]():
		return valueTryToGoInt16(val)
	case reflectType[int32]():
		return valueTryToGoInt32(val)
	case reflectType[int64]():
		return valueTryToGoInt64(val)
	case reflectType[int]():
		return valueTryToGoInt(val)
	case reflectType[uint8]():
		return valueTryToGoUint8(val)
	case reflectType[uint16]():
		return valueTryToGoUint16(val)
	case reflectType[uint32]():
		return valueTryToGoUint32(val)
	case reflectType[uint64]():
		return valueTryToGoUint64(val)
	case reflectType[uint]():
		return valueTryToGoUint(val)
	case reflectType[string]():
		return stringFromValue(option.Some(val))
	case reflectType[kwArgs]():
		return valueTryToKwArgs(val)
	case reflectType[option.Option[Value]]():
		return valueTryToOption(val, func(val Value) (Value, error) { return val, nil })
	case reflectType[option.Option[bool]]():
		return valueTryToOption(val, valueTryToGoBool)
	case reflectType[option.Option[int8]]():
		return valueTryToOption(val, valueTryToGoInt8)
	case reflectType[option.Option[int16]]():
		return valueTryToOption(val, valueTryToGoInt16)
	case reflectType[option.Option[int32]]():
		return valueTryToOption(val, valueTryToGoInt32)
	case reflectType[option.Option[int64]]():
		return valueTryToOption(val, valueTryToGoInt64)
	case reflectType[option.Option[int]]():
		return valueTryToOption(val, valueTryToGoInt)
	case reflectType[option.Option[uint]]():
		return valueTryToOption(val, valueTryToGoUint)
	case reflectType[option.Option[uint8]]():
		return valueTryToOption(val, valueTryToGoUint8)
	case reflectType[option.Option[uint16]]():
		return valueTryToOption(val, valueTryToGoUint16)
	case reflectType[option.Option[uint32]]():
		return valueTryToOption(val, valueTryToGoUint32)
	case reflectType[option.Option[uint64]]():
		return valueTryToOption(val, valueTryToGoUint64)
	case reflectType[option.Option[string]]():
		return valueTryToOption(val, func(val Value) (string, error) {
			return stringFromValue(option.Some(val))
		})
	case reflectType[[]Value]():
		return valueTryToValueSlice(val)
	}
	panic("unsupported destination type")
}

func valueTryToGoStringWithAsStr(val Value) (string, error) {
	optStr := val.asStr()
	if optStr.IsSome() {
		return optStr.Unwrap(), nil
	}
	return "", NewError(InvalidOperation, "value is not a string")
}

func valueTryToGoString(val Value) (string, error) {
	// TODO: compare benchmark with implementation using asStr().
	if v, ok := val.(stringValue); ok {
		return v.Str, nil
	}
	return "", NewError(InvalidOperation, "value is not a string")
}

func valueTryToOption[T any](val Value, fn func(val Value) (T, error)) (option.Option[T], error) {
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
