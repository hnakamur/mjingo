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

type value struct {
	kind valueKind
	data any
}
