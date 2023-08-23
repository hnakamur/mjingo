package internal

import (
	"bytes"
	"cmp"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type Value interface {
	String() string

	typ() valueType
	Kind() ValueKind
	IsUndefined() bool
	IsNone() bool
	IsSafe() bool
	IsTrue() bool
	GetAttrFast(key string) option.Option[Value]
	GetItemOpt(key Value) option.Option[Value]
	AsStr() option.Option[string]
	TryToI64() (int64, error)
	AsF64() option.Option[float64]
	AsSeq() option.Option[SeqObject]
	Clone() Value
	TryIter() (Iterator, error)
	Len() option.Option[uint]
	Call(state *State, args []Value) (Value, error)
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

var Undefined = undefinedValue{}
var None = noneValue{}

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

type undefinedValue struct{}
type BoolValue struct{ B bool }
type u64Value struct{ n uint64 }
type i64Value struct{ n int64 }
type f64Value struct{ f float64 }
type noneValue struct{}
type InvalidValue struct{ Detail string }
type u128Value struct{ hi, lo uint64 }
type i128Value struct {
	hi int64
	lo uint64
}
type stringValue struct {
	str    string
	strTyp stringType
}
type bytesValue struct{ b []byte }
type SeqValue struct{ items []Value }
type mapValue struct {
	m      *IndexMap
	mapTyp mapType
}
type dynamicValue struct {
	dy Object
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
var _ = Value(BoolValue{})
var _ = Value(u64Value{})
var _ = Value(i64Value{})
var _ = Value(f64Value{})
var _ = Value(noneValue{})
var _ = Value(InvalidValue{})
var _ = Value(u128Value{})
var _ = Value(i128Value{})
var _ = Value(stringValue{})
var _ = Value(bytesValue{})
var _ = Value(SeqValue{})
var _ = Value(mapValue{})
var _ = Value(dynamicValue{})

func (v undefinedValue) String() string { return "" }
func (v BoolValue) String() string      { return strconv.FormatBool(v.B) }
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
func (v InvalidValue) String() string { return fmt.Sprintf("<invalid value: %s>", v.Detail) }
func (v u128Value) String() string    { panic("not implemented yet") }
func (v i128Value) String() string    { panic("not implemented yet") }
func (v stringValue) String() string  { return v.str }
func (v bytesValue) String() string   { return string(v.b) } // TODO: equivalent impl as String::from_utf8_lossy
func (v SeqValue) String() string {
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
	l := v.m.Len()
	for i := uint(0); i < l; i++ {
		e, _ := v.m.EntryAt(i)
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(e.Key.AsStr().Unwrap())
		b.WriteString(": ")
		b.WriteString(e.Value.String()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("}")
	return b.String()
}
func (v dynamicValue) String() string { return fmt.Sprintf("%s", v.dy) }

func (undefinedValue) typ() valueType { return valueTypeUndefined }
func (BoolValue) typ() valueType      { return valueTypeBool }
func (u64Value) typ() valueType       { return valueTypeU64 }
func (i64Value) typ() valueType       { return valueTypeI64 }
func (f64Value) typ() valueType       { return valueTypeF64 }
func (noneValue) typ() valueType      { return valueTypeNone }
func (InvalidValue) typ() valueType   { return valueTypeInvalid }
func (u128Value) typ() valueType      { return valueTypeU128 }
func (i128Value) typ() valueType      { return valueTypeI128 }
func (stringValue) typ() valueType    { return valueTypeString }
func (bytesValue) typ() valueType     { return valueTypeBytes }
func (SeqValue) typ() valueType       { return valueTypeSeq }
func (mapValue) typ() valueType       { return valueTypeMap }
func (dynamicValue) typ() valueType   { return valueTypeDynamic }

func (undefinedValue) Kind() ValueKind { return ValueKindUndefined }
func (BoolValue) Kind() ValueKind      { return ValueKindBool }
func (u64Value) Kind() ValueKind       { return ValueKindNumber }
func (i64Value) Kind() ValueKind       { return ValueKindNumber }
func (f64Value) Kind() ValueKind       { return ValueKindNumber }
func (noneValue) Kind() ValueKind      { return ValueKindNone }
func (InvalidValue) Kind() ValueKind   { return ValueKindMap } // XXX: invalid values report themselves as maps which is a lie
func (u128Value) Kind() ValueKind      { return ValueKindNumber }
func (i128Value) Kind() ValueKind      { return ValueKindNumber }
func (stringValue) Kind() ValueKind    { return ValueKindString }
func (bytesValue) Kind() ValueKind     { return ValueKindBytes }
func (SeqValue) Kind() ValueKind       { return ValueKindSeq }
func (mapValue) Kind() ValueKind       { return ValueKindMap }
func (dynamicValue) Kind() ValueKind   { panic("not implemented for valueTypeDynamic") }

func (undefinedValue) IsUndefined() bool { return true }
func (BoolValue) IsUndefined() bool      { return false }
func (u64Value) IsUndefined() bool       { return false }
func (i64Value) IsUndefined() bool       { return false }
func (f64Value) IsUndefined() bool       { return false }
func (noneValue) IsUndefined() bool      { return false }
func (InvalidValue) IsUndefined() bool   { return false }
func (u128Value) IsUndefined() bool      { return false }
func (i128Value) IsUndefined() bool      { return false }
func (stringValue) IsUndefined() bool    { return false }
func (bytesValue) IsUndefined() bool     { return false }
func (SeqValue) IsUndefined() bool       { return false }
func (mapValue) IsUndefined() bool       { return false }
func (dynamicValue) IsUndefined() bool   { return false }

func (undefinedValue) IsNone() bool { return false }
func (BoolValue) IsNone() bool      { return false }
func (u64Value) IsNone() bool       { return false }
func (i64Value) IsNone() bool       { return false }
func (f64Value) IsNone() bool       { return false }
func (noneValue) IsNone() bool      { return true }
func (InvalidValue) IsNone() bool   { return false }
func (u128Value) IsNone() bool      { return false }
func (i128Value) IsNone() bool      { return false }
func (stringValue) IsNone() bool    { return false }
func (bytesValue) IsNone() bool     { return false }
func (SeqValue) IsNone() bool       { return false }
func (mapValue) IsNone() bool       { return false }
func (dynamicValue) IsNone() bool   { return false }

func (undefinedValue) IsSafe() bool { return false }
func (BoolValue) IsSafe() bool      { return false }
func (u64Value) IsSafe() bool       { return false }
func (i64Value) IsSafe() bool       { return false }
func (f64Value) IsSafe() bool       { return false }
func (noneValue) IsSafe() bool      { return false }
func (InvalidValue) IsSafe() bool   { return false }
func (u128Value) IsSafe() bool      { return false }
func (i128Value) IsSafe() bool      { return false }
func (v stringValue) IsSafe() bool  { return v.strTyp == stringTypeSafe }
func (bytesValue) IsSafe() bool     { return false }
func (SeqValue) IsSafe() bool       { return false }
func (mapValue) IsSafe() bool       { return false }
func (dynamicValue) IsSafe() bool   { return false }

func (undefinedValue) IsTrue() bool { return false }
func (v BoolValue) IsTrue() bool    { return v.B }
func (v u64Value) IsTrue() bool     { return v.n != 0 }
func (v i64Value) IsTrue() bool     { return v.n != 0 }
func (v f64Value) IsTrue() bool     { return v.f != 0.0 }
func (noneValue) IsTrue() bool      { return false }
func (InvalidValue) IsTrue() bool   { return false }
func (v u128Value) IsTrue() bool    { panic("not implemented") }
func (v i128Value) IsTrue() bool    { panic("not implemented") }
func (v stringValue) IsTrue() bool  { return len(v.str) != 0 }
func (v bytesValue) IsTrue() bool   { return len(v.b) != 0 }
func (v SeqValue) IsTrue() bool     { return len(v.items) != 0 }
func (v mapValue) IsTrue() bool     { return v.m.Len() != 0 }
func (v dynamicValue) IsTrue() bool { panic("not implemented for valueTypeDynamic") }

func (undefinedValue) GetAttrFast(_ string) option.Option[Value] { return option.None[Value]() }
func (BoolValue) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (u64Value) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (i64Value) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (f64Value) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (noneValue) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (InvalidValue) GetAttrFast(_ string) option.Option[Value]   { return option.None[Value]() }
func (u128Value) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (i128Value) GetAttrFast(_ string) option.Option[Value]      { return option.None[Value]() }
func (stringValue) GetAttrFast(_ string) option.Option[Value]    { return option.None[Value]() }
func (bytesValue) GetAttrFast(_ string) option.Option[Value]     { return option.None[Value]() }
func (SeqValue) GetAttrFast(_ string) option.Option[Value]       { return option.None[Value]() }
func (v mapValue) GetAttrFast(key string) option.Option[Value] {
	if val, ok := v.m.Get(KeyRefFromString(key)); ok {
		return option.Some(val)
	}
	return option.None[Value]()
}
func (dynamicValue) GetAttrFast(_ string) option.Option[Value] {
	panic("not implemented yet")
}

func (undefinedValue) GetItemOpt(_ Value) option.Option[Value] { return option.None[Value]() }
func (BoolValue) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (u64Value) GetItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (i64Value) GetItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (f64Value) GetItemOpt(_ Value) option.Option[Value]       { return option.None[Value]() }
func (noneValue) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (InvalidValue) GetItemOpt(_ Value) option.Option[Value]   { return option.None[Value]() }
func (u128Value) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (i128Value) GetItemOpt(_ Value) option.Option[Value]      { return option.None[Value]() }
func (stringValue) GetItemOpt(_ Value) option.Option[Value]    { return option.None[Value]() }
func (bytesValue) GetItemOpt(_ Value) option.Option[Value]     { return option.None[Value]() }
func (v SeqValue) GetItemOpt(key Value) option.Option[Value] {
	keyRf := valueKeyRef{val: key}
	if optIdx := keyRf.AasI64(); optIdx.IsSome() {
		idx := optIdx.Unwrap()
		if idx < math.MinInt || math.MaxInt < idx {
			return option.None[Value]()
		}
		seq := newSliceSeqObject(v.items)
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
func (v mapValue) GetItemOpt(key Value) option.Option[Value] {
	if v, ok := v.m.Get(KeyRefFromValue(key)); ok {
		return option.Some(v)
	}
	return option.None[Value]()
}
func (dynamicValue) GetItemOpt(_ Value) option.Option[Value] {
	panic("not implemented yet")
}

func (undefinedValue) AsStr() option.Option[string] { return option.None[string]() }
func (BoolValue) AsStr() option.Option[string]      { return option.None[string]() }
func (u64Value) AsStr() option.Option[string]       { return option.None[string]() }
func (i64Value) AsStr() option.Option[string]       { return option.None[string]() }
func (f64Value) AsStr() option.Option[string]       { return option.None[string]() }
func (noneValue) AsStr() option.Option[string]      { return option.None[string]() }
func (InvalidValue) AsStr() option.Option[string]   { return option.None[string]() }
func (u128Value) AsStr() option.Option[string]      { return option.None[string]() }
func (i128Value) AsStr() option.Option[string]      { return option.None[string]() }
func (v stringValue) AsStr() option.Option[string]  { return option.Some(v.str) }
func (bytesValue) AsStr() option.Option[string]     { return option.None[string]() }
func (SeqValue) AsStr() option.Option[string]       { return option.None[string]() }
func (v mapValue) AsStr() option.Option[string]     { return option.None[string]() }
func (dynamicValue) AsStr() option.Option[string] {
	panic("not implemented yet")
}

func (v undefinedValue) TryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v BoolValue) TryToI64() (int64, error) {
	if v.B {
		return 1, nil
	}
	return 0, nil
}
func (v u64Value) TryToI64() (int64, error) { return int64(v.n), nil }
func (v i64Value) TryToI64() (int64, error) { return v.n, nil }
func (v f64Value) TryToI64() (int64, error) {
	if float64(int64(v.f)) == v.f {
		return int64(v.f), nil
	}
	return 0, unsupportedConversion(v.typ(), "i64")
}
func (v noneValue) TryToI64() (int64, error)    { return 0, unsupportedConversion(v.typ(), "i64") }
func (v InvalidValue) TryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }
func (v u128Value) TryToI64() (int64, error)    { panic("not implemented yet") }
func (v i128Value) TryToI64() (int64, error)    { panic("not implemented yet") }
func (v stringValue) TryToI64() (int64, error)  { return 0, unsupportedConversion(v.typ(), "i64") }
func (v bytesValue) TryToI64() (int64, error)   { return 0, unsupportedConversion(v.typ(), "i64") }
func (v SeqValue) TryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v mapValue) TryToI64() (int64, error)     { return 0, unsupportedConversion(v.typ(), "i64") }
func (v dynamicValue) TryToI64() (int64, error) { return 0, unsupportedConversion(v.typ(), "i64") }

func (undefinedValue) AsF64() option.Option[float64] { return option.None[float64]() }
func (v BoolValue) AsF64() option.Option[float64] {
	if v.B {
		return option.Some(float64(1))
	}
	return option.None[float64]()
}
func (v u64Value) AsF64() option.Option[float64]    { return option.Some(float64(v.n)) }
func (v i64Value) AsF64() option.Option[float64]    { return option.Some(float64(v.n)) }
func (v f64Value) AsF64() option.Option[float64]    { return option.Some(v.f) }
func (noneValue) AsF64() option.Option[float64]     { return option.None[float64]() }
func (InvalidValue) AsF64() option.Option[float64]  { return option.None[float64]() }
func (u128Value) AsF64() option.Option[float64]     { panic("not implemented yet") }
func (i128Value) AsF64() option.Option[float64]     { panic("not implemented yet") }
func (v stringValue) AsF64() option.Option[float64] { return option.None[float64]() }
func (bytesValue) AsF64() option.Option[float64]    { return option.None[float64]() }
func (SeqValue) AsF64() option.Option[float64]      { return option.None[float64]() }
func (mapValue) AsF64() option.Option[float64]      { return option.None[float64]() }
func (dynamicValue) AsF64() option.Option[float64] {
	panic("not implemented yet")
}

func (undefinedValue) AsSeq() option.Option[SeqObject] { return option.None[SeqObject]() }
func (BoolValue) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (u64Value) AsSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (i64Value) AsSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (f64Value) AsSeq() option.Option[SeqObject]       { return option.None[SeqObject]() }
func (noneValue) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (InvalidValue) AsSeq() option.Option[SeqObject]   { return option.None[SeqObject]() }
func (u128Value) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (i128Value) AsSeq() option.Option[SeqObject]      { return option.None[SeqObject]() }
func (stringValue) AsSeq() option.Option[SeqObject]    { return option.None[SeqObject]() }
func (bytesValue) AsSeq() option.Option[SeqObject]     { return option.None[SeqObject]() }
func (v SeqValue) AsSeq() option.Option[SeqObject] {
	return option.Some(newSliceSeqObject(v.items))
}
func (mapValue) AsSeq() option.Option[SeqObject] { return option.None[SeqObject]() }
func (dynamicValue) AsSeq() option.Option[SeqObject] {
	panic("not implemented yet")
}

func (v undefinedValue) Clone() Value { return v }
func (v BoolValue) Clone() Value      { return v }
func (v u64Value) Clone() Value       { return v }
func (v i64Value) Clone() Value       { return v }
func (v f64Value) Clone() Value       { return v }
func (v noneValue) Clone() Value      { return v }
func (v InvalidValue) Clone() Value   { return v }
func (v u128Value) Clone() Value      { return v }
func (v i128Value) Clone() Value      { return v }
func (v stringValue) Clone() Value    { return v }
func (v bytesValue) Clone() Value {
	b := make([]byte, len(v.b))
	copy(b, v.b)
	return bytesValue{b: b}
}
func (v SeqValue) Clone() Value {
	items := make([]Value, len(v.items))
	for i, item := range v.items {
		// Is shallow copy OK?
		items[i] = item
	}
	return SeqValue{items: items}
}
func (v mapValue) Clone() Value {
	m := v.m.Clone()
	return mapValue{m: m, mapTyp: v.mapTyp}
}
func (v dynamicValue) Clone() Value {
	// TODO: implement real clone
	return v
}

func (undefinedValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v BoolValue) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v u64Value) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v i64Value) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v f64Value) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (noneValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v InvalidValue) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v u128Value) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v i128Value) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v stringValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &charsValueIteratorState{s: v.str}, len: uint(utf8.RuneCountInString(v.str))}, nil
}
func (v bytesValue) TryIter() (Iterator, error) {
	return Iterator{}, NewError(InvalidOperation, fmt.Sprintf("%s is not iteratble", v.Kind()))
}
func (v SeqValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &seqValueIteratorState{items: v.items}, len: uint(len(v.items))}, nil
}
func (v mapValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &mapValueIteratorState{keys: v.m.keys()}, len: uint(len(v.m.keys()))}, nil
}
func (v dynamicValue) TryIter() (Iterator, error) {
	panic("not implemented yet")
}

func unsupportedConversion(kind valueType, target string) error {
	return NewError(InvalidOperation,
		fmt.Sprintf("cannot convert %s to %s", kind, target))
}

type Iterator struct {
	iterState valueIteratorState
	len       uint
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
	for {
		optVal := i.Next()
		if optVal.IsNone() {
			break
		}
		if !f(optVal.Unwrap()) {
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
	// obj Object
}
type mapValueIteratorState struct {
	idx  uint
	keys []KeyRef
}

func (s *emptyValueIteratorState) advanceState() option.Option[Value] { return option.None[Value]() }
func (s *charsValueIteratorState) advanceState() option.Option[Value] {
	if s.offset < uint(len(s.s)) {
		r, size := utf8.DecodeRuneInString(s.s[s.offset:])
		s.offset += uint(size)
		return option.Some[Value](stringValue{str: string(r)})
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
		return option.Some[Value](stringValue{str: item})
	}
	return option.None[Value]()
}
func (s *dynSeqValueIteratorState) advanceState() option.Option[Value] { panic("not implemented") }
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
	v.items = append(v.items, val)
}

func (undefinedValue) Len() option.Option[uint] { return option.None[uint]() }
func (BoolValue) Len() option.Option[uint]      { return option.None[uint]() }
func (u64Value) Len() option.Option[uint]       { return option.None[uint]() }
func (i64Value) Len() option.Option[uint]       { return option.None[uint]() }
func (f64Value) Len() option.Option[uint]       { return option.None[uint]() }
func (noneValue) Len() option.Option[uint]      { return option.None[uint]() }
func (InvalidValue) Len() option.Option[uint]   { return option.None[uint]() }
func (u128Value) Len() option.Option[uint]      { return option.None[uint]() }
func (i128Value) Len() option.Option[uint]      { return option.None[uint]() }
func (v stringValue) Len() option.Option[uint] {
	return option.Some(uint(utf8.RuneCountInString(v.str)))
}
func (bytesValue) Len() option.Option[uint] { return option.None[uint]() }
func (v SeqValue) Len() option.Option[uint] { return option.Some(uint(len(v.items))) }
func (v mapValue) Len() option.Option[uint] { return option.Some(v.m.Len()) }
func (dynamicValue) Len() option.Option[uint] {
	panic("not implemented yet")
}

func Equal(v Value, other Value) bool {
	switch {
	case v.Kind() == ValueKindNone && other.Kind() == ValueKindNone:
		return true
	case v.Kind() == ValueKindUndefined && other.Kind() == ValueKindUndefined:
		return true
	case v.Kind() == ValueKindString && other.Kind() == ValueKindString:
		a := v.(stringValue).str
		b := other.(stringValue).str
		return a == b
	case v.Kind() == ValueKindBytes && other.Kind() == ValueKindBytes:
		a := v.(bytesValue).b
		b := other.(bytesValue).b
		return bytes.Equal(a, b)
	default:
		switch c := coerce(v, other).(type) {
		case f64CoerceResult:
			return c.lhs == c.rhs
		case i64CoerceResult:
			return c.lhs == c.rhs
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

// Cmp returns
// -1 if v is less than other,
//
//	0 if v equals other,
//
// +1 if v is greater than other.
func Cmp(v Value, other Value) int {
	var rv int
outer:
	switch {
	case v.Kind() == ValueKindNone && other.Kind() == ValueKindNone:
		rv = 0
	case v.Kind() == ValueKindUndefined && other.Kind() == ValueKindUndefined:
		rv = 0
	case v.Kind() == ValueKindString && other.Kind() == ValueKindString:
		a := v.(stringValue).str
		b := other.(stringValue).str
		rv = strings.Compare(a, b)
	case v.Kind() == ValueKindBytes && other.Kind() == ValueKindBytes:
		a := v.(bytesValue).b
		b := other.(bytesValue).b
		rv = bytes.Compare(a, b)
	default:
		switch c := coerce(v, other).(type) {
		case f64CoerceResult:
			return f64TotalCmp(c.lhs, c.rhs)
		case i64CoerceResult:
			return cmp.Compare(c.lhs, c.rhs)
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
					return option.Compare(optValA, optValB, Cmp)
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

func (v undefinedValue) Call(state *State, args []Value) (Value, error) {
	return notCallableValueType(v)
}
func (v BoolValue) Call(state *State, args []Value) (Value, error)    { return notCallableValueType(v) }
func (v u64Value) Call(state *State, args []Value) (Value, error)     { return notCallableValueType(v) }
func (v i64Value) Call(state *State, args []Value) (Value, error)     { return notCallableValueType(v) }
func (v f64Value) Call(state *State, args []Value) (Value, error)     { return notCallableValueType(v) }
func (v noneValue) Call(state *State, args []Value) (Value, error)    { return notCallableValueType(v) }
func (v InvalidValue) Call(state *State, args []Value) (Value, error) { return notCallableValueType(v) }
func (v u128Value) Call(state *State, args []Value) (Value, error)    { return notCallableValueType(v) }
func (v i128Value) Call(state *State, args []Value) (Value, error)    { return notCallableValueType(v) }
func (v stringValue) Call(state *State, args []Value) (Value, error)  { return notCallableValueType(v) }
func (v bytesValue) Call(state *State, args []Value) (Value, error)   { return notCallableValueType(v) }
func (v SeqValue) Call(state *State, args []Value) (Value, error)     { return notCallableValueType(v) }
func (v mapValue) Call(state *State, args []Value) (Value, error)     { return notCallableValueType(v) }
func (v dynamicValue) Call(state *State, args []Value) (Value, error) {
	if c, ok := v.dy.(Caller); ok {
		return c.Call(state, args)
	}
	return nil, NewError(InvalidOperation, "tried to call non callable object")
}

func notCallableValueType(v Value) (Value, error) {
	return nil, NewError(InvalidOperation,
		fmt.Sprintf("value of type %s is not callable", v.Kind()))
}
