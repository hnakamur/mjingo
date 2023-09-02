package value

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

	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/datast/option"
)

type Value interface {
	String() string
	DebugString() string

	typ() valueType
	Kind() ValueKind
	IsUndefined() bool
	IsNone() bool
	IsSafe() bool
	IsTrue() bool
	GetAttrFast(key string) option.Option[Value]
	GetItemOpt(key Value) option.Option[Value]
	AsStr() option.Option[string]
	TryToI128() (big.Int, error)
	TryToI64() (int64, error)
	TryToUint() (uint, error)
	AsF64() option.Option[float64]
	AsSeq() option.Option[SeqObject]
	Clone() Value
	TryIter() (Iterator, error)
	Len() option.Option[uint]
	// Call(state *State, args []Value) (Value, error)
	// CallMethod(state *State, name string, args []Value) (Value, error)
	Hash(h hash.Hash)
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

type ValueKind int

const (
	// The value is undefined
	ValueKindUndefined ValueKind = iota + 1
	// The value is the none singleton ([`()`])
	ValueKindNone
	// The value is a [`bool`]
	ValueKindBool
	// The value is a number of a supported type.
	ValueKindNumber
	// The value is a string.
	ValueKindString
	// The value is a byte array.
	ValueKindBytes
	// The value is an array of other values.
	ValueKindSeq
	// The value is a key/value mapping.
	ValueKindMap
)

var Undefined = UndefinedValue{}
var None = NoneValue{}

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
		return "seq"
	case ValueKindMap:
		return "map"
	default:
		panic(fmt.Sprintf("invalid valueKind: %d", k))
	}
}

type UndefinedValue struct{}
type BoolValue struct{ B bool }
type U64Value struct{ N uint64 }
type I64Value struct{ N int64 }
type F64Value struct{ F float64 }
type NoneValue struct{}
type InvalidValue struct{ Detail string }
type U128Value struct{ N big.Int }
type I128Value struct{ N big.Int }
type StringValue struct {
	Str  string
	Type StringType
}
type BytesValue struct{ B []byte }
type SeqValue struct{ Items []Value }
type MapValue struct {
	Map  *ValueMap
	Type MapType
}
type DynamicValue struct {
	Dy Object
}

// / The type of map
type MapType uint

const (
	// A regular map
	MapTypeNormal MapType = iota + 1
	// A map representing keyword arguments
	MapTypeKwargs
)

type StringType uint

const (
	StringTypeNormal StringType = iota
	StringTypeSafe
)

var _ = Value(UndefinedValue{})
var _ = Value(BoolValue{})
var _ = Value(U64Value{})
var _ = Value(I64Value{})
var _ = Value(F64Value{})
var _ = Value(NoneValue{})
var _ = Value(InvalidValue{})
var _ = Value(U128Value{})
var _ = Value(I128Value{})
var _ = Value(StringValue{})
var _ = Value(BytesValue{})
var _ = Value(SeqValue{})
var _ = Value(MapValue{})
var _ = Value(DynamicValue{})

func (v UndefinedValue) String() string { return "" }
func (v BoolValue) String() string      { return strconv.FormatBool(v.B) }
func (v U64Value) String() string       { return strconv.FormatUint(v.N, 10) }
func (v I64Value) String() string       { return strconv.FormatInt(v.N, 10) }
func (v F64Value) String() string {
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
func (v NoneValue) String() string    { return "none" }
func (v InvalidValue) String() string { return fmt.Sprintf("<invalid value: %s>", v.Detail) }
func (v U128Value) String() string    { return v.N.String() }
func (v I128Value) String() string    { return v.N.String() }
func (v StringValue) String() string  { return v.Str }
func (v BytesValue) String() string   { return string(v.B) } // TODO: equivalent impl as String::from_utf8_lossy
func (v SeqValue) String() string {
	var b strings.Builder
	b.WriteString("[")
	for i, item := range v.Items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item.DebugString()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("]")
	return b.String()
}
func (v MapValue) String() string {
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
		b.WriteString(e.Value.DebugString()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("}")
	return b.String()
}
func (v DynamicValue) String() string { return fmt.Sprintf("%s", v.Dy) }

func (v UndefinedValue) DebugString() string { return "Undefined" }
func (v BoolValue) DebugString() string      { return strconv.FormatBool(v.B) }
func (v U64Value) DebugString() string       { return strconv.FormatUint(v.N, 10) }
func (v I64Value) DebugString() string       { return strconv.FormatInt(v.N, 10) }
func (v F64Value) DebugString() string {
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
func (v NoneValue) DebugString() string    { return "None" }
func (v InvalidValue) DebugString() string { return fmt.Sprintf("<invalid value: %s>", v.Detail) }
func (v U128Value) DebugString() string    { return v.N.String() }
func (v I128Value) DebugString() string    { return v.N.String() }
func (v StringValue) DebugString() string  { return fmt.Sprintf("%q", v.Str) } // TODO: equivalent impl with Rust's std::fmt::Debug
func (v BytesValue) DebugString() string   { return string(v.B) }              // TODO: equivalent impl as String::from_utf8_lossy
func (v SeqValue) DebugString() string {
	var b strings.Builder
	b.WriteString("[")
	for i, item := range v.Items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(item.DebugString()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("]")
	return b.String()
}
func (v MapValue) DebugString() string {
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
		b.WriteString(e.Value.DebugString()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("}")
	return b.String()
}
func (v DynamicValue) DebugString() string { return fmt.Sprintf("%s", v.Dy) }

func (UndefinedValue) typ() valueType { return valueTypeUndefined }
func (BoolValue) typ() valueType      { return valueTypeBool }
func (U64Value) typ() valueType       { return valueTypeU64 }
func (I64Value) typ() valueType       { return valueTypeI64 }
func (F64Value) typ() valueType       { return valueTypeF64 }
func (NoneValue) typ() valueType      { return valueTypeNone }
func (InvalidValue) typ() valueType   { return valueTypeInvalid }
func (U128Value) typ() valueType      { return valueTypeU128 }
func (I128Value) typ() valueType      { return valueTypeI128 }
func (StringValue) typ() valueType    { return valueTypeString }
func (BytesValue) typ() valueType     { return valueTypeBytes }
func (SeqValue) typ() valueType       { return valueTypeSeq }
func (MapValue) typ() valueType       { return valueTypeMap }
func (DynamicValue) typ() valueType   { return valueTypeDynamic }

func (UndefinedValue) Kind() ValueKind { return ValueKindUndefined }
func (BoolValue) Kind() ValueKind      { return ValueKindBool }
func (U64Value) Kind() ValueKind       { return ValueKindNumber }
func (I64Value) Kind() ValueKind       { return ValueKindNumber }
func (F64Value) Kind() ValueKind       { return ValueKindNumber }
func (NoneValue) Kind() ValueKind      { return ValueKindNone }
func (InvalidValue) Kind() ValueKind {
	// XXX: invalid values report themselves as maps which is a lie
	return ValueKindMap
}
func (U128Value) Kind() ValueKind   { return ValueKindNumber }
func (I128Value) Kind() ValueKind   { return ValueKindNumber }
func (StringValue) Kind() ValueKind { return ValueKindString }
func (BytesValue) Kind() ValueKind  { return ValueKindBytes }
func (SeqValue) Kind() ValueKind    { return ValueKindSeq }
func (MapValue) Kind() ValueKind    { return ValueKindMap }
func (v DynamicValue) Kind() ValueKind {
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

func (UndefinedValue) IsUndefined() bool { return true }
func (BoolValue) IsUndefined() bool      { return false }
func (U64Value) IsUndefined() bool       { return false }
func (I64Value) IsUndefined() bool       { return false }
func (F64Value) IsUndefined() bool       { return false }
func (NoneValue) IsUndefined() bool      { return false }
func (InvalidValue) IsUndefined() bool   { return false }
func (U128Value) IsUndefined() bool      { return false }
func (I128Value) IsUndefined() bool      { return false }
func (StringValue) IsUndefined() bool    { return false }
func (BytesValue) IsUndefined() bool     { return false }
func (SeqValue) IsUndefined() bool       { return false }
func (MapValue) IsUndefined() bool       { return false }
func (DynamicValue) IsUndefined() bool   { return false }

func (UndefinedValue) IsNone() bool { return false }
func (BoolValue) IsNone() bool      { return false }
func (U64Value) IsNone() bool       { return false }
func (I64Value) IsNone() bool       { return false }
func (F64Value) IsNone() bool       { return false }
func (NoneValue) IsNone() bool      { return true }
func (InvalidValue) IsNone() bool   { return false }
func (U128Value) IsNone() bool      { return false }
func (I128Value) IsNone() bool      { return false }
func (StringValue) IsNone() bool    { return false }
func (BytesValue) IsNone() bool     { return false }
func (SeqValue) IsNone() bool       { return false }
func (MapValue) IsNone() bool       { return false }
func (DynamicValue) IsNone() bool   { return false }

func (UndefinedValue) IsSafe() bool { return false }
func (BoolValue) IsSafe() bool      { return false }
func (U64Value) IsSafe() bool       { return false }
func (I64Value) IsSafe() bool       { return false }
func (F64Value) IsSafe() bool       { return false }
func (NoneValue) IsSafe() bool      { return false }
func (InvalidValue) IsSafe() bool   { return false }
func (U128Value) IsSafe() bool      { return false }
func (I128Value) IsSafe() bool      { return false }
func (v StringValue) IsSafe() bool  { return v.Type == StringTypeSafe }
func (BytesValue) IsSafe() bool     { return false }
func (SeqValue) IsSafe() bool       { return false }
func (MapValue) IsSafe() bool       { return false }
func (DynamicValue) IsSafe() bool   { return false }

func (UndefinedValue) IsTrue() bool { return false }
func (v BoolValue) IsTrue() bool    { return v.B }
func (v U64Value) IsTrue() bool     { return v.N != 0 }
func (v I64Value) IsTrue() bool     { return v.N != 0 }
func (v F64Value) IsTrue() bool     { return v.F != 0.0 }
func (NoneValue) IsTrue() bool      { return false }
func (InvalidValue) IsTrue() bool   { return false }
func (v U128Value) IsTrue() bool {
	var zero big.Int
	return v.N.Cmp(&zero) != 0
}
func (v I128Value) IsTrue() bool {
	var zero big.Int
	return v.N.Cmp(&zero) != 0
}
func (v StringValue) IsTrue() bool { return len(v.Str) != 0 }
func (v BytesValue) IsTrue() bool  { return len(v.B) != 0 }
func (v SeqValue) IsTrue() bool    { return len(v.Items) != 0 }
func (v MapValue) IsTrue() bool    { return v.Map.Len() != 0 }
func (v DynamicValue) IsTrue() bool {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return true
	case ObjectKindSeq:
		return v.Dy.(SeqObject).ItemCount() != 0
	case ObjectKindStruct:
		return FieldCount(v.Dy.(StructObject)) != 0
	default:
		panic("unreachable")
	}
}

func (UndefinedValue) GetAttrFast(_ string) option.Option[Value] { return option.None[Value]() }
func (BoolValue) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (U64Value) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (I64Value) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (F64Value) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (NoneValue) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (InvalidValue) GetAttrFast(_ string) option.Option[Value]   { return option.None[Value]() }
func (U128Value) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (I128Value) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (StringValue) GetAttrFast(_ string) option.Option[Value]    { return option.None[Value]() }
func (BytesValue) GetAttrFast(_ string) option.Option[Value]     { return option.None[Value]() }
func (SeqValue) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (v MapValue) GetAttrFast(key string) option.Option[Value] {
	if val, ok := v.Map.Get(KeyRefFromString(key)); ok {
		return option.Some(val)
	}
	return option.None[Value]()
}
func (v DynamicValue) GetAttrFast(key string) option.Option[Value] {
	if s, ok := v.Dy.(StructObject); ok {
		return s.GetField(key)
	}
	return option.None[Value]()
}

func (UndefinedValue) GetItemOpt(_ Value) option.Option[Value] { return option.None[Value]() }
func (BoolValue) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (U64Value) GetItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (I64Value) GetItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (F64Value) GetItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (NoneValue) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (InvalidValue) GetItemOpt(_ Value) option.Option[Value]   { return option.None[Value]() }
func (U128Value) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (I128Value) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (StringValue) GetItemOpt(_ Value) option.Option[Value]    { return option.None[Value]() }
func (BytesValue) GetItemOpt(_ Value) option.Option[Value]     { return option.None[Value]() }
func (v SeqValue) GetItemOpt(key Value) option.Option[Value] {
	return getItemOptFromSeq(NewSliceSeqObject(v.Items), key)
}
func (v MapValue) GetItemOpt(key Value) option.Option[Value] {
	if v, ok := v.Map.Get(KeyRefFromValue(key)); ok {
		return option.Some(v)
	}
	return option.None[Value]()
}
func (v DynamicValue) GetItemOpt(key Value) option.Option[Value] {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return option.None[Value]()
	case ObjectKindSeq:
		return getItemOptFromSeq(v.Dy.(SeqObject), key)
	case ObjectKindStruct:
		if optKey := key.AsStr(); optKey.IsSome() {
			return v.Dy.(StructObject).GetField(optKey.Unwrap())
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

func (UndefinedValue) AsStr() option.Option[string] { return option.None[string]() }
func (BoolValue) AsStr() option.Option[string]      { return option.None[string]() }
func (U64Value) AsStr() option.Option[string]       { return option.None[string]() }
func (I64Value) AsStr() option.Option[string]       { return option.None[string]() }
func (F64Value) AsStr() option.Option[string]       { return option.None[string]() }
func (NoneValue) AsStr() option.Option[string]      { return option.None[string]() }
func (InvalidValue) AsStr() option.Option[string]   { return option.None[string]() }
func (U128Value) AsStr() option.Option[string]      { return option.None[string]() }
func (I128Value) AsStr() option.Option[string]      { return option.None[string]() }
func (v StringValue) AsStr() option.Option[string]  { return option.Some(v.Str) }
func (BytesValue) AsStr() option.Option[string]     { return option.None[string]() }
func (SeqValue) AsStr() option.Option[string]       { return option.None[string]() }
func (v MapValue) AsStr() option.Option[string]     { return option.None[string]() }
func (DynamicValue) AsStr() option.Option[string]   { return option.None[string]() }

func (v UndefinedValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v BoolValue) TryToI128() (big.Int, error) {
	var n big.Int
	if v.B {
		n.SetUint64(1)
	}
	return n, nil
}
func (v U64Value) TryToI128() (big.Int, error) {
	var n big.Int
	n.SetUint64(v.N)
	return n, nil
}
func (v I64Value) TryToI128() (big.Int, error) {
	var n big.Int
	n.SetInt64(v.N)
	return n, nil
}
func (v F64Value) TryToI128() (big.Int, error) {
	if float64(int64(v.F)) == v.F {
		var n big.Int
		n.SetInt64(int64(v.F))
		return n, nil
	}
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v NoneValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v InvalidValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v U128Value) TryToI128() (big.Int, error) {
	if v.N.Cmp(getI128Max()) > 0 {
		return big.Int{}, unsupportedConversion(v.typ(), "i128")
	}
	var n big.Int
	n.Set(&v.N)
	return n, nil
}
func (v I128Value) TryToI128() (big.Int, error) {
	var n big.Int
	n.Set(&v.N)
	return n, nil
}
func (v StringValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v BytesValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v SeqValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v MapValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}
func (v DynamicValue) TryToI128() (big.Int, error) {
	return big.Int{}, unsupportedConversion(v.typ(), "i128")
}

func (v UndefinedValue) TryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v BoolValue) TryToI64() (int64, error) {
	if v.B {
		return 1, nil
	}
	return 0, nil
}
func (v U64Value) TryToI64() (int64, error) { return int64(v.N), nil }
func (v I64Value) TryToI64() (int64, error) { return v.N, nil }
func (v F64Value) TryToI64() (int64, error) {
	if float64(int64(v.F)) == v.F {
		return int64(v.F), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v NoneValue) TryToI64() (int64, error)    { return 0, unsupportedConversion(v.typ(), "i64") }
func (v InvalidValue) TryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v U128Value) TryToI64() (int64, error) {
	if v.N.IsInt64() {
		return v.N.Int64(), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v I128Value) TryToI64() (int64, error) {
	if v.N.IsInt64() {
		return v.N.Int64(), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v StringValue) TryToI64() (int64, error)  { return 0, unsupportedConversion(v.typ(), "i64") }
func (v BytesValue) TryToI64() (int64, error)   { return 0, unsupportedConversion(v.typ(), "i64") }
func (v SeqValue) TryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v MapValue) TryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v DynamicValue) TryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }

func (v UndefinedValue) TryToUint() (uint, error) { return 0, unsupportedConversion(v.typ(), "uint") }
func (v BoolValue) TryToUint() (uint, error) {
	if v.B {
		return 1, nil
	}
	return 0, nil
}
func (v U64Value) TryToUint() (uint, error) {
	if v.N > math.MaxUint {
		return 0, unsupportedConversion(v.typ(), "uint")
	}
	return uint(v.N), nil
}
func (v I64Value) TryToUint() (uint, error) {
	if v.N < 0 {
		return 0, unsupportedConversion(v.typ(), "uint")
	}
	return uint(v.N), nil
}
func (v F64Value) TryToUint() (uint, error) {
	// MiniJinja uses int64 here, not uint.
	// https://github.com/mitsuhiko/minijinja/blob/1.0.7/minijinja/src/value/argtypes.rs#L438-L439
	// And it has comment "for the intention here see Key::from_borrowed_value"
	// but "from_borrowed_value" does not exist.
	if float64(int64(v.F)) == v.F && v.F >= 0 {
		return uint(v.F), nil
	}
	return 0, unsupportedConversion(v.typ(), "uint")
}
func (v NoneValue) TryToUint() (uint, error)    { return 0, unsupportedConversion(v.typ(), "uint") }
func (v InvalidValue) TryToUint() (uint, error) { return 0, unsupportedConversion(v.typ(), "uint") }
func (v U128Value) TryToUint() (uint, error) {
	if v.N.IsUint64() {
		n := v.N.Uint64()
		if n <= math.MaxUint {
			return uint(n), nil
		}
	}
	return 0, unsupportedConversion(v.typ(), "uint")
}
func (v I128Value) TryToUint() (uint, error) {
	if v.N.IsUint64() {
		n := v.N.Uint64()
		if n <= math.MaxUint {
			return uint(n), nil
		}
	}
	return 0, unsupportedConversion(v.typ(), "uint")
}
func (v StringValue) TryToUint() (uint, error)  { return 0, unsupportedConversion(v.typ(), "uint") }
func (v BytesValue) TryToUint() (uint, error)   { return 0, unsupportedConversion(v.typ(), "uint") }
func (v SeqValue) TryToUint() (uint, error)     { return 0, unsupportedConversion(v.typ(), "uint") }
func (v MapValue) TryToUint() (uint, error)     { return 0, unsupportedConversion(v.typ(), "uint") }
func (v DynamicValue) TryToUint() (uint, error) { return 0, unsupportedConversion(v.typ(), "uint") }

func (UndefinedValue) AsF64() option.Option[float64] { return option.None[float64]() }
func (v BoolValue) AsF64() option.Option[float64] {
	var f float64
	if v.B {
		f = 1
	}
	return option.Some(f)
}
func (v U64Value) AsF64() option.Option[float64]   { return option.Some(float64(v.N)) }
func (v I64Value) AsF64() option.Option[float64]   { return option.Some(float64(v.N)) }
func (v F64Value) AsF64() option.Option[float64]   { return option.Some(v.F) }
func (NoneValue) AsF64() option.Option[float64]    { return option.None[float64]() }
func (InvalidValue) AsF64() option.Option[float64] { return option.None[float64]() }
func (v U128Value) AsF64() option.Option[float64] {
	f, _ := v.N.Float64()
	return option.Some(f)
}
func (v I128Value) AsF64() option.Option[float64] {
	f, _ := v.N.Float64()
	return option.Some(f)
}
func (StringValue) AsF64() option.Option[float64]  { return option.None[float64]() }
func (BytesValue) AsF64() option.Option[float64]   { return option.None[float64]() }
func (SeqValue) AsF64() option.Option[float64]     { return option.None[float64]() }
func (MapValue) AsF64() option.Option[float64]     { return option.None[float64]() }
func (DynamicValue) AsF64() option.Option[float64] { return option.None[float64]() }

func (UndefinedValue) AsSeq() option.Option[SeqObject] { return option.None[SeqObject]() }
func (BoolValue) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (U64Value) AsSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (I64Value) AsSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (F64Value) AsSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (NoneValue) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (InvalidValue) AsSeq() option.Option[SeqObject]   { return option.None[SeqObject]() }
func (U128Value) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (I128Value) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (StringValue) AsSeq() option.Option[SeqObject]    { return option.None[SeqObject]() }
func (BytesValue) AsSeq() option.Option[SeqObject]     { return option.None[SeqObject]() }
func (v SeqValue) AsSeq() option.Option[SeqObject] {
	return option.Some(NewSliceSeqObject(v.Items))
}
func (MapValue) AsSeq() option.Option[SeqObject] { return option.None[SeqObject]() }
func (v DynamicValue) AsSeq() option.Option[SeqObject] {
	if seq, ok := v.Dy.(SeqObject); ok {
		return option.Some(seq)
	}
	return option.None[SeqObject]()
}

func (v UndefinedValue) Clone() Value { return v }
func (v BoolValue) Clone() Value      { return v }
func (v U64Value) Clone() Value       { return v }
func (v I64Value) Clone() Value       { return v }
func (v F64Value) Clone() Value       { return v }
func (v NoneValue) Clone() Value      { return v }
func (v InvalidValue) Clone() Value   { return v }
func (v U128Value) Clone() Value {
	c := v
	c.N.Set(&v.N)
	return c
}
func (v I128Value) Clone() Value {
	c := v
	c.N.Set(&v.N)
	return c
}
func (v StringValue) Clone() Value { return v }
func (v BytesValue) Clone() Value {
	b := make([]byte, len(v.B))
	copy(b, v.B)
	return BytesValue{B: b}
}
func (v SeqValue) Clone() Value {
	items := make([]Value, len(v.Items))
	for i, item := range v.Items {
		// Is shallow copy OK?
		items[i] = item
	}
	return SeqValue{Items: items}
}
func (v MapValue) Clone() Value {
	m := v.Map.Clone()
	return MapValue{Map: m, Type: v.Type}
}
func (v DynamicValue) Clone() Value {
	// TODO: implement real clone
	return v
}

func (UndefinedValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v BoolValue) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v U64Value) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v I64Value) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v F64Value) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (NoneValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v InvalidValue) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v U128Value) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v I128Value) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v StringValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &charsValueIteratorState{s: v.Str}, len: uint(utf8.RuneCountInString(v.Str))}, nil
}
func (v BytesValue) TryIter() (Iterator, error) {
	return Iterator{}, common.NewError(common.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v SeqValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &seqValueIteratorState{items: v.Items}, len: uint(len(v.Items))}, nil
}
func (v MapValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &mapValueIteratorState{keys: v.Map.Keys()}, len: uint(len(v.Map.Keys()))}, nil
}
func (v DynamicValue) TryIter() (Iterator, error) {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return Iterator{iterState: &emptyValueIteratorState{}}, nil
	case ObjectKindSeq:
		seqObj := v.Dy.(SeqObject)
		return Iterator{iterState: &dynSeqValueIteratorState{obj: seqObj}, len: seqObj.ItemCount()}, nil
	case ObjectKindStruct:
		obj := v.Dy.(StructObject)
		if optFields := obj.StaticFields(); optFields.IsSome() {
			return Iterator{iterState: &stringsValueIteratorState{items: optFields.Unwrap()}}, nil
		}
		return Iterator{iterState: &stringsValueIteratorState{items: obj.Fields()}}, nil
	default:
		panic("unreachable")
	}
}

func unsupportedConversion(kind valueType, target string) error {
	return common.NewError(common.InvalidOperation,
		fmt.Sprintf("cannot convert %s to %s", kind, target))
}

func I32TryFromValue(val Value) (int32, error) {
	n, err := val.TryToI64()
	if err != nil || n < math.MinInt32 || n > math.MaxInt32 {
		return 0, unsupportedConversion(val.typ(), "i32")
	}
	return int32(n), nil
}

type Iterator struct {
	iterState valueIteratorState
	len       uint
}

func IteratorFromSeqObject(s SeqObject) *Iterator {
	return &Iterator{iterState: &dynSeqValueIteratorState{obj: s}}
}

func IteratorFromStrings(items []string) *Iterator {
	return &Iterator{iterState: &stringsValueIteratorState{items: items}}
}

func (i Iterator) Chain(other Iterator) Iterator {
	return Iterator{
		iterState: &chainedValueIteratorState{
			states: []valueIteratorState{i.iterState, other.iterState},
		},
		len: i.len + other.len,
	}
}

func (i Iterator) Cloned() Iterator {
	return Iterator{
		iterState: &cloneValueIteratorState{
			state: i.iterState,
		},
		len: i.len,
	}
}

func (i *Iterator) Next() option.Option[Value] {
	optVal := i.iterState.advanceState()
	if optVal.IsSome() {
		i.len--
	}
	return optVal
}

func (i *Iterator) Len() uint {
	return i.len
}

// All returns if every element of the iterator matches a predicate.
// An empty iterator returns true.
func (i *Iterator) All(f func(Value) bool) bool {
	var item Value
	for i.Next().UnwrapTo(&item) {
		if !f(item) {
			return false
		}
	}
	return true
}

func (i *Iterator) CompareBy(other *Iterator, f func(a, b Value) int) int {
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

func (i *Iterator) Min() option.Option[Value] { return i.minBy(Cmp) }

func (i *Iterator) minBy(compare func(a, b Value) int) option.Option[Value] {
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

func (i *Iterator) Max() option.Option[Value] { return i.maxBy(Cmp) }

func (i *Iterator) maxBy(compare func(a, b Value) int) option.Option[Value] {
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

func (i *Iterator) Collect() []Value {
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
	obj SeqObject
}
type mapValueIteratorState struct {
	idx  uint
	keys []KeyRef
}

type cloneValueIteratorState struct {
	state valueIteratorState
}

func (s *cloneValueIteratorState) advanceState() option.Option[Value] {
	return option.Map(s.state.advanceState(), func(val Value) Value { return val.Clone() })
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
		return option.Some[Value](StringValue{Str: string(r)})
	}
	return option.None[Value]()
}
func (s *seqValueIteratorState) advanceState() option.Option[Value] {
	if s.idx < uint(len(s.items)) {
		item := s.items[s.idx]
		s.idx++
		return option.Some(item.Clone())
	}
	return option.None[Value]()
}
func (s *stringsValueIteratorState) advanceState() option.Option[Value] {
	if s.idx < uint(len(s.items)) {
		item := s.items[s.idx]
		s.idx++
		return option.Some[Value](StringValue{Str: item})
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

func (v *SeqValue) Append(val Value) {
	v.Items = append(v.Items, val)
}

func (UndefinedValue) Len() option.Option[uint] { return option.None[uint]() }
func (BoolValue) Len() option.Option[uint]      { return option.None[uint]() }
func (U64Value) Len() option.Option[uint]       { return option.None[uint]() }
func (I64Value) Len() option.Option[uint]       { return option.None[uint]() }
func (F64Value) Len() option.Option[uint]       { return option.None[uint]() }
func (NoneValue) Len() option.Option[uint]      { return option.None[uint]() }
func (InvalidValue) Len() option.Option[uint]   { return option.None[uint]() }
func (U128Value) Len() option.Option[uint]      { return option.None[uint]() }
func (I128Value) Len() option.Option[uint]      { return option.None[uint]() }
func (v StringValue) Len() option.Option[uint] {
	return option.Some(uint(utf8.RuneCountInString(v.Str)))
}
func (BytesValue) Len() option.Option[uint] { return option.None[uint]() }
func (v SeqValue) Len() option.Option[uint] { return option.Some(uint(len(v.Items))) }
func (v MapValue) Len() option.Option[uint] { return option.Some(v.Map.Len()) }
func (v DynamicValue) Len() option.Option[uint] {
	switch v.Dy.Kind() {
	case ObjectKindPlain:
		return option.None[uint]()
	case ObjectKindSeq:
		return option.Some(v.Dy.(SeqObject).ItemCount())
	case ObjectKindStruct:
		return option.Some(FieldCount(v.Dy.(StructObject)))
	default:
		panic("unreachable")
	}
}

func Equal(v, other Value) bool {
	switch {
	case v.Kind() == ValueKindNone && other.Kind() == ValueKindNone:
		return true
	case v.Kind() == ValueKindUndefined && other.Kind() == ValueKindUndefined:
		return true
	case v.Kind() == ValueKindString && other.Kind() == ValueKindString:
		a := v.(StringValue).Str
		b := other.(StringValue).Str
		return a == b
	case v.Kind() == ValueKindBytes && other.Kind() == ValueKindBytes:
		a := v.(BytesValue).B
		b := other.(BytesValue).B
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
			if optA, optB := v.AsSeq(), other.AsSeq(); optA.IsSome() && optB.IsSome() {
				iterA, err := v.TryIter()
				if err != nil {
					return false
				}
				iterB, err := v.TryIter()
				if err != nil {
					return false
				}
				return iterA.All(func(itemA Value) bool {
					itemB := iterB.Next().Unwrap()
					return Equal(itemA, itemB)
				})
			} else if v.Kind() == ValueKindMap && other.Kind() == ValueKindMap {
				if v.Len() != other.Len() {
					return false
				}
				iterA, err := v.TryIter()
				if err != nil {
					return false
				}
				return iterA.All(func(key Value) bool {
					optValA := v.GetItemOpt(key)
					optValB := other.GetItemOpt(key)
					if optValA.IsSome() && optValB.IsSome() {
						return Equal(optValA.Unwrap(), optValB.Unwrap())
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

func valueEqual(v, other Value) bool {
	return Cmp(v, other) == 0
}

// Cmp returns
// -1 if v is less than other,
//
//	0 if v equals other,
//
// +1 if v is greater than other.
func Cmp(v, other Value) int {
	var rv int
outer:
	switch {
	case v.Kind() == ValueKindNone && other.Kind() == ValueKindNone:
		rv = 0
	case v.Kind() == ValueKindUndefined && other.Kind() == ValueKindUndefined:
		rv = 0
	case v.Kind() == ValueKindString && other.Kind() == ValueKindString:
		a := v.(StringValue).Str
		b := other.(StringValue).Str
		rv = strings.Compare(a, b)
	case v.Kind() == ValueKindBytes && other.Kind() == ValueKindBytes:
		a := v.(BytesValue).B
		b := other.(BytesValue).B
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
			if optA, optB := v.AsSeq(), other.AsSeq(); optA.IsSome() && optB.IsSome() {
				iterA, err := v.TryIter()
				if err != nil {
					break outer
				}
				iterB, err := other.TryIter()
				if err != nil {
					break outer
				}
				return iterA.CompareBy(&iterB, Cmp)
			} else if v.Kind() == ValueKindMap && other.Kind() == ValueKindMap {
				iterA, err := v.TryIter()
				if err != nil {
					break outer
				}
				iterB, err := other.TryIter()
				if err != nil {
					break outer
				}
				return iterA.CompareBy(&iterB, func(keyA, keyB Value) int {
					if rv := Cmp(keyA, keyB); rv != 0 {
						return 0
					}
					optValA := v.GetItemOpt(keyA)
					optValB := other.GetItemOpt(keyB)
					return optValA.Compare(optValB, Cmp)
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
	rightInt := int64(math.Float64bits(left))
	leftInt ^= int64(uint64(leftInt>>63) >> 1)
	rightInt ^= int64(uint64(rightInt>>63) >> 1)
	return cmp.Compare(leftInt, rightInt)
}

func GetItem(val, key Value) (Value, error) {
	if val.IsUndefined() {
		return nil, common.NewError(common.UndefinedError, "")
	}
	return val.GetItemOpt(key).UnwrapOr(Undefined), nil
}

func BoolTryFromValue(v Value) (bool, error) {
	if boolVal, ok := v.(BoolValue); ok {
		return boolVal.B, nil
	}
	return false, unsupportedConversion(v.typ(), "bool")
}

func boolTryFromOptionValue(v option.Option[Value]) (bool, error) {
	if v.IsNone() {
		return false, common.NewError(common.MissingArgument, "")
	}
	return BoolTryFromValue(v.Unwrap())
}

func GetAttr(val Value, key string) (Value, error) {
	switch v := val.(type) {
	case UndefinedValue:
		return nil, common.NewError(common.UndefinedError, "")
	case MapValue:
		if v2, ok := v.Map.Get(KeyRefFromString(key)); ok {
			return v2.Clone(), nil
		}
	case DynamicValue:
		if obj, ok := v.Dy.(StructObject); ok {
			if optField := obj.GetField(key); optField.IsSome() {
				return optField.Unwrap(), nil
			}
		}
	}
	return Undefined, nil
}

func valueGetItemByIndex(val Value, idx uint) (Value, error) {
	return GetItem(val, ValueFromU64(uint64(idx)))
}

func GetPath(val Value, path string) (Value, error) {
	rv := val.Clone()
	for _, part := range strings.Split(path, ".") {
		num, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			rv, err = GetAttr(val, part)
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

func (v UndefinedValue) Hash(h hash.Hash) { valueHash(v, h) }
func (v BoolValue) Hash(h hash.Hash)      { valueHash(v, h) }
func (v U64Value) Hash(h hash.Hash)       { valueHash(v, h) }
func (v I64Value) Hash(h hash.Hash)       { valueHash(v, h) }
func (v F64Value) Hash(h hash.Hash)       { valueHash(v, h) }
func (v NoneValue) Hash(h hash.Hash)      { valueHash(v, h) }
func (v InvalidValue) Hash(h hash.Hash)   { valueHash(v, h) }
func (v U128Value) Hash(h hash.Hash)      { valueHash(v, h) }
func (v I128Value) Hash(h hash.Hash)      { valueHash(v, h) }
func (v StringValue) Hash(h hash.Hash)    { valueHash(v, h) }
func (v BytesValue) Hash(h hash.Hash)     { valueHash(v, h) }
func (v SeqValue) Hash(h hash.Hash)       { valueHash(v, h) }
func (v MapValue) Hash(h hash.Hash)       { valueHash(v, h) }
func (v DynamicValue) Hash(h hash.Hash)   { valueHash(v, h) }

func valueHash(val Value, h hash.Hash) {
	switch v := val.(type) {
	case NoneValue, UndefinedValue:
		h.Write([]byte{0})
	case StringValue:
		io.WriteString(h, v.Str)
	case BoolValue:
		b := byte(8)
		if v.B {
			b = byte(1)
		}
		h.Write([]byte{b})
	case InvalidValue:
		io.WriteString(h, v.Detail)
	case BytesValue:
		h.Write(v.B)
	case SeqValue:
		binary.Write(h, binary.BigEndian, uint64(len(v.Items)))
		for _, item := range v.Items {
			valueHash(item, h)
		}
	case MapValue:
		l := v.Map.Len()
		for i := uint(0); i < l; i++ {
			entry, _ := v.Map.EntryAt(i)
			keyRefHash(entry.Key, h)
			valueHash(entry.Value, h)
		}
	case DynamicValue:
		switch v.Dy.Kind() {
		case ObjectKindPlain:
			h.Write([]byte{0})
		case ObjectKindSeq:
			var item Value
			for iter := IteratorFromSeqObject(v.Dy.(SeqObject)); iter.Next().UnwrapTo(&item); {
				valueHash(item, h)
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
	case U64Value, I64Value, F64Value, U128Value, I128Value:
		n, err := val.Clone().TryToI64()
		if err != nil {
			val.AsF64().Hash(h, f64Hash)
		} else {
			binary.Write(h, binary.BigEndian, n)
		}
	}
}

func f64Hash(f float64, h hash.Hash) {
	binary.Write(h, binary.BigEndian, math.Float64bits(f))
}

func (v UndefinedValue) Equal(other any) bool { return valueEqualAny(v, other) }
func (v BoolValue) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v U64Value) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v I64Value) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v F64Value) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v NoneValue) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v InvalidValue) Equal(other any) bool   { return valueEqualAny(v, other) }
func (v U128Value) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v I128Value) Equal(other any) bool      { return valueEqualAny(v, other) }
func (v StringValue) Equal(other any) bool    { return valueEqualAny(v, other) }
func (v BytesValue) Equal(other any) bool     { return valueEqualAny(v, other) }
func (v SeqValue) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v MapValue) Equal(other any) bool       { return valueEqualAny(v, other) }
func (v DynamicValue) Equal(other any) bool   { return valueEqualAny(v, other) }
