package mjingo

type keyRef interface {
	typ() keyRefType
	asStr() option[string]
	asI64() option[int64]
}

type valueKeyRef struct{ val value }
type strKeyRef struct{ str string }

func (valueKeyRef) typ() keyRefType { return keyRefTypeValue }
func (strKeyRef) typ() keyRefType   { return keyRefTypeStr }

func (k valueKeyRef) asStr() option[string] { return k.val.asStr() }
func (k strKeyRef) asStr() option[string]   { return option[string]{valid: true, data: k.str} }

func (k valueKeyRef) asI64() option[int64] {
	if i, err := k.val.tryToI64(); err != nil {
		return option[int64]{}
	} else {
		return option[int64]{valid: true, data: i}
	}
}
func (k strKeyRef) asI64() option[int64] { return option[int64]{} }

type keyRefType uint

const (
	keyRefTypeValue keyRefType = iota
	keyRefTypeStr
)
