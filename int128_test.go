package mjingo

import (
	"math/big"
	"testing"
)

func TestU128TryFromBigInt(t *testing.T) {
	t.Run("u128Max", func(t *testing.T) {
		input := mustNewBigIntFromString("340282366920938463463374607431768211455", 10)
		u, err := U128TryFromBigInt(input)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := u.BigInt(), input; got.Cmp(want) != 0 {
			t.Errorf("result mismatch, got=%v, want=%v", got, want)
		}
	})
	t.Run("error", func(t *testing.T) {
		testCases := []*big.Int{
			mustNewBigIntFromString("340282366920938463463374607431768211456", 10),
			mustNewBigIntFromString("-1", 10),
		}
		for _, tc := range testCases {
			_, err := U128TryFromBigInt(tc)
			if err == nil {
				t.Errorf("should get an error but no error for input %v", tc)
			}
		}
	})
}
