package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

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

type ValueFromGoValueOption func(*valueFromGoValueConfig)

type valueFromGoValueConfig struct {
	structTag string
}

func WithStructTag(tag string) ValueFromGoValueOption {
	return func(cfg *valueFromGoValueConfig) {
		cfg.structTag = tag
	}
}

func ValueFromGoValue(val any, opts ...ValueFromGoValueOption) Value {
	var config valueFromGoValueConfig
	for _, opt := range opts {
		opt(&config)
	}
	return valueFromGoValueHelper(val, &config, 0)
}

const maxNestLevelForValueFromGoValue = 100

func valueFromGoValueHelper(val any, config *valueFromGoValueConfig, level uint) Value {
	if level >= maxNestLevelForValueFromGoValue {
		return InvalidValue{Detail: "nested level too deep"}
	}
	switch v := val.(type) {
	case bool:
		return mapErrToInvalidValue(serializeBool(v))
	case uint8:
		return mapErrToInvalidValue(serializeU8(v))
	case uint16:
		return mapErrToInvalidValue(serializeU16(v))
	case uint32:
		return mapErrToInvalidValue(serializeU32(v))
	case uint64:
		return mapErrToInvalidValue(serializeU64(v))
	case uint:
		return mapErrToInvalidValue(serializeUint(v))
	case int8:
		return mapErrToInvalidValue(serializeI8(v))
	case int16:
		return mapErrToInvalidValue(serializeI16(v))
	case int32:
		return mapErrToInvalidValue(serializeI32(v))
	case int64:
		return mapErrToInvalidValue(serializeI64(v))
	case int:
		return mapErrToInvalidValue(serializeInt(v))
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			f, err := v.Float64()
			if err != nil {
				return mapErrToInvalidValue(nil, errors.New("invalid json.Number"))
			}
			return mapErrToInvalidValue(serializeF64(f))
		}
		return mapErrToInvalidValue(serializeI64(n))
	case big.Int:
		if isI128(&v) {
			return mapErrToInvalidValue(serializeI128(v))
		}
		return mapErrToInvalidValue(serializeU128(v))
	case float32:
		return mapErrToInvalidValue(serializeF32(v))
	case float64:
		return mapErrToInvalidValue(serializeF64(v))
	case string:
		return mapErrToInvalidValue(serializeStr(v))
	case nil:
		return mapErrToInvalidValue(serializeNone())
	default:
		ty := reflect.TypeOf(v)
		k := ty.Kind()
		switch k {
		case reflect.Struct:
			return ValueFromObject(structObjectWithReflect(reflect.ValueOf(v), config, level))
		case reflect.Array, reflect.Slice:
			return ValueFromObject(sqeObjectFromGoReflectSeq(reflect.ValueOf(v), config, level))
		case reflect.Map:
			return valueFromGoMapReflect(reflect.ValueOf(v), config, level)
		case reflect.Ptr:
			return valueFromGoValueHelper(reflect.ValueOf(v).Elem().Interface(), config, level+1)
		}
		return mapErrToInvalidValue(nil, fmt.Errorf("unsupported type: %T, ty=%+v, kind=%s", val, ty, k))
	}
}

func mapErrToInvalidValue(val Value, err error) Value {
	if err != nil {
		return InvalidValue{Detail: err.Error()}
	}
	return val
}

func valueFromGoMapReflect(val reflect.Value, config *valueFromGoValueConfig, level uint) Value {
	m := NewIndexMap()
	iter := val.MapRange()
	for iter.Next() {
		key := valueFromGoValueHelper(iter.Key().Interface(), config, level+1)
		v := valueFromGoValueHelper(iter.Value().Interface(), config, level+1)
		m.Set(KeyRefFromValue(key), v)
	}
	return ValueFromIndexMap(m)
}

type reflectStructObject struct {
	val            reflect.Value
	config         *valueFromGoValueConfig
	level          uint
	fieldNames     []string
	nameToFieldIdx map[string]int
}

var _ = (Object)((*reflectStructObject)(nil))
var _ = (StructObject)((*reflectStructObject)(nil))

func structObjectWithReflect(val reflect.Value, config *valueFromGoValueConfig, level uint) *reflectStructObject {
	return &reflectStructObject{val: val, config: config, level: level}
}

func (*reflectStructObject) Kind() ObjectKind { return ObjectKindStruct }

func (o *reflectStructObject) GetField(name string) option.Option[Value] {
	o.collectFieldNames()
	idx, ok := o.nameToFieldIdx[name]
	if !ok {
		return option.None[Value]()
	}
	fv := o.val.Field(idx)
	val := valueFromGoValueHelper(fv.Interface(), o.config, o.level+1)
	return option.Some(val)
}

func (o *reflectStructObject) StaticFields() option.Option[[]string] {
	o.collectFieldNames()
	return option.Some(o.fieldNames)
}

func (o *reflectStructObject) collectFieldNames() {
	if o.nameToFieldIdx == nil {
		o.nameToFieldIdx = make(map[string]int)
		ty := o.val.Type()
		n := ty.NumField()
		for i := 0; i < n; i++ {
			f := ty.Field(i)
			if f.IsExported() {
				name := o.keyNameForField(f)
				o.fieldNames = append(o.fieldNames, name)
				o.nameToFieldIdx[name] = i
			}
		}
	}
}

func (o *reflectStructObject) keyNameForField(f reflect.StructField) string {
	if o.config.structTag != "" {
		if tagVal, ok := f.Tag.Lookup(o.config.structTag); ok {
			nameInTag, _, _ := strings.Cut(tagVal, ",")
			if nameInTag != "" {
				return nameInTag
			}
		}
	}
	return f.Name
}

func (o *reflectStructObject) Fields() []string { return nil }

type reflectSeqObject struct {
	val    reflect.Value
	config *valueFromGoValueConfig
	level  uint
}

var _ = (Object)((*reflectSeqObject)(nil))
var _ = (SeqObject)((*reflectSeqObject)(nil))

func sqeObjectFromGoReflectSeq(val reflect.Value, config *valueFromGoValueConfig, level uint) *reflectSeqObject {
	return &reflectSeqObject{val: val, config: config, level: level}
}

func (*reflectSeqObject) Kind() ObjectKind { return ObjectKindSeq }

func (o *reflectSeqObject) GetItem(idx uint) option.Option[Value] {
	if idx >= o.ItemCount() {
		return option.None[Value]()
	}
	val := valueFromGoValueHelper(o.val.Index(int(idx)).Interface(), o.config, o.level+1)
	return option.Some(val)
}

func (o *reflectSeqObject) ItemCount() uint {
	return uint(o.val.Len())
}
