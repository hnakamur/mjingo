package mjingo

type seqObject interface {
	getItem(idx uint) option[value]
	itemCount() uint
}

func newSliceSeqObject(values []value) seqObject {
	return &sliceSeqObject{values: values}
}

type sliceSeqObject struct {
	values []value
}

func (s *sliceSeqObject) getItem(idx uint) option[value] {
	if idx >= uint(len(s.values)) {
		return option[value]{}
	}
	return option[value]{valid: true, data: s.values[idx]}
}

func (s *sliceSeqObject) itemCount() uint {
	return uint(len(s.values))
}
