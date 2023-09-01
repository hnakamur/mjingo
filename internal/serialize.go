package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"reflect"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

func serializeBool(v bool) (Value, error) {
	return ValueFromBool(v), nil
}

func serializeI8(v int8) (Value, error) {
	return ValueFromI64(int64(v)), nil
}

func serializeI16(v int16) (Value, error) {
	return ValueFromI64(int64(v)), nil
}

func serializeI32(v int32) (Value, error) {
	return ValueFromI64(int64(v)), nil
}

func serializeI64(v int64) (Value, error) {
	return ValueFromI64(v), nil
}

func serializeInt(v int) (Value, error) {
	return ValueFromI64(int64(v)), nil
}

func serializeI128(v big.Int) (Value, error) {
	if isI128(&v) {
		return ValueFromI128(v), nil
	}
	return nil, errors.New("value out of range of i128")
}

func serializeU8(v uint8) (Value, error) {
	return ValueFromU64(uint64(v)), nil
}

func serializeU16(v uint16) (Value, error) {
	return ValueFromU64(uint64(v)), nil
}

func serializeU32(v uint32) (Value, error) {
	return ValueFromU64(uint64(v)), nil
}

func serializeU64(v uint64) (Value, error) {
	return ValueFromU64(v), nil
}

func serializeUint(v uint) (Value, error) {
	return ValueFromU64(uint64(v)), nil
}

func serializeU128(v big.Int) (Value, error) {
	if isU128(&v) {
		return ValueFromU128(v), nil
	}
	return nil, errors.New("value out of range of u128")
}

func serializeF32(v float32) (Value, error) {
	return ValueFromF64(float64(v)), nil
}

func serializeF64(v float64) (Value, error) {
	return ValueFromF64(v), nil
}

func serializeRune(v rune) (Value, error) {
	return ValueFromString(string(v)), nil
}

func serializeStr(v string) (Value, error) {
	return ValueFromString(v), nil
}

func serializeBytes(v []byte) (Value, error) {
	return ValueFromBytes(v), nil
}

func serializeNone() (Value, error) {
	return None, nil
}

func ValueTryFromGoValue(val any) (Value, error) {
	log.Printf("ValueTryFromGoValue val=%+v %T", val, val)
	switch v := val.(type) {
	case bool:
		return serializeBool(v)
	case uint8:
		return serializeU8(v)
	case uint16:
		return serializeU16(v)
	case uint32:
		return serializeU32(v)
	case uint64:
		return serializeU64(v)
	case uint:
		return serializeUint(v)
	case int8:
		return serializeI8(v)
	case int16:
		return serializeI16(v)
	case int32:
		return serializeI32(v)
	case int64:
		return serializeI64(v)
	case int:
		return serializeInt(v)
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			f, err := v.Float64()
			if err != nil {
				return nil, errors.New("invalid json.Number")
			}
			return serializeF64(f)
		}
		return serializeI64(n)
	case big.Int:
		if isI128(&v) {
			return serializeI128(v)
		}
		return serializeU128(v)
	case float32:
		return serializeF32(v)
	case float64:
		return serializeF64(v)
	case string:
		return serializeStr(v)
	case nil:
		return serializeNone()
	case []any:
		items := make([]Value, len(v))
		for i := range v {
			item, err := ValueTryFromGoValue(v[i])
			if err != nil {
				return nil, err
			}
			items[i] = item
		}
		return ValueFromSlice(items), nil
	case map[any]any:
		m := NewIndexMap()
		for goKey, goVal := range v {
			key, err := ValueTryFromGoValue(goKey)
			if err != nil {
				return nil, err
			}
			val2, err := ValueTryFromGoValue(goVal)
			if err != nil {
				return nil, err
			}
			m.Set(KeyRefFromValue(key), val2)
		}
		return ValueFromIndexMap(m), nil
	case map[string]any:
		m := NewIndexMap()
		for goKey, goVal := range v {
			key, err := ValueTryFromGoValue(goKey)
			if err != nil {
				return nil, err
			}
			val2, err := ValueTryFromGoValue(goVal)
			if err != nil {
				return nil, err
			}
			m.Set(KeyRefFromValue(key), val2)
		}
		return ValueFromIndexMap(m), nil
	default:
		ty := reflect.TypeOf(v)
		k := ty.Kind()
		switch k {
		case reflect.Struct:
			return ValueFromObject(StructObjectWithReflect(reflect.ValueOf(v))), nil
		case reflect.Array, reflect.Slice:
			return ValueFromObject(sqeObjectFromGoReflectSeq(reflect.ValueOf(v))), nil
		case reflect.Map:
			return valueTryFromGoMapReflect(reflect.ValueOf(v))
		case reflect.Ptr:
			return ValueTryFromGoValue(reflect.ValueOf(v).Elem().Interface())
		}
		return nil, fmt.Errorf("unsupported type: %T, ty=%+v, kind=%s", val, ty, k)
	}
}

type reflectStructObject struct {
	val reflect.Value
}

var _ = (Object)(reflectStructObject{})
var _ = (StructObject)(reflectStructObject{})

func StructObjectWithReflect(val reflect.Value) reflectStructObject {
	return reflectStructObject{val: val}
}

func (reflectStructObject) Kind() ObjectKind { return ObjectKindStruct }

func (o reflectStructObject) GetField(name string) option.Option[Value] {
	ty := o.val.Type()
	f, ok := ty.FieldByName(name)
	if !ok {
		return option.None[Value]()
	}
	fv := o.val.FieldByIndex(f.Index)
	val, err := ValueTryFromGoValue(fv.Interface())
	if err != nil {
		val = InvalidValue{Detail: err.Error()}
	}
	return option.Some(val)
}

func (o reflectStructObject) StaticFields() option.Option[[]string] {
	ty := o.val.Type()
	n := ty.NumField()
	var fields []string
	for i := 0; i < n; i++ {
		f := ty.Field(i)
		if f.IsExported() {
			fields = append(fields, f.Name)
		}
	}
	return option.Some(fields)
}

func (o reflectStructObject) Fields() []string { return nil }

type reflectSeqObject struct {
	val reflect.Value
}

var _ = (Object)(reflectSeqObject{})
var _ = (SeqObject)(reflectSeqObject{})

func sqeObjectFromGoReflectSeq(val reflect.Value) reflectStructObject {
	return reflectStructObject{val: val}
}

func (reflectSeqObject) Kind() ObjectKind { return ObjectKindSeq }

func (o reflectSeqObject) GetItem(idx uint) option.Option[Value] {
	if idx >= o.ItemCount() {
		return option.None[Value]()
	}
	val, err := ValueTryFromGoValue(o.val.Index(int(idx)))
	if err != nil {
		val = InvalidValue{Detail: err.Error()}
	}
	return option.Some(val)
}

func (o reflectSeqObject) ItemCount() uint {
	return uint(o.val.Len())
}

func valueTryFromGoMapReflect(val reflect.Value) (Value, error) {
	m := NewIndexMap()
	iter := val.MapRange()
	for iter.Next() {
		key, err := ValueTryFromGoValue(iter.Key())
		if err != nil {
			return nil, err
		}
		v, err := ValueTryFromGoValue(iter.Value())
		if err != nil {
			return nil, err
		}
		m.Set(KeyRefFromValue(key), v)
	}
	return ValueFromIndexMap(m), nil
}
