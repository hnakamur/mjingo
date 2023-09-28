package mjingo

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/hnakamur/mjingo/internal/rustfmt"
	"github.com/hnakamur/mjingo/option"
)

func serializeBool(v bool) (Value, error) {
	return valueFromBool(v), nil
}

func serializeI8(v int8) (Value, error) {
	return valueFromI64(int64(v)), nil
}

func serializeI16(v int16) (Value, error) {
	return valueFromI64(int64(v)), nil
}

func serializeI32(v int32) (Value, error) {
	return valueFromI64(int64(v)), nil
}

func serializeI64(v int64) (Value, error) {
	return valueFromI64(v), nil
}

func serializeInt(v int) (Value, error) {
	return valueFromI64(int64(v)), nil
}

func serializeI128(v I128) (Value, error) {
	return valueFromI128(v), nil
}

func serializeU8(v uint8) (Value, error) {
	return valueFromU64(uint64(v)), nil
}

func serializeU16(v uint16) (Value, error) {
	return valueFromU64(uint64(v)), nil
}

func serializeU32(v uint32) (Value, error) {
	return valueFromU64(uint64(v)), nil
}

func serializeU64(v uint64) (Value, error) {
	return valueFromU64(v), nil
}

func serializeUint(v uint) (Value, error) {
	return valueFromU64(uint64(v)), nil
}

func serializeU128(v U128) (Value, error) {
	return valueFromU128(v), nil
}

func serializeF32(v float32) (Value, error) {
	return valueFromF64(float64(v)), nil
}

func serializeF64(v float64) (Value, error) {
	return valueFromF64(v), nil
}

func serializeRune(v rune) (Value, error) {
	return valueFromString(string(v)), nil
}

func serializeStr(v string) (Value, error) {
	return valueFromString(v), nil
}

func serializeBytes(v []byte) (Value, error) {
	return valueFromBytes(v), nil
}

func serializeNone() (Value, error) {
	return none, nil
}

// ValueFromGoValueOption is the option type to configure the behavior of
// [ValueFromGoValue].
type ValueFromGoValueOption func(*valueFromGoValueConfig)

type valueFromGoValueConfig struct {
	structTag string
}

// WithStructTag sets the struct tag which is used to reference a struct field.
// If a struct tag value exists with the specified tag name, the value before comma
// is used as a field name instead.
func WithStructTag(tag string) ValueFromGoValueOption {
	return func(cfg *valueFromGoValueConfig) {
		cfg.structTag = tag
	}
}

// ValueFromGoValue creates a value from a Go value.
//
// Supported scalar types are bool, uint8, uint16, uint32, uint64, uint, int8, int16,
// int32, int64, int, json.Number, I128, U128, float32, float64, string, nil, Value.
//
// And struct, slice, pointer, and map of these types are supported.
func ValueFromGoValue(val any, opts ...ValueFromGoValueOption) Value {
	var config valueFromGoValueConfig
	for _, opt := range opts {
		opt(&config)
	}
	return valueFromGoValueHelper(val, &config, 0)
}

const maxNestLevelForValueFromGoValue = 100

func canConvertibleToValue(ty reflect.Type) bool {
	switch ty.Kind() {
	case reflect.Bool, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uint, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Int, reflect.Float32, reflect.Float64, reflect.String,
		reflect.Struct, reflect.Array, reflect.Slice, reflect.Map, reflect.Ptr:
		return true
	case reflect.Interface:
		return ty == reflectType[Object]()
	default:
		return false
	}
}

func valueFromGoValueHelper(val any, config *valueFromGoValueConfig, level uint) Value {
	if level >= maxNestLevelForValueFromGoValue {
		return Value{data: invalidValue{Detail: "nested level too deep"}}
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
				return mapErrToInvalidValue(Value{}, errors.New("invalid json.Number"))
			}
			return mapErrToInvalidValue(serializeF64(f))
		}
		return mapErrToInvalidValue(serializeI64(n))
	case *I128:
		return mapErrToInvalidValue(serializeI128(*v))
	case *U128:
		return mapErrToInvalidValue(serializeU128(*v))
	case I128:
		return mapErrToInvalidValue(serializeI128(v))
	case U128:
		return mapErrToInvalidValue(serializeU128(v))
	case float32:
		return mapErrToInvalidValue(serializeF32(v))
	case float64:
		return mapErrToInvalidValue(serializeF64(v))
	// case rune: // Cannot do this because of "duplicate case rune in type switch"
	// 	return mapErrToInvalidValue(serializeRune(v))
	case string:
		return mapErrToInvalidValue(serializeStr(v))
	case nil:
		return mapErrToInvalidValue(serializeNone())
	case Value:
		return v
	case Object:
		return ValueFromObject(v)
	case map[string]any:
		return valueFromStrKeyGoMap(v)
	case []Value:
		return valueFromSlice(v)
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
		return mapErrToInvalidValue(Value{}, fmt.Errorf("unsupported type: %T, ty=%+v, kind=%s", val, ty, k))
	}
}

func valueFromStrKeyGoMap[V any](m map[string]V) Value {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	vm := newValueMap()
	for _, key := range keys {
		valVal := ValueFromGoValue(m[key])
		vm.Set(keyRefFromString(key), valVal)
	}
	return valueFromIndexMap(vm)
}

func valueFromGoMap[K interface {
	comparable
	cmp.Ordered
}, V any](m map[K]V) Value {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	vm := newValueMap()
	for _, key := range keys {
		keyVal := ValueFromGoValue(key)
		valVal := ValueFromGoValue(m[key])
		vm.Set(keyRefFromValue(keyVal), valVal)
	}
	return valueFromIndexMap(vm)
}

func mapErrToInvalidValue(val Value, err error) Value {
	if err != nil {
		return Value{data: invalidValue{Detail: err.Error()}}
	}
	return val
}

func valueFromGoMapReflect(val reflect.Value, config *valueFromGoValueConfig, level uint) Value {
	entries := make([]valueMapEntry, 0, val.Len())
	for iter := val.MapRange(); iter.Next(); {
		key := valueFromGoValueHelper(iter.Key().Interface(), config, level+1)
		v := valueFromGoValueHelper(iter.Value().Interface(), config, level+1)
		entries = append(entries, valueMapEntry{Key: keyRefFromValue(key), Value: v})
	}
	slices.SortFunc(entries, func(a, b valueMapEntry) int {
		return keyRefCmp(a.Key, b.Key)
	})

	return valueFromIndexMap(valueMapFromEntries(entries))
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

func (s *reflectStructObject) GetField(name string) option.Option[Value] {
	s.collectFieldNames()
	idx, ok := s.nameToFieldIdx[name]
	if !ok {
		return option.None[Value]()
	}
	fv := s.val.Field(idx)
	val := valueFromGoValueHelper(fv.Interface(), s.config, s.level+1)
	return option.Some(val)
}

func (s *reflectStructObject) StaticFields() option.Option[[]string] {
	s.collectFieldNames()
	return option.Some(s.fieldNames)
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

func (s *reflectStructObject) keyNameForField(f reflect.StructField) string {
	if s.config.structTag != "" {
		if tagVal, ok := f.Tag.Lookup(s.config.structTag); ok {
			nameInTag, _, _ := strings.Cut(tagVal, ",")
			if nameInTag != "" {
				return nameInTag
			}
		}
	}
	return f.Name
}

func (*reflectStructObject) Fields() []string { return nil }

type reflectSeqObject struct {
	val    reflect.Value
	config *valueFromGoValueConfig
	level  uint
}

var _ Object = (*reflectSeqObject)(nil)
var _ SeqObject = (*reflectSeqObject)(nil)
var _ rustfmt.Formatter = (*reflectSeqObject)(nil)

func sqeObjectFromGoReflectSeq(val reflect.Value, config *valueFromGoValueConfig, level uint) *reflectSeqObject {
	return &reflectSeqObject{val: val, config: config, level: level}
}

func (*reflectSeqObject) Kind() ObjectKind { return ObjectKindSeq }

func (s *reflectSeqObject) GetItem(idx uint) option.Option[Value] {
	if idx >= s.ItemCount() {
		return option.None[Value]()
	}
	val := valueFromGoValueHelper(s.val.Index(int(idx)).Interface(), s.config, s.level+1)
	return option.Some(val)
}

func (s *reflectSeqObject) ItemCount() uint {
	return uint(s.val.Len())
}

func (*reflectSeqObject) SupportsCustomVerb(verb rune) bool {
	return verb == rustfmt.DebugVerb || verb == rustfmt.DisplayVerb
}

func (s *reflectSeqObject) Format(f fmt.State, verb rune) {
	switch verb {
	case rustfmt.DisplayVerb, rustfmt.DebugVerb:
		l := s.ItemCount()
		items := make([]any, l)
		for i := uint(0); i < l; i++ {
			items[i] = s.GetItem(i).Unwrap()
		}
		rustfmt.NewDebugList(items).Format(f, verb)
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods reflectSeqObject
		type reflectSeqObject hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), reflectSeqObject(*s))
	}
}
