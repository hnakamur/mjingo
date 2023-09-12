package mjingo

import (
	"reflect"

	"github.com/hnakamur/mjingo/option"
)

type BoxedFunc = func(State, []Value) (Value, error)

func BoxedFuncFromFunc(fn any) BoxedFunc {
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

	argTypes := buildArgTypesOfFunc(fn)
	if err := checkArgTypes(argTypes); err != nil {
		panic(err.Error())
	}
	fnVal := reflect.ValueOf(fn)
	return func(state State, values []Value) (Value, error) {
		goVals, err := argsToGoValuesReflect(state, values, argTypes)
		if err != nil {
			return nil, err
		}
		reflectVals := make([]reflect.Value, len(goVals))
		for i, goVal := range goVals {
			if fnType.IsVariadic() && i == fnType.NumIn()-1 {
				reflectVals[i] = reflect.ValueOf(goVal).Convert(sliceTypeForRestTypeReflect(argTypes[i]))
			} else {
				reflectVals[i] = reflect.ValueOf(goVal)
			}
		}
		var retVals []reflect.Value
		if fnType.IsVariadic() {
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

func (f funcObject) Call(state State, args []Value) (Value, error) {
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
	switch v := val.(type) {
	case undefinedValue:
		return valueFromIndexMap(newValueMap()), nil
	case mapValue:
		return valueFromIndexMap(v.Map), nil
	}
	return nil, NewError(InvalidOperation, "")
}
