package mjingo

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"

	"github.com/hnakamur/mjingo/internal/datast/option"
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
		if s, ok := start.(i64Value); ok {
			startVal = s.N
		} else {
			panic("opsSlice start must be an i64")
		}
	}
	stopVal := option.None[int64]()
	if !stop.isNone() {
		if s, ok := stop.(i64Value); ok {
			stopVal = option.Some(s.N)
		} else {
			panic("opsSlice stop must be an i64")
		}
	}
	stepVal := int64(1)
	if !step.isNone() {
		if s, ok := step.(i64Value); ok {
			stepVal = s.N
			if stepVal < 0 {
				return nil, newError(InvalidOperation,
					"cannot slice by negative step size")
			}
			if stepVal == 0 {
				return nil, newError(InvalidOperation,
					"cannot slice by step size of 0")
			}
		} else {
			panic("opsSlice step must be an i64")
		}
	}

	var maybeSeq seqObject
	switch v := val.(type) {
	case stringValue:
		chars := []rune(v.Str)
		startIdx, stopIdx := opGetOffsetAndLen(startVal, stopVal, func() uint { return uint(len(chars)) })
		sliced := make([]rune, 0, len(chars))
		for i := startIdx; i < stopIdx; i += uint(stepVal) {
			sliced = append(sliced, chars[i])
		}
		return stringValue{Str: string(sliced)}, nil
	case undefinedValue, noneValue:
		return seqValue{Items: []Value{}}, nil
	case seqValue:
		maybeSeq = newSliceSeqObject(v.Items)
	case dynamicValue:
		if obj, ok := v.Dy.(seqObject); ok {
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
		return seqValue{Items: sliced}, nil
	}
	return nil, newError(InvalidOperation,
		fmt.Sprintf("value of type %s cannot be sliced", val.typ()))
}

func opNeg(val Value) (Value, error) {
	if val.kind() != valueKindNumber {
		return nil, newError(InvalidOperation, "")
	}
	if v, ok := val.(f64Value); ok {
		return f64Value{F: -v.F}, nil
	}

	x, err := val.tryToI128()
	if err != nil {
		return nil, newError(InvalidOperation, "")
	}
	x.Neg(&x)
	return i128AsValue(&x), nil
}

func opAdd(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var n big.Int
		i128WrappingAdd(&n, &c.lhs, &c.rhs)
		return i128AsValue(&n), nil
	case f64CoerceResult:
		return f64Value{F: c.lhs + c.rhs}, nil
	case strCoerceResult:
		return stringValue{Str: c.lhs + c.rhs}, nil
	}
	return nil, impossibleOp("+", lhs, rhs)
}

func opSub(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var n big.Int
		n.Sub(&c.lhs, &c.rhs)
		if isI128(&n) {
			return i128AsValue(&n), nil
		}
		return nil, failedOp("-", lhs, rhs)
	case f64CoerceResult:
		return f64Value{F: c.lhs - c.rhs}, nil
	}
	return nil, impossibleOp("-", lhs, rhs)
}

func opMul(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var n big.Int
		n.Mul(&c.lhs, &c.rhs)
		if isI128(&n) {
			return i128AsValue(&n), nil
		}
		return nil, failedOp("*", lhs, rhs)
	case f64CoerceResult:
		return f64Value{F: c.lhs * c.rhs}, nil
	}
	return nil, impossibleOp("*", lhs, rhs)
}

func opDiv(lhs, rhs Value) (Value, error) {
	optA := lhs.asF64()
	optB := rhs.asF64()
	if optA.IsSome() && optB.IsSome() {
		d := optA.Unwrap() / optB.Unwrap()
		return valueFromF64(d), nil
	}
	return nil, impossibleOp("/", lhs, rhs)
}

func opIntDiv(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var zero big.Int
		if c.rhs.Cmp(&zero) == 0 {
			return nil, failedOp("//", lhs, rhs)
		}
		var div, mod big.Int
		div.DivMod(&c.lhs, &c.rhs, &mod)
		if isI128(&div) {
			return i128AsValue(&div), nil
		}
		return nil, failedOp("//", lhs, rhs)
	case f64CoerceResult:
		// TODO: div_euclid
		return f64Value{F: math.Floor(c.lhs / c.rhs)}, nil
	}
	return nil, impossibleOp("//", lhs, rhs)
}

func opRem(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var zero big.Int
		if c.rhs.Cmp(&zero) == 0 {
			return nil, failedOp("%", lhs, rhs)
		}
		var div, mod big.Int
		div.DivMod(&c.lhs, &c.rhs, &mod)
		if isI128(&mod) {
			return i128AsValue(&mod), nil
		}
		return nil, failedOp("%", lhs, rhs)
	case f64CoerceResult:
		return f64Value{F: math.Remainder(c.lhs, c.rhs)}, nil
	}
	return nil, impossibleOp("%", lhs, rhs)
}

func opPow(lhs, rhs Value) (Value, error) {
	switch c := coerce(lhs, rhs).(type) {
	case i128CoerceResult:
		var exp uint32
		if !c.rhs.IsUint64() || c.rhs.Uint64() > math.MaxUint32 {
			return nil, failedOp("**", lhs, rhs)
		}
		exp = uint32(c.rhs.Uint64())
		var n big.Int
		if i128CheckedPow(&n, &c.lhs, exp) == nil {
			return nil, failedOp("**", lhs, rhs)
		}
		return i128AsValue(&n), nil
	case f64CoerceResult:
		return f64Value{F: math.Pow(c.lhs, c.rhs)}, nil
	}
	return nil, impossibleOp("**", lhs, rhs)
}

func opStringConcat(left, right Value) Value {
	return stringValue{Str: fmt.Sprintf("%s%s", left, right)}
}

// / Implements a containment operation on values.
func opContains(container Value, val Value) (Value, error) {
	// Special case where if the container is undefined, it cannot hold
	// values.  For strict containment checks the vm has a special case.
	if container.isUndefined() {
		return valueFromBool(false), nil
	}
	var rv bool
	if optContainerStr := container.asStr(); optContainerStr.IsSome() {
		containerStr := optContainerStr.Unwrap()
		var valStr string
		if optValStr := val.asStr(); optValStr.IsSome() {
			valStr = optValStr.Unwrap()
		} else {
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
	} else if mapVal, ok := container.(mapValue); ok {
		_, ok := mapVal.Map.Get(keyRefFromValue(val.clone()))
		rv = ok
	} else {
		return nil, newError(InvalidOperation,
			"cannot perform a containment check on this value")
	}
	return valueFromBool(rv), nil
}

type coerceResult interface {
	typ() coerceResultType
}

type i128CoerceResult struct {
	lhs big.Int
	rhs big.Int
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

func coerce(a, b Value) coerceResult {
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
		return i128CoerceResult{lhs: aVal, rhs: bVal}
	}
}

var i128Min, i128Max, u128Max, twoPow128 big.Int

func getI128Min() *big.Int {
	return sync.OnceValue(func() *big.Int {
		if _, ok := i128Min.SetString("-170141183460469231731687303715884105728", 10); !ok {
			panic("set i128Min")
		}
		return &i128Min
	})()
}

func getI128Max() *big.Int {
	return sync.OnceValue(func() *big.Int {
		if _, ok := i128Max.SetString("170141183460469231731687303715884105727", 10); !ok {
			panic("set i128Max")
		}
		return &i128Max
	})()
}

func getU128Max() *big.Int {
	return sync.OnceValue(func() *big.Int {
		if _, ok := u128Max.SetString("340282366920938463463374607431768211455", 10); !ok {
			panic("set u128Max")
		}
		return &u128Max
	})()
}

func getTwoPow128() *big.Int {
	return sync.OnceValue(func() *big.Int {
		if _, ok := twoPow128.SetString("340282366920938463463374607431768211456", 10); !ok {
			panic("set twoPow128")
		}
		return &twoPow128
	})()
}

func isI128(n *big.Int) bool {
	return n.Cmp(getI128Min()) >= 0 && n.Cmp(getI128Max()) <= 0
}

func isU128(n *big.Int) bool {
	var zero big.Int
	return n.Cmp(&zero) >= 0 && n.Cmp(getI128Max()) <= 0
}

func i128WrappingAdd(ret, lhs, rhs *big.Int) *big.Int {
	ret.Add(lhs, rhs)
	if ret.Cmp(getI128Min()) < 0 {
		ret.Add(ret, getTwoPow128())
		return ret
	}
	if ret.Cmp(getI128Max()) > 0 {
		ret.Sub(ret, getTwoPow128())
		return ret
	}
	return ret
}

func castU128AsI128(ret, input *big.Int) *big.Int {
	ret.Set(input)
	if input.Cmp(getI128Max()) > 0 {
		ret.Sub(ret, getTwoPow128())
	}
	return ret
}

func i128AsValue(val *big.Int) Value {
	if val.IsInt64() {
		return i64Value{N: val.Int64()}
	}
	return i128Value{N: *val}
}

func i128CheckedMul(ret, lhs, rhs *big.Int) *big.Int {
	ret.Mul(lhs, rhs)
	if isI128(ret) {
		return ret
	}
	return nil
}

func i128CheckedPow(ret, base *big.Int, exp uint32) *big.Int {
	// ported from https://github.com/rust-lang/rust/blob/1.72.0/library/core/src/num/int_macros.rs#L875-L899
	ret.SetUint64(1)
	if exp == 0 {
		return ret
	}
	base2 := &big.Int{}
	base2.Set(base)
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

func failedOp(op string, lhs, rhs Value) error {
	return newError(InvalidOperation,
		fmt.Sprintf("unable to calculate %s %s %s", lhs, op, rhs))
}

func impossibleOp(op string, lhs, rhs Value) error {
	return newError(InvalidOperation,
		fmt.Sprintf("tried to use %s operator on unsupported types %s and %s", op, lhs, rhs))
}