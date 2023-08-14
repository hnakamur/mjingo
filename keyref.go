package mjingo

type keyRefKind uint

const (
	keyRefKindValue keyRefKind = iota
	keyRefKindStr
)

type keyRef struct {
	kind keyRefKind
	data any
}

type valueKeyRefData = value
type strKeyRefData = string

func (k *keyRef) asStr() option[string] {
	switch k.kind {
	case keyRefKindValue:
		data := k.data.(valueKeyRefData)
		return data.asStr()
	case keyRefKindStr:
		data := k.data.(strKeyRefData)
		return option[string]{valid: true, data: data}
	default:
		panic("invalid keyRef kind")
	}
}

func (k *keyRef) asI64() option[int64] {
	if k.kind == keyRefKindValue {
		data := k.data.(valueKeyRefData)
		if i, err := data.tryToI64(); err != nil {
			return option[int64]{}
		} else {
			return option[int64]{valid: true, data: i}
		}
	}
	return option[int64]{}
}
