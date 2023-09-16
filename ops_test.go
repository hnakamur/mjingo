package mjingo

import (
	"math/big"
	"testing"
)

func TestOpAdd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testCases := []struct {
			lhs, rhs any
			want     Value
		}{
			{lhs: 1, rhs: 2, want: ValueFromGoValue(3)},
			{lhs: "foo", rhs: "bar", want: ValueFromGoValue("foobar")},
		}
		for _, tc := range testCases {
			lhs := ValueFromGoValue(tc.lhs)
			rhs := ValueFromGoValue(tc.rhs)
			got, err := opAdd(lhs, rhs)
			if err != nil {
				t.Errorf("should not get error but got error, got=%s, lhs=%v, rhs=%v", err, tc.lhs, tc.rhs)
			} else {
				if want := tc.want; got != want {
					t.Errorf("result mismatch, got=%v, want=%v, lhs=%v, rhs=%v", got, want, tc.lhs, tc.rhs)
				}
			}
		}
	})
	t.Run("error", func(t *testing.T) {
		testCases := []struct {
			lhs, rhs   any
			wantErrMsg string
		}{
			{lhs: "a", rhs: 42,
				wantErrMsg: "invalid operation: tried to use + operator on unsupported types string and number"},
			{lhs: new(I128).MustSetString("170141183460469231731687303715884105727", 10), rhs: 1,
				wantErrMsg: "invalid operation: unable to calculate 170141183460469231731687303715884105727 + 1"},
		}
		for _, tc := range testCases {
			lhs := ValueFromGoValue(tc.lhs)
			rhs := ValueFromGoValue(tc.rhs)
			_, err := opAdd(lhs, rhs)
			if err != nil {
				if got, want := err.Error(), tc.wantErrMsg; got != want {
					t.Errorf("error mismatch, got=%s, want=%s, lhs=%v, rhs=%v", got, want, tc.lhs, tc.rhs)
				}
			} else {
				if got, want := err.Error(), tc.wantErrMsg; got != want {
					t.Errorf("should get error but not, want=%s, lhs=%v, rhs=%v", want, tc.lhs, tc.rhs)
				}
			}
		}
	})
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
