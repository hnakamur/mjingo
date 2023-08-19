package valu

import "github.com/hnakamur/mjingo/internal/datast/option"

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

func (s *sliceSeqObject) GetItem(idx uint) option.Option[Value] {
	if idx >= uint(len(s.values)) {
		return option.None[Value]()
	}
	return option.Some(s.values[idx])
}

func (s *sliceSeqObject) ItemCount() uint {
	return uint(len(s.values))
}
