package internal

import "github.com/hnakamur/mjingo/internal/datast/option"

type Object interface {
	Kind() ObjectKind
}

type CallMethoder interface {
	CallMethod(state *State, name string, args []Value) (Value, error)
}

type Caller interface {
	Call(state *State, args []Value) (Value, error)
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

func newSliceSeqObject(values []Value) SeqObject {
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