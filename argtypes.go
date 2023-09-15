package mjingo

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/option"
)

type JustOneArgTypes interface {
	ScalarTypes | OptionalTypes | SliceTypes | *State | Kwargs
}
type FirstArgTypes interface {
	ScalarTypes | OptionalTypes | SliceTypes | *State
}
type MiddleArgTypes interface {
	ScalarTypes | OptionalTypes | SliceTypes
}
type FixedArityLastArgTypes interface {
	ScalarTypes | OptionalTypes | SliceTypes | Kwargs
}
type VariadicArgElemTypes interface {
	ScalarTypes
}

type RetValTypes interface {
	ScalarTypes | SliceTypes
}

type ScalarTypes interface {
	Value | bool | uint8 | uint16 | uint32 | uint64 | uint |
		int8 | int16 | int32 | int64 | int | I128 | U128 |
		float32 | float64 | string
}

type SliceTypes interface {
	[]Value | []bool | []uint8 | []uint16 | []uint32 | []uint64 | []uint |
		[]int8 | []int16 | []int32 | []int64 | []int | []I128 | []U128 |
		[]float32 | []float64 | []string
}

type OptionalTypes interface {
	option.Option[Value] | option.Option[bool] | option.Option[uint8] |
		option.Option[uint16] | option.Option[uint32] | option.Option[uint64] |
		option.Option[uint] | option.Option[int8] | option.Option[int16] |
		option.Option[int32] | option.Option[int64] | option.Option[int] |
		option.Option[big.Int] | option.Option[float32] | option.Option[float64] |
		option.Option[string]
}

func valueFromBytes(val []byte) Value {
	return Value{data: bytesValue{B: val}}
}

func valueFromString(val string) Value {
	return Value{data: stringValue{Str: val, Type: stringTypeNormal}}
}

// ValueFromSafeString creates a value from a safe string.
//
// A safe string is one that will bypass auto escaping.  For instance if you
// want to have the template engine render some HTML without the user having to
// supply the `|safe` filter, you can use a value of this type instead.
func ValueFromSafeString(s string) Value {
	return Value{data: stringValue{Str: s, Type: stringTypeSafe}}
}

func valueFromBool(val bool) Value {
	return Value{data: boolValue{B: val}}
}

func valueFromI64(n int64) Value {
	return Value{data: i64Value{N: n}}
}

func valueFromI128(n I128) Value {
	return Value{data: i128Value{N: n}}
}

func valueFromU64(n uint64) Value {
	return Value{data: u64Value{N: n}}
}

func valueFromU128(n U128) Value {
	return Value{data: u128Value{N: n}}
}

func valueFromF64(f float64) Value {
	return Value{data: f64Value{F: f}}
}

func valueFromSlice(values []Value) Value {
	return Value{data: seqValue{Items: values}}
}

func valueFromIndexMap(m *valueMap) Value {
	return Value{data: mapValue{Map: m, Type: mapTypeNormal}}
}

func valueFromKwargs(a Kwargs) Value {
	return Value{data: mapValue{Map: &a.values, Type: mapTypeKwargs}}
}

func valueFromObject(dy Object) Value {
	return Value{data: dynamicValue{Dy: dy}}
}

// Kwargs is the utility to accept keyword arguments.
//
// Keyword arguments are represented as regular values as the last argument
// in an argument list.  This can be quite complex to use manually so this
// type is added as a utility.  You can use [Kwargs.GetValue] to fetch a
// single keyword argument and then use [Kwargs.AssertAllUsed]
// to make sure extra arguments create an error.
type Kwargs struct {
	values valueMap
	used   hashset.StrHashSet
}

func newKwargs(m valueMap) Kwargs {
	return Kwargs{
		values: m,
		used:   *hashset.NewStrHashSet(),
	}
}

func valueTryToKwargs(val Value) (Kwargs, error) {
	if val.data == nil {
		return newKwargs(*newValueMap()), nil
	}
	switch v := val.data.(type) {
	case undefinedValue:
		return newKwargs(*newValueMap()), nil
	case mapValue:
		if v.Type == mapTypeKwargs {
			return newKwargs(*v.Map.Clone()), nil
		}
	}
	return Kwargs{}, NewError(InvalidOperation, "")
}

// PeekValue gets a single argument from the kwargs but don't mark it as used.
// The caller can convert a Value to a Go value with [ValueTryToGoValue].
func (a *Kwargs) PeekValue(key string) option.Option[Value] {
	val, ok := a.values.Get(keyRefFromString(key))
	if ok {
		return option.Some(val)
	}
	return option.None[Value]()
}

// GetValue gets a single argument from the kwargs and marks it as used.
// The caller can convert a Value to a Go value with [ValueTryToGoValue].
func (a *Kwargs) GetValue(key string) option.Option[Value] {
	optVal := a.PeekValue(key)
	if optVal.IsSome() {
		a.used.Add(key)
	}
	return optVal
}

// AssertAllUsed asserts that all kwargs were used.
func (a *Kwargs) AssertAllUsed() error {
	for _, keyRf := range a.values.Keys() {
		if optKey := keyRf.AsStr(); optKey.IsSome() {
			key := optKey.Unwrap()
			if !a.used.Contains(key) {
				return NewError(TooManyArguments,
					fmt.Sprintf("unknown keyword argument '%s'", key))
			}
		} else {
			return NewError(InvalidOperation, "non string keys passed to kwargs")
		}
	}
	return nil
}

func valueTryToGoUint(val Value) (uint, error) {
	return val.tryToUint()
}

func valueTryToValueSlice(val Value) ([]Value, error) {
	iter, err := val.tryIter()
	if err != nil {
		return nil, err
	}
	return iter.Collect(), nil
}

func ConvertArgToGoValue[T JustOneArgTypes](state *State, values []Value) (T, []Value, error) {
	var v T
	values, err := convertArgToGoVarTo(state, values, &v)
	return v, values, err
}

func convertArgToGoVarTo(state *State, values []Value, destPtr any) ([]Value, error) {
	switch p := destPtr.(type) {
	case **State:
		*p = state
		return values, nil
	case *Value:
		return convertArgToGoVarHelper[Value](values, p, valueTryToValue)
	case *bool:
		return convertArgToGoVarHelper[bool](values, p, valueTryToGoBool)
	case *int8:
		return convertArgToGoVarHelper[int8](values, p, valueTryToGoInt8)
	case *int16:
		return convertArgToGoVarHelper[int16](values, p, valueTryToGoInt16)
	case *int32:
		return convertArgToGoVarHelper[int32](values, p, valueTryToGoInt32)
	case *int64:
		return convertArgToGoVarHelper[int64](values, p, valueTryToGoInt64)
	case *int:
		return convertArgToGoVarHelper[int](values, p, valueTryToGoInt)
	case *uint8:
		return convertArgToGoVarHelper[uint8](values, p, valueTryToGoUint8)
	case *uint16:
		return convertArgToGoVarHelper[uint16](values, p, valueTryToGoUint16)
	case *uint32:
		return convertArgToGoVarHelper[uint32](values, p, valueTryToGoUint32)
	case *uint64:
		return convertArgToGoVarHelper[uint64](values, p, valueTryToGoUint64)
	case *uint:
		return convertArgToGoVarHelper[uint](values, p, valueTryToGoUint)
	case *float32:
		return convertArgToGoVarHelper[float32](values, p, valueTryToGoFloat32)
	case *float64:
		return convertArgToGoVarHelper[float64](values, p, valueTryToGoFloat64)
	case *string:
		return convertArgToGoVarHelper[string](values, p, valueTryToGoString)
	case *Kwargs: // Kwargs are taken from the last element
		if len(values) == 0 {
			*p = newKwargs(*newValueMap())
			return values, nil
		}
		kwargs, err := valueTryToKwargs(values[len(values)-1])
		if err != nil {
			return nil, err
		}
		*p = kwargs
		return values[:len(values)-1], nil
	case *option.Option[Value]:
		return convertArgToGoOptionVarHelper[Value](values, p, valueTryToValue)
	case *option.Option[bool]:
		return convertArgToGoOptionVarHelper[bool](values, p, valueTryToGoBool)
	case *option.Option[int8]:
		return convertArgToGoOptionVarHelper[int8](values, p, valueTryToGoInt8)
	case *option.Option[int16]:
		return convertArgToGoOptionVarHelper[int16](values, p, valueTryToGoInt16)
	case *option.Option[int32]:
		return convertArgToGoOptionVarHelper[int32](values, p, valueTryToGoInt32)
	case *option.Option[int64]:
		return convertArgToGoOptionVarHelper[int64](values, p, valueTryToGoInt64)
	case *option.Option[int]:
		return convertArgToGoOptionVarHelper[int](values, p, valueTryToGoInt)
	case *option.Option[uint8]:
		return convertArgToGoOptionVarHelper[uint8](values, p, valueTryToGoUint8)
	case *option.Option[uint16]:
		return convertArgToGoOptionVarHelper[uint16](values, p, valueTryToGoUint16)
	case *option.Option[uint32]:
		return convertArgToGoOptionVarHelper[uint32](values, p, valueTryToGoUint32)
	case *option.Option[uint64]:
		return convertArgToGoOptionVarHelper[uint64](values, p, valueTryToGoUint64)
	case *option.Option[uint]:
		return convertArgToGoOptionVarHelper[uint](values, p, valueTryToGoUint)
	case *option.Option[float32]:
		return convertArgToGoOptionVarHelper[float32](values, p, valueTryToGoFloat32)
	case *option.Option[float64]:
		return convertArgToGoOptionVarHelper[float64](values, p, valueTryToGoFloat64)
	case *option.Option[string]:
		return convertArgToGoOptionVarHelper[string](values, p, valueTryToGoString)
	case *[]Value:
		return convertArgToGoSliceVarHelper[Value](values, p, valueSliceTryToGoSlice[[]Value, Value])
	case *[]bool:
		return convertArgToGoSliceVarHelper[bool](values, p, valueSliceTryToGoSlice[[]bool, bool])
	case *[]int8:
		return convertArgToGoSliceVarHelper[int8](values, p, valueSliceTryToGoSlice[[]int8, int8])
	case *[]int16:
		return convertArgToGoSliceVarHelper[int16](values, p, valueSliceTryToGoSlice[[]int16, int16])
	case *[]int32:
		return convertArgToGoSliceVarHelper[int32](values, p, valueSliceTryToGoSlice[[]int32, int32])
	case *[]int64:
		return convertArgToGoSliceVarHelper[int64](values, p, valueSliceTryToGoSlice[[]int64, int64])
	case *[]int:
		return convertArgToGoSliceVarHelper[int](values, p, valueSliceTryToGoSlice[[]int, int])
	case *[]uint8:
		return convertArgToGoSliceVarHelper[uint8](values, p, valueSliceTryToGoSlice[[]uint8, uint8])
	case *[]uint16:
		return convertArgToGoSliceVarHelper[uint16](values, p, valueSliceTryToGoSlice[[]uint16, uint16])
	case *[]uint32:
		return convertArgToGoSliceVarHelper[uint32](values, p, valueSliceTryToGoSlice[[]uint32, uint32])
	case *[]uint64:
		return convertArgToGoSliceVarHelper[uint64](values, p, valueSliceTryToGoSlice[[]uint64, uint64])
	case *[]uint:
		return convertArgToGoSliceVarHelper[uint](values, p, valueSliceTryToGoSlice[[]uint, uint])
	case *[]float32:
		return convertArgToGoSliceVarHelper[float32](values, p, valueSliceTryToGoSlice[[]float32, float32])
	case *[]float64:
		return convertArgToGoSliceVarHelper[float64](values, p, valueSliceTryToGoSlice[[]float64, float64])
	case *[]string:
		return convertArgToGoSliceVarHelper[string](values, p, valueSliceTryToGoSlice[[]string, string])
	}
	panic(fmt.Sprintf("unsupported go variable type: %T", destPtr))
}

func ConvertArgToGoValueVariadic[S ~[]E, E ScalarTypes](values []Value) (S, error) {
	var s S
	if err := convertArgToGoVariadicVarTo(values, &s); err != nil {
		return s, err
	}
	return s, nil
}

func convertArgToGoVariadicVarTo(values []Value, destPtr any) error {
	switch p := destPtr.(type) {
	case *[]Value:
		return convertArgToGoVariadicVarHelper[Value](values, p, valueSliceTryToGoSlice[[]Value, Value])
	case *[]bool:
		return convertArgToGoVariadicVarHelper[bool](values, p, valueSliceTryToGoSlice[[]bool, bool])
	case *[]int8:
		return convertArgToGoVariadicVarHelper[int8](values, p, valueSliceTryToGoSlice[[]int8, int8])
	case *[]int16:
		return convertArgToGoVariadicVarHelper[int16](values, p, valueSliceTryToGoSlice[[]int16, int16])
	case *[]int32:
		return convertArgToGoVariadicVarHelper[int32](values, p, valueSliceTryToGoSlice[[]int32, int32])
	case *[]int64:
		return convertArgToGoVariadicVarHelper[int64](values, p, valueSliceTryToGoSlice[[]int64, int64])
	case *[]int:
		return convertArgToGoVariadicVarHelper[int](values, p, valueSliceTryToGoSlice[[]int, int])
	case *[]uint8:
		return convertArgToGoVariadicVarHelper[uint8](values, p, valueSliceTryToGoSlice[[]uint8, uint8])
	case *[]uint16:
		return convertArgToGoVariadicVarHelper[uint16](values, p, valueSliceTryToGoSlice[[]uint16, uint16])
	case *[]uint32:
		return convertArgToGoVariadicVarHelper[uint32](values, p, valueSliceTryToGoSlice[[]uint32, uint32])
	case *[]uint64:
		return convertArgToGoVariadicVarHelper[uint64](values, p, valueSliceTryToGoSlice[[]uint64, uint64])
	case *[]uint:
		return convertArgToGoVariadicVarHelper[uint](values, p, valueSliceTryToGoSlice[[]uint, uint])
	case *[]float32:
		return convertArgToGoVariadicVarHelper[float32](values, p, valueSliceTryToGoSlice[[]float32, float32])
	case *[]float64:
		return convertArgToGoVariadicVarHelper[float64](values, p, valueSliceTryToGoSlice[[]float64, float64])
	case *[]string:
		return convertArgToGoVariadicVarHelper[string](values, p, valueSliceTryToGoSlice[[]string, string])
	}
	panic(fmt.Sprintf("unsupported go variable type: %T", destPtr))
}

func convertArgToGoVarHelper[T any](values []Value, dest *T, f func(Value) (T, error)) ([]Value, error) {
	if len(values) == 0 {
		return nil, NewError(MissingArgument, "")
	}
	v, err := f(values[0])
	if err != nil {
		return nil, err
	}
	*dest = v
	return values[1:], nil
}

func convertArgToGoOptionVarHelper[T any](values []Value, dest *option.Option[T], f func(Value) (T, error)) ([]Value, error) {
	if len(values) == 0 {
		*dest = option.None[T]()
		return values, nil
	}
	v, err := f(values[0])
	if err != nil {
		return nil, err
	}
	*dest = option.Some(v)
	return values[1:], nil
}

func convertArgToGoSliceVarHelper[T ScalarTypes](values []Value, dest *[]T, f func([]Value) ([]T, error)) ([]Value, error) {
	if len(values) == 0 {
		return nil, NewError(MissingArgument, "")
	}
	argVals, err := valueTryToValueSlice(values[0])
	if err != nil {
		return nil, err
	}
	v, err := f(argVals)
	if err != nil {
		return nil, err
	}
	*dest = v
	return values[1:], nil
}

func convertArgToGoVariadicVarHelper[T ScalarTypes](values []Value, dest *[]T, f func([]Value) ([]T, error)) error {
	v, err := f(values)
	if err != nil {
		return err
	}
	*dest = v
	return nil
}

func valueSliceTryToGoSlice[S ~[]E, E ScalarTypes](values []Value) (S, error) {
	slice := make(S, len(values))
	for i, val := range values {
		if err := valueTryToGoValueNoReflect(val, &slice[i]); err != nil {
			return nil, err
		}
	}
	return slice, nil
}

func argsToGoValuesNoReflect(state *State, values []Value, destPtrs []any, variadic bool) error {
	i := 0
	for j, destPtr := range destPtrs {
		kind := findArgTypeKindFromDestPtr(destPtr)
		switch kind {
		case argTypeKindState:
			*(destPtr.(**State)) = state
		case argTypeKindPrimitive:
			if i >= len(values) {
				return NewError(MissingArgument, "")
			}
			err := valueTryToGoValueNoReflect(values[i], destPtr)
			if err != nil {
				return err
			}
			i++
		case argTypeKindKwargs:
			if i < len(values) {
				goVal, err := valueTryToKwargs(values[i])
				if err != nil {
					return err
				}
				*(destPtr.(*Kwargs)) = goVal
				i++
			} else {
				*(destPtr.(*Kwargs)) = newKwargs(*newValueMap())
			}
		case argTypeKindOption:
			if i < len(values) {
				if err := valueTryToOptionValueNoReflect(values[i], destPtr); err != nil {
					return err
				}
				i++
			} else {
				// Do nothing since None is zero value of option.Option[T].
				// Caller must prepare zero value at destPtr.
			}
		case argTypeKindSlice:
			if variadic && j == len(destPtrs)-1 {
				if err := valueSliceTryToGoSliceTo(values[i:], destPtr); err != nil {
					return err
				}
				i = len(values)
			} else {
				if i >= len(values) {
					return NewError(MissingArgument, "")
				}
				argVals, err := valueTryToValueSlice(values[i])
				if err != nil {
					return err
				}
				if err := valueSliceTryToGoSliceTo(argVals, destPtr); err != nil {
					return err
				}
				i++
			}
		}
	}
	if i < len(values) {
		return NewError(TooManyArguments, "")
	}
	return nil
}

func argsToGoValuesReflect(state *State, values []Value, argTypes []reflect.Type, variadic bool) ([]any, error) {
	var goVals []any
	i := 0
	for j, argType := range argTypes {
		kind := findArgTypeKind(argType)
		switch kind {
		case argTypeKindState:
			goVals = append(goVals, state)
		case argTypeKindPrimitive:
			if i >= len(values) {
				return nil, NewError(MissingArgument, "")
			}
			goVal, err := valueTryToGoValueReflect(values[i], argType)
			if err != nil {
				return nil, err
			}
			goVals = append(goVals, goVal)
			i++
		case argTypeKindKwargs:
			var goVal Kwargs
			if i < len(values) {
				var err error
				goVal, err = valueTryToKwargs(values[i])
				if err != nil {
					return nil, err
				}
				i++
			} else {
				goVal = newKwargs(*newValueMap())
			}
			goVals = append(goVals, goVal)
		case argTypeKindOption:
			var goVal any
			if i < len(values) {
				var err error
				goVal, err = valueTryToGoValueReflect(values[i], argType)
				if err != nil {
					return nil, err
				}
				i++
			} else {
				// None is zero value of option.Option[T].
				goVal = reflect.Zero(argType).Interface()
			}
			goVals = append(goVals, goVal)
		case argTypeKindSlice:
			if variadic && j == len(argTypes)-1 {
				goVal, err := valueSliceTryToGoSliceReflect(values[i:], argType)
				if err != nil {
					return nil, err
				}
				goVals = append(goVals, goVal)
				i = len(values)
			} else {
				if i >= len(values) {
					return nil, NewError(MissingArgument, "")
				}
				argVals, err := valueTryToValueSlice(values[i])
				if err != nil {
					return nil, err
				}
				goVal, err := valueSliceTryToGoSliceReflect(argVals, argType)
				if err != nil {
					return nil, err
				}
				goVals = append(goVals, goVal)
				i++
			}
		}
	}
	if i < len(values) {
		return nil, NewError(TooManyArguments, "")
	}
	return goVals, nil
}

func buildArgTypesOfFunc(fn any) []reflect.Type {
	typ := reflect.TypeOf(fn)
	numIn := typ.NumIn()
	argTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		argTypes[i] = typ.In(i)
	}
	return argTypes
}

func checkArgTypes(argTypes []reflect.Type, variadic bool) error {
	seenOptional := false
	for i, argType := range argTypes {
		kind := findArgTypeKind(argType)
		switch kind {
		case argTypeKindState:
			if i != 0 {
				return NewError(InvalidOperation,
					"argument of State type must be the first argument")
			}
		case argTypeKindPrimitive:
			if seenOptional {
				return NewError(InvalidOperation,
					"argument of non-optional type cannot be after argument of optional type")
			}
		case argTypeKindSlice:
			if seenOptional && !(variadic && i == len(argTypes)-1) {
				return NewError(InvalidOperation,
					"argument of non-optional type cannot be after argument of optional type")
			}
		case argTypeKindOption:
			seenOptional = true
		case argTypeKindKwargs:
			if i != len(argTypes)-1 {
				return NewError(InvalidOperation,
					fmt.Sprintf("argument of %s type must be the last argument", kind))
			}
		default:
			return NewError(InvalidOperation, fmt.Sprintf("argument type %T is unsupported", reflect.Zero(argType).Interface()))
		}
	}
	return nil
}

type argTypeKind int

const (
	argTypeKindUnsupported argTypeKind = iota
	argTypeKindState
	argTypeKindPrimitive
	argTypeKindSlice
	argTypeKindOption
	argTypeKindKwargs
)

func (k argTypeKind) String() string {
	switch k {
	case argTypeKindState:
		return "State"
	case argTypeKindPrimitive:
		return "primitive"
	case argTypeKindSlice:
		return "slice"
	case argTypeKindOption:
		return "Option"
	case argTypeKindKwargs:
		return "Kwargs"
	}
	return "unsupported"
}

func findArgTypeKindFromDestPtr(destPtr any) argTypeKind {
	switch destPtr.(type) {
	case **State:
		return argTypeKindState
	case *Value, *bool, *int8,
		*int16, *int32, *int64,
		*int, *uint8, *uint16,
		*uint32, *uint64, *uint,
		*float32, *float64, *string:
		return argTypeKindPrimitive
	case *[]Value, *[]bool, *[]int8,
		*[]int16, *[]int32, *[]int64,
		*[]int, *[]uint8, *[]uint16,
		*[]uint32, *[]uint64, *[]uint,
		*[]float32, *[]float64, *[]string:
		return argTypeKindSlice
	case *option.Option[Value], *option.Option[bool],
		*option.Option[int8], *option.Option[int16],
		*option.Option[int32], *option.Option[int64],
		*option.Option[int], *option.Option[uint8],
		*option.Option[uint16], *option.Option[uint32],
		*option.Option[uint64], *option.Option[uint],
		*option.Option[float32], *option.Option[float64],
		*option.Option[string]:
		return argTypeKindOption
	case *Kwargs:
		return argTypeKindKwargs
	}
	return argTypeKindUnsupported
}

func findArgTypeKind(argType reflect.Type) argTypeKind {
	switch argType {
	case reflectType[*State]():
		return argTypeKindState
	case reflectType[Value](), reflectType[bool](), reflectType[int8](),
		reflectType[int16](), reflectType[int32](), reflectType[int64](),
		reflectType[int](), reflectType[uint8](), reflectType[uint16](),
		reflectType[uint32](), reflectType[uint64](), reflectType[uint](),
		reflectType[float32](), reflectType[float64](), reflectType[string]():
		return argTypeKindPrimitive
	case reflectType[[]Value](), reflectType[[]bool](), reflectType[[]int8](),
		reflectType[[]int16](), reflectType[[]int32](), reflectType[[]int64](),
		reflectType[[]int](), reflectType[[]uint8](), reflectType[[]uint16](),
		reflectType[[]uint32](), reflectType[[]uint64](), reflectType[[]uint](),
		reflectType[[]float32](), reflectType[[]float64](), reflectType[[]string]():
		return argTypeKindSlice
	case reflectType[option.Option[Value]](), reflectType[option.Option[bool]](),
		reflectType[option.Option[int8]](), reflectType[option.Option[int16]](),
		reflectType[option.Option[int32]](), reflectType[option.Option[int64]](),
		reflectType[option.Option[int]](), reflectType[option.Option[uint8]](),
		reflectType[option.Option[uint16]](), reflectType[option.Option[uint32]](),
		reflectType[option.Option[uint64]](), reflectType[option.Option[uint]](),
		reflectType[option.Option[float32]](), reflectType[option.Option[float64]](),
		reflectType[option.Option[string]]():
		return argTypeKindOption
	case reflectType[Kwargs]():
		return argTypeKindKwargs
	}
	return argTypeKindUnsupported
}
