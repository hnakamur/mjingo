package internal

import (
	"errors"
	"fmt"
	"math/big"
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

// bool, for JSON booleans
// float64, for JSON numbers
// string, for JSON strings
// []interface{}, for JSON arrays
// map[string]interface{}, for JSON objects
// nil for JSON null

func ValueTryFromGoValue(val any) (Value, error) {
	switch v := val.(type) {
	case bool:
		return serializeBool(v)
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
		return nil, fmt.Errorf("unsupported type: %T", val)
	}
}
