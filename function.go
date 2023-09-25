package mjingo

import (
	"fmt"
	"reflect"

	"github.com/hnakamur/mjingo/internal/rustfmt"
	"github.com/hnakamur/mjingo/option"
)

type namedBoxedFunc struct {
	name string
	fn   BoxedFunc
}

// BoxedFunc is the type of a boxed function.
//
// A boxed function can be registered as global functions to the environment via
// [Environment.AddFunction].
//
// A boxed function can be created by wrapping a "unboxed" function using one of BoxedFuncFrom* function.
// Or a user can define a boxed function directly.
// In that case, [ConvertArgToGoValue] and [ConvertVariadicArgsToGoValue] can be used to convert a Value
// to Go's data type.
type BoxedFunc = func(*State, []Value) (Value, error)

// 1 argument functions

// BoxedFuncFromFixedArity1ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity1ArgNoErrFunc[A JustOneArgTypes, R RetValTypes](f func(A) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret := f(a)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromFixedArity1ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity1ArgWithErrFunc[A JustOneArgTypes, R RetValTypes](f func(A) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret, err := f(a)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic1ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic1ArgNoErrFunc[A VariadicArgElemTypes, R RetValTypes](f func(...A) R) BoxedFunc {
	return func(_state *State, values []Value) (Value, error) {
		a, err := ConvertVariadicArgsToGoValue[[]A, A](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic1ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic1ArgWithErrFunc[A VariadicArgElemTypes, R RetValTypes](f func(...A) (R, error)) BoxedFunc {
	return func(_state *State, values []Value) (Value, error) {
		a, err := ConvertVariadicArgsToGoValue[[]A, A](values)
		if err != nil {
			return Value{}, err
		}
		ret, err := f(a...)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// 2 argument functions

// BoxedFuncFromFixedArity2ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity2ArgNoErrFunc[A FirstArgTypes, B FixedArityLastArgTypes, R RetValTypes](f func(A, B) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret := f(a, b)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromFixedArity2ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity2ArgWithErrFunc[A FirstArgTypes, B FixedArityLastArgTypes, R RetValTypes](f func(A, B) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret, err := f(a, b)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic2ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic2ArgNoErrFunc[A FirstArgTypes, B VariadicArgElemTypes, R RetValTypes](f func(A, ...B) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, err := ConvertVariadicArgsToGoValue[[]B, B](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic2ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic2ArgWithErrFunc[A FirstArgTypes, B VariadicArgElemTypes, R RetValTypes](f func(A, ...B) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, err := ConvertVariadicArgsToGoValue[[]B, B](values)
		if err != nil {
			return Value{}, err
		}
		ret, err := f(a, b...)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// 3 argument functions

// BoxedFuncFromFixedArity3ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity3ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C FixedArityLastArgTypes, R RetValTypes](f func(A, B, C) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret := f(a, b, c)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromFixedArity3ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity3ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C FixedArityLastArgTypes, R RetValTypes](f func(A, B, C) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret, err := f(a, b, c)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic3ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic3ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C VariadicArgElemTypes, R RetValTypes](f func(A, B, ...C) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, err := ConvertVariadicArgsToGoValue[[]C, C](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b, c...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic3ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic3ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C VariadicArgElemTypes, R RetValTypes](f func(A, B, ...C) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, err := ConvertVariadicArgsToGoValue[[]C, C](values)
		if err != nil {
			return Value{}, err
		}
		ret, err := f(a, b, c...)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// 4 argument functions

// BoxedFuncFromFixedArity4ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity4ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D FixedArityLastArgTypes, R RetValTypes](f func(A, B, C, D) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret := f(a, b, c, d)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromFixedArity4ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity4ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D FixedArityLastArgTypes, R RetValTypes](f func(A, B, C, D) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret, err := f(a, b, c, d)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic4ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic4ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D VariadicArgElemTypes, R RetValTypes](f func(A, B, C, ...D) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, err := ConvertVariadicArgsToGoValue[[]D, D](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b, c, d...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic4ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic4ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D VariadicArgElemTypes, R RetValTypes](f func(A, B, C, ...D) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, err := ConvertVariadicArgsToGoValue[[]D, D](values)
		if err != nil {
			return Value{}, err
		}
		ret, err := f(a, b, c, d...)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// 5 argument functions

// BoxedFuncFromFixedArity5ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity5ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E FixedArityLastArgTypes, R RetValTypes](f func(A, B, C, D, E) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return Value{}, err
		}
		e, values, err := ConvertArgToGoValue[E](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret := f(a, b, c, d, e)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromFixedArity5ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromFixedArity5ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E FixedArityLastArgTypes, R RetValTypes](f func(A, B, C, D, E) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return Value{}, err
		}
		e, values, err := ConvertArgToGoValue[E](state, values)
		if err != nil {
			return Value{}, err
		}
		if len(values) > 0 {
			return Value{}, NewError(TooManyArguments, "")
		}
		ret, err := f(a, b, c, d, e)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic5ArgNoErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic5ArgNoErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E VariadicArgElemTypes, R RetValTypes](f func(A, B, C, D, ...E) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return Value{}, err
		}
		e, err := ConvertVariadicArgsToGoValue[[]E, E](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b, c, d, e...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromVariadic5ArgWithErrFunc creates a boxed function which wraps f.
func BoxedFuncFromVariadic5ArgWithErrFunc[A FirstArgTypes, B MiddleArgTypes, C MiddleArgTypes, D MiddleArgTypes, E VariadicArgElemTypes, R RetValTypes](f func(A, B, C, D, ...E) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, values, err := ConvertArgToGoValue[B](state, values)
		if err != nil {
			return Value{}, err
		}
		c, values, err := ConvertArgToGoValue[C](state, values)
		if err != nil {
			return Value{}, err
		}
		d, values, err := ConvertArgToGoValue[D](state, values)
		if err != nil {
			return Value{}, err
		}
		e, err := ConvertVariadicArgsToGoValue[[]E, E](values)
		if err != nil {
			return Value{}, err
		}
		ret, err := f(a, b, c, d, e...)
		if err != nil {
			return Value{}, err
		}
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

// BoxedFuncFromFuncReflect creates a boxed function which wraps f using Go's reflect package.
//
// This may be slower so caller should prefer generic BoxedFuncFrom* functions.
func BoxedFuncFromFuncReflect(fn any) BoxedFunc {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("argument must be a function")
	}

	numOut := fnType.NumOut()
	if numOut != 1 && numOut != 2 {
		panic("return value count must be 1 or 2")
	}
	if !canConvertibleToValue(fnType.Out(0)) {
		panic("first return value type is unsupported")
	}
	if numOut == 2 {
		assertType[error](fnType.Out(1), "type of second return value must be error")
	}

	variadic := fnType.IsVariadic()
	argTypes := buildArgTypesOfFunc(fn)
	if err := checkArgTypes(argTypes, variadic); err != nil {
		panic(err.Error())
	}
	fnVal := reflect.ValueOf(fn)
	return func(state *State, values []Value) (Value, error) {
		goVals, err := argsToGoValuesReflect(state, values, argTypes, variadic)
		if err != nil {
			return Value{}, err
		}
		reflectVals := make([]reflect.Value, len(goVals))
		for i, goVal := range goVals {
			reflectVals[i] = reflect.ValueOf(goVal)
		}
		var retVals []reflect.Value
		if variadic {
			retVals = fnVal.CallSlice(reflectVals)
		} else {
			retVals = fnVal.Call(reflectVals)
		}
		switch len(retVals) {
		case 1:
			return ValueFromGoValue(retVals[0].Interface()), nil
		case 2:
			retVal0 := ValueFromGoValue(retVals[0].Interface())
			retVal1 := retVals[1].Interface()
			if retVal1 != nil {
				return retVal0, retVal1.(error)
			}
			return retVal0, nil
		}
		panic("unreachable")
	}
}

func valueFromBoxedFunc(name string, f BoxedFunc) Value {
	return ValueFromObject(funcObject{name: name, f: f})
}

type funcObject struct {
	name string
	f    BoxedFunc
}

var _ = (Object)(funcObject{})
var _ = (Caller)(funcObject{})
var _ = (rustfmt.Formatter)(funcObject{})

func (funcObject) Kind() ObjectKind { return ObjectKindPlain }

func (fo funcObject) Call(state *State, args []Value) (Value, error) {
	return fo.f(state, args)
}

// SupportRustFormat implements rustfmt.Formatter.
func (funcObject) SupportsCustomVerb(verb rune) bool {
	return verb == rustfmt.DebugVerb || verb == rustfmt.DisplayVerb
}

// Format implements rustfmt.Formatter.
func (fo funcObject) Format(f fmt.State, verb rune) {
	switch verb {
	case rustfmt.DisplayVerb, rustfmt.DebugVerb:
		fmt.Fprintf(f, "<function %s>", fo.name)
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods funcObject
		type funcObject hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), funcObject(fo))
	}
}

func rangeFunc(lower uint32, upper, step option.Option[uint32]) ([]uint32, error) {
	var iUpper uint32
	if upper.IsSome() {
		iUpper = upper.Unwrap()
	} else {
		iUpper = lower
		lower = 0
	}

	iStep := uint32(1)
	if step.IsSome() {
		iStep = step.Unwrap()
		if iStep == 0 {
			return nil, NewError(InvalidOperation, "cannot create range with step of 0")
		}
	}

	n := (iUpper - lower) / iStep
	if n > 10000 {
		return nil, NewError(InvalidOperation, "range has too many elements")
	}

	rv := make([]uint32, 0, n)
	for i := lower; i < iUpper; i += iStep {
		rv = append(rv, i)
	}
	return rv, nil
}

func dictFunc(val Value) (Value, error) {
	switch v := val.data.(type) {
	case undefinedValue:
		return valueFromIndexMap(newValueMap()), nil
	case mapValue:
		return valueFromIndexMap(v.Map), nil
	}
	return Value{}, NewError(InvalidOperation, "")
}
