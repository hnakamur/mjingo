package mjingo

import (
	"fmt"
	"hash"
	"io"

	"github.com/hnakamur/mjingo/internal/rustfmt"
	"github.com/hnakamur/mjingo/option"
)

type keyRef interface {
	typ() keyRefType
	AsStr() option.Option[string]
	AsI64() option.Option[int64]
	AsValue() Value
	Hash(h hash.Hash)
	Equal(other any) bool
}

func keyRefFromValue(val Value) keyRef {
	return valueKeyRef{val: val}
}

func keyRefFromString(val string) keyRef {
	return strKeyRef{str: val}
}

type valueKeyRef struct{ val Value }
type strKeyRef struct{ str string }

var _ rustfmt.Formatter = valueKeyRef{}
var _ rustfmt.Formatter = strKeyRef{}

func (valueKeyRef) typ() keyRefType { return keyRefTypeValue }
func (strKeyRef) typ() keyRefType   { return keyRefTypeStr }

func (k valueKeyRef) AsStr() option.Option[string] { return valueAsOptionString(k.val) }
func (k strKeyRef) AsStr() option.Option[string]   { return option.Some(k.str) }

func (k valueKeyRef) AsI64() option.Option[int64] {
	if i, err := k.val.tryToI64(); err != nil {
		return option.None[int64]()
	} else {
		return option.Some(i)
	}
}
func (k strKeyRef) AsI64() option.Option[int64] { return option.None[int64]() }

func (k valueKeyRef) AsValue() Value { return k.val.clone() }
func (k strKeyRef) AsValue() Value   { return valueFromString(k.str) }

func (k valueKeyRef) Hash(h hash.Hash) { keyRefHash(k, h) }
func (k strKeyRef) Hash(h hash.Hash)   { keyRefHash(k, h) }

func (k valueKeyRef) Equal(other any) bool { return keyRefEqualAny(k, other) }
func (k strKeyRef) Equal(other any) bool   { return keyRefEqualAny(k, other) }

func keyRefEqualAny(a keyRef, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if bb, ok := b.(keyRef); ok {
		return keyRefEqual(a, bb)
	}
	return false
}

func keyRefEqual(a, b keyRef) bool {
	if optAStr, optBStr := a.AsStr(), b.AsStr(); optAStr.IsSome() && optBStr.IsSome() {
		return optAStr.Unwrap() == optBStr.Unwrap()
	}
	return valueCmp(a.AsValue(), b.AsValue()) == 0
}

func keyRefHash(k keyRef, h hash.Hash) {
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

// Format implements rustfmt.Formatter.
func (k valueKeyRef) Format(f fmt.State, verb rune) {
	switch verb {
	case rustfmt.DisplayVerb, rustfmt.DebugVerb:
		k.val.Format(f, verb)
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods valueKeyRef
		type valueKeyRef hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), valueKeyRef(k))
	}
}

// SupportRustFormat implements rustfmt.Formatter.
func (valueKeyRef) SupportsCustomVerb(verb rune) bool {
	return verb == rustfmt.DebugVerb || verb == rustfmt.DisplayVerb
}

// Format implements rustfmt.Formatter.
func (k strKeyRef) Format(f fmt.State, verb rune) {
	switch verb {
	case rustfmt.DisplayVerb:
		fmt.Fprintf(f, "%s", k.str)
	case rustfmt.DebugVerb:
		fmt.Fprintf(f, "%q", k.str)
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods strKeyRef
		type strKeyRef hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), strKeyRef(k))
	}
}

// SupportRustFormat implements rustfmt.Formatter.
func (strKeyRef) SupportsCustomVerb(verb rune) bool {
	return verb == rustfmt.DebugVerb || verb == rustfmt.DisplayVerb
}
