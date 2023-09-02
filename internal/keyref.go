package internal

import (
	"hash"
	"io"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type KeyRef interface {
	typ() keyRefType
	AsStr() option.Option[string]
	AsI64() option.Option[int64]
	AsValue() Value
	Hash(h hash.Hash)
	Equal(other any) bool
}

func KeyRefFromValue(val Value) KeyRef {
	return valueKeyRef{val: val}
}

func KeyRefFromString(val string) KeyRef {
	return strKeyRef{str: val}
}

type valueKeyRef struct{ val Value }
type strKeyRef struct{ str string }

func (valueKeyRef) typ() keyRefType { return keyRefTypeValue }
func (strKeyRef) typ() keyRefType   { return keyRefTypeStr }

func (k valueKeyRef) AsStr() option.Option[string] { return k.val.AsStr() }
func (k strKeyRef) AsStr() option.Option[string]   { return option.Some(k.str) }

func (k valueKeyRef) AsI64() option.Option[int64] {
	if i, err := k.val.TryToI64(); err != nil {
		return option.None[int64]()
	} else {
		return option.Some(i)
	}
}
func (k strKeyRef) AsI64() option.Option[int64] { return option.None[int64]() }

func (k valueKeyRef) AsValue() Value { return k.val.Clone() }
func (k strKeyRef) AsValue() Value   { return ValueFromString(k.str) }

func (k valueKeyRef) Hash(h hash.Hash) { keyRefHash(k, h) }
func (k strKeyRef) Hash(h hash.Hash)   { keyRefHash(k, h) }

func (k valueKeyRef) Equal(other any) bool { return keyRefEqualAny(k, other) }
func (k strKeyRef) Equal(other any) bool   { return keyRefEqualAny(k, other) }

func keyRefEqualAny(a KeyRef, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if bb, ok := b.(KeyRef); ok {
		return keyRefEqual(a, bb)
	}
	return false
}

func keyRefEqual(a, b KeyRef) bool {
	if optAStr, optBStr := a.AsStr(), b.AsStr(); optAStr.IsSome() && optBStr.IsSome() {
		return optAStr.Unwrap() == optBStr.Unwrap()
	}
	return Cmp(a.AsValue(), b.AsValue()) == 0
}

func keyRefHash(k KeyRef, h hash.Hash) {
	if optStr := k.AsStr(); optStr.IsSome() {
		io.WriteString(h, optStr.Unwrap())
	} else {
		valueHash(k.AsValue(), h)
	}
}

type keyRefType uint

const (
	keyRefTypeValue keyRefType = iota
	keyRefTypeStr
)
