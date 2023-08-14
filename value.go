package mjingo

import (
	"fmt"
	"math"
)

type valueKind int

const (
	valueKindUndefined valueKind = iota + 1
	valueKindBool
	valueKindU64
	valueKindI64
	valueKindF64
	valueKindNone
	valueKindInvalid
	valueKindU128
	valueKindI128
	valueKindString
	valueKindBytes
	valueKindSeq
	valueKindMap
	valueKindDynamic
)

var valueUndefined = value{kind: valueKindUndefined}
var valueNone = value{kind: valueKindNone}

func (k valueKind) String() string {
	switch k {
	case valueKindUndefined:
		return "undefined"
	case valueKindBool:
		return "bool"
	case valueKindU64:
		return "u64"
	case valueKindI64:
		return "i64"
	case valueKindF64:
		return "f64"
	case valueKindNone:
		return "none"
	case valueKindInvalid:
		return "invalid"
	case valueKindU128:
		return "u128"
	case valueKindI128:
		return "i128"
	case valueKindString:
		return "string"
	case valueKindBytes:
		return "bytes"
	case valueKindSeq:
		return "seq"
	case valueKindMap:
		return "map"
	case valueKindDynamic:
		return "dynamic"
	default:
		panic(fmt.Sprintf("invalid valueKind: %d", k))
	}
}

type boolValueData = bool
type u64ValueData = uint64
type i64ValueData = int64
type f64ValueData = float64
type invalidValueData = string

type u128ValueData = struct {
	hi uint64
	lo uint64
}

type i128ValueData = struct {
	hi int64
	lo uint64
}

type stringValueData = string
type bytesValueData = []byte
type seqValueData = []value
type mapValueData = map[string]value

type value struct {
	kind valueKind
	data any
}

func (v *value) isUndefined() bool {
	return v.kind == valueKindUndefined
}

func (v *value) isNone() bool {
	return v.kind == valueKindNone
}

func (v *value) getAttrFast(key string) option[value] {
	switch v.kind {
	case valueKindMap:
		items := v.data.(mapValueData)
		if v, ok := items[key]; ok {
			return option[value]{valid: true, data: v}
		}
	default:
		panic(fmt.Sprintf("not implemented for valueKind: %s", v.kind))
	}
	return option[value]{}
}

func (v *value) getItemOpt(key value) option[value] {
	keyRf := keyRef{kind: keyRefKindValue, data: key}
	var seq seqObject
	switch v.kind {
	case valueKindMap:
		items := v.data.(mapValueData)

		// implementation here is different from minijinja.
		if keyData := keyRf.asStr(); keyData.valid {
			if v, ok := items[keyData.data]; ok {
				return option[value]{valid: true, data: v}
			}
		} else {
			panic(fmt.Sprintf("value.getItemOpt does not support non string key: %+v", key))
		}
	case valueKindSeq:
		items := v.data.(seqValueData)
		seq = newSliceSeqObject(items)
	default:
		panic(fmt.Sprintf("not implemented for valueKind: %s", v.kind))
	}

	if idx := keyRf.asI64(); idx.valid {
		if idx.data < math.MinInt || math.MaxInt < idx.data {
			return option[value]{}
		}
		var i uint
		if idx.data < 0 {
			c := seq.itemCount()
			if uint(-idx.data) > c {
				return option[value]{}
			}
			i = c - uint(-idx.data)
		} else {
			i = uint(idx.data)
		}
		return seq.getItem(i)
	}
	return option[value]{}
}

func (v value) asStr() option[string] {
	switch v.kind {
	case valueKindString:
		data := v.data.(stringValueData)
		return option[string]{valid: true, data: data}
	default:
		return option[string]{}
	}
}

func (v value) tryToI64() (int64, error) {
	switch v.kind {
	case valueKindBool:
		data := v.data.(boolValueData)
		if data {
			return 1, nil
		} else {
			return 0, nil
		}
	case valueKindI64:
		data := v.data.(i64ValueData)
		return data, nil
	case valueKindU64:
		data := v.data.(u64ValueData)
		return int64(data), nil
	case valueKindF64:
		data := v.data.(f64ValueData)
		if float64(int64(data)) == data {
			return int64(data), nil
		}
	case valueKindI128:
		panic("not implemented")
	case valueKindU128:
		panic("not implemented")
	}
	return 0, unsupportedConversion(v.kind, "i64")
}

func unsupportedConversion(kind valueKind, target string) error {
	return &Error{
		kind: InvalidOperation,
		detail: option[string]{
			valid: true,
			data:  fmt.Sprintf("cannot convert %s to %s", kind, target),
		},
	}
}

func valueMapWithCapacity(capacity uint) map[string]value {
	return make(map[string]value, untrustedSizeHint(capacity))
}
