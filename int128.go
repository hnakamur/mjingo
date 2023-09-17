package mjingo

import (
	"fmt"
	"math/big"
	"sync"
)

// I128 represents an integer in the range between
// -170141183460469231731687303715884105728 and 170141183460469231731687303715884105727
// (both ends inclusive).
// The zero value for an I128 represents the value 0.
//
// Operations always take pointer arguments (*I128) rather
// than I128 values, and each unique Int value requires
// its own unique *I128 pointer. To "copy" an I128 value,
// an existing (or newly allocated) I128 must be set to
// a new value using the [I128.Set] method; shallow copies
// of I128s are not supported and may lead to errors.
//
// Note that methods may leak the I128's value through timing side-channels.
// Because of this and because of the scope and complexity of the
// implementation, I128 is not well-suited to implement cryptographic operations.
type I128 struct{ n big.Int }

// I128FromInt64 allocates and returns a new I128 set to x.
func I128FromInt64(x int64) *I128 {
	return &I128{n: *big.NewInt(x)}
}

// I128FromUint64 allocates and returns a new I128 set to x.
func I128FromUint64(x uint64) *I128 {
	var rv I128
	rv.n.SetUint64(x)
	return &rv
}

// I128TryFromBigInt allocates and returns a new I128 set to x.
// If x is out of range of I128, it returns an error.
func I128TryFromBigInt(x *big.Int) (*I128, error) {
	if !isI128(x) {
		return nil, NewError(InvalidOperation, "cannot convert to I128")
	}
	var rv I128
	rv.n.Set(x)
	return &rv, nil
}

// CheckedAbs sets z to |x| (the absolute value of x) and returns z.
// If the operation overflows, the value of z is undefined but the returned value is nil.
func (z *I128) CheckedAbs(x *I128) *I128 {
	z.n.Abs(&x.n)
	return z.checkedVal()
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (x *I128) Cmp(y *I128) int { return x.n.Cmp(&y.n) }

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

// CheckedAdd sets z to the sum x+y and returns z if the result is in the range of I128.
// If the operation overflows, the value of z is undefined but the returned value is nil.
func (z *I128) CheckedAdd(x, y *I128) *I128 {
	z.n.Add(&x.n, &y.n)
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

// Set sets z to x and returns z.
func (z *I128) Set(x *I128) *I128 {
	z.n.Set(&x.n)
	return z
}

// IsInt64 reports whether x can be represented as an int64.
func (x *I128) IsInt64() bool { return x.n.IsInt64() }

// IsUint64 reports whether x can be represented as a uint64.
func (x *I128) IsUint64() bool { return x.n.IsUint64() }

// Int64 returns the int64 representation of x.
// If x cannot be represented in an int64, the result is undefined.
func (x *I128) Int64() int64 { return x.n.Int64() }

// Uint64 returns the uint64 representation of x.
// If x cannot be represented in a uint64, the result is undefined.
func (x *I128) Uint64() uint64 { return x.n.Uint64() }

// SetInt64 sets z to x and returns z.
func (z *I128) SetInt64(x int64) *I128 {
	z.n.SetInt64(x)
	return z
}

// SetUint64 sets z to x and returns z.
func (z *I128) SetUint64(x uint64) *I128 {
	z.n.SetUint64(x)
	return z
}

func (z *I128) mustSetString(s string, base int) *I128 {
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

// String returns the decimal representation of x in base 10.
func (x *I128) String() string { return x.n.String() }

// BigInt returns a new big.Int whose value is copied from x.
func (x *I128) BigInt() big.Int {
	var rv big.Int
	rv.Set(&x.n)
	return rv
}

// Format implements fmt.Formatter.
func (x I128) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte(x.String())) }

// U128 represents an integer in the range between
// 0 and 340282366920938463463374607431768211455
// (both ends inclusive).
// The zero value for an U128 represents the value 0.
//
// Operations always take pointer arguments (*U128) rather
// than U128 values, and each unique Int value requires
// its own unique *U128 pointer. To "copy" an U128 value,
// an existing (or newly allocated) U128 must be set to
// a new value using the [U128.Set] method; shallow copies
// of U128s are not supported and may lead to errors.
//
// Note that methods may leak the U128's value through timing side-channels.
// Because of this and because of the scope and complexity of the
// implementation, U128 is not well-suited to implement cryptographic operations.
type U128 struct{ n big.Int }

// U128FromUint64 allocates and returns a new U128 set to x.
func U128FromUint64(x uint64) *U128 {
	var rv U128
	rv.n.SetUint64(x)
	return &rv
}

// U128TryFromInt64 allocates and returns a new U128 set to x.
// If x is out of range of U128, it returns an error.
func U128TryFromInt64(x int64) (*U128, error) {
	if x < 0 {
		return nil, NewError(InvalidOperation, "cannot convert to U128")
	}
	var rv U128
	rv.n.SetInt64(x)
	return &rv, nil
}

// U128TryFromBigInt allocates and returns a new U128 set to x.
// If x is out of range of U128, it returns an error.
func U128TryFromBigInt(x *big.Int) (*U128, error) {
	if !isU128(x) {
		return nil, NewError(InvalidOperation, "cannot convert to U128")
	}
	var rv U128
	rv.n.Set(x)
	return &rv, nil
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (x *U128) Cmp(y *U128) int { return x.n.Cmp(&y.n) }

// Set sets z to x and returns z.
func (z *U128) Set(x *U128) *U128 {
	z.n.Set(&x.n)
	return z
}

// IsInt64 reports whether x can be represented as an int64.
func (x *U128) IsInt64() bool { return x.n.IsInt64() }

// IsUint64 reports whether x can be represented as a uint64.
func (x *U128) IsUint64() bool { return x.n.IsUint64() }

// Int64 returns the int64 representation of x.
// If x cannot be represented in an int64, the result is undefined.
func (x *U128) Int64() int64 { return x.n.Int64() }

// Uint64 returns the uint64 representation of x.
// If x cannot be represented in a uint64, the result is undefined.
func (x *U128) Uint64() uint64 { return x.n.Uint64() }

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
// If the input is out of range of U128, SetString fails.
func (z *U128) SetString(s string, base int) (*U128, bool) {
	r, ok := z.n.SetString(s, base)
	if !ok || !isU128(r) {
		return nil, false
	}
	return z, true
}

// String returns the decimal representation of x in base 10.
func (x *U128) String() string { return x.n.String() }

// Format implements fmt.Formatter.
func (x U128) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte(x.String())) }

// BigInt returns a new big.Int whose value is copied from x.
func (x *U128) BigInt() big.Int {
	var rv big.Int
	rv.Set(&x.n)
	return rv
}

var i128MinAbs = mustNewBigIntFromString("170141183460469231731687303715884105728", 10)
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
