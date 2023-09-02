package internal

import (
	"fmt"
	"math/big"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
)

// type argsFromValuesFn[T any] func(state *virtualMachineState, args []value) (T, error)

type Unit struct{}

func ValueFromBytes(val []byte) Value {
	return bytesValue{b: val}
}

func ValueFromString(val string) Value {
	return stringValue{str: val, strTyp: stringTypeNormal}
}

// Creates a value from a safe string.
//
// A safe string is one that will bypass auto escaping.  For instance if you
// want to have the template engine render some HTML without the user having to
// supply the `|safe` filter, you can use a value of this type instead.
func ValueFromSafeString(s string) Value {
	return stringValue{str: s, strTyp: stringTypeSafe}
}

func ValueFromUnit(_ Unit) Value {
	return None
}

func ValueFromBool(val bool) Value {
	return boolValue{b: val}
}

func ValueFromI64(n int64) Value {
	return i64Value{n: n}
}

func ValueFromI128(n big.Int) Value {
	return i128Value{n: n}
}

func ValueFromU64(n uint64) Value {
	return u64Value{n: n}
}

func ValueFromU128(n big.Int) Value {
	return u128Value{n: n}
}

func ValueFromF64(f float64) Value {
	return f64Value{f: f}
}

func ValueFromSlice(values []Value) Value {
	return seqValue{items: values}
}

func ValueFromIndexMap(m *ValueMap) Value {
	return mapValue{m: m, mapTyp: mapTypeNormal}
}

func ValueFromKwargs(a Kwargs) Value {
	return mapValue{m: &a.Values, mapTyp: mapTypeKwargs}
}

func ValueFromObject(dy Object) Value {
	return dynamicValue{dy: dy}
}

func ValueFromFunc(f FuncFunc) Value {
	return dynamicValue{dy: FuncObject{f: f}}
}

func StringFromValue(value option.Option[Value]) (string, error) {
	if value.IsSome() {
		optStr := value.Unwrap().AsStr()
		if optStr.IsSome() {
			return optStr.Unwrap(), nil
		}
		return "", NewError(InvalidOperation, "value is not a string")
	}
	return "", NewError(MissingArgument, "")
}

func StringTryFromValue(val Value) (string, error) {
	if v, ok := val.(stringValue); ok {
		return v.str, nil
	}
	return "", NewError(InvalidOperation, "value is not a string")
}

type rest[T any] struct {
	args []T
}

type Kwargs struct {
	Values ValueMap
	Used   hashset.StrHashSet
}

func NewKwargs(m ValueMap) Kwargs {
	return Kwargs{
		Values: m,
		Used:   *hashset.NewStrHashSet(),
	}
}

func KwargsTryFromValue(val Value) (Kwargs, error) {
	switch v := val.(type) {
	case undefinedValue:
		return NewKwargs(*NewValueMap()), nil
	case mapValue:
		if v.mapTyp == mapTypeKwargs {
			return NewKwargs(*v.m.Clone()), nil
		}
	}
	return Kwargs{}, NewError(InvalidOperation, "")
}

// Get a single argument from the kwargs but don't mark it as used.
func (a *Kwargs) peekValue(key string) option.Option[Value] {
	val, ok := a.Values.Get(KeyRefFromString(key))
	if ok {
		return option.Some(val)
	}
	return option.None[Value]()
}

// Gets a single argument from the kwargs and marks it as used.
func (a *Kwargs) getValue(key string) option.Option[Value] {
	optVal := a.peekValue(key)
	if optVal.IsSome() {
		a.Used.Add(key)
	}
	return optVal
}

// Asserts that all kwargs were used.
func (a *Kwargs) assertAllUsed() error {
	for _, keyRf := range a.Values.Keys() {
		if optKey := keyRf.AsStr(); optKey.IsSome() {
			key := optKey.Unwrap()
			if !a.Used.Contains(key) {
				return NewError(TooManyArguments,
					fmt.Sprintf("unknown keyword argument '%s'", key))
			}
		} else {
			return NewError(InvalidOperation, "non string keys passed to kwargs")
		}
	}
	return nil
}
