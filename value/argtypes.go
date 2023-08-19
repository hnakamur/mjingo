package value

// type argsFromValuesFn[T any] func(state *virtualMachineState, args []value) (T, error)

type Unit struct{}

func FromBytes(val []byte) Value {
	return bytesValue{b: val}
}

func FromString(val string) Value {
	return stringValue{str: val, strTyp: stringTypeNormal}
}

// Creates a value from a safe string.
//
// A safe string is one that will bypass auto escaping.  For instance if you
// want to have the template engine render some HTML without the user having to
// supply the `|safe` filter, you can use a value of this type instead.
func FromSafeString(s string) Value {
	return stringValue{str: s, strTyp: stringTypeSafe}
}

func FromUnit(_ Unit) Value {
	return None
}

func FromBool(val bool) Value {
	return BoolValue{B: val}
}

func FromI64(n int64) Value {
	return i64Value{n: n}
}

func FromF64(f float64) Value {
	return f64Value{f: f}
}

func FromSlice(values []Value) Value {
	return SeqValue{items: values}
}

func FromValueIndexMap(m *valueIndexMap) Value {
	return mapValue{m: m, mapTyp: mapTypeNormal}
}

type rest[T any] struct {
	args []T
}

type kwargs struct {
	values valueIndexMap
	used   map[string]struct{}
}
