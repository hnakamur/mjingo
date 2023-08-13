package mjingo

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

type value struct {
	kind valueKind
	data any
}
