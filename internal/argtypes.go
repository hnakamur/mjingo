package internal

import "github.com/hnakamur/mjingo/internal/datast/option"

// type argsFromValuesFn[T any] func(state *virtualMachineState, args []value) (T, error)

type Unit struct{}

func ValueFromBytes(val []byte) Value {
	return bytesValue{b: val}
}

func ValueFromString(val string) Value {
	return stringValue{str: val, strTyp: stringTypeNormal}
}

// Creates a value from a safe string.
//
// A safe string is one that will bypass auto escaping.  For instance if you
// want to have the template engine render some HTML without the user having to
// supply the `|safe` filter, you can use a value of this type instead.
func ValueFromSafeString(s string) Value {
	return stringValue{str: s, strTyp: stringTypeSafe}
}

func ValueFromUnit(_ Unit) Value {
	return None
}

func ValueFromBool(val bool) Value {
	return BoolValue{B: val}
}

func ValueFromI64(n int64) Value {
	return i64Value{n: n}
}

func ValueFromF64(f float64) Value {
	return f64Value{f: f}
}

func ValueFromSlice(values []Value) Value {
	return SeqValue{items: values}
}

func ValueFromIndexMap(m *IndexMap) Value {
	return mapValue{m: m, mapTyp: mapTypeNormal}
}

func ValueFromKwargs(a Kwargs) Value {
	return mapValue{m: &a.Values, mapTyp: mapTypeKwargs}
}

func ValueFromObject(dy Object) Value {
	return dynamicValue{dy: dy}
}

func StringFromValue(value option.Option[Value]) (string, error) {
	if option.IsSome(value) {
		optStr := option.Unwrap(value).AsStr()
		if option.IsSome(optStr) {
			return option.Unwrap(optStr), nil
		}
		return "", NewError(InvalidOperation, "value is not a string")
	}
	return "", NewError(MissingArgument, "")
}

func StringTryFromValue(val Value) (string, error) {
	if v, ok := val.(stringValue); ok {
		return v.str, nil
	}
	return "", NewError(InvalidOperation, "value is not a string")
}

type rest[T any] struct {
	args []T
}

type Kwargs struct {
	Values IndexMap
	Used   map[string]struct{}
}
