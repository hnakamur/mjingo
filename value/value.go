package value

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/datast/indexmap"
	"github.com/hnakamur/mjingo/internal/datast/option"
)

type Value interface {
	String() string

	typ() valueType
	kind() valueKind
	IsUndefined() bool
	IsNone() bool
	IsTrue() bool
	GetAttrFast(key string) option.Option[Value]
	GetItemOpt(key Value) option.Option[Value]
	AsStr() option.Option[string]
	TryToI64() (int64, error)
	AsF64() option.Option[float64]
	AsSeq() option.Option[SeqObject]
	Clone() Value
	TryIter() (Iterator, error)
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
	// TODO: implement
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
	l := indexmap.Len(v.m)
	for i := uint(0); i < l; i++ {
		e, _ := indexmap.EntryAt(v.m, i)
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}
		b.WriteString(option.Unwrap(e.Key.AsStr()))
		b.WriteString(": ")
		b.WriteString(e.Value.String()) // MiniJinja uses fmt::Debug instead of fmt::Display here
	}
	b.WriteString("}")
	return b.String()
}
func (v dynamicValue) String() string { panic("not implemented yet") }

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

func (undefinedValue) kind() valueKind { return valueKindUndefined }
func (BoolValue) kind() valueKind      { return valueKindBool }
func (u64Value) kind() valueKind       { return valueKindNumber }
func (i64Value) kind() valueKind       { return valueKindNumber }
func (f64Value) kind() valueKind       { return valueKindNumber }
func (noneValue) kind() valueKind      { return valueKindNone }
func (InvalidValue) kind() valueKind   { return valueKindMap } // XXX: invalid values report themselves as maps which is a lie
func (u128Value) kind() valueKind      { return valueKindNumber }
func (i128Value) kind() valueKind      { return valueKindNumber }
func (stringValue) kind() valueKind    { return valueKindString }
func (bytesValue) kind() valueKind     { return valueKindBytes }
func (SeqValue) kind() valueKind       { return valueKindSeq }
func (mapValue) kind() valueKind       { return valueKindMap }
func (dynamicValue) kind() valueKind   { panic("not implemented for valueTypeDynamic") }

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
func (v mapValue) IsTrue() bool     { return indexmap.Len(v.m) != 0 }
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
	if val, ok := indexmap.Get[KeyRef, Value](v.m, KeyRefFromString(key)); ok {
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
	if optIdx := keyRf.AasI64(); option.IsSome(optIdx) {
		idx := option.Unwrap(optIdx)
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
	if v, ok := indexmap.Get[KeyRef, Value](v.m, KeyRefFromValue(key)); ok {
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
	m := indexmap.Clone(v.m)
	return mapValue{m: m, mapTyp: v.mapTyp}
}
func (dynamicValue) Clone() Value {
	panic("not implemented yet")
}

func (undefinedValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v BoolValue) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v u64Value) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v i64Value) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v f64Value) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (noneValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &emptyValueIteratorState{}}, nil
}
func (v InvalidValue) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v u128Value) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v i128Value) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v stringValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &charsValueIteratorState{s: v.str}, len: uint(utf8.RuneCountInString(v.str))}, nil
}
func (v bytesValue) TryIter() (Iterator, error) {
	return Iterator{}, internal.NewError(internal.InvalidOperation, fmt.Sprintf("%s is not iteratble", v.kind()))
}
func (v SeqValue) TryIter() (Iterator, error) {
	return Iterator{iterState: &seqValueIteratorState{items: v.items}, len: uint(len(v.items))}, nil
}
func (v mapValue) TryIter() (Iterator, error) { panic("not implemented yet") }
func (v dynamicValue) TryIter() (Iterator, error) {
	panic("not implemented yet")
}

func unsupportedConversion(kind valueType, target string) error {
	return internal.NewError(internal.InvalidOperation,
		fmt.Sprintf("cannot convert %s to %s", kind, target))
}

type Iterator struct {
	iterState valueIteratorState
	len       uint
}

func (i *Iterator) Next() option.Option[Value] {
	optVal := i.iterState.advanceState()
	if option.IsSome(optVal) {
		i.len--
	}
	return optVal
}

func (i *Iterator) Len() uint {
	return i.len
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
	idx uint
	// TODO: implement ordered map
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
func (s *mapValueIteratorState) advanceState() option.Option[Value]    { panic("not implemented") }

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
