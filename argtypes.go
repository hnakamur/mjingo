package mjingo

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/option"
)

type FirstArgTypes interface {
	ScalarTypes | SliceTypes | *State
}
type MiddleArgTypes interface {
	ScalarTypes | SliceTypes
}
type LastArgTypes interface {
	ScalarTypes | SliceTypes | Kwargs
}

type RetValTypes interface {
	ScalarTypes | SliceTypes
}

type ScalarTypes interface {
	Value | bool | uint8 | uint16 | uint32 | uint64 | uint |
		int8 | int16 | int32 | int64 | int | big.Int |
		string
}

type SliceTypes interface {
	[]Value | []bool | []uint8 | []uint16 | []uint32 | []uint64 | []uint |
		[]int8 | []int16 | []int32 | []int64 | []int | []big.Int |
		[]string
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

func valueFromI128(n big.Int) Value {
	return Value{data: i128Value{N: n}}
}

func valueFromU64(n uint64) Value {
	return Value{data: u64Value{N: n}}
}

func valueFromU128(n big.Int) Value {
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

func ArgsTo1GoValue[A any](state *State, values []Value) (A, error) {
	var a A
	goVals, err := argsToGoValuesReflect(state, values, []reflect.Type{reflectType[A]()})
	if err != nil {
		return a, err
	}
	if goVals[0] != nil {
		a = goVals[0].(A)
	}
	return a, nil
}

func ArgsTo2GoValues[A any, B any](state *State, values []Value) (A, B, error) {
	var a A
	var b B
	goVals, err := argsToGoValuesReflect(state, values,
		[]reflect.Type{reflectType[A](), reflectType[B]()})
	if err != nil {
		return a, b, err
	}
	if goVals[0] != nil {
		a = goVals[0].(A)
	}
	if goVals[1] != nil {
		b = goVals[1].(B)
	}
	return a, b, nil
}

func ArgsTo3GoValues[A any, B any, C any](state *State, values []Value) (A, B, C, error) {
	var a A
	var b B
	var c C
	goVals, err := argsToGoValuesReflect(state, values,
		[]reflect.Type{reflectType[A](), reflectType[B](), reflectType[C]()})
	if err != nil {
		return a, b, c, err
	}
	if goVals[0] != nil {
		a = goVals[0].(A)
	}
	if goVals[1] != nil {
		b = goVals[1].(B)
	}
	if goVals[2] != nil {
		c = goVals[2].(C)
	}
	return a, b, c, nil
}

func ArgsTo4GoValues[A any, B any, C any, D any](state *State, values []Value) (A, B, C, D, error) {
	var a A
	var b B
	var c C
	var d D
	goVals, err := argsToGoValuesReflect(state, values,
		[]reflect.Type{reflectType[A](), reflectType[B](), reflectType[C](), reflectType[D]()})
	if err != nil {
		return a, b, c, d, err
	}
	if goVals[0] != nil {
		a = goVals[0].(A)
	}
	if goVals[1] != nil {
		b = goVals[1].(B)
	}
	if goVals[2] != nil {
		c = goVals[2].(C)
	}
	if goVals[3] != nil {
		d = goVals[3].(D)
	}
	return a, b, c, d, nil
}

func ArgsTo5GoValues[A any, B any, C any, D any, E any](state *State, values []Value) (A, B, C, D, E, error) {
	var a A
	var b B
	var c C
	var d D
	var e E
	goVals, err := argsToGoValuesReflect(state, values,
		[]reflect.Type{reflectType[A](), reflectType[B](), reflectType[C](), reflectType[D](),
			reflectType[E]()})
	if err != nil {
		return a, b, c, d, e, err
	}
	if goVals[0] != nil {
		a = goVals[0].(A)
	}
	if goVals[1] != nil {
		b = goVals[1].(B)
	}
	if goVals[2] != nil {
		c = goVals[2].(C)
	}
	if goVals[3] != nil {
		d = goVals[3].(D)
	}
	if goVals[4] != nil {
		e = goVals[4].(E)
	}
	return a, b, c, d, e, nil
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

func argsToGoValuesReflect(state *State, values []Value, argTypes []reflect.Type) ([]any, error) {
	var goVals []any
	i := 0
	for _, argType := range argTypes {
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
		case argTypeKindRest:
			sliceType := sliceTypeForRestTypeReflect(argType)
			sliceVal, err := valueSliceTryToGoSliceReflect(values[i:], sliceType)
			if err != nil {
				return nil, err
			}
			goVal := reflect.ValueOf(sliceVal).Convert(argType).Interface()
			goVals = append(goVals, goVal)
			i = len(values)
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
		if typ.IsVariadic() && i == numIn-1 {
			argTypes[i] = restTypeFromSliceTypeReflect(typ.In(i))
		} else {
			argTypes[i] = typ.In(i)
		}
	}
	return argTypes
}

func checkArgTypes(argTypes []reflect.Type) error {
	seenOptional := false
	for i, argType := range argTypes {
		kind := findArgTypeKind(argType)
		switch kind {
		case argTypeKindState:
			if i != 0 {
				return NewError(InvalidOperation,
					"argument of State type must be the first argument")
			}
		case argTypeKindPrimitive, argTypeKindSlice:
			if seenOptional {
				return NewError(InvalidOperation,
					"argument of non-optional type cannot be after argument of optional type")
			}
		case argTypeKindOption:
			seenOptional = true
		case argTypeKindRest, argTypeKindKwargs:
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

// Rest is a utility type to capture remaining arguments.
type Rest[T any] []T

type argTypeKind int

const (
	argTypeKindUnsupported argTypeKind = iota
	argTypeKindState
	argTypeKindPrimitive
	argTypeKindSlice
	argTypeKindOption
	argTypeKindRest
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
	case argTypeKindRest:
		return "Rest"
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
	case *Rest[Value], *Rest[bool],
		*Rest[int8], *Rest[int16],
		*Rest[int32], *Rest[int64],
		*Rest[int], *Rest[uint8],
		*Rest[uint16], *Rest[uint32],
		*Rest[uint64], *Rest[uint],
		*Rest[float32], *Rest[float64],
		*Rest[string]:
		return argTypeKindRest
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
	case reflectType[Rest[Value]](), reflectType[Rest[bool]](),
		reflectType[Rest[int8]](), reflectType[Rest[int16]](),
		reflectType[Rest[int32]](), reflectType[Rest[int64]](),
		reflectType[Rest[int]](), reflectType[Rest[uint8]](),
		reflectType[Rest[uint16]](), reflectType[Rest[uint32]](),
		reflectType[Rest[uint64]](), reflectType[Rest[uint]](),
		reflectType[Rest[float32]](), reflectType[Rest[float64]](),
		reflectType[Rest[string]]():
		return argTypeKindRest
	case reflectType[Kwargs]():
		return argTypeKindKwargs
	}
	return argTypeKindUnsupported
}

func sliceTypeForRestTypeReflect(typ reflect.Type) reflect.Type {
	switch typ {
	case reflectType[Rest[Value]]():
		return reflectType[[]Value]()
	case reflectType[Rest[bool]]():
		return reflectType[[]bool]()
	case reflectType[Rest[int8]]():
		return reflectType[[]int8]()
	case reflectType[Rest[int16]]():
		return reflectType[[]int16]()
	case reflectType[Rest[int32]]():
		return reflectType[[]int32]()
	case reflectType[Rest[int64]]():
		return reflectType[[]int64]()
	case reflectType[Rest[int]]():
		return reflectType[[]int]()
	case reflectType[Rest[uint8]]():
		return reflectType[[]uint8]()
	case reflectType[Rest[uint16]]():
		return reflectType[[]uint16]()
	case reflectType[Rest[uint32]]():
		return reflectType[[]uint32]()
	case reflectType[Rest[uint64]]():
		return reflectType[[]uint64]()
	case reflectType[Rest[uint]]():
		return reflectType[[]uint]()
	case reflectType[Rest[float32]]():
		return reflectType[[]float32]()
	case reflectType[Rest[float64]]():
		return reflectType[[]float64]()
	case reflectType[Rest[string]]():
		return reflectType[[]string]()
	}
	panic("unreachable")
}

func restTypeFromSliceTypeReflect(typ reflect.Type) reflect.Type {
	switch typ {
	case reflectType[[]Value]():
		return reflectType[Rest[Value]]()
	case reflectType[[]bool]():
		return reflectType[Rest[bool]]()
	case reflectType[[]int8]():
		return reflectType[Rest[int8]]()
	case reflectType[[]int16]():
		return reflectType[Rest[int16]]()
	case reflectType[[]int32]():
		return reflectType[Rest[int32]]()
	case reflectType[[]int64]():
		return reflectType[Rest[int64]]()
	case reflectType[[]int]():
		return reflectType[Rest[int]]()
	case reflectType[[]uint8]():
		return reflectType[Rest[uint8]]()
	case reflectType[[]uint16]():
		return reflectType[Rest[uint16]]()
	case reflectType[[]uint32]():
		return reflectType[Rest[uint32]]()
	case reflectType[[]uint64]():
		return reflectType[Rest[uint64]]()
	case reflectType[[]uint]():
		return reflectType[Rest[uint]]()
	case reflectType[[]float32]():
		return reflectType[Rest[float32]]()
	case reflectType[[]float64]():
		return reflectType[Rest[float64]]()
	case reflectType[[]string]():
		return reflectType[Rest[string]]()
	}
	panic("unreachable")
}
