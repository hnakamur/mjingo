package mjingo

import (
	"fmt"
	"math"
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

func opsAdd(lhs, rhs value) (value, error) {
	if cMayRes := coerce(lhs, rhs); cMayRes.valid {
		cRes := cMayRes.data
		switch cRes.typ {
		case coerceResultTypeI64:
			data := cRes.data.(i64CoerceResultData)
			return value{
				typ:  valueTypeI64,
				data: data.lhs + data.rhs,
			}, nil
		case coerceResultTypeF64:
			data := cRes.data.(f64CoerceResultData)
			return value{
				typ:  valueTypeF64,
				data: data.lhs + data.rhs,
			}, nil
		case coerceResultTypeStr:
			data := cRes.data.(strCoerceResultData)
			return value{
				typ:  valueTypeString,
				data: data.lhs + data.rhs,
			}, nil
		}
	}
	return value{}, impossibleOp("+", lhs, rhs)
}

func opsSub(lhs, rhs value) (value, error) {
	if cMayRes := coerce(lhs, rhs); cMayRes.valid {
		cRes := cMayRes.data
		switch cRes.typ {
		case coerceResultTypeI64:
			data := cRes.data.(i64CoerceResultData)
			if data.lhs < data.rhs {
				return value{}, failedOp("-", lhs, rhs)
			}
			return value{
				typ:  valueTypeI64,
				data: data.lhs - data.rhs,
			}, nil
		case coerceResultTypeF64:
			data := cRes.data.(f64CoerceResultData)
			return value{
				typ:  valueTypeF64,
				data: data.lhs - data.rhs,
			}, nil
		}
	}
	return value{}, impossibleOp("-", lhs, rhs)
}

type coerceResultType int

const (
	// I64 here (for now) instead of I128 in MiniJinja
	coerceResultTypeI64 coerceResultType = iota + 1
	coerceResultTypeF64
	coerceResultTypeStr
)

type coerceResult struct {
	typ  coerceResultType
	data any
}

type i64CoerceResultData struct {
	lhs int64
	rhs int64
}

type f64CoerceResultData struct {
	lhs float64
	rhs float64
}

type strCoerceResultData struct {
	lhs string
	rhs string
}

func coerce(a, b value) option[coerceResult] {
	switch {
	case a.typ == valueTypeU64 && b.typ == valueTypeU64:
		aVal := a.data.(u64ValueData)
		bVal := b.data.(u64ValueData)
		if aVal > math.MaxInt64 || bVal > math.MaxInt64 {
			return option[coerceResult]{}
		}
		return option[coerceResult]{
			valid: true,
			data: coerceResult{
				typ:  coerceResultTypeI64,
				data: i64CoerceResultData{lhs: int64(aVal), rhs: int64(bVal)},
			},
		}
	case a.typ == valueTypeI64 && b.typ == valueTypeI64:
		aVal := a.data.(i64ValueData)
		bVal := b.data.(i64ValueData)
		return option[coerceResult]{
			valid: true,
			data: coerceResult{
				typ:  coerceResultTypeI64,
				data: i64CoerceResultData{lhs: aVal, rhs: bVal},
			},
		}
	case a.typ == valueTypeString && b.typ == valueTypeString:
		aVal := a.data.(stringValueData)
		bVal := b.data.(stringValueData)
		return option[coerceResult]{
			valid: true,
			data: coerceResult{
				typ:  coerceResultTypeStr,
				data: strCoerceResultData{lhs: aVal, rhs: bVal},
			},
		}
	case a.typ == valueTypeF64 || b.typ == valueTypeF64:
		var aVal, bVal float64
		if a.typ == valueTypeF64 {
			aVal = a.data.(f64ValueData)
			if bMayVal := b.asF64(); bMayVal.valid {
				bVal = bMayVal.data
			} else {
				return option[coerceResult]{}
			}
		} else if b.typ == valueTypeF64 {
			bVal = b.data.(f64ValueData)
			if aMayVal := a.asF64(); aMayVal.valid {
				aVal = aMayVal.data
			} else {
				return option[coerceResult]{}
			}
		}
		return option[coerceResult]{
			valid: true,
			data: coerceResult{
				typ:  coerceResultTypeF64,
				data: f64CoerceResultData{lhs: aVal, rhs: bVal},
			},
		}
	case a.typ == valueTypeI128 || a.typ == valueTypeU128 || b.typ == valueTypeI128 || b.typ == valueTypeU128:
		panic("not implemented")
	default:
		// everything else goes up to i64 (different from i128 in MiniJinja)
		aVal, err := a.tryToI64()
		if err != nil {
			return option[coerceResult]{}
		}
		bVal, err := b.tryToI64()
		if err != nil {
			return option[coerceResult]{}
		}
		return option[coerceResult]{
			valid: true,
			data: coerceResult{
				typ:  coerceResultTypeI64,
				data: i64CoerceResultData{lhs: aVal, rhs: bVal},
			},
		}
	}
}

func failedOp(op string, lhs, rhs value) error {
	return &Error{
		kind: InvalidOperation,
		detail: option[string]{
			valid: true,
			data: fmt.Sprintf("unable to calculate %s %s %s",
				lhs, op, rhs),
		},
	}
}

func impossibleOp(op string, lhs, rhs value) error {
	return &Error{
		kind: InvalidOperation,
		detail: option[string]{
			valid: true,
			data: fmt.Sprintf("tried to use %s operator on unsupported types %s and %s",
				op, lhs, rhs),
		},
	}
}
