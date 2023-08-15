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

func opsSlice(val, start, stop, step value) (value, error) {
	startVal := int64(0)
	if !start.isNone() {
		if s, ok := start.(i64Value); ok {
			startVal = s.n
		} else {
			panic("opsSlice start must be an i64")
		}
	}
	stopVal := option[int64]{}
	if !stop.isNone() {
		if s, ok := stop.(i64Value); ok {
			stopVal = option[int64]{valid: true, data: s.n}
		} else {
			panic("opsSlice stop must be an i64")
		}
	}
	stepVal := int64(1)
	if !step.isNone() {
		if s, ok := step.(i64Value); ok {
			stepVal = s.n
			if stepVal < 0 {
				return nil, &Error{
					typ: InvalidOperation,
					detail: option[string]{
						valid: true,
						data:  "cannot slice by negative step size",
					},
				}
			}
			if stepVal == 0 {
				return nil, &Error{
					typ: InvalidOperation,
					detail: option[string]{
						valid: true,
						data:  "cannot slice by step size of 0",
					},
				}
			}
		} else {
			panic("opsSlice step must be an i64")
		}
	}

	var maybeSeq seqObject
	switch v := val.(type) {
	case stringValue:
		chars := []rune(v.s)
		startIdx, stopIdx := opsGetOffsetAndLen(startVal, stopVal, func() uint { return uint(len(chars)) })
		sliced := make([]rune, 0, len(chars))
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			sliced = append(sliced, chars[i])
		}
		return stringValue{s: string(sliced)}, nil
	case undefinedValue, noneValue:
		return seqValue{items: []value{}}, nil
	case seqValue:
		maybeSeq = newSliceSeqObject(v.items)
	case dynamicValue:
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
		return seqValue{items: sliced}, nil
	}
	return nil, &Error{
		typ: InvalidOperation,
		detail: option[string]{
			valid: true,
			data:  fmt.Sprintf("value of type %s cannot be sliced", val.typ()),
		},
	}
}

func opsNeg(val value) (value, error) {
	if val.kind() != valueKindNumber {
		return nil, &Error{typ: InvalidOperation}
	}
	if v, ok := val.(f64Value); ok {
		return f64Value{f: -v.f}, nil
	}

	if val.typ() == valueTypeI128 || val.typ() == valueTypeU128 {
		panic("not implemented")
	}

	if x, err := val.tryToI64(); err != nil {
		return nil, err
	} else {
		return i64Value{n: -x}, nil
	}
}

func opsAdd(lhs, rhs value) (value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i64CoerceResult:
		return i64Value{n: c.lhs + c.rhs}, nil
	case f64CoerceResult:
		return f64Value{f: c.lhs + c.rhs}, nil
	case strCoerceResult:
		return stringValue{s: c.lhs + c.rhs}, nil
	}
	return nil, impossibleOp("+", lhs, rhs)
}

func opsSub(lhs, rhs value) (value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i64CoerceResult:
		if c.lhs < c.rhs {
			return nil, failedOp("-", lhs, rhs)
		}
		return i64Value{n: c.lhs - c.rhs}, nil
	case f64CoerceResult:
		return f64Value{f: c.lhs - c.rhs}, nil
	}
	return nil, impossibleOp("-", lhs, rhs)
}

func opsPow(lhs, rhs value) (value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i64CoerceResult:
		if c.rhs < 0 {
			return nil, failedOp("**", lhs, rhs)
		}
		// TODO: checked_pow
		acc := int64(1)
		for i := int64(0); i < c.rhs; i++ {
			acc *= c.lhs
		}
		return i64Value{n: acc}, nil
	case f64CoerceResult:
		return f64Value{f: math.Pow(c.lhs, c.rhs)}, nil
	}
	return nil, impossibleOp("**", lhs, rhs)
}

func opsStringConcat(left, right value) value {
	return stringValue{s: fmt.Sprintf("%s%s", left, right)}
}

type coerceResult interface {
	typ() coerceResultType
}

type i64CoerceResult struct {
	lhs int64
	rhs int64
}

type f64CoerceResult struct {
	lhs float64
	rhs float64
}

type strCoerceResult struct {
	lhs string
	rhs string
}

func (i64CoerceResult) typ() coerceResultType { return coerceResultTypeI64 }
func (f64CoerceResult) typ() coerceResultType { return coerceResultTypeF64 }
func (strCoerceResult) typ() coerceResultType { return coerceResultTypeStr }

/*
i64CoerceResult
f64CoerceResult
strCoerceResult
*/

type coerceResultType int

const (
	// I64 here (for now) instead of I128 in MiniJinja
	coerceResultTypeI64 coerceResultType = iota + 1
	coerceResultTypeF64
	coerceResultTypeStr
)

func coerce(a, b value) coerceResult {
	switch {
	case a.typ() == valueTypeU64 && b.typ() == valueTypeU64:
		aVal := a.(u64Value).n
		bVal := b.(u64Value).n
		if aVal > math.MaxInt64 || bVal > math.MaxInt64 {
			return nil
		}
		return i64CoerceResult{lhs: int64(aVal), rhs: int64(bVal)}
	case a.typ() == valueTypeI64 && b.typ() == valueTypeI64:
		aVal := a.(i64Value).n
		bVal := b.(i64Value).n
		return i64CoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeString && b.typ() == valueTypeString:
		aVal := a.(stringValue).s
		bVal := b.(stringValue).s
		return strCoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeF64 || b.typ() == valueTypeF64:
		var aVal, bVal float64
		if af, ok := a.(f64Value); ok {
			aVal = af.f
			if bMayVal := b.asF64(); bMayVal.valid {
				bVal = bMayVal.data
			} else {
				return nil
			}
		} else if bf, ok := b.(f64Value); ok {
			bVal = bf.f
			if aMayVal := a.asF64(); aMayVal.valid {
				aVal = aMayVal.data
			} else {
				return nil
			}
		}
		return f64CoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeI128 || a.typ() == valueTypeU128 || b.typ() == valueTypeI128 || b.typ() == valueTypeU128:
		panic("not implemented")
	default:
		// everything else goes up to i64 (different from i128 in MiniJinja)
		aVal, err := a.tryToI64()
		if err != nil {
			return nil
		}
		bVal, err := b.tryToI64()
		if err != nil {
			return nil
		}
		return i64CoerceResult{lhs: aVal, rhs: bVal}
	}
}

func failedOp(op string, lhs, rhs value) error {
	return &Error{
		typ: InvalidOperation,
		detail: option[string]{
			valid: true,
			data: fmt.Sprintf("unable to calculate %s %s %s",
				lhs, op, rhs),
		},
	}
}

func impossibleOp(op string, lhs, rhs value) error {
	return &Error{
		typ: InvalidOperation,
		detail: option[string]{
			valid: true,
			data: fmt.Sprintf("tried to use %s operator on unsupported types %s and %s",
				op, lhs, rhs),
		},
	}
}
