package internal

import (
	"math"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

// intCheckedPow, intCheckedRemEuclid, uintCheckedPow, and uintCheckedRemEuclid
// are ported from Rust's standaard library.

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func intCheckedSub[T Int](lhs, rhs T) option.Option[T] {
	ret := lhs - rhs
	if (rhs > 0 && ret > lhs) || (rhs < 0 && ret < lhs) {
		return option.None[T]()
	}
	return option.Some(ret)
}

func intCheckedMul[T Int](lhs, rhs T) option.Option[T] {
	ret := lhs * rhs
	if lhs != 0 && ret/lhs != rhs {
		return option.None[T]()
	}
	return option.Some(ret)
}

func intCheckedPow[T Int](base T, exp uint32) option.Option[T] {
	if exp == 0 {
		return option.Some[T](1)
	}
	acc := T(1)
	for exp > 1 {
		if exp&1 == 1 {
			if tmp := intCheckedMul(acc, base); tmp.IsSome() {
				acc = tmp.Unwrap()
			} else {
				return option.None[T]()
			}
		}
		exp /= 2
		if tmp := intCheckedMul(base, base); tmp.IsSome() {
			base = tmp.Unwrap()
		} else {
			return option.None[T]()
		}
	}
	return intCheckedMul(acc, base)
}

func intCheckedRemEuclid[T Int](lhs, rhs T) option.Option[T] {
	if rhs == 0 || (lhs == intMinValue[T]() && rhs == -1) {
		return option.None[T]()
	}
	return option.Some(intRemEuclid(lhs, rhs))
}

func intRemEuclid[T Int](lhs, rhs T) T {
	r := lhs % rhs
	if r < 0 {
		return r + intWrappingAbs(rhs)
	}
	return r
}

func intWrappingAbs[T Int](n T) T {
	if n < 0 {
		return -n
	}
	return n
}

func intMinValue[T Int]() T {
	var r T
	switch ret := any(r).(type) {
	case int:
		ret = math.MinInt
		return T(ret)
	case int8:
		ret = math.MinInt8
		return T(ret)
	case int16:
		ret = math.MinInt16
		return T(ret)
	case int32:
		ret = math.MinInt32
		return T(ret)
	case int64:
		ret = math.MinInt64
		return T(ret)
	default:
		panic("unreachable")
	}
}

func intMaxValue[T Int]() T {
	var r T
	switch ret := any(r).(type) {
	case int:
		ret = math.MaxInt
		return T(ret)
	case int8:
		ret = math.MaxInt8
		return T(ret)
	case int16:
		ret = math.MaxInt16
		return T(ret)
	case int32:
		ret = math.MaxInt32
		return T(ret)
	case int64:
		ret = math.MaxInt64
		return T(ret)
	default:
		panic("unreachable")
	}
}

func uintCheckedSub[T Uint](lhs, rhs T) option.Option[T] {
	ret := lhs - rhs
	if ret > lhs {
		return option.None[T]()
	}
	return option.Some(ret)
}

func uintCheckedMul[T Uint](lhs, rhs T) option.Option[T] {
	ret := lhs * rhs
	if lhs != 0 && ret/lhs != rhs {
		return option.None[T]()
	}
	return option.Some(ret)
}

func uintCheckedPow[T Uint](base T, exp uint32) option.Option[T] {
	if exp == 0 {
		return option.Some[T](1)
	}
	acc := T(1)
	for exp > 1 {
		if exp&1 == 1 {
			if tmp := uintCheckedMul(acc, base); tmp.IsSome() {
				acc = tmp.Unwrap()
			} else {
				return option.None[T]()
			}
		}
		exp /= 2
		if tmp := uintCheckedMul(base, base); tmp.IsSome() {
			base = tmp.Unwrap()
		} else {
			return option.None[T]()
		}
	}
	return uintCheckedMul(acc, base)
}

func uintCheckedRemEuclid[T Uint](lhs, rhs T) option.Option[T] {
	if rhs == 0 {
		return option.None[T]()
	}
	return option.Some(uintRemEuclid(lhs, rhs))
}

func uintRemEuclid[T Uint](lhs, rhs T) T {
	return lhs % rhs
}

func uintWrappingAbs[T Uint](n T) T {
	return n
}

func uintMinValue[T Uint]() T {
	return T(0)
}

func uintMaxValue[T Uint]() T {
	var r T
	switch ret := any(r).(type) {
	case uint:
		ret = math.MaxUint
		return T(ret)
	case uint8:
		ret = math.MaxUint8
		return T(ret)
	case uint16:
		ret = math.MaxUint16
		return T(ret)
	case uint32:
		ret = math.MaxUint32
		return T(ret)
	case uint64:
		ret = math.MaxUint64
		return T(ret)
	default:
		panic("unreachable")
	}
}
