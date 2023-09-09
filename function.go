package mjingo

import (
	"reflect"
	"slices"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type boxedFunc = func(*vmState, []Value) (Value, error)

type funcObject struct{ f boxedFunc }

func valueFromFunc(f boxedFunc) Value {
	return valueFromObject(funcObject{f: f})
}

func boxedFuncFromFunc(fn any) boxedFunc {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("argument must be a function")
	}

	numOut := fnType.NumOut()
	if numOut != 1 && numOut != 2 {
		panic("return value count must be 1 or 2")
	}
	if numOut == 2 {
		assertType[error](fnType.Out(1), "type of seond return value must be error")
	}

	numIn := fnType.NumIn()
	if numIn < 1 && numIn > 5 {
		panic("only functions with argument count between 1 and 5 are supported")
	}
	optCount := checkFuncArgTypes(fnType)

	fnVal := reflect.ValueOf(fn)
	return func(state *vmState, values []Value) (Value, error) {
		reflectVals := make([]reflect.Value, 0, numIn)
		inOffset := 0
		if fnType.In(0) == reflectType[*vmState]() {
			reflectVals = append(reflectVals, reflect.ValueOf(state))
			inOffset++
		}
		wantValuesLen := numIn - inOffset
		if fnType.IsVariadic() {
			wantValuesLen--
		}
		if len(values) < wantValuesLen-optCount {
			return nil, newError(MissingArgument, "")
		}
		if len(values) > wantValuesLen && !fnType.IsVariadic() {
			return nil, newError(TooManyArguments, "")
		}
		var inValues []Value
		if len(inValues) >= wantValuesLen {
			inValues = values
		} else {
			inValues = slices.Clone(values)
			for i := len(inValues); i < wantValuesLen; i++ {
				inValues = append(inValues, nil)
			}
		}

		for i, val := range inValues {
			var argType reflect.Type
			if fnType.IsVariadic() && i+inOffset >= numIn-1 {
				argType = fnType.In(numIn - 1).Elem()
			} else {
				argType = fnType.In(i + inOffset)
			}
			goVal, err := goValueFromValue(val, argType)
			if err != nil {
				return nil, err
			}
			reflectVals = append(reflectVals, reflect.ValueOf(goVal))
		}
		retVals := fnVal.Call(reflectVals)
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

var _ = (object)(funcObject{})
var _ = (caller)(funcObject{})

func (funcObject) Kind() objectKind { return objectKindPlain }

func (f funcObject) Call(state *vmState, args []Value) (Value, error) {
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
			return nil, newError(InvalidOperation, "cannot create range with step of 0")
		}
	}

	n := (iUpper - lower) / iStep
	if n > 10000 {
		return nil, newError(InvalidOperation, "range has too many elements")
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
	return nil, newError(InvalidOperation, "")
}
