package mjingo

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type value interface {
	String() string

	typ() valueType
	kind() valueKind
	isUndefined() bool
	isNone() bool
	getAttrFast(key string) option[value]
	getItemOpt(key value) option[value]
	asStr() option[string]
	tryToI64() (int64, error)
	asF64() option[float64]
	// clone() value TODO: implment
}

type valueType int

const (
	valueTypeUndefined valueType = iota + 1
	valueTypeBool
	valueTypeU64
	valueTypeI64
	valueTypeF64
	valueTypeNone
	valueTypeInvalid
	valueTypeU128
	valueTypeI128
	valueTypeString
	valueTypeBytes
	valueTypeSeq
	valueTypeMap
	valueTypeDynamic
)

type valueKind int

const (
	// The value is undefined
	valueKindUndefined valueKind = iota + 1
	// The value is the none singleton ([`()`])
	valueKindNone
	// The value is a [`bool`]
	valueKindBool
	// The value is a number of a supported type.
	valueKindNumber
	// The value is a string.
	valueKindString
	// The value is a byte array.
	valueKindBytes
	// The value is an array of other values.
	valueKindSeq
	// The value is a key/value mapping.
	valueKindMap
)

var valueUndefined = undefinedValue{}
var valueNone = noneValue{}

func (t valueType) String() string {
	switch t {
	case valueTypeUndefined:
		return "undefined"
	case valueTypeBool:
		return "bool"
	case valueTypeU64:
		return "u64"
	case valueTypeI64:
		return "i64"
	case valueTypeF64:
		return "f64"
	case valueTypeNone:
		return "none"
	case valueTypeInvalid:
		return "invalid"
	case valueTypeU128:
		return "u128"
	case valueTypeI128:
		return "i128"
	case valueTypeString:
		return "string"
	case valueTypeBytes:
		return "bytes"
	case valueTypeSeq:
		return "seq"
	case valueTypeMap:
		return "map"
	case valueTypeDynamic:
		return "dynamic"
	default:
		panic(fmt.Sprintf("invalid valueType: %d", t))
	}
}

type undefinedValue struct{}
type boolValue struct{ b bool }
type u64Value struct{ n uint64 }
type i64Value struct{ n int64 }
type f64Value struct{ f float64 }
type noneValue struct{}
type invalidValue struct{ detail string }
type u128Value struct{ hi, lo uint64 }
type i128Value struct {
	hi int64
	lo uint64
}
type stringValue struct{ s string }
type bytesValue struct{ b []byte }
type seqValue struct{ items []value }
type mapValue struct {
	// TODO: use an ordered map
	// TODO: use keyRef as key
	m map[string]value
}
type dynamicValue struct {
	// TODO: implement
}

var _ = value(undefinedValue{})
var _ = value(boolValue{})
var _ = value(u64Value{})
var _ = value(i64Value{})
var _ = value(f64Value{})
var _ = value(noneValue{})
var _ = value(invalidValue{})
var _ = value(u128Value{})
var _ = value(i128Value{})
var _ = value(stringValue{})
var _ = value(bytesValue{})
var _ = value(seqValue{})
var _ = value(mapValue{})
var _ = value(dynamicValue{})

func (v undefinedValue) String() string { return "" }
func (v boolValue) String() string      { return strconv.FormatBool(v.b) }
func (v u64Value) String() string       { return strconv.FormatUint(v.n, 10) }
func (v i64Value) String() string       { return strconv.FormatInt(v.n, 10) }
func (v f64Value) String() string {
	f := v.f
	if math.IsNaN(f) {
		return "NaN"
	} else if math.IsInf(f, 1) {
		return "inf"
	} else if math.IsInf(f, -1) {
		return "-inf"
	} else {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if strings.ContainsRune(s, '.') {
			return s
		}
		return s + ".0"
	}
}
func (v noneValue) String() string    { return "none" }
func (v invalidValue) String() string { return fmt.Sprintf("<invalid value: %s>", v.detail) }
func (v u128Value) String() string    { panic("not implemented yet") }
func (v i128Value) String() string    { panic("not implemented yet") }
func (v stringValue) String() string  { return v.s }
func (v bytesValue) String() string   { return string(v.b) } // TODO: equivalent impl as String::from_utf8_lossy
func (v seqValue) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, item := range v.items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item.String()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("]")
	return b.String()
}
func (v mapValue) String() string {
	var b strings.Builder
	b.WriteString("{")
	first := true
	for key, val := range v.m {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(val.String()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("}")
	return b.String()
}
func (v dynamicValue) String() string { panic("not implemented yet") }

func (undefinedValue) typ() valueType { return valueTypeUndefined }
func (boolValue) typ() valueType      { return valueTypeBool }
func (u64Value) typ() valueType       { return valueTypeU64 }
func (i64Value) typ() valueType       { return valueTypeI64 }
func (f64Value) typ() valueType       { return valueTypeF64 }
func (noneValue) typ() valueType      { return valueTypeNone }
func (invalidValue) typ() valueType   { return valueTypeInvalid }
func (u128Value) typ() valueType      { return valueTypeU128 }
func (i128Value) typ() valueType      { return valueTypeI128 }
func (stringValue) typ() valueType    { return valueTypeString }
func (bytesValue) typ() valueType     { return valueTypeBytes }
func (seqValue) typ() valueType       { return valueTypeSeq }
func (mapValue) typ() valueType       { return valueTypeMap }
func (dynamicValue) typ() valueType   { return valueTypeDynamic }

func (undefinedValue) kind() valueKind { return valueKindUndefined }
func (boolValue) kind() valueKind      { return valueKindBool }
func (u64Value) kind() valueKind       { return valueKindNumber }
func (i64Value) kind() valueKind       { return valueKindNumber }
func (f64Value) kind() valueKind       { return valueKindNumber }
func (noneValue) kind() valueKind      { return valueKindNone }
func (invalidValue) kind() valueKind   { return valueKindMap } // XXX: invalid values report themselves as maps which is a lie
func (u128Value) kind() valueKind      { return valueKindNumber }
func (i128Value) kind() valueKind      { return valueKindNumber }
func (stringValue) kind() valueKind    { return valueKindString }
func (bytesValue) kind() valueKind     { return valueKindBytes }
func (seqValue) kind() valueKind       { return valueKindSeq }
func (mapValue) kind() valueKind       { return valueKindMap }
func (dynamicValue) kind() valueKind   { panic("not implemented for valueTypeDynamic") }

func (undefinedValue) isNone() bool { return false }
func (boolValue) isNone() bool      { return false }
func (u64Value) isNone() bool       { return false }
func (i64Value) isNone() bool       { return false }
func (f64Value) isNone() bool       { return false }
func (noneValue) isNone() bool      { return true }
func (invalidValue) isNone() bool   { return false }
func (u128Value) isNone() bool      { return false }
func (i128Value) isNone() bool      { return false }
func (stringValue) isNone() bool    { return false }
func (bytesValue) isNone() bool     { return false }
func (seqValue) isNone() bool       { return false }
func (mapValue) isNone() bool       { return false }
func (dynamicValue) isNone() bool   { return false }

func (undefinedValue) isUndefined() bool { return true }
func (boolValue) isUndefined() bool      { return false }
func (u64Value) isUndefined() bool       { return false }
func (i64Value) isUndefined() bool       { return false }
func (f64Value) isUndefined() bool       { return false }
func (noneValue) isUndefined() bool      { return false }
func (invalidValue) isUndefined() bool   { return false }
func (u128Value) isUndefined() bool      { return false }
func (i128Value) isUndefined() bool      { return false }
func (stringValue) isUndefined() bool    { return false }
func (bytesValue) isUndefined() bool     { return false }
func (seqValue) isUndefined() bool       { return false }
func (mapValue) isUndefined() bool       { return false }
func (dynamicValue) isUndefined() bool   { return false }

func (undefinedValue) getAttrFast(key string) option[value] { return option[value]{} }
func (boolValue) getAttrFast(key string) option[value]      { return option[value]{} }
func (u64Value) getAttrFast(key string) option[value]       { return option[value]{} }
func (i64Value) getAttrFast(key string) option[value]       { return option[value]{} }
func (f64Value) getAttrFast(key string) option[value]       { return option[value]{} }
func (noneValue) getAttrFast(key string) option[value]      { return option[value]{} }
func (invalidValue) getAttrFast(key string) option[value]   { return option[value]{} }
func (u128Value) getAttrFast(key string) option[value]      { return option[value]{} }
func (i128Value) getAttrFast(key string) option[value]      { return option[value]{} }
func (stringValue) getAttrFast(key string) option[value]    { return option[value]{} }
func (bytesValue) getAttrFast(key string) option[value]     { return option[value]{} }
func (seqValue) getAttrFast(key string) option[value]       { return option[value]{} }
func (v mapValue) getAttrFast(key string) option[value] {
	if val, ok := v.m[key]; ok {
		return option[value]{valid: true, data: val}
	}
	return option[value]{}
}
func (dynamicValue) getAttrFast(key string) option[value] {
	panic("not implemented yet")
}

func (undefinedValue) getItemOpt(key value) option[value] { return option[value]{} }
func (boolValue) getItemOpt(key value) option[value]      { return option[value]{} }
func (u64Value) getItemOpt(key value) option[value]       { return option[value]{} }
func (i64Value) getItemOpt(key value) option[value]       { return option[value]{} }
func (f64Value) getItemOpt(key value) option[value]       { return option[value]{} }
func (noneValue) getItemOpt(key value) option[value]      { return option[value]{} }
func (invalidValue) getItemOpt(key value) option[value]   { return option[value]{} }
func (u128Value) getItemOpt(key value) option[value]      { return option[value]{} }
func (i128Value) getItemOpt(key value) option[value]      { return option[value]{} }
func (stringValue) getItemOpt(key value) option[value]    { return option[value]{} }
func (bytesValue) getItemOpt(key value) option[value]     { return option[value]{} }
func (v seqValue) getItemOpt(key value) option[value] {
	keyRf := keyRef{kind: keyRefKindValue, data: key}
	if idx := keyRf.asI64(); idx.valid {
		if idx.data < math.MinInt || math.MaxInt < idx.data {
			return option[value]{}
		}
		seq := newSliceSeqObject(v.items)
		var i uint
		if idx.data < 0 {
			c := seq.itemCount()
			if uint(-idx.data) > c {
				return option[value]{}
			}
			i = c - uint(-idx.data)
		} else {
			i = uint(idx.data)
		}
		return seq.getItem(i)
	}
	return option[value]{}
}
func (v mapValue) getItemOpt(key value) option[value] {
	keyRf := keyRef{kind: keyRefKindValue, data: key}
	// implementation here is different from minijinja.
	if keyData := keyRf.asStr(); keyData.valid {
		if v, ok := v.m[keyData.data]; ok {
			return option[value]{valid: true, data: v}
		}
		return option[value]{}
	}
	panic(fmt.Sprintf("value.getItemOpt does not support non string key: %+v", key))
}
func (dynamicValue) getItemOpt(key value) option[value] {
	panic("not implemented yet")
}

func (undefinedValue) asStr() option[string] { return option[string]{} }
func (boolValue) asStr() option[string]      { return option[string]{} }
func (u64Value) asStr() option[string]       { return option[string]{} }
func (i64Value) asStr() option[string]       { return option[string]{} }
func (f64Value) asStr() option[string]       { return option[string]{} }
func (noneValue) asStr() option[string]      { return option[string]{} }
func (invalidValue) asStr() option[string]   { return option[string]{} }
func (u128Value) asStr() option[string]      { return option[string]{} }
func (i128Value) asStr() option[string]      { return option[string]{} }
func (v stringValue) asStr() option[string]  { return option[string]{valid: true, data: v.s} }
func (bytesValue) asStr() option[string]     { return option[string]{} }
func (seqValue) asStr() option[string]       { return option[string]{} }
func (v mapValue) asStr() option[string]     { return option[string]{} }
func (dynamicValue) asStr() option[string] {
	panic("not implemented yet")
}

func (v undefinedValue) tryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v boolValue) tryToI64() (int64, error) {
	if v.b {
		return 1, nil
	}
	return 0, nil
}
func (v u64Value) tryToI64() (int64, error) { return int64(v.n), nil }
func (v i64Value) tryToI64() (int64, error) { return v.n, nil }
func (v f64Value) tryToI64() (int64, error) {
	if float64(int64(v.f)) == v.f {
		return int64(v.f), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v noneValue) tryToI64() (int64, error)    { return 0, unsupportedConversion(v.typ(), "i64") }
func (v invalidValue) tryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v u128Value) tryToI64() (int64, error)    { panic("not implemented yet") }
func (v i128Value) tryToI64() (int64, error)    { panic("not implemented yet") }
func (v stringValue) tryToI64() (int64, error)  { return 0, unsupportedConversion(v.typ(), "i64") }
func (v bytesValue) tryToI64() (int64, error)   { return 0, unsupportedConversion(v.typ(), "i64") }
func (v seqValue) tryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v mapValue) tryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v dynamicValue) tryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }

func (undefinedValue) asF64() option[float64] { return option[float64]{} }
func (v boolValue) asF64() option[float64] {
	if v.b {
		return option[float64]{valid: true, data: float64(1)}
	}
	return option[float64]{}
}
func (v u64Value) asF64() option[float64]    { return option[float64]{valid: true, data: float64(v.n)} }
func (v i64Value) asF64() option[float64]    { return option[float64]{valid: true, data: float64(v.n)} }
func (v f64Value) asF64() option[float64]    { return option[float64]{valid: true, data: v.f} }
func (noneValue) asF64() option[float64]     { return option[float64]{} }
func (invalidValue) asF64() option[float64]  { return option[float64]{} }
func (u128Value) asF64() option[float64]     { panic("not implemented yet") }
func (i128Value) asF64() option[float64]     { panic("not implemented yet") }
func (v stringValue) asF64() option[float64] { return option[float64]{} }
func (bytesValue) asF64() option[float64]    { return option[float64]{} }
func (seqValue) asF64() option[float64]      { return option[float64]{} }
func (mapValue) asF64() option[float64]      { return option[float64]{} }
func (dynamicValue) asF64() option[float64] {
	panic("not implemented yet")
}

func unsupportedConversion(kind valueType, target string) error {
	return &Error{
		typ: InvalidOperation,
		detail: option[string]{
			valid: true,
			data:  fmt.Sprintf("cannot convert %s to %s", kind, target),
		},
	}
}

func valueMapWithCapacity(capacity uint) map[string]value {
	return make(map[string]value, untrustedSizeHint(capacity))
}
