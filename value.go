package mjingo

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/internal/rustfmt"
	"github.com/hnakamur/mjingo/option"
)

// Value provides a dynamic value type abstraction.
//
// This struct gives access to a dynamically typed value which is used by
// the template engine during execution.
//
// For the most part the existence of the value type can be ignored as
// mjingo will perform the necessary conversions for you.  For instance
// if you write a filter that converts a string you can directly declare the
// filter to take a string.  However for some more
// advanced use cases it's useful to know that this type exists.
//
// # Basic Value Conversions
//
// Values are typically created via [ValueFromGoValue] function.
//
// The special [Undefined] value also exists but does not
// have a Go equivalent.  It can be created via the [Undefined]
// variable.
type Value struct {
	data valueData
}

var _ rustfmt.Formatter = (*Value)(nil)

func (v Value) String() string      { return v.data.String() }
func (v Value) DebugString() string { return v.data.debugString() }
func (v Value) typ() valueType      { return v.data.typ() }

// Kind returns the Kind of the value.
//
// This can be used to determine what's in the value before trying to
// perform operations on it.
func (v Value) Kind() ValueKind { return v.data.kind() }

func (v Value) isUndefined() bool { return v.data.isUndefined() }
func (v Value) isNone() bool      { return v.data.isNone() }
func (v Value) isSafe() bool      { return v.data.isSafe() }
func (v Value) isTrue() bool      { return v.data.isTrue() }
func (v Value) getAttrFast(key string) option.Option[Value] {
	return v.data.getAttrFast(key)
}
func (v Value) getItemOpt(key Value) option.Option[Value] {
	return v.data.getItemOpt(key)
}
func (v Value) tryToI128() (*I128, error)       { return v.data.tryToI128() }
func (v Value) tryToI64() (int64, error)        { return v.data.tryToI64() }
func (v Value) tryToUint() (uint, error)        { return v.data.tryToUint() }
func (v Value) asF64() option.Option[float64]   { return v.data.asF64() }
func (v Value) asSeq() option.Option[SeqObject] { return v.data.asSeq() }
func (v Value) clone() Value                    { return Value{data: v.data.clone()} }
func (v Value) tryIter() (iterator, error)      { return v.data.tryIter() }
func (v Value) len() option.Option[uint]        { return v.data.len() }
func (v Value) hash(h hash.Hash)                { v.data.hash(h) }
func (v Value) Equal(other any) bool {
	if otherVal, ok := other.(Value); ok {
		return valueEqual(v, otherVal)
	}
	return false
}

type valueData interface {
	fmt.Stringer
	debugString() string

	typ() valueType
	kind() ValueKind
	isUndefined() bool
	isNone() bool
	isSafe() bool
	isTrue() bool
	getAttrFast(key string) option.Option[Value]
	getItemOpt(key Value) option.Option[Value]
	tryToI128() (*I128, error)
	tryToI64() (int64, error)
	tryToUint() (uint, error)
	asF64() option.Option[float64]
	asSeq() option.Option[SeqObject]
	clone() valueData
	tryIter() (iterator, error)
	len() option.Option[uint]
	hash(h hash.Hash)
	Equal(other any) bool
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

// ValueKind describes the kind of value.
type ValueKind int

const (
	// ValueKindUndefined represents the value is undefined
	ValueKindUndefined ValueKind = iota + 1
	// ValueKindNone represents the value is the none (None).
	//
	// Note this is different from a None value of option.Option.
	ValueKindNone
	// ValueKindBool repreesnts the value is a bool
	ValueKindBool
	// ValueKindNumber represents the value is a number of a supported type.
	ValueKindNumber
	// ValueKindString represents the value is a string.
	ValueKindString
	// ValueKindBytes represents the value is a byte array.
	ValueKindBytes
	// ValueKindSeq represents the value is an array of other values.
	ValueKindSeq
	// ValueKindMap represents the value is a key/value mapping.
	ValueKindMap
)

// Undefined is the undefined value.
//
// This constant variable exists because the undefined type does not exist in Go
// and this is the only way to construct it.
var Undefined Value

var none = Value{data: noneValue{}}

func init() {
	Undefined = Value{data: undefinedValue{}}
}

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

func (t valueType) Format(f fmt.State, verb rune) {
	switch verb {
	case 's':
		io.WriteString(f, t.String())
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods valueType
		type valueType hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), valueType(t))
	}
}

func (k ValueKind) String() string {
	switch k {
	case ValueKindUndefined:
		return "undefined"
	case ValueKindBool:
		return "bool"
	case ValueKindNumber:
		return "number"
	case ValueKindNone:
		return "none"
	case ValueKindString:
		return "string"
	case ValueKindBytes:
		return "bytes"
	case ValueKindSeq:
		return "sequence"
	case ValueKindMap:
		return "map"
	default:
		panic(fmt.Sprintf("invalid valueKind: %d", k))
	}
}

type undefinedValue struct{}
type boolValue struct{ B bool }
type u64Value struct{ N uint64 }
type i64Value struct{ N int64 }
type f64Value struct{ F float64 }
type noneValue struct{}
type invalidValue struct{ Detail string }
type u128Value struct{ N U128 }
type i128Value struct{ N I128 }
type stringValue struct {
	Str  string
	Type stringType
}
type bytesValue struct{ B []byte }
type seqValue struct{ Items []Value }
type mapValue struct {
	Map  *valueMap
	Type mapType
}
type dynamicValue struct {
	Dy Object
}

// / The type of map
type mapType uint

const (
	// A regular map
	mapTypeNormal mapType = iota + 1
	// A map representing keyword arguments
	mapTypeKwargs
)

type stringType uint

const (
	stringTypeNormal stringType = iota
	stringTypeSafe
)

var _ = valueData(undefinedValue{})
var _ = valueData(boolValue{})
var _ = valueData(u64Value{})
var _ = valueData(i64Value{})
var _ = valueData(f64Value{})
var _ = valueData(noneValue{})
var _ = valueData(invalidValue{})
var _ = valueData(u128Value{})
var _ = valueData(i128Value{})
var _ = valueData(stringValue{})
var _ = valueData(bytesValue{})
var _ = valueData(seqValue{})
var _ = valueData(mapValue{})
var _ = valueData(dynamicValue{})

func (v undefinedValue) String() string { return "" }
func (v boolValue) String() string      { return strconv.FormatBool(v.B) }
func (v u64Value) String() string       { return strconv.FormatUint(v.N, 10) }
func (v i64Value) String() string       { return strconv.FormatInt(v.N, 10) }
func (v f64Value) String() string {
	f := v.F
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
func (v invalidValue) String() string { return fmt.Sprintf("<invalid value: %s>", v.Detail) }
func (v u128Value) String() string    { return v.N.String() }
func (v i128Value) String() string    { return v.N.String() }
func (v stringValue) String() string  { return v.Str }
func (v bytesValue) String() string   { return string(v.B) } // TODO: equivalent impl as String::from_utf8_lossy
func (v seqValue) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, item := range v.Items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item.DebugString())
	}
	b.WriteString("]")
	return b.String()
}
func (v mapValue) String() string {
	var b strings.Builder
	b.WriteString("{")
	first := true
	l := v.Map.Len()
	for i := uint(0); i < l; i++ {
		e, _ := v.Map.EntryAt(i)
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(e.Key.AsValue().DebugString())
		b.WriteString(": ")
		b.WriteString(e.Value.DebugString())
	}
	b.WriteString("}")
	return b.String()
}
func (v dynamicValue) String() string {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return fmt.Sprintf("%+v", v.Dy)
	case ObjectKindSeq:
		seq := v.Dy.(SeqObject)
		var b strings.Builder
		b.WriteString("[")
		l := seq.ItemCount()
		for i := uint(0); i < l; i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			item := seq.GetItem(i).Unwrap()
			b.WriteString(item.DebugString())
		}
		b.WriteString("]")
		return b.String()
	case ObjectKindStruct:
		if m, ok := v.Dy.(*macro); ok {
			return m.String()
		}
		obj := v.Dy.(StructObject)
		fields := staticOrDynamicFields(obj)
		var b strings.Builder
		b.WriteString("{")
		for i, field := range fields {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(field)
			b.WriteString(": ")
			b.WriteString(obj.GetField(field).Unwrap().String())
		}
		b.WriteString("}")
		return b.String()
	default:
		panic("unreachable")
	}
}

func (v undefinedValue) debugString() string { return "undefined" }
func (v boolValue) debugString() string      { return strconv.FormatBool(v.B) }
func (v u64Value) debugString() string       { return strconv.FormatUint(v.N, 10) }
func (v i64Value) debugString() string       { return strconv.FormatInt(v.N, 10) }
func (v f64Value) debugString() string {
	f := v.F
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
func (v noneValue) debugString() string    { return "none" }
func (v invalidValue) debugString() string { return fmt.Sprintf("<invalid value: %s>", v.Detail) }
func (v u128Value) debugString() string    { return v.N.String() }
func (v i128Value) debugString() string    { return v.N.String() }
func (v stringValue) debugString() string  { return fmt.Sprintf("%q", v.Str) }
func (v bytesValue) debugString() string   { return string(v.B) } // TODO: equivalent impl as String::from_utf8_lossy
func (v seqValue) debugString() string {
	var b strings.Builder
	b.WriteString("[")
	for i, item := range v.Items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item.DebugString())
	}
	b.WriteString("]")
	return b.String()
}
func (v mapValue) debugString() string {
	var b strings.Builder
	b.WriteString("{")
	first := true
	l := v.Map.Len()
	for i := uint(0); i < l; i++ {
		e, _ := v.Map.EntryAt(i)
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(e.Key.AsValue().DebugString())
		b.WriteString(": ")
		b.WriteString(e.Value.DebugString())
	}
	b.WriteString("}")
	return b.String()
}
func (v dynamicValue) debugString() string { return v.String() }

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

func (undefinedValue) kind() ValueKind { return ValueKindUndefined }
func (boolValue) kind() ValueKind      { return ValueKindBool }
func (u64Value) kind() ValueKind       { return ValueKindNumber }
func (i64Value) kind() ValueKind       { return ValueKindNumber }
func (f64Value) kind() ValueKind       { return ValueKindNumber }
func (noneValue) kind() ValueKind      { return ValueKindNone }
func (invalidValue) kind() ValueKind {
	// XXX: invalid values report themselves as maps which is a lie
	return ValueKindMap
}
func (u128Value) kind() ValueKind   { return ValueKindNumber }
func (i128Value) kind() ValueKind   { return ValueKindNumber }
func (stringValue) kind() ValueKind { return ValueKindString }
func (bytesValue) kind() ValueKind  { return ValueKindBytes }
func (seqValue) kind() ValueKind    { return ValueKindSeq }
func (mapValue) kind() ValueKind    { return ValueKindMap }
func (v dynamicValue) kind() ValueKind {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		// XXX: basic objects should probably not report as map
		return ValueKindMap
	case ObjectKindSeq:
		return ValueKindSeq
	case ObjectKindStruct:
		return ValueKindMap
	default:
		panic("unreachable")
	}
}

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

func (undefinedValue) isSafe() bool { return false }
func (boolValue) isSafe() bool      { return false }
func (u64Value) isSafe() bool       { return false }
func (i64Value) isSafe() bool       { return false }
func (f64Value) isSafe() bool       { return false }
func (noneValue) isSafe() bool      { return false }
func (invalidValue) isSafe() bool   { return false }
func (u128Value) isSafe() bool      { return false }
func (i128Value) isSafe() bool      { return false }
func (v stringValue) isSafe() bool  { return v.Type == stringTypeSafe }
func (bytesValue) isSafe() bool     { return false }
func (seqValue) isSafe() bool       { return false }
func (mapValue) isSafe() bool       { return false }
func (dynamicValue) isSafe() bool   { return false }

func (undefinedValue) isTrue() bool { return false }
func (v boolValue) isTrue() bool    { return v.B }
func (v u64Value) isTrue() bool     { return v.N != 0 }
func (v i64Value) isTrue() bool     { return v.N != 0 }
func (v f64Value) isTrue() bool     { return v.F != 0.0 }
func (noneValue) isTrue() bool      { return false }
func (invalidValue) isTrue() bool   { return false }
func (v u128Value) isTrue() bool {
	var zero U128
	return v.N.Cmp(&zero) != 0
}
func (v i128Value) isTrue() bool {
	var zero I128
	return v.N.Cmp(&zero) != 0
}
func (v stringValue) isTrue() bool { return len(v.Str) != 0 }
func (v bytesValue) isTrue() bool  { return len(v.B) != 0 }
func (v seqValue) isTrue() bool    { return len(v.Items) != 0 }
func (v mapValue) isTrue() bool    { return v.Map.Len() != 0 }
func (v dynamicValue) isTrue() bool {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return true
	case ObjectKindSeq:
		return v.Dy.(SeqObject).ItemCount() != 0
	case ObjectKindStruct:
		return fieldCount(v.Dy.(StructObject)) != 0
	default:
		panic("unreachable")
	}
}

func (undefinedValue) getAttrFast(_ string) option.Option[Value] { return option.None[Value]() }
func (boolValue) getAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (u64Value) getAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (i64Value) getAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (f64Value) getAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (noneValue) getAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (invalidValue) getAttrFast(_ string) option.Option[Value]   { return option.None[Value]() }
func (u128Value) getAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (i128Value) getAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (stringValue) getAttrFast(_ string) option.Option[Value]    { return option.None[Value]() }
func (bytesValue) getAttrFast(_ string) option.Option[Value]     { return option.None[Value]() }
func (seqValue) getAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (v mapValue) getAttrFast(key string) option.Option[Value] {
	if val, ok := v.Map.Get(keyRefFromString(key)); ok {
		return option.Some(val)
	}
	return option.None[Value]()
}
func (v dynamicValue) getAttrFast(key string) option.Option[Value] {
	if s, ok := v.Dy.(StructObject); ok {
		return s.GetField(key)
	}
	return option.None[Value]()
}

func (undefinedValue) getItemOpt(_ Value) option.Option[Value] { return option.None[Value]() }
func (boolValue) getItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (u64Value) getItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (i64Value) getItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (f64Value) getItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (noneValue) getItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (invalidValue) getItemOpt(_ Value) option.Option[Value]   { return option.None[Value]() }
func (u128Value) getItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (i128Value) getItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (v stringValue) getItemOpt(key Value) option.Option[Value] {
	idx, err := key.tryToI64()
	if err != nil {
		return option.None[Value]()
	}
	chars := []rune(v.Str)
	if idx < 0 {
		if -idx > int64(len(chars)) {
			return option.None[Value]()
		}
		idx = int64(len(chars)) + idx
	}
	if idx >= int64(len(chars)) {
		return option.None[Value]()
	}
	return option.Some(valueFromString(string(chars[idx])))
}
func (bytesValue) getItemOpt(_ Value) option.Option[Value] { return option.None[Value]() }
func (v seqValue) getItemOpt(key Value) option.Option[Value] {
	return getItemOptFromSeq(newSliceSeqObject(v.Items), key)
}
func (v mapValue) getItemOpt(key Value) option.Option[Value] {
	if v, ok := v.Map.Get(keyRefFromValue(key)); ok {
		return option.Some(v)
	}
	return option.None[Value]()
}
func (v dynamicValue) getItemOpt(key Value) option.Option[Value] {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return option.None[Value]()
	case ObjectKindSeq:
		return getItemOptFromSeq(v.Dy.(SeqObject), key)
	case ObjectKindStruct:
		if strKey := ""; valueAsOptionString(key).UnwrapTo(&strKey) {
			return v.Dy.(StructObject).GetField(strKey)
		}
		return option.None[Value]()
	default:
		panic("unreachable")
	}
}

func getItemOptFromSeq(seq SeqObject, key Value) option.Option[Value] {
	keyRf := valueKeyRef{val: key}
	if optIdx := keyRf.AsI64(); optIdx.IsSome() {
		idx := optIdx.Unwrap()
		if idx < math.MinInt || math.MaxInt < idx {
			return option.None[Value]()
		}
		var i uint
		if idx < 0 {
			c := seq.ItemCount()
			if uint(-idx) > c {
				return option.None[Value]()
			}
			i = c - uint(-idx)
		} else {
			i = uint(idx)
		}
		return seq.GetItem(i)
	}
	return option.None[Value]()
}

func (v undefinedValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v boolValue) tryToI128() (*I128, error) {
	var n int64
	if v.B {
		n = 1
	}
	return I128FromInt64(n), nil
}
func (v u64Value) tryToI128() (*I128, error) { return I128FromUint64(v.N), nil }
func (v i64Value) tryToI128() (*I128, error) { return I128FromInt64(v.N), nil }
func (v f64Value) tryToI128() (*I128, error) {
	if float64(int64(v.F)) == v.F {
		return I128FromInt64(int64(v.F)), nil
	}
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v noneValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v invalidValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v u128Value) tryToI128() (*I128, error) {
	if v.N.n.Cmp(i128Max) > 0 {
		return nil, unsupportedConversion(v.typ(), "i128")
	}
	return I128TryFromBigInt(&v.N.n)
}
func (v i128Value) tryToI128() (*I128, error) { return I128TryFromBigInt(&v.N.n) }
func (v stringValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v bytesValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v seqValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v mapValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}
func (v dynamicValue) tryToI128() (*I128, error) {
	return nil, unsupportedConversion(v.typ(), "i128")
}

func (v undefinedValue) tryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v boolValue) tryToI64() (int64, error) {
	if v.B {
		return 1, nil
	}
	return 0, nil
}
func (v u64Value) tryToI64() (int64, error) {
	if v.N > math.MaxInt64 {
		return 0, unsupportedConversion(v.typ(), "i64")
	}
	return int64(v.N), nil
}
func (v i64Value) tryToI64() (int64, error) { return v.N, nil }
func (v f64Value) tryToI64() (int64, error) {
	if float64(int64(v.F)) == v.F {
		return int64(v.F), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v noneValue) tryToI64() (int64, error)    { return 0, unsupportedConversion(v.typ(), "i64") }
func (v invalidValue) tryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v u128Value) tryToI64() (int64, error) {
	if v.N.IsInt64() {
		return v.N.Int64(), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v i128Value) tryToI64() (int64, error) {
	if v.N.IsInt64() {
		return v.N.Int64(), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v stringValue) tryToI64() (int64, error)  { return 0, unsupportedConversion(v.typ(), "i64") }
func (v bytesValue) tryToI64() (int64, error)   { return 0, unsupportedConversion(v.typ(), "i64") }
func (v seqValue) tryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v mapValue) tryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v dynamicValue) tryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }

func (v undefinedValue) tryToUint() (uint, error) { return 0, unsupportedConversion(v.typ(), "uint") }
func (v boolValue) tryToUint() (uint, error) {
	if v.B {
		return 1, nil
	}
	return 0, nil
}
func (v u64Value) tryToUint() (uint, error) {
	if v.N > math.MaxUint {
		return 0, unsupportedConversion(v.typ(), "uint")
	}
	return uint(v.N), nil
}
func (v i64Value) tryToUint() (uint, error) {
	if v.N < 0 {
		return 0, unsupportedConversion(v.typ(), "uint")
	}
	return uint(v.N), nil
}
func (v f64Value) tryToUint() (uint, error) {
	// MiniJinja uses int64 here, not uint.
	// https://github.com/mitsuhiko/minijinja/blob/1.0.7/minijinja/src/value/argtypes.rs#L438-L439
	// And it has comment "for the intention here see Key::from_borrowed_value"
	// but "from_borrowed_value" does not exist.
	if float64(int64(v.F)) == v.F && v.F >= 0 {
		return uint(v.F), nil
	}
	return 0, unsupportedConversion(v.typ(), "uint")
}
func (v noneValue) tryToUint() (uint, error)    { return 0, unsupportedConversion(v.typ(), "uint") }
func (v invalidValue) tryToUint() (uint, error) { return 0, unsupportedConversion(v.typ(), "uint") }
func (v u128Value) tryToUint() (uint, error) {
	if v.N.IsUint64() {
		n := v.N.Uint64()
		if n <= math.MaxUint {
			return uint(n), nil
		}
	}
	return 0, unsupportedConversion(v.typ(), "uint")
}
func (v i128Value) tryToUint() (uint, error) {
	if v.N.IsUint64() {
		n := v.N.Uint64()
		if n <= math.MaxUint {
			return uint(n), nil
		}
	}
	return 0, unsupportedConversion(v.typ(), "uint")
}
func (v stringValue) tryToUint() (uint, error)  { return 0, unsupportedConversion(v.typ(), "uint") }
func (v bytesValue) tryToUint() (uint, error)   { return 0, unsupportedConversion(v.typ(), "uint") }
func (v seqValue) tryToUint() (uint, error)     { return 0, unsupportedConversion(v.typ(), "uint") }
func (v mapValue) tryToUint() (uint, error)     { return 0, unsupportedConversion(v.typ(), "uint") }
func (v dynamicValue) tryToUint() (uint, error) { return 0, unsupportedConversion(v.typ(), "uint") }

func (undefinedValue) asF64() option.Option[float64] { return option.None[float64]() }
func (v boolValue) asF64() option.Option[float64] {
	var f float64
	if v.B {
		f = 1
	}
	return option.Some(f)
}
func (v u64Value) asF64() option.Option[float64]   { return option.Some(float64(v.N)) }
func (v i64Value) asF64() option.Option[float64]   { return option.Some(float64(v.N)) }
func (v f64Value) asF64() option.Option[float64]   { return option.Some(v.F) }
func (noneValue) asF64() option.Option[float64]    { return option.None[float64]() }
func (invalidValue) asF64() option.Option[float64] { return option.None[float64]() }
func (v u128Value) asF64() option.Option[float64] {
	f, _ := v.N.n.Float64()
	return option.Some(f)
}
func (v i128Value) asF64() option.Option[float64] {
	f, _ := v.N.n.Float64()
	return option.Some(f)
}
func (stringValue) asF64() option.Option[float64]  { return option.None[float64]() }
func (bytesValue) asF64() option.Option[float64]   { return option.None[float64]() }
func (seqValue) asF64() option.Option[float64]     { return option.None[float64]() }
func (mapValue) asF64() option.Option[float64]     { return option.None[float64]() }
func (dynamicValue) asF64() option.Option[float64] { return option.None[float64]() }

func (undefinedValue) asSeq() option.Option[SeqObject] { return option.None[SeqObject]() }
func (boolValue) asSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (u64Value) asSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (i64Value) asSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (f64Value) asSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (noneValue) asSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (invalidValue) asSeq() option.Option[SeqObject]   { return option.None[SeqObject]() }
func (u128Value) asSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (i128Value) asSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (stringValue) asSeq() option.Option[SeqObject]    { return option.None[SeqObject]() }
func (bytesValue) asSeq() option.Option[SeqObject]     { return option.None[SeqObject]() }
func (v seqValue) asSeq() option.Option[SeqObject] {
	return option.Some(newSliceSeqObject(v.Items))
}
func (mapValue) asSeq() option.Option[SeqObject] { return option.None[SeqObject]() }
func (v dynamicValue) asSeq() option.Option[SeqObject] {
	if seq, ok := v.Dy.(SeqObject); ok {
		return option.Some(seq)
	}
	return option.None[SeqObject]()
}

func (v undefinedValue) clone() valueData { return v }
func (v boolValue) clone() valueData      { return v }
func (v u64Value) clone() valueData       { return v }
func (v i64Value) clone() valueData       { return v }
func (v f64Value) clone() valueData       { return v }
func (v noneValue) clone() valueData      { return v }
func (v invalidValue) clone() valueData   { return v }
func (v u128Value) clone() valueData {
	c := v
	c.N.Set(&v.N)
	return c
}
func (v i128Value) clone() valueData {
	c := v
	c.N.Set(&v.N)
	return c
}
func (v stringValue) clone() valueData { return v }
func (v bytesValue) clone() valueData {
	b := make([]byte, len(v.B))
	copy(b, v.B)
	return bytesValue{B: b}
}
func (v seqValue) clone() valueData {
	items := make([]Value, len(v.Items))
	for i, item := range v.Items {
		// Is shallow copy OK?
		items[i] = item
	}
	return seqValue{Items: items}
}
func (v mapValue) clone() valueData {
	m := v.Map.Clone()
	return mapValue{Map: m, Type: v.Type}
}
func (v dynamicValue) clone() valueData {
	// TODO: implement real clone
	return v
}

func (undefinedValue) tryIter() (iterator, error) {
	return iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v boolValue) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (v u64Value) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (v i64Value) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (v f64Value) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (noneValue) tryIter() (iterator, error) {
	return iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v invalidValue) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (v u128Value) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (v i128Value) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (v stringValue) tryIter() (iterator, error) {
	return iterator{iterState: &charsValueIteratorState{s: v.Str}, len: uint(utf8.RuneCountInString(v.Str))}, nil
}
func (v bytesValue) tryIter() (iterator, error) {
	return iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iterable", v.kind()))
}
func (v seqValue) tryIter() (iterator, error) {
	return iterator{iterState: &seqValueIteratorState{items: v.Items}, len: uint(len(v.Items))}, nil
}
func (v mapValue) tryIter() (iterator, error) {
	return iterator{iterState: &mapValueIteratorState{keys: v.Map.Keys()}, len: uint(len(v.Map.Keys()))}, nil
}
func (v dynamicValue) tryIter() (iterator, error) {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return iterator{iterState: &emptyValueIteratorState{}}, nil
	case ObjectKindSeq:
		seqObj := v.Dy.(SeqObject)
		return iterator{iterState: &dynSeqValueIteratorState{obj: seqObj}, len: seqObj.ItemCount()}, nil
	case ObjectKindStruct:
		obj := v.Dy.(StructObject)
		fields := staticOrDynamicFields(obj)
		return iterator{iterState: &stringsValueIteratorState{items: fields}}, nil
	default:
		panic("unreachable")
	}
}

func unsupportedConversion(kind valueType, target string) error {
	return NewError(InvalidOperation,
		fmt.Sprintf("cannot convert %s to %s", kind, target))
}

func valueTryToGoInt8(val Value) (int8, error) {
	n, err := val.tryToI64()
	if err != nil || n < math.MinInt8 || n > math.MaxInt8 {
		return 0, unsupportedConversion(val.typ(), "int8")
	}
	return int8(n), nil
}

func valueTryToGoInt16(val Value) (int16, error) {
	n, err := val.tryToI64()
	if err != nil || n < math.MinInt16 || n > math.MaxInt16 {
		return 0, unsupportedConversion(val.typ(), "int16")
	}
	return int16(n), nil
}

func valueTryToGoInt32(val Value) (int32, error) {
	n, err := val.tryToI64()
	if err != nil || n < math.MinInt32 || n > math.MaxInt32 {
		return 0, unsupportedConversion(val.typ(), "int32")
	}
	return int32(n), nil
}

func valueTryToGoInt64(val Value) (int64, error) { return val.tryToI64() }

func valueTryToGoInt(val Value) (int, error) {
	n, err := val.tryToI64()
	if err != nil || n < math.MinInt || n > math.MaxInt8 {
		return 0, unsupportedConversion(val.typ(), "int")
	}
	return int(n), nil
}

func valueTryToGoUint8(val Value) (uint8, error) {
	n, err := val.tryToI64()
	if err != nil || n < 0 || n > math.MaxUint8 {
		return 0, unsupportedConversion(val.typ(), "uint8")
	}
	return uint8(n), nil
}

func valueTryToGoUint16(val Value) (uint16, error) {
	n, err := val.tryToI64()
	if err != nil || n < 0 || n > math.MaxUint16 {
		return 0, unsupportedConversion(val.typ(), "uint16")
	}
	return uint16(n), nil
}

func valueTryToGoUint32(val Value) (uint32, error) {
	n, err := val.tryToI64()
	if err != nil || n < 0 || n > math.MaxUint32 {
		return 0, unsupportedConversion(val.typ(), "uint32")
	}
	return uint32(n), nil
}

func valueTryToGoUint64(val Value) (uint64, error) {
	n, err := val.tryToI128()
	if err != nil || !n.IsUint64() {
		return 0, unsupportedConversion(val.typ(), "uint64")
	}
	return n.Uint64(), nil
}

func valueTryToGoI128(val Value) (I128, error) {
	i, err := val.tryToI128()
	if err != nil {
		return I128{}, err
	}
	return *i, nil
}

func valueTryToGoU128(val Value) (U128, error) {
	if u, ok := val.data.(u128Value); ok {
		u2, _ := U128TryFromBigInt(&u.N.n)
		return *u2, nil
	}
	n, err := val.tryToI128()
	var zero I128
	if err != nil || n.Cmp(&zero) < 0 {
		return U128{}, unsupportedConversion(val.typ(), "uint64")
	}
	u, _ := U128TryFromBigInt(&n.n)
	return *u, nil
}

func valueTryToGoFloat32(val Value) (float32, error) {
	var f float64
	if val.asF64().UnwrapTo(&f) {
		return float32(f), nil
	}
	return 0, unsupportedConversion(val.typ(), "float32")
}

func valueTryToGoFloat64(val Value) (float64, error) {
	var f float64
	if val.asF64().UnwrapTo(&f) {
		return f, nil
	}
	return 0, unsupportedConversion(val.typ(), "float64")
}

type iterator struct {
	iterState valueIteratorState
	len       uint
}

func iteratorFromSeqObject(s SeqObject) *iterator {
	return &iterator{iterState: &dynSeqValueIteratorState{obj: s}}
}

func iteratorFromStrings(items []string) *iterator {
	return &iterator{iterState: &stringsValueIteratorState{items: items}}
}

func (i iterator) Chain(other iterator) iterator {
	return iterator{
		iterState: &chainedValueIteratorState{
			states: []valueIteratorState{i.iterState, other.iterState},
		},
		len: i.len + other.len,
	}
}

func (i iterator) Cloned() iterator {
	return iterator{
		iterState: &cloneValueIteratorState{
			state: i.iterState,
		},
		len: i.len,
	}
}

func (i *iterator) Next() option.Option[Value] {
	optVal := i.iterState.advanceState()
	if optVal.IsSome() {
		i.len--
	}
	return optVal
}

func (i *iterator) Len() uint {
	return i.len
}

// All returns if every element of the iterator matches a predicate.
// An empty iterator returns true.
func (i *iterator) All(f func(Value) bool) bool {
	for item := (Value{}); i.Next().UnwrapTo(&item); {
		if !f(item) {
			return false
		}
	}
	return true
}

func (i *iterator) CompareBy(other *iterator, f func(a, b Value) int) int {
	for {
		optA := i.Next()
		optB := other.Next()
		if optA.IsNone() {
			if optB.IsSome() {
				return -1
			}
			break
		}
		if optB.IsNone() {
			if optA.IsSome() {
				return 1
			}
			break
		}
		if c := f(optA.Unwrap(), optB.Unwrap()); c != 0 {
			return c
		}
	}
	return 0
}

func (i *iterator) Min() option.Option[Value] { return i.minBy(valueCmp) }

func (i *iterator) minBy(compare func(a, b Value) int) option.Option[Value] {
	rv := option.None[Value]()
	for {
		optItem := i.Next()
		if optItem.IsNone() {
			break
		}
		if rv.IsNone() || compare(optItem.Unwrap(), rv.Unwrap()) < 0 {
			rv = optItem
		}
	}
	return rv
}

func (i *iterator) Max() option.Option[Value] { return i.maxBy(valueCmp) }

func (i *iterator) maxBy(compare func(a, b Value) int) option.Option[Value] {
	rv := option.None[Value]()
	for {
		optItem := i.Next()
		if optItem.IsNone() {
			break
		}
		if rv.IsNone() || compare(optItem.Unwrap(), rv.Unwrap()) > 0 {
			rv = optItem
		}
	}
	return rv
}

func (i *iterator) Collect() []Value {
	items := make([]Value, 0, i.Len())
	for item := (Value{}); i.Next().UnwrapTo(&item); {
		items = append(items, item)
	}
	return items
}

type valueIteratorState interface {
	advanceState() option.Option[Value]
}

type emptyValueIteratorState struct{}
type charsValueIteratorState struct {
	offset uint
	s      string
}
type seqValueIteratorState struct {
	idx   uint
	items []Value
}
type stringsValueIteratorState struct {
	idx   uint
	items []string
}
type dynSeqValueIteratorState struct {
	idx uint
	obj SeqObject
}
type mapValueIteratorState struct {
	idx  uint
	keys []keyRef
}

type cloneValueIteratorState struct {
	state valueIteratorState
}

func (s *cloneValueIteratorState) advanceState() option.Option[Value] {
	return option.Map(s.state.advanceState(), func(val Value) Value { return val.clone() })
}

type chainedValueIteratorState struct {
	states []valueIteratorState
}

func (s *chainedValueIteratorState) advanceState() option.Option[Value] {
	var rv option.Option[Value]
	if len(s.states) > 0 {
		rv = s.states[0].advanceState()
		for rv.IsNone() && len(s.states) > 1 {
			clear(s.states[0:1])
			s.states = s.states[1:]
			rv = s.states[0].advanceState()
		}
	}
	return rv
}

func (s *emptyValueIteratorState) advanceState() option.Option[Value] { return option.None[Value]() }
func (s *charsValueIteratorState) advanceState() option.Option[Value] {
	if s.offset < uint(len(s.s)) {
		r, size := utf8.DecodeRuneInString(s.s[s.offset:])
		s.offset += uint(size)
		return option.Some(valueFromString(string(r)))
	}
	return option.None[Value]()
}
func (s *seqValueIteratorState) advanceState() option.Option[Value] {
	if s.idx < uint(len(s.items)) {
		item := s.items[s.idx]
		s.idx++
		return option.Some(item.clone())
	}
	return option.None[Value]()
}
func (s *stringsValueIteratorState) advanceState() option.Option[Value] {
	if s.idx < uint(len(s.items)) {
		item := s.items[s.idx]
		s.idx++
		return option.Some(valueFromString(item))
	}
	return option.None[Value]()
}
func (s *dynSeqValueIteratorState) advanceState() option.Option[Value] {
	val := s.obj.GetItem(s.idx)
	s.idx++
	return val
}
func (s *mapValueIteratorState) advanceState() option.Option[Value] {
	if s.idx < uint(len(s.keys)) {
		key := s.keys[s.idx]
		s.idx++
		return option.Some(key.AsValue())
	}
	return option.None[Value]()
}

var _ = valueIteratorState((*emptyValueIteratorState)(nil))
var _ = valueIteratorState((*charsValueIteratorState)(nil))
var _ = valueIteratorState((*seqValueIteratorState)(nil))
var _ = valueIteratorState((*stringsValueIteratorState)(nil))
var _ = valueIteratorState((*dynSeqValueIteratorState)(nil))
var _ = valueIteratorState((*mapValueIteratorState)(nil))
var _ = valueIteratorState((*cloneValueIteratorState)(nil))
var _ = valueIteratorState((*chainedValueIteratorState)(nil))

type valueIteratorStateType int

const (
	valueIteratorStateTypeEmpty valueIteratorStateType = iota + 1
	valueIteratorStateTypeChars
	valueIteratorStateTypeSeq
	valueIteratorStateTypeStrings
	valueIteratorStateTypeDynSeq
	valueIteratorStateTypeMap
)

func (v *seqValue) Append(val Value) {
	v.Items = append(v.Items, val)
}

func (undefinedValue) len() option.Option[uint] { return option.None[uint]() }
func (boolValue) len() option.Option[uint]      { return option.None[uint]() }
func (u64Value) len() option.Option[uint]       { return option.None[uint]() }
func (i64Value) len() option.Option[uint]       { return option.None[uint]() }
func (f64Value) len() option.Option[uint]       { return option.None[uint]() }
func (noneValue) len() option.Option[uint]      { return option.None[uint]() }
func (invalidValue) len() option.Option[uint]   { return option.None[uint]() }
func (u128Value) len() option.Option[uint]      { return option.None[uint]() }
func (i128Value) len() option.Option[uint]      { return option.None[uint]() }
func (v stringValue) len() option.Option[uint] {
	return option.Some(uint(utf8.RuneCountInString(v.Str)))
}
func (bytesValue) len() option.Option[uint] { return option.None[uint]() }
func (v seqValue) len() option.Option[uint] { return option.Some(uint(len(v.Items))) }
func (v mapValue) len() option.Option[uint] { return option.Some(v.Map.Len()) }
func (v dynamicValue) len() option.Option[uint] {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return option.None[uint]()
	case ObjectKindSeq:
		return option.Some(v.Dy.(SeqObject).ItemCount())
	case ObjectKindStruct:
		return option.Some(fieldCount(v.Dy.(StructObject)))
	default:
		panic("unreachable")
	}
}

func valueEqual(v, other Value) bool { return valueDataEqual(v.data, other.data) }

func valueDataEqual(v, other valueData) bool {
	switch {
	case v.kind() == ValueKindNone && other.kind() == ValueKindNone:
		return true
	case v.kind() == ValueKindUndefined && other.kind() == ValueKindUndefined:
		return true
	case v.kind() == ValueKindString && other.kind() == ValueKindString:
		a := v.(stringValue).Str
		b := other.(stringValue).Str
		return a == b
	case v.kind() == ValueKindBytes && other.kind() == ValueKindBytes:
		a := v.(bytesValue).B
		b := other.(bytesValue).B
		return bytes.Equal(a, b)
	default:
		switch c := coerceData(v, other).(type) {
		case f64CoerceResult:
			return c.lhs == c.rhs
		case i128CoerceResult:
			return c.lhs.Cmp(&c.rhs) == 0
		case strCoerceResult:
			return c.lhs == c.rhs
		default:
			if optA, optB := v.asSeq(), other.asSeq(); optA.IsSome() && optB.IsSome() {
				iterA, err := v.tryIter()
				if err != nil {
					return false
				}
				iterB, err := v.tryIter()
				if err != nil {
					return false
				}
				return iterA.All(func(itemA Value) bool {
					itemB := iterB.Next().Unwrap()
					return valueEqual(itemA, itemB)
				})
			} else if v.kind() == ValueKindMap && other.kind() == ValueKindMap {
				if v.len() != other.len() {
					return false
				}
				iterA, err := v.tryIter()
				if err != nil {
					return false
				}
				return iterA.All(func(key Value) bool {
					optValA := v.getItemOpt(key)
					optValB := other.getItemOpt(key)
					if optValA.IsSome() && optValB.IsSome() {
						return valueEqual(optValA.Unwrap(), optValB.Unwrap())
					}
					return false
				})
			}
		}
	}
	return false
}

func valueDataEqualAny(v valueData, other any) bool {
	if o, ok := other.(valueData); ok {
		return valueDataEqual(v, o)
	}
	return false
}

// valueCmp returns
// -1 if v is less than other,
//
//	0 if v equals other,
//
// +1 if v is greater than other.
func valueCmp(v, other Value) int {
	var rv int
outer:
	switch {
	case v.Kind() == ValueKindNone && other.Kind() == ValueKindNone:
		rv = 0
	case v.Kind() == ValueKindUndefined && other.Kind() == ValueKindUndefined:
		rv = 0
	case v.Kind() == ValueKindString && other.Kind() == ValueKindString:
		a := v.data.(stringValue).Str
		b := other.data.(stringValue).Str
		rv = strings.Compare(a, b)
	case v.Kind() == ValueKindBytes && other.Kind() == ValueKindBytes:
		a := v.data.(bytesValue).B
		b := other.data.(bytesValue).B
		rv = bytes.Compare(a, b)
	default:
		switch c := coerce(v, other).(type) {
		case f64CoerceResult:
			rv = f64TotalCmp(c.lhs, c.rhs)
		case i128CoerceResult:
			rv = c.lhs.Cmp(&c.rhs)
		case strCoerceResult:
			rv = strings.Compare(c.lhs, c.rhs)
		default:
			if optA, optB := v.asSeq(), other.asSeq(); optA.IsSome() && optB.IsSome() {
				iterA, err := v.tryIter()
				if err != nil {
					break outer
				}
				iterB, err := other.tryIter()
				if err != nil {
					break outer
				}
				rv = iterA.CompareBy(&iterB, valueCmp)
			} else if v.Kind() == ValueKindMap && other.Kind() == ValueKindMap {
				iterA, err := v.tryIter()
				if err != nil {
					break outer
				}
				iterB, err := other.tryIter()
				if err != nil {
					break outer
				}
				rv = iterA.CompareBy(&iterB, func(keyA, keyB Value) int {
					if rv := valueCmp(keyA, keyB); rv != 0 {
						return 0
					}
					optValA := v.getItemOpt(keyA)
					optValB := other.getItemOpt(keyB)
					return optValA.Compare(optValB, valueCmp)
				})
			}
		}
	}
	if rv != 0 {
		return rv
	}
	return cmp.Compare(v.Kind(), other.Kind())
}

func f64TotalCmp(left, right float64) int {
	leftInt := int64(math.Float64bits(left))
	rightInt := int64(math.Float64bits(right))
	leftInt ^= int64(uint64(leftInt>>63) >> 1)
	rightInt ^= int64(uint64(rightInt>>63) >> 1)
	return cmp.Compare(leftInt, rightInt)
}

func getItem(val, key Value) (Value, error) {
	if val.isUndefined() {
		return Value{}, NewError(UndefinedError, "")
	}
	return val.getItemOpt(key).UnwrapOr(Undefined), nil
}

func valueTryToGoBool(v Value) (bool, error) {
	if boolVal, ok := v.data.(boolValue); ok {
		return boolVal.B, nil
	}
	return false, unsupportedConversion(v.typ(), "bool")
}

func boolTryFromOptionValue(v option.Option[Value]) (bool, error) {
	if v.IsNone() {
		return false, NewError(MissingArgument, "")
	}
	return valueTryToGoBool(v.Unwrap())
}

func getAttr(val Value, key string) (Value, error) {
	switch v := val.data.(type) {
	case undefinedValue:
		return Value{}, NewError(UndefinedError, "")
	case mapValue:
		if v2, ok := v.Map.Get(keyRefFromString(key)); ok {
			return v2.clone(), nil
		}
	case dynamicValue:
		if obj, ok := v.Dy.(StructObject); ok {
			if optField := obj.GetField(key); optField.IsSome() {
				return optField.Unwrap(), nil
			}
		}
	}
	return Undefined, nil
}

func valueGetItemByIndex(val Value, idx uint) (Value, error) {
	return getItem(val, valueFromU64(uint64(idx)))
}

func getPath(val Value, path string) (Value, error) {
	rv := val.clone()
	for _, part := range strings.Split(path, ".") {
		num, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			rv, err = getAttr(rv, part)
			if err != nil {
				return Value{}, err
			}
		} else {
			rv, err = valueGetItemByIndex(rv, uint(num))
			if err != nil {
				return Value{}, err
			}
		}
	}
	return rv, nil
}

func (v undefinedValue) hash(h hash.Hash) { valueDataHash(v, h) }
func (v boolValue) hash(h hash.Hash)      { valueDataHash(v, h) }
func (v u64Value) hash(h hash.Hash)       { valueDataHash(v, h) }
func (v i64Value) hash(h hash.Hash)       { valueDataHash(v, h) }
func (v f64Value) hash(h hash.Hash)       { valueDataHash(v, h) }
func (v noneValue) hash(h hash.Hash)      { valueDataHash(v, h) }
func (v invalidValue) hash(h hash.Hash)   { valueDataHash(v, h) }
func (v u128Value) hash(h hash.Hash)      { valueDataHash(v, h) }
func (v i128Value) hash(h hash.Hash)      { valueDataHash(v, h) }
func (v stringValue) hash(h hash.Hash)    { valueDataHash(v, h) }
func (v bytesValue) hash(h hash.Hash)     { valueDataHash(v, h) }
func (v seqValue) hash(h hash.Hash)       { valueDataHash(v, h) }
func (v mapValue) hash(h hash.Hash)       { valueDataHash(v, h) }
func (v dynamicValue) hash(h hash.Hash)   { valueDataHash(v, h) }

func valueHash(val Value, h hash.Hash) { valueDataHash(val.data, h) }

func valueDataHash(val valueData, h hash.Hash) {
	switch v := val.(type) {
	case noneValue, undefinedValue:
		h.Write([]byte{0})
	case stringValue:
		io.WriteString(h, v.Str)
	case boolValue:
		b := byte(8)
		if v.B {
			b = byte(1)
		}
		h.Write([]byte{b})
	case invalidValue:
		io.WriteString(h, v.Detail)
	case bytesValue:
		h.Write(v.B)
	case seqValue:
		binary.Write(h, binary.BigEndian, uint64(len(v.Items)))
		for _, item := range v.Items {
			valueDataHash(item.data, h)
		}
	case mapValue:
		l := v.Map.Len()
		for i := uint(0); i < l; i++ {
			entry, _ := v.Map.EntryAt(i)
			keyRefHash(entry.Key, h)
			valueDataHash(entry.Value.data, h)
		}
	case dynamicValue:
		switch v.Dy.Kind() {
		case ObjectKindPlain:
			h.Write([]byte{0})
		case ObjectKindSeq:
			for iter, item := iteratorFromSeqObject(v.Dy.(SeqObject)), (Value{}); iter.Next().UnwrapTo(&item); {
				valueDataHash(item.data, h)
			}
		case ObjectKindStruct:
			structObj := v.Dy.(StructObject)
			var fields []string
			if !structObj.StaticFields().UnwrapTo(&fields) {
				fields = structObj.Fields()
			}
			for _, field := range fields {
				io.WriteString(h, field)
				structObj.GetField(field).Hash(h, valueHash)
			}
		}
	case u64Value, i64Value, f64Value, u128Value, i128Value:
		// TODO: fix for u128 and i128
		n, err := val.clone().tryToI64()
		if err != nil {
			val.asF64().Hash(h, f64Hash)
		} else {
			binary.Write(h, binary.BigEndian, n)
		}
	}
}

func f64Hash(f float64, h hash.Hash) {
	binary.Write(h, binary.BigEndian, math.Float64bits(f))
}

func (v undefinedValue) Equal(other any) bool { return valueDataEqualAny(v, other) }
func (v boolValue) Equal(other any) bool      { return valueDataEqualAny(v, other) }
func (v u64Value) Equal(other any) bool       { return valueDataEqualAny(v, other) }
func (v i64Value) Equal(other any) bool       { return valueDataEqualAny(v, other) }
func (v f64Value) Equal(other any) bool       { return valueDataEqualAny(v, other) }
func (v noneValue) Equal(other any) bool      { return valueDataEqualAny(v, other) }
func (v invalidValue) Equal(other any) bool   { return valueDataEqualAny(v, other) }
func (v u128Value) Equal(other any) bool      { return valueDataEqualAny(v, other) }
func (v i128Value) Equal(other any) bool      { return valueDataEqualAny(v, other) }
func (v stringValue) Equal(other any) bool    { return valueDataEqualAny(v, other) }
func (v bytesValue) Equal(other any) bool     { return valueDataEqualAny(v, other) }
func (v seqValue) Equal(other any) bool       { return valueDataEqualAny(v, other) }
func (v mapValue) Equal(other any) bool       { return valueDataEqualAny(v, other) }
func (v dynamicValue) Equal(other any) bool   { return valueDataEqualAny(v, other) }

func valueAsGoString(val Value) (string, bool) {
	strVal, ok := val.data.(stringValue)
	return strVal.Str, ok
}

func valueAsOptionString(val Value) option.Option[string] {
	strVal, ok := val.data.(stringValue)
	if ok {
		return option.Some(strVal.Str)
	}
	return option.None[string]()
}

func (Value) SupportsCustomVerb(verb rune) bool {
	return verb == rustfmt.DebugVerb || verb == rustfmt.DisplayVerb
}

func (v Value) Format(f fmt.State, verb rune) {
	switch verb {
	case rustfmt.DisplayVerb:
		switch d := v.data.(type) {
		case undefinedValue:
			// do nothing
		case boolValue:
			fmt.Fprintf(f, "%v", d.B)
		case u64Value:
			fmt.Fprintf(f, "%d", d.N)
		case i64Value:
			fmt.Fprintf(f, "%d", d.N)
		case f64Value:
			io.WriteString(f, v.data.String())
		case noneValue:
			io.WriteString(f, "none")
		case invalidValue:
			fmt.Fprintf(f, "<invalid value: %s>", d.Detail)
		case u128Value:
			io.WriteString(f, d.N.String())
		case i128Value:
			io.WriteString(f, d.N.String())
		case stringValue:
			io.WriteString(f, d.Str)
		case bytesValue:
			// https://go.dev/ref/spec#Conversions_to_and_from_a_string_type
			// Values outside the range of valid Unicode code points are converted to "\uFFFD"
			io.WriteString(f, string(d.B))
		case seqValue:
			rustfmt.NewDebugList(slicex.Map(d.Items, func(v Value) any { return v })).Format(f, rustfmt.DebugVerb)
		case mapValue:
			rustfmt.NewDebugMap(*d.Map).Format(f, rustfmt.DebugVerb)
		case dynamicValue:
			rustfmt.FormatValue(f, verb, d.Dy, "%v")
		default:
			panic("not implemented yet")
			// fmt.Fprintf(f, fmt.FormatString(f, verb), v.data)
		}
	case rustfmt.DebugVerb:
		switch d := v.data.(type) {
		case undefinedValue:
			io.WriteString(f, "undefined")
		case boolValue:
			fmt.Fprintf(f, "%v", d.B)
		case u64Value:
			fmt.Fprintf(f, "%d", d.N)
		case i64Value:
			fmt.Fprintf(f, "%d", d.N)
		case f64Value:
			io.WriteString(f, v.data.String())
		case noneValue:
			io.WriteString(f, "none")
		case invalidValue:
			fmt.Fprintf(f, "<invalid value: %s>", d.Detail)
		case u128Value:
			io.WriteString(f, d.N.String())
		case i128Value:
			io.WriteString(f, d.N.String())
		case stringValue:
			fmt.Fprintf(f, "%q", d.Str)
		case bytesValue:
			rustfmt.NewDebugList(slicex.Map(d.B, func(v byte) any { return fmt.Sprintf("%q", v) })).Format(f, verb)
		case seqValue:
			rustfmt.NewDebugList(slicex.Map(d.Items, func(v Value) any { return v })).Format(f, verb)
		case mapValue:
			rustfmt.NewDebugMap(*d.Map).Format(f, verb)
		case dynamicValue:
			rustfmt.FormatValue(f, verb, d.Dy, "%v")
		default:
			fmt.Fprintf(f, fmt.FormatString(f, verb), d)
		}
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods Value
		type Value hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), Value(v))
	}
}
