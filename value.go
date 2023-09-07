package mjingo

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type Value interface {
	fmt.Stringer
	debugString() string

	typ() valueType
	kind() valueKind
	isUndefined() bool
	isNone() bool
	isSafe() bool
	isTrue() bool
	getAttrFast(key string) option.Option[Value]
	getItemOpt(key Value) option.Option[Value]
	asStr() option.Option[string]
	tryToI128() (big.Int, error)
	tryToI64() (int64, error)
	tryToUint() (uint, error)
	asF64() option.Option[float64]
	asSeq() option.Option[seqObject]
	clone() Value
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

// Undefined is the undefined value.
//
// This constant variable exists because the undefined type does not exist in Go
// and this is the only way to construct it.
var Undefined Value

var none = noneValue{}

func init() {
	Undefined = undefinedValue{}
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

func (k valueKind) String() string {
	switch k {
	case valueKindUndefined:
		return "undefined"
	case valueKindBool:
		return "bool"
	case valueKindNumber:
		return "number"
	case valueKindNone:
		return "none"
	case valueKindString:
		return "string"
	case valueKindBytes:
		return "bytes"
	case valueKindSeq:
		return "seq"
	case valueKindMap:
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
type u128Value struct{ N big.Int }
type i128Value struct{ N big.Int }
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
	Dy object
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

var _ = Value(undefinedValue{})
var _ = Value(boolValue{})
var _ = Value(u64Value{})
var _ = Value(i64Value{})
var _ = Value(f64Value{})
var _ = Value(noneValue{})
var _ = Value(invalidValue{})
var _ = Value(u128Value{})
var _ = Value(i128Value{})
var _ = Value(stringValue{})
var _ = Value(bytesValue{})
var _ = Value(seqValue{})
var _ = Value(mapValue{})
var _ = Value(dynamicValue{})

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
		b.WriteString(item.debugString())
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
		b.WriteString(e.Key.AsValue().debugString())
		b.WriteString(": ")
		b.WriteString(e.Value.debugString())
	}
	b.WriteString("}")
	return b.String()
}
func (v dynamicValue) String() string { return fmt.Sprintf("%s", v.Dy) }

func (v undefinedValue) debugString() string { return "Undefined" }
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
func (v noneValue) debugString() string    { return "None" }
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
		b.WriteString(item.debugString())
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
		b.WriteString(e.Key.AsValue().debugString())
		b.WriteString(": ")
		b.WriteString(e.Value.debugString())
	}
	b.WriteString("}")
	return b.String()
}
func (v dynamicValue) debugString() string { return fmt.Sprintf("%s", v.Dy) }

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
func (invalidValue) kind() valueKind {
	// XXX: invalid values report themselves as maps which is a lie
	return valueKindMap
}
func (u128Value) kind() valueKind   { return valueKindNumber }
func (i128Value) kind() valueKind   { return valueKindNumber }
func (stringValue) kind() valueKind { return valueKindString }
func (bytesValue) kind() valueKind  { return valueKindBytes }
func (seqValue) kind() valueKind    { return valueKindSeq }
func (mapValue) kind() valueKind    { return valueKindMap }
func (v dynamicValue) kind() valueKind {
	switch v.Dy.Kind() {
	case objectKindPlain:
		// XXX: basic objects should probably not report as map
		return valueKindMap
	case objectKindSeq:
		return valueKindSeq
	case objectKindStruct:
		return valueKindMap
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
	var zero big.Int
	return v.N.Cmp(&zero) != 0
}
func (v i128Value) isTrue() bool {
	var zero big.Int
	return v.N.Cmp(&zero) != 0
}
func (v stringValue) isTrue() bool { return len(v.Str) != 0 }
func (v bytesValue) isTrue() bool  { return len(v.B) != 0 }
func (v seqValue) isTrue() bool    { return len(v.Items) != 0 }
func (v mapValue) isTrue() bool    { return v.Map.Len() != 0 }
func (v dynamicValue) isTrue() bool {
	switch v.Dy.Kind() {
	case objectKindPlain:
		return true
	case objectKindSeq:
		return v.Dy.(seqObject).ItemCount() != 0
	case objectKindStruct:
		return fieldCount(v.Dy.(structObject)) != 0
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
	if s, ok := v.Dy.(structObject); ok {
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
func (stringValue) getItemOpt(_ Value) option.Option[Value]    { return option.None[Value]() }
func (bytesValue) getItemOpt(_ Value) option.Option[Value]     { return option.None[Value]() }
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
	case objectKindPlain:
		return option.None[Value]()
	case objectKindSeq:
		return getItemOptFromSeq(v.Dy.(seqObject), key)
	case objectKindStruct:
		if optKey := key.asStr(); optKey.IsSome() {
			return v.Dy.(structObject).GetField(optKey.Unwrap())
		}
		return option.None[Value]()
	default:
		panic("unreachable")
	}
}

func getItemOptFromSeq(seq seqObject, key Value) option.Option[Value] {
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

func (undefinedValue) asStr() option.Option[string] { return option.None[string]() }
func (boolValue) asStr() option.Option[string]      { return option.None[string]() }
func (u64Value) asStr() option.Option[string]       { return option.None[string]() }
func (i64Value) asStr() option.Option[string]       { return option.None[string]() }
func (f64Value) asStr() option.Option[string]       { return option.None[string]() }
func (noneValue) asStr() option.Option[string]      { return option.None[string]() }
func (invalidValue) asStr() option.Option[string]   { return option.None[string]() }
func (u128Value) asStr() option.Option[string]      { return option.None[string]() }
func (i128Value) asStr() option.Option[string]      { return option.None[string]() }
func (v stringValue) asStr() option.Option[string]  { return option.Some(v.Str) }
func (bytesValue) asStr() option.Option[string]     { return option.None[string]() }
func (seqValue) asStr() option.Option[string]       { return option.None[string]() }
func (v mapValue) asStr() option.Option[string]     { return option.None[string]() }
func (dynamicValue) asStr() option.Option[string]   { return option.None[string]() }

func (v undefinedValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v boolValue) tryToI128() (big.Int, error) {
	var n big.Int
	if v.B {
		n.SetUint64(1)
	}
	return n, nil
}
func (v u64Value) tryToI128() (big.Int, error) {
	var n big.Int
	n.SetUint64(v.N)
	return n, nil
}
func (v i64Value) tryToI128() (big.Int, error) {
	var n big.Int
	n.SetInt64(v.N)
	return n, nil
}
func (v f64Value) tryToI128() (big.Int, error) {
	if float64(int64(v.F)) == v.F {
		var n big.Int
		n.SetInt64(int64(v.F))
		return n, nil
	}
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v noneValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v invalidValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v u128Value) tryToI128() (big.Int, error) {
	if v.N.Cmp(getI128Max()) > 0 {
		return big.Int{}, unsupportedConversion(v.typ(), "i128")
	}
	var n big.Int
	n.Set(&v.N)
	return n, nil
}
func (v i128Value) tryToI128() (big.Int, error) {
	var n big.Int
	n.Set(&v.N)
	return n, nil
}
func (v stringValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v bytesValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v seqValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v mapValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v dynamicValue) tryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}

func (v undefinedValue) tryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v boolValue) tryToI64() (int64, error) {
	if v.B {
		return 1, nil
	}
	return 0, nil
}
func (v u64Value) tryToI64() (int64, error) { return int64(v.N), nil }
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
	f, _ := v.N.Float64()
	return option.Some(f)
}
func (v i128Value) asF64() option.Option[float64] {
	f, _ := v.N.Float64()
	return option.Some(f)
}
func (stringValue) asF64() option.Option[float64]  { return option.None[float64]() }
func (bytesValue) asF64() option.Option[float64]   { return option.None[float64]() }
func (seqValue) asF64() option.Option[float64]     { return option.None[float64]() }
func (mapValue) asF64() option.Option[float64]     { return option.None[float64]() }
func (dynamicValue) asF64() option.Option[float64] { return option.None[float64]() }

func (undefinedValue) asSeq() option.Option[seqObject] { return option.None[seqObject]() }
func (boolValue) asSeq() option.Option[seqObject]      { return option.None[seqObject]() }
func (u64Value) asSeq() option.Option[seqObject]       { return option.None[seqObject]() }
func (i64Value) asSeq() option.Option[seqObject]       { return option.None[seqObject]() }
func (f64Value) asSeq() option.Option[seqObject]       { return option.None[seqObject]() }
func (noneValue) asSeq() option.Option[seqObject]      { return option.None[seqObject]() }
func (invalidValue) asSeq() option.Option[seqObject]   { return option.None[seqObject]() }
func (u128Value) asSeq() option.Option[seqObject]      { return option.None[seqObject]() }
func (i128Value) asSeq() option.Option[seqObject]      { return option.None[seqObject]() }
func (stringValue) asSeq() option.Option[seqObject]    { return option.None[seqObject]() }
func (bytesValue) asSeq() option.Option[seqObject]     { return option.None[seqObject]() }
func (v seqValue) asSeq() option.Option[seqObject] {
	return option.Some(newSliceSeqObject(v.Items))
}
func (mapValue) asSeq() option.Option[seqObject] { return option.None[seqObject]() }
func (v dynamicValue) asSeq() option.Option[seqObject] {
	if seq, ok := v.Dy.(seqObject); ok {
		return option.Some(seq)
	}
	return option.None[seqObject]()
}

func (v undefinedValue) clone() Value { return v }
func (v boolValue) clone() Value      { return v }
func (v u64Value) clone() Value       { return v }
func (v i64Value) clone() Value       { return v }
func (v f64Value) clone() Value       { return v }
func (v noneValue) clone() Value      { return v }
func (v invalidValue) clone() Value   { return v }
func (v u128Value) clone() Value {
	c := v
	c.N.Set(&v.N)
	return c
}
func (v i128Value) clone() Value {
	c := v
	c.N.Set(&v.N)
	return c
}
func (v stringValue) clone() Value { return v }
func (v bytesValue) clone() Value {
	b := make([]byte, len(v.B))
	copy(b, v.B)
	return bytesValue{B: b}
}
func (v seqValue) clone() Value {
	items := make([]Value, len(v.Items))
	for i, item := range v.Items {
		// Is shallow copy OK?
		items[i] = item
	}
	return seqValue{Items: items}
}
func (v mapValue) clone() Value {
	m := v.Map.Clone()
	return mapValue{Map: m, Type: v.Type}
}
func (v dynamicValue) clone() Value {
	// TODO: implement real clone
	return v
}

func (undefinedValue) tryIter() (iterator, error) {
	return iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v boolValue) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v u64Value) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v i64Value) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v f64Value) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (noneValue) tryIter() (iterator, error) {
	return iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v invalidValue) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v u128Value) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v i128Value) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v stringValue) tryIter() (iterator, error) {
	return iterator{iterState: &charsValueIteratorState{s: v.Str}, len: uint(utf8.RuneCountInString(v.Str))}, nil
}
func (v bytesValue) tryIter() (iterator, error) {
	return iterator{}, newError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v seqValue) tryIter() (iterator, error) {
	return iterator{iterState: &seqValueIteratorState{items: v.Items}, len: uint(len(v.Items))}, nil
}
func (v mapValue) tryIter() (iterator, error) {
	return iterator{iterState: &mapValueIteratorState{keys: v.Map.Keys()}, len: uint(len(v.Map.Keys()))}, nil
}
func (v dynamicValue) tryIter() (iterator, error) {
	switch v.Dy.Kind() {
	case objectKindPlain:
		return iterator{iterState: &emptyValueIteratorState{}}, nil
	case objectKindSeq:
		seqObj := v.Dy.(seqObject)
		return iterator{iterState: &dynSeqValueIteratorState{obj: seqObj}, len: seqObj.ItemCount()}, nil
	case objectKindStruct:
		obj := v.Dy.(structObject)
		if optFields := obj.StaticFields(); optFields.IsSome() {
			return iterator{iterState: &stringsValueIteratorState{items: optFields.Unwrap()}}, nil
		}
		return iterator{iterState: &stringsValueIteratorState{items: obj.Fields()}}, nil
	default:
		panic("unreachable")
	}
}

func unsupportedConversion(kind valueType, target string) error {
	return newError(InvalidOperation,
		fmt.Sprintf("cannot convert %s to %s", kind, target))
}

func i32TryFromValue(val Value) (int32, error) {
	n, err := val.tryToI64()
	if err != nil || n < math.MinInt32 || n > math.MaxInt32 {
		return 0, unsupportedConversion(val.typ(), "i32")
	}
	return int32(n), nil
}

func u32TryFromValue(val Value) (uint32, error) {
	n, err := val.tryToI64()
	if err != nil || n < 0 || n > math.MaxUint32 {
		return 0, unsupportedConversion(val.typ(), "u32")
	}
	return uint32(n), nil
}

type iterator struct {
	iterState valueIteratorState
	len       uint
}

func iteratorFromSeqObject(s seqObject) *iterator {
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
	var item Value
	for i.Next().UnwrapTo(&item) {
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
	var item Value
	for i.Next().UnwrapTo(&item) {
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
	obj seqObject
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
			clear(s.states[len(s.states)-1:])
			s.states = s.states[:len(s.states)-1]
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
		return option.Some[Value](stringValue{Str: string(r)})
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
		return option.Some[Value](stringValue{Str: item})
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
	case objectKindPlain:
		return option.None[uint]()
	case objectKindSeq:
		return option.Some(v.Dy.(seqObject).ItemCount())
	case objectKindStruct:
		return option.Some(fieldCount(v.Dy.(structObject)))
	default:
		panic("unreachable")
	}
}

func valueEqual(v, other Value) bool {
	switch {
	case v.kind() == valueKindNone && other.kind() == valueKindNone:
		return true
	case v.kind() == valueKindUndefined && other.kind() == valueKindUndefined:
		return true
	case v.kind() == valueKindString && other.kind() == valueKindString:
		a := v.(stringValue).Str
		b := other.(stringValue).Str
		return a == b
	case v.kind() == valueKindBytes && other.kind() == valueKindBytes:
		a := v.(bytesValue).B
		b := other.(bytesValue).B
		return bytes.Equal(a, b)
	default:
		switch c := coerce(v, other).(type) {
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
			} else if v.kind() == valueKindMap && other.kind() == valueKindMap {
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

func valueEqualAny(v Value, other any) bool {
	if v == nil && other == nil {
		return true
	}
	if o, ok := other.(Value); ok {
		return valueEqual(v, o)
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
	case v.kind() == valueKindNone && other.kind() == valueKindNone:
		rv = 0
	case v.kind() == valueKindUndefined && other.kind() == valueKindUndefined:
		rv = 0
	case v.kind() == valueKindString && other.kind() == valueKindString:
		a := v.(stringValue).Str
		b := other.(stringValue).Str
		rv = strings.Compare(a, b)
	case v.kind() == valueKindBytes && other.kind() == valueKindBytes:
		a := v.(bytesValue).B
		b := other.(bytesValue).B
		rv = bytes.Compare(a, b)
	default:
		switch c := coerce(v, other).(type) {
		case f64CoerceResult:
			return f64TotalCmp(c.lhs, c.rhs)
		case i128CoerceResult:
			return c.lhs.Cmp(&c.rhs)
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
				return iterA.CompareBy(&iterB, valueCmp)
			} else if v.kind() == valueKindMap && other.kind() == valueKindMap {
				iterA, err := v.tryIter()
				if err != nil {
					break outer
				}
				iterB, err := other.tryIter()
				if err != nil {
					break outer
				}
				return iterA.CompareBy(&iterB, func(keyA, keyB Value) int {
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
	return cmp.Compare(v.kind(), other.kind())
}

func f64TotalCmp(left, right float64) int {
	leftInt := int64(math.Float64bits(left))
	rightInt := int64(math.Float64bits(left))
	leftInt ^= int64(uint64(leftInt>>63) >> 1)
	rightInt ^= int64(uint64(rightInt>>63) >> 1)
	return cmp.Compare(leftInt, rightInt)
}

func getItem(val, key Value) (Value, error) {
	if val.isUndefined() {
		return nil, newError(UndefinedError, "")
	}
	return val.getItemOpt(key).UnwrapOr(Undefined), nil
}

func boolTryFromValue(v Value) (bool, error) {
	if boolVal, ok := v.(boolValue); ok {
		return boolVal.B, nil
	}
	return false, unsupportedConversion(v.typ(), "bool")
}

func boolTryFromOptionValue(v option.Option[Value]) (bool, error) {
	if v.IsNone() {
		return false, newError(MissingArgument, "")
	}
	return boolTryFromValue(v.Unwrap())
}

func getAttr(val Value, key string) (Value, error) {
	switch v := val.(type) {
	case undefinedValue:
		return nil, newError(UndefinedError, "")
	case mapValue:
		if v2, ok := v.Map.Get(keyRefFromString(key)); ok {
			return v2.clone(), nil
		}
	case dynamicValue:
		if obj, ok := v.Dy.(structObject); ok {
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
			rv, err = getAttr(val, part)
			if err != nil {
				return nil, err
			}
		} else {
			rv, err = valueGetItemByIndex(val, uint(num))
			if err != nil {
				return nil, err
			}
		}
	}
	return rv, nil
}

func (v undefinedValue) hash(h hash.Hash) { valueHash(v, h) }
func (v boolValue) hash(h hash.Hash)      { valueHash(v, h) }
func (v u64Value) hash(h hash.Hash)       { valueHash(v, h) }
func (v i64Value) hash(h hash.Hash)       { valueHash(v, h) }
func (v f64Value) hash(h hash.Hash)       { valueHash(v, h) }
func (v noneValue) hash(h hash.Hash)      { valueHash(v, h) }
func (v invalidValue) hash(h hash.Hash)   { valueHash(v, h) }
func (v u128Value) hash(h hash.Hash)      { valueHash(v, h) }
func (v i128Value) hash(h hash.Hash)      { valueHash(v, h) }
func (v stringValue) hash(h hash.Hash)    { valueHash(v, h) }
func (v bytesValue) hash(h hash.Hash)     { valueHash(v, h) }
func (v seqValue) hash(h hash.Hash)       { valueHash(v, h) }
func (v mapValue) hash(h hash.Hash)       { valueHash(v, h) }
func (v dynamicValue) hash(h hash.Hash)   { valueHash(v, h) }

func valueHash(val Value, h hash.Hash) {
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
			valueHash(item, h)
		}
	case mapValue:
		l := v.Map.Len()
		for i := uint(0); i < l; i++ {
			entry, _ := v.Map.EntryAt(i)
			keyRefHash(entry.Key, h)
			valueHash(entry.Value, h)
		}
	case dynamicValue:
		switch v.Dy.Kind() {
		case objectKindPlain:
			h.Write([]byte{0})
		case objectKindSeq:
			var item Value
			for iter := iteratorFromSeqObject(v.Dy.(seqObject)); iter.Next().UnwrapTo(&item); {
				valueHash(item, h)
			}
		case objectKindStruct:
			structObj := v.Dy.(structObject)
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

func (v undefinedValue) Equal(other any) bool { return valueEqualAny(v, other) }
func (v boolValue) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v u64Value) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v i64Value) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v f64Value) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v noneValue) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v invalidValue) Equal(other any) bool   { return valueEqualAny(v, other) }
func (v u128Value) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v i128Value) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v stringValue) Equal(other any) bool    { return valueEqualAny(v, other) }
func (v bytesValue) Equal(other any) bool     { return valueEqualAny(v, other) }
func (v seqValue) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v mapValue) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v dynamicValue) Equal(other any) bool   { return valueEqualAny(v, other) }
