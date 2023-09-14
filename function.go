package mjingo

import (
	"reflect"

	"github.com/hnakamur/mjingo/option"
)

type BoxedFunc = func(*State, []Value) (Value, error)

// 1 argument functions

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

func BoxedFuncFromVariadic1ArgNoErrFunc[A VariadicArgElemTypes, R RetValTypes](f func(...A) R) BoxedFunc {
	return func(_state *State, values []Value) (Value, error) {
		a, err := ConvertArgToGoValueVariadic[[]A, A](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

func BoxedFuncFromVariadic1ArgWithErrFunc[A VariadicArgElemTypes, R RetValTypes](f func(...A) (R, error)) BoxedFunc {
	return func(_state *State, values []Value) (Value, error) {
		a, err := ConvertArgToGoValueVariadic[[]A, A](values)
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

func BoxedFuncFromVariadic2ArgNoErrFunc[A FirstArgTypes, B VariadicArgElemTypes, R RetValTypes](f func(A, ...B) R) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, err := ConvertArgToGoValueVariadic[[]B, B](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

func BoxedFuncFromVariadic2ArgWithErrFunc[A FirstArgTypes, B VariadicArgElemTypes, R RetValTypes](f func(A, ...B) (R, error)) BoxedFunc {
	return func(state *State, values []Value) (Value, error) {
		a, values, err := ConvertArgToGoValue[A](state, values)
		if err != nil {
			return Value{}, err
		}
		b, err := ConvertArgToGoValueVariadic[[]B, B](values)
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
		c, err := ConvertArgToGoValueVariadic[[]C, C](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b, c...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

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
		c, err := ConvertArgToGoValueVariadic[[]C, C](values)
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
		d, err := ConvertArgToGoValueVariadic[[]D, D](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b, c, d...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

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
		d, err := ConvertArgToGoValueVariadic[[]D, D](values)
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
		e, err := ConvertArgToGoValueVariadic[[]E, E](values)
		if err != nil {
			return Value{}, err
		}
		ret := f(a, b, c, d, e...)
		retVal := ValueFromGoValue(ret)
		return retVal, nil
	}
}

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
		e, err := ConvertArgToGoValueVariadic[[]E, E](values)
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

func valueFromBoxedFunc(f BoxedFunc) Value {
	return valueFromObject(funcObject{f: f})
}

type funcObject struct{ f BoxedFunc }

var _ = (Object)(funcObject{})
var _ = (Caller)(funcObject{})

func (funcObject) Kind() ObjectKind { return ObjectKindPlain }

func (f funcObject) Call(state *State, args []Value) (Value, error) {
	return f.f(state, args)
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
