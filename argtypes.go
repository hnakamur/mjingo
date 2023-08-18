package mjingo

type argsFromValuesFn[T any] func(state *virtualMachineState, args []value) (T, error)

func valueFromBytes(val []byte) value {
	return bytesValue{b: val}
}

func valueFromString(val string) value {
	return stringValue{str: val}
}

func valueFromUnit(_ unit) value {
	return valueNone
}

func valueFromBool(val bool) value {
	return boolValue{b: val}
}

func valueFromValueSlice(values []value) value {
	return seqValue{items: values}
}

func valueFromValueIndexMap(m *valueIndexMap) value {
	return mapValue{m: m, mapTyp: mapTypeNormal}
}

type rest[T any] struct {
	args []T
}

type kwargs struct {
	values valueIndexMap
	used   map[string]struct{}
}
