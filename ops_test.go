package mjingo

import (
	"math/big"
	"testing"
)

func TestI128WrappingAdd(t *testing.T) {
	testCases := []struct {
		lhs, rhs, want string
	}{
		{lhs: "2", rhs: "3", want: "5"},
		{lhs: "170141183460469231731687303715884105727", rhs: "1", want: "-170141183460469231731687303715884105728"},
		{lhs: "170141183460469231731687303715884105727", rhs: "2", want: "-170141183460469231731687303715884105727"},
		{lhs: "-170141183460469231731687303715884105728", rhs: "-1", want: "170141183460469231731687303715884105727"},
		{lhs: "-170141183460469231731687303715884105728", rhs: "-2", want: "170141183460469231731687303715884105726"},
	}
	for _, tc := range testCases {
		var lhs, rhs, rv I128
		setBigIntString(t, &lhs.n, tc.lhs)
		setBigIntString(t, &rhs.n, tc.rhs)
		i128WrappingAdd(&rv, &lhs, &rhs)
		if got, want := rv.String(), tc.want; got != want {
			t.Errorf("result mismatch, got=%s, want=%s, lhs=%s, rhs=%s", got, want, tc.lhs, tc.rhs)
		}
	}
}

func TestCaseU128AsI128(t *testing.T) {
	testCases := []struct {
		input, want string
	}{
		{input: "0", want: "0"},
		{input: "340282366920938463463374607431768211455", want: "-1"},
		{input: "170141183460469231731687303715884105727", want: "170141183460469231731687303715884105727"},
		{input: "170141183460469231731687303715884105728", want: "-170141183460469231731687303715884105728"},
		{input: "170141183460469231731687303715884105729", want: "-170141183460469231731687303715884105727"},
	}
	for _, tc := range testCases {
		var u U128
		var rv I128
		setBigIntString(t, &u.n, tc.input)
		castU128AsI128(&rv, &u)
		if got, want := rv.String(), tc.want; got != want {
			t.Errorf("result mismatch, got=%s, want=%s, input=%s", got, want, tc.input)
		}
	}
}

func setBigIntString(t *testing.T, n *big.Int, s string) {
	if _, ok := n.SetString(s, 10); !ok {
		t.Fatalf("invalid BigInt string: %s", s)
	}
}
