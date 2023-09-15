package mjingo

import (
	"fmt"
	"math/big"
	"sync"
)

// I128 represents an integer in the range between
// -170141183460469231731687303715884105728 and 170141183460469231731687303715884105727
// (both ends inclusive).
type I128 struct{ n big.Int }

func I128FromInt64(n int64) *I128 {
	return &I128{n: *big.NewInt(n)}
}

func I128FromUint64(n uint64) *I128 {
	var rv I128
	rv.n.SetUint64(n)
	return &rv
}

func I128TryFromBigInt(n *big.Int) (*I128, error) {
	if !isI128(n) {
		return nil, NewError(InvalidOperation, "cannot convert to I128")
	}
	var rv I128
	rv.n.Set(n)
	return &rv, nil
}

// CheckedAbs sets z to |x| (the absolute value of x) and returns z.
// If the operation overflows, the value of z is undefined but the returned value is nil.
func (z *I128) CheckedAbs(x *I128) *I128 {
	z.n.Abs(&x.n)
	return z.checkedVal()
}

func (i *I128) Cmp(x *I128) int { return i.n.Cmp(&x.n) }

// CheckedDiv sets sets z to the quotient x/y and returns z if y != 0 and the result is in the range of I128.
// If the operation overflows, the value of z is undefined but the returned value is nil.
// Div implements Euclidean division (unlike Go); see [math/big.Int.DivMod] for more details.
func (z *I128) CheckedDiv(x, y *I128) *I128 {
	var zero I128
	if y.Cmp(&zero) == 0 {
		return nil
	}
	z.n.Div(&x.n, &y.n)
	return z.checkedVal()
}

// CheckedMod sets sets z to the modulus x%y and returns z if y != 0 and the result is in the range of I128.
// If the operation overflows, the value of z is undefined but the returned value is nil.
// Mod implements Euclidean modulus (unlike Go); see [math/big.Int.DivMod] for more details.
func (z *I128) CheckedMod(x, y *I128) *I128 {
	var zero I128
	if y.Cmp(&zero) == 0 {
		return nil
	}
	z.n.Mod(&x.n, &y.n)
	return z.checkedVal()
}

// CheckedMul sets z to the product x*y and returns z if the result is in the range of I128.
// If the operation overflows, the value of z is undefined but the returned value is nil.
func (z *I128) CheckedMul(x, y *I128) *I128 {
	z.n.Mul(&x.n, &y.n)
	return z.checkedVal()
}

// CheckedSub sets z to the difference x-y and returns z if the result is in the range of I128.
// If the operation overflows, the value of z is undefined but the returned value is nil.
func (z *I128) CheckedSub(x, y *I128) *I128 {
	z.n.Sub(&x.n, &y.n)
	return z.checkedVal()
}

// CheckedNeg sets z to -x and returns z if the result is in the range of I128.
// If the operation overflows, the value of z is undefined but the returned value is nil.
func (z *I128) CheckedNeg(x *I128) *I128 {
	z.n.Neg(&x.n)
	return z.checkedVal()
}

func (z *I128) checkedVal() *I128 {
	if isI128(&z.n) {
		return z
	}
	return nil
}

func (i *I128) Set(x *I128)        { i.n.Set(&x.n) }
func (i *I128) IsInt64() bool      { return i.n.IsInt64() }
func (i *I128) Int64() int64       { return i.n.Int64() }
func (i *I128) IsUint64() bool     { return i.n.IsUint64() }
func (i *I128) Uint64() uint64     { return i.n.Uint64() }
func (i *I128) SetInt64(x int64)   { i.n.SetInt64(x) }
func (i *I128) SetUint64(x uint64) { i.n.SetUint64(x) }

// MustSetString sets z to the value of s, interpreted in the given base,
// and returns z or panic on failure. The entire string
// (not just a prefix) must be valid for success. If MustSetString fails,
// it panics.
//
// The base argument must be 0 or a value between 2 and MaxBase.
// For base 0, the number prefix determines the actual base: A prefix of
// “0b” or “0B” selects base 2, “0”, “0o” or “0O” selects base 8,
// and “0x” or “0X” selects base 16. Otherwise, the selected base is 10
// and no prefix is accepted.
//
// For bases <= 36, lower and upper case letters are considered the same:
// The letters 'a' to 'z' and 'A' to 'Z' represent digit values 10 to 35.
// For bases > 36, the upper case letters 'A' to 'Z' represent the digit
// values 36 to 61.
//
// For base 0, an underscore character “_” may appear between a base
// prefix and an adjacent digit, and between successive digits; such
// underscores do not change the value of the number.
// Incorrect placement of underscores is reported as an error if there
// are no other errors. If base != 0, underscores are not recognized
// and act like any other character that is not a valid digit.
//
// If the input is out of range of I128, MustSetString fails.
func (z *I128) MustSetString(s string, base int) *I128 {
	_, ok := z.SetString(s, base)
	if !ok {
		panic("overflow in I128.MustSetString")
	}
	return z
}

// SetString sets z to the value of s, interpreted in the given base,
// and returns z and a boolean indicating success. The entire string
// (not just a prefix) must be valid for success. If SetString fails,
// the value of z is undefined but the returned value is nil.
//
// The base argument must be 0 or a value between 2 and MaxBase.
// For base 0, the number prefix determines the actual base: A prefix of
// “0b” or “0B” selects base 2, “0”, “0o” or “0O” selects base 8,
// and “0x” or “0X” selects base 16. Otherwise, the selected base is 10
// and no prefix is accepted.
//
// For bases <= 36, lower and upper case letters are considered the same:
// The letters 'a' to 'z' and 'A' to 'Z' represent digit values 10 to 35.
// For bases > 36, the upper case letters 'A' to 'Z' represent the digit
// values 36 to 61.
//
// For base 0, an underscore character “_” may appear between a base
// prefix and an adjacent digit, and between successive digits; such
// underscores do not change the value of the number.
// Incorrect placement of underscores is reported as an error if there
// are no other errors. If base != 0, underscores are not recognized
// and act like any other character that is not a valid digit.
//
// If the input is out of range of I128, SetString fails.
func (z *I128) SetString(s string, base int) (*I128, bool) {
	r, ok := z.n.SetString(s, base)
	if !ok || !isI128(r) {
		return nil, false
	}
	return z, true
}

func (i *I128) String() string { return i.n.String() }
func (i *I128) BigInt() big.Int {
	var rv big.Int
	rv.Set(&i.n)
	return rv
}

type U128 struct{ n big.Int }

func U128FromUint64(n uint64) *U128 {
	var rv U128
	rv.n.SetUint64(n)
	return &rv
}

func U128TryFromInt64(n int64) (*U128, error) {
	if n < 0 {
		return nil, NewError(InvalidOperation, "cannot convert to U128")
	}
	var rv U128
	rv.n.SetInt64(n)
	return &rv, nil
}

func U128TryFromBigInt(n *big.Int) (*U128, error) {
	if !isU128(n) {
		return nil, NewError(InvalidOperation, "cannot convert to U128")
	}
	var rv U128
	rv.n.Set(n)
	return &rv, nil
}

func (u *U128) Cmp(x *U128) int { return u.n.Cmp(&x.n) }
func (u *U128) Set(x *U128)     { u.n.Set(&x.n) }
func (u *U128) IsInt64() bool   { return u.n.IsInt64() }
func (u *U128) Int64() int64    { return u.n.Int64() }
func (u *U128) IsUint64() bool  { return u.n.IsUint64() }
func (u *U128) Uint64() uint64  { return u.n.Uint64() }
func (u *U128) String() string  { return u.n.String() }
func (u *U128) BigInt() big.Int {
	var rv big.Int
	rv.Set(&u.n)
	return rv
}

var i128Min = mustNewBigIntFromString("-170141183460469231731687303715884105728", 10)
var i128Max = mustNewBigIntFromString("170141183460469231731687303715884105727", 10)

func mustNewBigIntFromString(s string, base int) *big.Int {
	n, ok := new(big.Int).SetString(s, base)
	if !ok {
		panic(fmt.Sprintf("failed to set big.Int by string %s and base %d", s, base))
	}
	return n
}

var u128Max, twoPow128 *big.Int

func getU128Max() *big.Int {
	return sync.OnceValue(func() *big.Int {
		u128Max = mustNewBigIntFromString("340282366920938463463374607431768211455", 10)
		return u128Max
	})()
}

func getTwoPow128() *big.Int {
	return sync.OnceValue(func() *big.Int {
		twoPow128 = mustNewBigIntFromString("340282366920938463463374607431768211456", 10)
		return twoPow128
	})()
}

func isI128(n *big.Int) bool {
	return n.Cmp(i128Min) >= 0 && n.Cmp(i128Max) <= 0
}

func isU128(n *big.Int) bool {
	var zero big.Int
	return n.Cmp(&zero) >= 0 && n.Cmp(getU128Max()) <= 0
}
