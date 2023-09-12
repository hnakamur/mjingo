package mjingo

import (
	"reflect"

	"github.com/hnakamur/mjingo/option"
)

func valueTryToGoValueReflect(val Value, destType reflect.Type) (any, error) {
	switch destType {
	case reflectType[Value]():
		return valueTryToValue(val)
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
	case reflectType[float32]():
		return valueTryToGoFloat32(val)
	case reflectType[float64]():
		return valueTryToGoFloat64(val)
	case reflectType[string]():
		return valueTryToGoString(val)
	case reflectType[Kwargs]():
		return valueTryToKwargs(val)
	case reflectType[option.Option[Value]]():
		return valueTryToOption(val, valueTryToValue)
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
	case reflectType[option.Option[float32]]():
		return valueTryToOption(val, valueTryToGoFloat32)
	case reflectType[option.Option[float64]]():
		return valueTryToOption(val, valueTryToGoFloat64)
	case reflectType[option.Option[string]]():
		return valueTryToOption(val, valueTryToGoString)
	case reflectType[[]Value]():
		return valueTryToValueSlice(val)
	}
	panic("unsupported destination type")
}

func ValueTryToGoValue[T any](val Value) (T, error) {
	var ret T
	err := valueTryToGoValueNoReflect(val, &ret)
	return ret, err
}

func valueTryToGoValueNoReflect(val Value, destPtr any) error {
	switch p := destPtr.(type) {
	case *Value:
		*p = val
	case *bool:
		return valueTryToGoValueHelper[bool](val, p, valueTryToGoBool)
	case *int8:
		return valueTryToGoValueHelper[int8](val, p, valueTryToGoInt8)
	case *int16:
		return valueTryToGoValueHelper[int16](val, p, valueTryToGoInt16)
	case *int32:
		return valueTryToGoValueHelper[int32](val, p, valueTryToGoInt32)
	case *int64:
		return valueTryToGoValueHelper[int64](val, p, valueTryToGoInt64)
	case *int:
		return valueTryToGoValueHelper[int](val, p, valueTryToGoInt)
	case *uint8:
		return valueTryToGoValueHelper[uint8](val, p, valueTryToGoUint8)
	case *uint16:
		return valueTryToGoValueHelper[uint16](val, p, valueTryToGoUint16)
	case *uint32:
		return valueTryToGoValueHelper[uint32](val, p, valueTryToGoUint32)
	case *uint64:
		return valueTryToGoValueHelper[uint64](val, p, valueTryToGoUint64)
	case *uint:
		return valueTryToGoValueHelper[uint](val, p, valueTryToGoUint)
	case *float32:
		return valueTryToGoValueHelper[float32](val, p, valueTryToGoFloat32)
	case *float64:
		return valueTryToGoValueHelper[float64](val, p, valueTryToGoFloat64)
	case *string:
		return valueTryToGoValueHelper[string](val, p, valueTryToGoString)
	case *Kwargs:
		return valueTryToGoValueHelper[Kwargs](val, p, valueTryToKwargs)
	}
	panic("unsupported destination type")
}

func valueTryToGoValueHelper[T any](val Value, dest *T, f func(Value) (T, error)) error {
	v, err := f(val)
	if err != nil {
		return err
	}
	*dest = v
	return nil
}

func valueSliceTryToGoSliceReflect(values []Value, destType reflect.Type) (any, error) {
	slice := reflect.MakeSlice(destType, 0, len(values))
	elemType := destType.Elem()
	for _, val := range values {
		elem, err := valueTryToGoValueReflect(val, elemType)
		if err != nil {
			return nil, err
		}
		slice = reflect.Append(slice, reflect.ValueOf(elem))
	}
	return slice.Interface(), nil
}

func valueTryToValue(val Value) (Value, error) { return val, nil }

func valueTryToGoString(val Value) (string, error) {
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
