package mjingo

import "fmt"

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
