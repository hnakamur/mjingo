package internal

import (
	"fmt"
	"math"
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

func opsGetOffsetAndLen(start int64, stop option.Option[int64], end func() uint) (uint, uint) {
	var startIdx uint
	var stopIdx uint
	if start < 0 || (stop.IsNone() || stop.Unwrap() < 0) {
		endIdx := end()
		if start < 0 {
			startIdx = uint(int64(endIdx) + start)
		} else {
			startIdx = uint(start)
		}
		if stop.IsSome() {
			if stop.Unwrap() < 0 {
				stopIdx = uint(int64(endIdx) + stop.Unwrap())
			} else {
				stopIdx = uint(stop.Unwrap())
			}
		} else {
			stopIdx = endIdx
		}
	} else {
		startIdx = uint(start)
		stopIdx = uint(stop.Unwrap())
	}
	if stopIdx > startIdx {
		stopIdx -= startIdx
	} else {
		stopIdx = 0
	}
	return startIdx, stopIdx
}

func Slice(val, start, stop, step Value) (Value, error) {
	startVal := int64(0)
	if !start.IsNone() {
		if s, ok := start.(i64Value); ok {
			startVal = s.n
		} else {
			panic("opsSlice start must be an i64")
		}
	}
	stopVal := option.None[int64]()
	if !stop.IsNone() {
		if s, ok := stop.(i64Value); ok {
			stopVal = option.Some(s.n)
		} else {
			panic("opsSlice stop must be an i64")
		}
	}
	stepVal := int64(1)
	if !step.IsNone() {
		if s, ok := step.(i64Value); ok {
			stepVal = s.n
			if stepVal < 0 {
				return nil, NewError(InvalidOperation,
					"cannot slice by negative step size")
			}
			if stepVal == 0 {
				return nil, NewError(InvalidOperation,
					"cannot slice by step size of 0")
			}
		} else {
			panic("opsSlice step must be an i64")
		}
	}

	var maybeSeq SeqObject
	switch v := val.(type) {
	case stringValue:
		chars := []rune(v.str)
		startIdx, stopIdx := opsGetOffsetAndLen(startVal, stopVal, func() uint { return uint(len(chars)) })
		sliced := make([]rune, 0, len(chars))
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			sliced = append(sliced, chars[i])
		}
		return stringValue{str: string(sliced)}, nil
	case undefinedValue, noneValue:
		return SeqValue{items: []Value{}}, nil
	case SeqValue:
		maybeSeq = newSliceSeqObject(v.items)
	case dynamicValue:
		panic("not implemented")
	}

	if maybeSeq != nil {
		startIdx, stopIdx := opsGetOffsetAndLen(startVal, stopVal, func() uint { return maybeSeq.ItemCount() })
		sliced := make([]Value, 0, maybeSeq.ItemCount())
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			if item := maybeSeq.GetItem(i); item.IsSome() {
				sliced = append(sliced, item.Unwrap())
			}
		}
		return SeqValue{items: sliced}, nil
	}
	return nil, NewError(InvalidOperation,
		fmt.Sprintf("value of type %s cannot be sliced", val.typ()))
}

func Neg(val Value) (Value, error) {
	if val.Kind() != ValueKindNumber {
		return nil, NewError(InvalidOperation, "")
	}
	if v, ok := val.(f64Value); ok {
		return f64Value{f: -v.f}, nil
	}

	if val.typ() == valueTypeI128 || val.typ() == valueTypeU128 {
		panic("not implemented")
	}

	if x, err := val.TryToI64(); err != nil {
		return nil, err
	} else {
		return i64Value{n: -x}, nil
	}
}

func Add(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i64CoerceResult:
		return i64Value{n: c.lhs + c.rhs}, nil
	case f64CoerceResult:
		return f64Value{f: c.lhs + c.rhs}, nil
	case strCoerceResult:
		return stringValue{str: c.lhs + c.rhs}, nil
	}
	return nil, impossibleOp("+", lhs, rhs)
}

func Sub(lhs, rhs Value) (Value, error) {
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

func Mul(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i64CoerceResult:
		// TODO: checked_mul
		return i64Value{n: c.lhs * c.rhs}, nil
	case f64CoerceResult:
		return f64Value{f: c.lhs * c.rhs}, nil
	}
	return nil, impossibleOp("*", lhs, rhs)
}

func Div(lhs, rhs Value) (Value, error) {
	optA := lhs.AsF64()
	optB := rhs.AsF64()
	if optA.IsSome() && optB.IsSome() {
		d := optA.Unwrap() / optB.Unwrap()
		return ValueFromF64(d), nil
	}
	return nil, impossibleOp("/", lhs, rhs)
}

func IntDiv(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i64CoerceResult:
		if c.rhs == 0 {
			return nil, failedOp("//", lhs, rhs)
		}
		// TODO: div_euclid
		return i64Value{n: c.lhs / c.rhs}, nil
	case f64CoerceResult:
		// TODO: div_euclid
		return f64Value{f: math.Floor(c.lhs / c.rhs)}, nil
	}
	return nil, impossibleOp("//", lhs, rhs)
}

func Rem(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i64CoerceResult:
		if c.rhs == 0 {
			return nil, failedOp("%", lhs, rhs)
		}
		// TODO: checked_rem_euclid
		return i64Value{n: c.lhs % c.rhs}, nil
	case f64CoerceResult:
		// TODO: checked_rem_euclid
		return f64Value{f: math.Remainder(c.lhs, c.rhs)}, nil
	}
	return nil, impossibleOp("%", lhs, rhs)
}

func Pow(lhs, rhs Value) (Value, error) {
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

func StringConcat(left, right Value) Value {
	return stringValue{str: fmt.Sprintf("%s%s", left, right)}
}

// / Implements a containment operation on values.
func Contains(container Value, val Value) (Value, error) {
	// Special case where if the container is undefined, it cannot hold
	// values.  For strict containment checks the vm has a special case.
	if container.IsUndefined() {
		return ValueFromBool(false), nil
	}
	var rv bool
	if optContainerStr := container.AsStr(); optContainerStr.IsSome() {
		containerStr := optContainerStr.Unwrap()
		var valStr string
		if optValStr := val.AsStr(); optValStr.IsSome() {
			valStr = optValStr.Unwrap()
		} else {
			valStr = val.String()
		}
		rv = strings.Contains(containerStr, valStr)
	} else if optSeq := container.AsSeq(); optSeq.IsSome() {
		seq := optSeq.Unwrap()
		n := seq.ItemCount()
		for i := uint(0); i < n; i++ {
			elem := seq.GetItem(i).Unwrap()
			if elem == val {
				rv = true
				break
			}
		}
	} else if mapVal, ok := container.(mapValue); ok {
		_, ok := mapVal.m.Get(KeyRefFromValue(val.Clone()))
		rv = ok
	} else {
		return nil, NewError(InvalidOperation,
			"cannot perform a containment check on this value")
	}
	return ValueFromBool(rv), nil
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

func coerce(a, b Value) coerceResult {
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
		aVal := a.(stringValue).str
		bVal := b.(stringValue).str
		return strCoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeF64 || b.typ() == valueTypeF64:
		var aVal, bVal float64
		if af, ok := a.(f64Value); ok {
			aVal = af.f
			if bMayVal := b.AsF64(); bMayVal.IsSome() {
				bVal = bMayVal.Unwrap()
			} else {
				return nil
			}
		} else if bf, ok := b.(f64Value); ok {
			bVal = bf.f
			if aMayVal := a.AsF64(); aMayVal.IsSome() {
				aVal = aMayVal.Unwrap()
			} else {
				return nil
			}
		}
		return f64CoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeI128 || a.typ() == valueTypeU128 || b.typ() == valueTypeI128 || b.typ() == valueTypeU128:
		panic("not implemented")
	default:
		// everything else goes up to i64 (different from i128 in MiniJinja)
		aVal, err := a.TryToI64()
		if err != nil {
			return nil
		}
		bVal, err := b.TryToI64()
		if err != nil {
			return nil
		}
		return i64CoerceResult{lhs: aVal, rhs: bVal}
	}
}

func failedOp(op string, lhs, rhs Value) error {
	return NewError(InvalidOperation,
		fmt.Sprintf("unable to calculate %s %s %s", lhs, op, rhs))
}

func impossibleOp(op string, lhs, rhs Value) error {
	return NewError(InvalidOperation,
		fmt.Sprintf("tried to use %s operator on unsupported types %s and %s", op, lhs, rhs))
}
