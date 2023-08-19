package valu

import "github.com/hnakamur/mjingo/internal/datast/option"

type KeyRef interface {
	typ() keyRefType
	AsStr() option.Option[string]
	AasI64() option.Option[int64]
}

func KeyRefFromValue(val Value) ValueKeyRef {
	return ValueKeyRef{val: val}
}

func KeyRefFromString(val string) StrKeyRef {
	return StrKeyRef{str: val}
}

type ValueKeyRef struct{ val Value }
type StrKeyRef struct{ str string }

func (ValueKeyRef) typ() keyRefType { return keyRefTypeValue }
func (StrKeyRef) typ() keyRefType   { return keyRefTypeStr }

func (k ValueKeyRef) AsStr() option.Option[string] { return k.val.AsStr() }
func (k StrKeyRef) AsStr() option.Option[string]   { return option.Some(k.str) }

func (k ValueKeyRef) AasI64() option.Option[int64] {
	if i, err := k.val.TryToI64(); err != nil {
		return option.None[int64]()
	} else {
		return option.Some(i)
	}
}
func (k StrKeyRef) AasI64() option.Option[int64] { return option.None[int64]() }

type keyRefType uint

const (
	keyRefTypeValue keyRefType = iota
	keyRefTypeStr
)
