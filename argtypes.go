package mjingo

import (
	"fmt"
	"math/big"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
)

// type argsFromValuesFn[T any] func(state *virtualMachineState, args []value) (T, error)

type Unit struct{}

func ValueFromBytes(val []byte) Value {
	return BytesValue{B: val}
}

func ValueFromString(val string) Value {
	return StringValue{Str: val, Type: StringTypeNormal}
}

// Creates a value from a safe string.
//
// A safe string is one that will bypass auto escaping.  For instance if you
// want to have the template engine render some HTML without the user having to
// supply the `|safe` filter, you can use a value of this type instead.
func ValueFromSafeString(s string) Value {
	return StringValue{Str: s, Type: StringTypeSafe}
}

func ValueFromBool(val bool) Value {
	return BoolValue{B: val}
}

func ValueFromI64(n int64) Value {
	return I64Value{N: n}
}

func ValueFromI128(n big.Int) Value {
	return I128Value{N: n}
}

func ValueFromU64(n uint64) Value {
	return U64Value{N: n}
}

func ValueFromU128(n big.Int) Value {
	return U128Value{N: n}
}

func ValueFromF64(f float64) Value {
	return F64Value{F: f}
}

func ValueFromSlice(values []Value) Value {
	return SeqValue{Items: values}
}

func ValueFromIndexMap(m *ValueMap) Value {
	return MapValue{Map: m, Type: MapTypeNormal}
}

func ValueFromKwargs(a Kwargs) Value {
	return MapValue{Map: &a.Values, Type: MapTypeKwargs}
}

func ValueFromObject(dy Object) Value {
	return DynamicValue{Dy: dy}
}

func StringFromValue(val option.Option[Value]) (string, error) {
	if val.IsSome() {
		optStr := val.Unwrap().AsStr()
		if optStr.IsSome() {
			return optStr.Unwrap(), nil
		}
		return "", NewError(InvalidOperation, "value is not a string")
	}
	return "", NewError(MissingArgument, "")
}

func StringTryFromValue(val Value) (string, error) {
	if v, ok := val.(StringValue); ok {
		return v.Str, nil
	}
	return "", NewError(InvalidOperation, "value is not a string")
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
	case UndefinedValue:
		return NewKwargs(*NewValueMap()), nil
	case MapValue:
		if v.Type == MapTypeKwargs {
			return NewKwargs(*v.Map.Clone()), nil
		}
	}
	return Kwargs{}, NewError(InvalidOperation, "")
}

// Get a single argument from the kwargs but don't mark it as used.
func (a *Kwargs) PeekValue(key string) option.Option[Value] {
	val, ok := a.Values.Get(KeyRefFromString(key))
	if ok {
		return option.Some(val)
	}
	return option.None[Value]()
}

// Gets a single argument from the kwargs and marks it as used.
func (a *Kwargs) GetValue(key string) option.Option[Value] {
	optVal := a.PeekValue(key)
	if optVal.IsSome() {
		a.Used.Add(key)
	}
	return optVal
}

// Asserts that all kwargs were used.
func (a *Kwargs) AssertAllUsed() error {
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
