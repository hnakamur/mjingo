package mjingo

import (
	"fmt"
)

func opsGetOffsetAndLen(start int64, stop option[int64], end func() uint) (uint, uint) {
	var startIdx uint
	var stopIdx uint
	if start < 0 || (!stop.valid || stop.data < 0) {
		endIdx := end()
		if start < 0 {
			startIdx = uint(int64(endIdx) + start)
		} else {
			startIdx = uint(start)
		}
		if stop.valid {
			if stop.data < 0 {
				stopIdx = uint(int64(endIdx) + stop.data)
			} else {
				stopIdx = uint(stop.data)
			}
		} else {
			stopIdx = endIdx
		}
	} else {
		startIdx = uint(start)
		stopIdx = uint(stop.data)
	}
	if stopIdx > startIdx {
		stopIdx -= startIdx
	} else {
		stopIdx = 0
	}
	return startIdx, stopIdx
}

func opsSlice(v, start, stop, step value) (value, error) {
	startVal := int64(0)
	if !start.isNone() {
		if start.typ != valueTypeI64 {
			panic("opsSlice start must be an i64")
		}
		startVal = start.data.(i64ValueData)
	}
	stopVal := option[int64]{}
	if !stop.isNone() {
		if stop.typ != valueTypeI64 {
			panic("opsSlice stop must be an i64")
		}
		stopVal = option[int64]{valid: true, data: stop.data.(i64ValueData)}
	}
	stepVal := int64(1)
	if !step.isNone() {
		if step.typ != valueTypeI64 {
			panic("opsSlice step must be an i64")
		}
		stepVal = step.data.(i64ValueData)
		if stepVal < 0 {
			return value{}, &Error{
				kind: InvalidOperation,
				detail: option[string]{
					valid: true,
					data:  "cannot slice by negative step size",
				},
			}
		}
		if stepVal == 0 {
			return value{}, &Error{
				kind: InvalidOperation,
				detail: option[string]{
					valid: true,
					data:  "cannot slice by step size of 0",
				},
			}
		}
	}

	var maybeSeq seqObject
	switch v.typ {
	case valueTypeString:
		data := v.data.(stringValueData)
		chars := []rune(data)
		startIdx, stopIdx := opsGetOffsetAndLen(startVal, stopVal, func() uint { return uint(len(chars)) })
		sliced := make([]rune, 0, len(chars))
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			sliced = append(sliced, chars[i])
		}
		return value{
			typ:  valueTypeString,
			data: string(sliced),
		}, nil
	case valueTypeUndefined, valueTypeNone:
		return value{
			typ:  valueTypeSeq,
			data: []value{},
		}, nil
	case valueTypeSeq:
		data := v.data.(seqValueData)
		maybeSeq = newSliceSeqObject(data)
	case valueTypeDynamic:
		panic("not implemented")
	}

	if maybeSeq != nil {
		startIdx, stopIdx := opsGetOffsetAndLen(startVal, stopVal, func() uint { return maybeSeq.itemCount() })
		sliced := make([]value, 0, maybeSeq.itemCount())
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			if item := maybeSeq.getItem(i); item.valid {
				sliced = append(sliced, item.data)
			}
		}
		return value{
			typ:  valueTypeSeq,
			data: sliced,
		}, nil
	}
	return value{}, &Error{
		kind: InvalidOperation,
		detail: option[string]{
			valid: true,
			data:  fmt.Sprintf("value of type %s cannot be sliced", v.typ),
		},
	}
}

func opsNeg(v value) (value, error) {
	if v.kind() != valueKindNumber {
		return value{}, &Error{kind: InvalidOperation}
	}
	if v.typ == valueTypeF64 {
		data := v.data.(f64ValueData)
		return value{typ: valueTypeF64, data: -data}, nil
	}

	if v.typ == valueTypeI128 || v.typ == valueTypeU128 {
		panic("not implemented")
	}

	x, err := v.tryToI64()
	if err != nil {
		return value{}, err
	}
	return value{typ: valueTypeI64, data: -x}, nil
}
