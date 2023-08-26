package internal

import (
	"math"
	"testing"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

func TestIntCheckedSub(t *testing.T) {
	testCases := []struct {
		lhs, rhs int64
		want     option.Option[int64]
	}{
		{lhs: 5, rhs: 3, want: option.Some[int64](2)},
		{lhs: 5, rhs: 5, want: option.Some[int64](0)},
		{lhs: 5, rhs: 6, want: option.Some[int64](-1)},
		{lhs: 0, rhs: math.MaxInt64, want: option.Some[int64](math.MinInt64 + 1)},
		{lhs: -1, rhs: math.MaxInt64, want: option.Some[int64](math.MinInt64)},
		{lhs: -2, rhs: math.MaxInt64, want: option.None[int64]()},
		{lhs: -1, rhs: math.MinInt64, want: option.Some[int64](math.MaxInt64)},
		{lhs: 0, rhs: math.MinInt64, want: option.None[int64]()},
	}
	for _, tc := range testCases {
		if got, want := intCheckedSub(tc.lhs, tc.rhs), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, lhs=%d, rhs=%d", got, want, tc.lhs, tc.rhs)
		}
	}
}

func TestIntCheckedMul(t *testing.T) {
	testCases := []struct {
		lhs, rhs int64
		want     option.Option[int64]
	}{
		{lhs: 2, rhs: math.MaxInt64 / 2, want: option.Some[int64](math.MaxInt64 - 1)},
		{lhs: 2, rhs: math.MaxInt64/2 + 1, want: option.None[int64]()},
		{lhs: 2, rhs: math.MinInt64 / 2, want: option.Some[int64](math.MinInt64)},
		{lhs: 2, rhs: math.MinInt64/2 - 1, want: option.None[int64]()},
	}
	for _, tc := range testCases {
		if got, want := intCheckedMul(tc.lhs, tc.rhs), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, lhs=%d, rhs=%d", got, want, tc.lhs, tc.rhs)
		}
	}
}

func TestIntCheckedRemEuclid(t *testing.T) {
	testCases := []struct {
		lhs, rhs int64
		want     option.Option[int64]
	}{
		{lhs: 5, rhs: 3, want: option.Some[int64](2)},
		{lhs: 5, rhs: -3, want: option.Some[int64](2)},
		{lhs: -5, rhs: 3, want: option.Some[int64](1)},
		{lhs: -5, rhs: -3, want: option.Some[int64](1)},
		{lhs: 5, rhs: 0, want: option.None[int64]()},
		{lhs: math.MinInt64, rhs: -1, want: option.None[int64]()},
	}
	for _, tc := range testCases {
		if got, want := intCheckedRemEuclid(tc.lhs, tc.rhs), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, lhs=%d, rhs=%d", got, want, tc.lhs, tc.rhs)
		}
	}
}

func TestIntCheckedPow(t *testing.T) {
	testCases := []struct {
		base int64
		exp  uint32
		want option.Option[int64]
	}{
		{base: 2, exp: 62, want: option.Some[int64](4611686018427387904)},
		{base: 2, exp: 63, want: option.None[int64]()},
		{base: -2, exp: 62, want: option.Some[int64](4611686018427387904)},
		{base: -2, exp: 63, want: option.Some[int64](math.MinInt64)},
	}
	for _, tc := range testCases {
		if got, want := intCheckedPow(tc.base, tc.exp), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, base=%d, exp=%d", got, want, tc.base, tc.exp)
		}
	}
}

func TestIntMaxValue(t *testing.T) {
	if got, want := intMaxValue[int](), math.MaxInt; got != want {
		t.Errorf("intMaxValue[int]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := intMaxValue[int8](), int8(math.MaxInt8); got != want {
		t.Errorf("intMaxValue[int8]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := intMaxValue[int16](), int16(math.MaxInt16); got != want {
		t.Errorf("intMaxValue[int16]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := intMaxValue[int32](), int32(math.MaxInt32); got != want {
		t.Errorf("intMaxValue[int32]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := intMaxValue[int64](), int64(math.MaxInt64); got != want {
		t.Errorf("intMaxValue[int64]() mismatch, got=%d, want=%d", got, want)
	}
}

func TestUintCheckedSub(t *testing.T) {
	testCases := []struct {
		lhs, rhs uint64
		want     option.Option[uint64]
	}{
		{lhs: 5, rhs: 3, want: option.Some[uint64](2)},
		{lhs: 5, rhs: 5, want: option.Some[uint64](0)},
		{lhs: 5, rhs: 6, want: option.None[uint64]()},
		{lhs: math.MaxUint64, rhs: 0, want: option.Some[uint64](math.MaxUint64)},
	}
	for _, tc := range testCases {
		if got, want := uintCheckedSub(tc.lhs, tc.rhs), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, lhs=%d, rhs=%d", got, want, tc.lhs, tc.rhs)
		}
	}
}

func TestUintCheckedMul(t *testing.T) {
	testCases := []struct {
		lhs, rhs uint64
		want     option.Option[uint64]
	}{
		{lhs: 2, rhs: math.MaxUint64 / 2, want: option.Some[uint64](math.MaxUint64 - 1)},
		{lhs: 2, rhs: math.MaxUint64/2 + 1, want: option.None[uint64]()},
	}
	for _, tc := range testCases {
		if got, want := uintCheckedMul(tc.lhs, tc.rhs), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, lhs=%d, rhs=%d", got, want, tc.lhs, tc.rhs)
		}
	}
}

func TestUintCheckedRemEuclid(t *testing.T) {
	testCases := []struct {
		lhs, rhs uint64
		want     option.Option[uint64]
	}{
		{lhs: 5, rhs: 3, want: option.Some[uint64](2)},
		{lhs: 5, rhs: 0, want: option.None[uint64]()},
	}
	for _, tc := range testCases {
		if got, want := uintCheckedRemEuclid(tc.lhs, tc.rhs), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, lhs=%d, rhs=%d", got, want, tc.lhs, tc.rhs)
		}
	}
}

func TestUintCheckedPow(t *testing.T) {
	testCases := []struct {
		base uint64
		exp  uint32
		want option.Option[uint64]
	}{
		{base: 2, exp: 62, want: option.Some[uint64](4611686018427387904)},
		{base: 2, exp: 63, want: option.Some[uint64](9223372036854775808)},
		{base: 2, exp: 64, want: option.None[uint64]()},
	}
	for _, tc := range testCases {
		if got, want := uintCheckedPow(tc.base, tc.exp), tc.want; got != want {
			t.Errorf("result mismatch, got=%v, want=%v, base=%d, exp=%d", got, want, tc.base, tc.exp)
		}
	}
}
func TestUintMaxValue(t *testing.T) {
	if got, want := uintMaxValue[uint](), uint(math.MaxUint); got != want {
		t.Errorf("uintMaxValue[uint]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := uintMaxValue[uint8](), uint8(math.MaxUint8); got != want {
		t.Errorf("uintMaxValue[uint8]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := uintMaxValue[uint16](), uint16(math.MaxUint16); got != want {
		t.Errorf("uintMaxValue[uint16]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := uintMaxValue[uint32](), uint32(math.MaxUint32); got != want {
		t.Errorf("uintMaxValue[uint32]() mismatch, got=%d, want=%d", got, want)
	}
	if got, want := uintMaxValue[uint64](), uint64(math.MaxUint64); got != want {
		t.Errorf("uintMaxValue[uint64]() mismatch, got=%d, want=%d", got, want)
	}
}
