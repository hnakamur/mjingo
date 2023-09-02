package mjingo

import (
	"fmt"
	"math/big"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
)

func valueFromBytes(val []byte) Value {
	return bytesValue{B: val}
}

func valueFromString(val string) Value {
	return stringValue{Str: val, Type: stringTypeNormal}
}

// ValueFromSafeString creates a value from a safe string.
//
// A safe string is one that will bypass auto escaping.  For instance if you
// want to have the template engine render some HTML without the user having to
// supply the `|safe` filter, you can use a value of this type instead.
func ValueFromSafeString(s string) Value {
	return stringValue{Str: s, Type: stringTypeSafe}
}

func valueFromBool(val bool) Value {
	return boolValue{B: val}
}

func valueFromI64(n int64) Value {
	return i64Value{N: n}
}

func valueFromI128(n big.Int) Value {
	return i128Value{N: n}
}

func valueFromU64(n uint64) Value {
	return u64Value{N: n}
}

func valueFromU128(n big.Int) Value {
	return u128Value{N: n}
}

func valueFromF64(f float64) Value {
	return f64Value{F: f}
}

func valueFromSlice(values []Value) Value {
	return seqValue{Items: values}
}

func valueFromIndexMap(m *valueMap) Value {
	return mapValue{Map: m, Type: mapTypeNormal}
}

func valueFromKwargs(a kwArgs) Value {
	return mapValue{Map: &a.Values, Type: mapTypeKwargs}
}

func valueFromObject(dy object) Value {
	return dynamicValue{Dy: dy}
}

func stringFromValue(val option.Option[Value]) (string, error) {
	if val.IsSome() {
		optStr := val.Unwrap().asStr()
		if optStr.IsSome() {
			return optStr.Unwrap(), nil
		}
		return "", newError(InvalidOperation, "value is not a string")
	}
	return "", newError(MissingArgument, "")
}

func stringTryFromValue(val Value) (string, error) {
	if v, ok := val.(stringValue); ok {
		return v.Str, nil
	}
	return "", newError(InvalidOperation, "value is not a string")
}

type kwArgs struct {
	Values valueMap
	Used   hashset.StrHashSet
}

func newKwArgs(m valueMap) kwArgs {
	return kwArgs{
		Values: m,
		Used:   *hashset.NewStrHashSet(),
	}
}

func kwArgsTryFromValue(val Value) (kwArgs, error) {
	switch v := val.(type) {
	case undefinedValue:
		return newKwArgs(*newValueMap()), nil
	case mapValue:
		if v.Type == mapTypeKwargs {
			return newKwArgs(*v.Map.Clone()), nil
		}
	}
	return kwArgs{}, newError(InvalidOperation, "")
}

// Get a single argument from the kwargs but don't mark it as used.
func (a *kwArgs) PeekValue(key string) option.Option[Value] {
	val, ok := a.Values.Get(keyRefFromString(key))
	if ok {
		return option.Some(val)
	}
	return option.None[Value]()
}

// Gets a single argument from the kwargs and marks it as used.
func (a *kwArgs) GetValue(key string) option.Option[Value] {
	optVal := a.PeekValue(key)
	if optVal.IsSome() {
		a.Used.Add(key)
	}
	return optVal
}

// Asserts that all kwargs were used.
func (a *kwArgs) AssertAllUsed() error {
	for _, keyRf := range a.Values.Keys() {
		if optKey := keyRf.AsStr(); optKey.IsSome() {
			key := optKey.Unwrap()
			if !a.Used.Contains(key) {
				return newError(TooManyArguments,
					fmt.Sprintf("unknown keyword argument '%s'", key))
			}
		} else {
			return newError(InvalidOperation, "non string keys passed to kwargs")
		}
	}
	return nil
}
