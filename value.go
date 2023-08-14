package mjingo

import (
	"fmt"
	"math"
)

type valueType int

const (
	valueTypeUndefined valueType = iota + 1
	valueTypeBool
	valueTypeU64
	valueTypeI64
	valueTypeF64
	valueTypeNone
	valueTypeInvalid
	valueTypeU128
	valueTypeI128
	valueTypeString
	valueTypeBytes
	valueTypeSeq
	valueTypeMap
	valueTypeDynamic
)

type valueKind int

const (
	// The value is undefined
	valueKindUndefined valueKind = iota + 1
	// The value is the none singleton ([`()`])
	valueKindNone
	// The value is a [`bool`]
	valueKindBool
	// The value is a number of a supported type.
	valueKindNumber
	// The value is a string.
	valueKindString
	// The value is a byte array.
	valueKindBytes
	// The value is an array of other values.
	valueKindSeq
	// The value is a key/value mapping.
	valueKindMap
)

var valueUndefined = value{typ: valueTypeUndefined}
var valueNone = value{typ: valueTypeNone}

func (t valueType) String() string {
	switch t {
	case valueTypeUndefined:
		return "undefined"
	case valueTypeBool:
		return "bool"
	case valueTypeU64:
		return "u64"
	case valueTypeI64:
		return "i64"
	case valueTypeF64:
		return "f64"
	case valueTypeNone:
		return "none"
	case valueTypeInvalid:
		return "invalid"
	case valueTypeU128:
		return "u128"
	case valueTypeI128:
		return "i128"
	case valueTypeString:
		return "string"
	case valueTypeBytes:
		return "bytes"
	case valueTypeSeq:
		return "seq"
	case valueTypeMap:
		return "map"
	case valueTypeDynamic:
		return "dynamic"
	default:
		panic(fmt.Sprintf("invalid valueType: %d", t))
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
	typ  valueType
	data any
}

func (v *value) isUndefined() bool {
	return v.typ == valueTypeUndefined
}

func (v *value) isNone() bool {
	return v.typ == valueTypeNone
}

func (v *value) getAttrFast(key string) option[value] {
	switch v.typ {
	case valueTypeMap:
		items := v.data.(mapValueData)
		if v, ok := items[key]; ok {
			return option[value]{valid: true, data: v}
		}
	default:
		panic(fmt.Sprintf("not implemented for valueType: %s", v.typ))
	}
	return option[value]{}
}

func (v *value) getItemOpt(key value) option[value] {
	keyRf := keyRef{kind: keyRefKindValue, data: key}
	var seq seqObject
	switch v.typ {
	case valueTypeMap:
		items := v.data.(mapValueData)

		// implementation here is different from minijinja.
		if keyData := keyRf.asStr(); keyData.valid {
			if v, ok := items[keyData.data]; ok {
				return option[value]{valid: true, data: v}
			}
		} else {
			panic(fmt.Sprintf("value.getItemOpt does not support non string key: %+v", key))
		}
	case valueTypeSeq:
		items := v.data.(seqValueData)
		seq = newSliceSeqObject(items)
	default:
		panic(fmt.Sprintf("not implemented for valueType: %s", v.typ))
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
	switch v.typ {
	case valueTypeString:
		data := v.data.(stringValueData)
		return option[string]{valid: true, data: data}
	default:
		return option[string]{}
	}
}

func (v value) tryToI64() (int64, error) {
	switch v.typ {
	case valueTypeBool:
		data := v.data.(boolValueData)
		if data {
			return 1, nil
		} else {
			return 0, nil
		}
	case valueTypeI64:
		data := v.data.(i64ValueData)
		return data, nil
	case valueTypeU64:
		data := v.data.(u64ValueData)
		return int64(data), nil
	case valueTypeF64:
		data := v.data.(f64ValueData)
		if float64(int64(data)) == data {
			return int64(data), nil
		}
	case valueTypeI128:
		panic("not implemented")
	case valueTypeU128:
		panic("not implemented")
	}
	return 0, unsupportedConversion(v.typ, "i64")
}

func unsupportedConversion(kind valueType, target string) error {
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

func (v value) kind() valueKind {
	switch v.typ {
	case valueTypeUndefined:
		return valueKindUndefined
	case valueTypeBool:
		return valueKindBool
	case valueTypeU64, valueTypeI64, valueTypeF64, valueTypeU128, valueTypeI128:
		return valueKindNumber
	case valueTypeNone:
		return valueKindNone
	case valueTypeInvalid:
		// XXX: invalid values report themselves as maps which is a lie
		return valueKindMap
	case valueTypeString:
		return valueKindString
	case valueTypeBytes:
		return valueKindBytes
	case valueTypeSeq:
		return valueKindSeq
	case valueTypeMap:
		return valueKindMap
	case valueTypeDynamic:
		panic("not implemented for valueTypeDynamic")
	default:
		panic(fmt.Sprintf("invalid valueType: %d", v.typ))
	}
}
