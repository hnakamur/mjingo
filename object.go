package mjingo

import "github.com/hnakamur/mjingo/option"

type object interface {
	Kind() objectKind
}

type objectKind uint

const (
	objectKindPlain objectKind = iota + 1
	objectKindSeq
	objectKindStruct
)

type seqObject interface {
	GetItem(idx uint) option.Option[Value]
	ItemCount() uint
}

func newSliceSeqObject(values []Value) seqObject {
	return &sliceSeqObject{values: values}
}

type sliceSeqObject struct {
	values []Value
}

func (s *sliceSeqObject) Kind() objectKind { return objectKindSeq }

func (s *sliceSeqObject) GetItem(idx uint) option.Option[Value] {
	if idx >= uint(len(s.values)) {
		return option.None[Value]()
	}
	return option.Some(s.values[idx])
}

func (s *sliceSeqObject) ItemCount() uint {
	return uint(len(s.values))
}

type structObject interface {
	GetField(name string) option.Option[Value]
	StaticFields() option.Option[[]string]
	Fields() []string
}

func fieldCount(s structObject) uint {
	optFields := s.StaticFields()
	if optFields.IsSome() {
		return uint(len(optFields.Unwrap()))
	}
	return uint(len(s.Fields()))
}

func staticOrDynamicFields(s structObject) []string {
	optFields := s.StaticFields()
	if optFields.IsSome() {
		return optFields.Unwrap()
	}
	return s.Fields()
}
