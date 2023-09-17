package mjingo

import (
	"fmt"
	"math"
	"strings"

	"github.com/hnakamur/mjingo/option"
)

func opGetOffsetAndLen(start int64, stop option.Option[int64], end func() uint) (uint, uint) {
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

func opSlice(val, start, stop, step Value) (Value, error) {
	startVal := int64(0)
	if !start.isNone() {
		if s, ok := start.data.(i64Value); ok {
			startVal = s.N
		} else {
			panic("opsSlice start must be an i64")
		}
	}
	stopVal := option.None[int64]()
	if !stop.isNone() {
		if v, err := stop.tryToI64(); err == nil {
			stopVal = option.Some(v)
		} else {
			return Value{}, NewError(InvalidOperation,
				"cannot convert slice stop index to i64")
		}
	}
	stepVal := int64(1)
	if !step.isNone() {
		if s, ok := step.data.(i64Value); ok {
			stepVal = s.N
			if stepVal < 0 {
				return Value{}, NewError(InvalidOperation,
					"cannot slice by negative step size")
			}
			if stepVal == 0 {
				return Value{}, NewError(InvalidOperation,
					"cannot slice by step size of 0")
			}
		} else {
			panic("opsSlice step must be an i64")
		}
	}

	var maybeSeq SeqObject
	switch v := val.data.(type) {
	case stringValue:
		chars := []rune(v.Str)
		startIdx, stopIdx := opGetOffsetAndLen(startVal, stopVal, func() uint { return uint(len(chars)) })
		sliced := make([]rune, 0, len(chars))
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			sliced = append(sliced, chars[i])
		}
		return valueFromString(string(sliced)), nil
	case undefinedValue, noneValue:
		return valueFromSlice([]Value{}), nil
	case seqValue:
		maybeSeq = newSliceSeqObject(v.Items)
	case dynamicValue:
		if obj, ok := v.Dy.(SeqObject); ok {
			maybeSeq = obj
		}
	}

	if maybeSeq != nil {
		startIdx, stopIdx := opGetOffsetAndLen(startVal, stopVal, func() uint { return maybeSeq.ItemCount() })
		sliced := make([]Value, 0, maybeSeq.ItemCount())
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			if item := maybeSeq.GetItem(i); item.IsSome() {
				sliced = append(sliced, item.Unwrap())
			}
		}
		return valueFromSlice(sliced), nil
	}
	return Value{}, NewError(InvalidOperation,
		fmt.Sprintf("value of type %s cannot be sliced", val.typ()))
}

func opNeg(val Value) (Value, error) {
	if val.Kind() != ValueKindNumber {
		return Value{}, NewError(InvalidOperation, "")
	}
	if v, ok := val.data.(f64Value); ok {
		return valueFromF64(-v.F), nil
	}

	// special case for the largest i128 that can still be
	// represented.
	if v, ok := val.data.(u128Value); ok && v.N.n.Cmp(i128MinAbs) == 0 {
		return val, nil
	}

	x, err := val.tryToI128()
	if err != nil {
		return Value{}, NewError(InvalidOperation, "")
	}
	if x.CheckedNeg(x) == nil {
		return Value{}, failedOpUnary("-", val)
	}
	return i128AsValue(x), nil
}

func opAdd(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var n I128
		if n.CheckedAdd(&c.lhs, &c.rhs) == nil {
			return Value{}, failedOp("+", lhs, rhs)
		}
		return i128AsValue(&n), nil
	case f64CoerceResult:
		return valueFromF64(c.lhs + c.rhs), nil
	case strCoerceResult:
		return valueFromString(c.lhs + c.rhs), nil
	}
	return Value{}, impossibleOp("+", lhs, rhs)
}

func opSub(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var n I128
		if n.CheckedSub(&c.lhs, &c.rhs) == nil {
			return Value{}, failedOp("-", lhs, rhs)
		}
		return i128AsValue(&n), nil
	case f64CoerceResult:
		return valueFromF64(c.lhs - c.rhs), nil
	}
	return Value{}, impossibleOp("-", lhs, rhs)
}

func opMul(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var n I128
		if n.CheckedMul(&c.lhs, &c.rhs) == nil {
			return Value{}, failedOp("*", lhs, rhs)
		}
		return i128AsValue(&n), nil
	case f64CoerceResult:
		return valueFromF64(c.lhs * c.rhs), nil
	}
	return Value{}, impossibleOp("*", lhs, rhs)
}

func opDiv(lhs, rhs Value) (Value, error) {
	optA := lhs.asF64()
	optB := rhs.asF64()
	if optA.IsSome() && optB.IsSome() {
		d := optA.Unwrap() / optB.Unwrap()
		return valueFromF64(d), nil
	}
	return Value{}, impossibleOp("/", lhs, rhs)
}

func opIntDiv(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var div I128
		if div.CheckedDiv(&c.lhs, &c.rhs) == nil {
			return Value{}, failedOp("//", lhs, rhs)
		}
		return i128AsValue(&div), nil
	case f64CoerceResult:
		return valueFromF64(math.Floor(c.lhs / c.rhs)), nil
	}
	return Value{}, impossibleOp("//", lhs, rhs)
}

func opRem(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var mod I128
		if mod.CheckedMod(&c.lhs, &c.rhs) == nil {
			return Value{}, failedOp("%", lhs, rhs)
		}
		return i128AsValue(&mod), nil
	case f64CoerceResult:
		return valueFromF64(math.Remainder(c.lhs, c.rhs)), nil
	}
	return Value{}, impossibleOp("%", lhs, rhs)
}

func opPow(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var exp uint32
		if !c.rhs.IsUint64() || c.rhs.Uint64() > math.MaxUint32 {
			return Value{}, failedOp("**", lhs, rhs)
		}
		exp = uint32(c.rhs.Uint64())
		var n I128
		if i128CheckedPow(&n, &c.lhs, exp) == nil {
			return Value{}, failedOp("**", lhs, rhs)
		}
		return i128AsValue(&n), nil
	case f64CoerceResult:
		return valueFromF64(math.Pow(c.lhs, c.rhs)), nil
	}
	return Value{}, impossibleOp("**", lhs, rhs)
}

func opStringConcat(left, right Value) Value {
	return valueFromString(fmt.Sprintf("%s%s", left, right))
}

// / Implements a containment operation on values.
func opContains(container Value, val Value) (Value, error) {
	// Special case where if the container is undefined, it cannot hold
	// values.  For strict containment checks the vm has a special case.
	if container.isUndefined() {
		return valueFromBool(false), nil
	}
	var rv bool
	if containerStr := ""; valueAsOptionString(container).UnwrapTo(&containerStr) {
		var valStr string
		if !valueAsOptionString(val).UnwrapTo(&valStr) {
			valStr = val.String()
		}
		rv = strings.Contains(containerStr, valStr)
	} else if optSeq := container.asSeq(); optSeq.IsSome() {
		seq := optSeq.Unwrap()
		n := seq.ItemCount()
		for i := uint(0); i < n; i++ {
			elem := seq.GetItem(i).Unwrap()
			if elem == val {
				rv = true
				break
			}
		}
	} else if mapVal, ok := container.data.(mapValue); ok {
		_, ok := mapVal.Map.Get(keyRefFromValue(val.clone()))
		rv = ok
	} else {
		return Value{}, NewError(InvalidOperation,
			"cannot perform a containment check on this value")
	}
	return valueFromBool(rv), nil
}

type coerceResult interface {
	typ() coerceResultType
}

type i128CoerceResult struct {
	lhs I128
	rhs I128
}

type f64CoerceResult struct {
	lhs float64
	rhs float64
}

type strCoerceResult struct {
	lhs string
	rhs string
}

func (i128CoerceResult) typ() coerceResultType { return coerceResultTypeI128 }
func (f64CoerceResult) typ() coerceResultType  { return coerceResultTypeF64 }
func (strCoerceResult) typ() coerceResultType  { return coerceResultTypeStr }

type coerceResultType int

const (
	// I64 here (for now) instead of I128 in MiniJinja
	coerceResultTypeI128 coerceResultType = iota + 1
	coerceResultTypeF64
	coerceResultTypeStr
)

func coerce(a, b Value) coerceResult { return coerceData(a.data, b.data) }

func coerceData(a, b valueData) coerceResult {
	switch {
	case a.typ() == valueTypeU64 && b.typ() == valueTypeU64:
		aVal := a.(u64Value).N
		bVal := b.(u64Value).N
		var rv i128CoerceResult
		rv.lhs.SetUint64(aVal)
		rv.rhs.SetUint64(bVal)
		return rv
	case a.typ() == valueTypeU128 && b.typ() == valueTypeU128:
		aVal := a.(u128Value)
		bVal := b.(u128Value)
		var rv i128CoerceResult
		castU128AsI128(&rv.lhs, &aVal.N)
		castU128AsI128(&rv.rhs, &bVal.N)
		return rv
	case a.typ() == valueTypeString && b.typ() == valueTypeString:
		aVal := a.(stringValue).Str
		bVal := b.(stringValue).Str
		return strCoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeI64 && b.typ() == valueTypeI64:
		aVal := a.(i64Value).N
		bVal := b.(i64Value).N
		var rv i128CoerceResult
		rv.lhs.SetInt64(aVal)
		rv.rhs.SetInt64(bVal)
		return rv
	case a.typ() == valueTypeI128 && b.typ() == valueTypeI128:
		aVal := a.(i128Value).N
		bVal := b.(i128Value).N
		return i128CoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeF64 && b.typ() == valueTypeF64:
		aVal := a.(f64Value).F
		bVal := b.(f64Value).F
		return f64CoerceResult{lhs: aVal, rhs: bVal}
	case a.typ() == valueTypeF64 || b.typ() == valueTypeF64:
		var aVal, bVal float64
		if af, ok := a.(f64Value); ok {
			aVal = af.F
			if bMayVal := b.asF64(); bMayVal.IsSome() {
				bVal = bMayVal.Unwrap()
			} else {
				return nil
			}
		} else if bf, ok := b.(f64Value); ok {
			bVal = bf.F
			if aMayVal := a.asF64(); aMayVal.IsSome() {
				aVal = aMayVal.Unwrap()
			} else {
				return nil
			}
		}
		return f64CoerceResult{lhs: aVal, rhs: bVal}
	default:
		// everything else goes up to i64 (different from i128 in MiniJinja)
		aVal, err := a.tryToI128()
		if err != nil {
			return nil
		}
		bVal, err := b.tryToI128()
		if err != nil {
			return nil
		}
		return i128CoerceResult{lhs: *aVal, rhs: *bVal}
	}
}

func castU128AsI128(ret *I128, input *U128) *I128 {
	ret.n.Set(&input.n)
	if input.n.Cmp(i128Max) > 0 {
		ret.n.Sub(&ret.n, getTwoPow128())
	}
	return ret
}

func i128AsValue(val *I128) Value {
	if val.IsInt64() {
		return valueFromI64(val.Int64())
	}
	return valueFromI128(*val)
}

func i128CheckedMul(ret, lhs, rhs *I128) *I128 {
	ret.n.Mul(&lhs.n, &rhs.n)
	if isI128(&ret.n) {
		return ret
	}
	return nil
}

func i128CheckedPow(ret, base *I128, exp uint32) *I128 {
	// ported from https://github.com/rust-lang/rust/blob/1.72.0/library/core/src/num/int_macros.rs#L875-L899
	ret.n.SetUint64(1)
	if exp == 0 {
		return ret
	}
	base2, _ := I128TryFromBigInt(&base.n)
	for exp > 1 {
		if exp&1 == 1 {
			ret = i128CheckedMul(ret, ret, base2)
			if ret == nil {
				return nil
			}
		}
		exp /= 2
		base2 = i128CheckedMul(base2, base2, base2)
		if base2 == nil {
			return nil
		}
	}
	return i128CheckedMul(ret, ret, base2)
}

func failedOpUnary(op string, v Value) error {
	return NewError(InvalidOperation,
		fmt.Sprintf("unable to calculate %s%s", op, v))
}

func failedOp(op string, lhs, rhs Value) error {
	return NewError(InvalidOperation,
		fmt.Sprintf("unable to calculate %s %s %s", lhs, op, rhs))
}

func impossibleOp(op string, lhs, rhs Value) error {
	return NewError(InvalidOperation,
		fmt.Sprintf("tried to use %s operator on unsupported types %s and %s",
			op, lhs.Kind(), rhs.Kind()))
}
