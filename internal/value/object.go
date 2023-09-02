package value

import "github.com/hnakamur/mjingo/internal/datast/option"

type Object interface {
	Kind() ObjectKind
}

type ObjectKind uint

const (
	ObjectKindPlain ObjectKind = iota + 1
	ObjectKindSeq
	ObjectKindStruct
)

type SeqObject interface {
	GetItem(idx uint) option.Option[Value]
	ItemCount() uint
}

func NewSliceSeqObject(values []Value) SeqObject {
	return &sliceSeqObject{values: values}
}

type sliceSeqObject struct {
	values []Value
}

func (s *sliceSeqObject) Kind() ObjectKind { return ObjectKindSeq }

func (s *sliceSeqObject) GetItem(idx uint) option.Option[Value] {
	if idx >= uint(len(s.values)) {
		return option.None[Value]()
	}
	return option.Some(s.values[idx])
}

func (s *sliceSeqObject) ItemCount() uint {
	return uint(len(s.values))
}

type StructObject interface {
	GetField(name string) option.Option[Value]
	StaticFields() option.Option[[]string]
	Fields() []string
}

func FieldCount(s StructObject) uint {
	optFields := s.StaticFields()
	if optFields.IsSome() {
		return uint(len(optFields.Unwrap()))
	}
	return uint(len(s.Fields()))
}

func StaticOrDynamicFields(s StructObject) []string {
	optFields := s.StaticFields()
	if optFields.IsSome() {
		return optFields.Unwrap()
	}
	return s.Fields()
}
