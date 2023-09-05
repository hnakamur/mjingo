package mjingo

import (
	"reflect"
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type BoxedTest = func(*vmState, []Value) (bool, error)

func boxedTestFromFunc(fn any) BoxedTest {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("argument must be a function")
	}

	numOut := fnType.NumOut()
	if numOut != 1 && numOut != 2 {
		panic("return value count must be 1 or 2")
	}
	assertType(fnType.Out(0), (*bool)(nil), "type of first return value must be bool")
	if numOut == 2 {
		assertType(fnType.Out(1), (*error)(nil), "type of seond return value must be error")
	}

	numIn := fnType.NumIn()
	if numIn < 1 && numIn > 3 {
		panic("only functions with argument count between 1 and 3 are supported")
	}
	assertFuncArgTypes(fnType)

	fnVal := reflect.ValueOf(fn)
	return func(state *vmState, values []Value) (bool, error) {
		reflectVals := make([]reflect.Value, 0, numIn)
		inOffset := 0
		if fnType.In(0) == typeFromPtr((**vmState)(nil)) {
			reflectVals = append(reflectVals, reflect.ValueOf(state))
			inOffset++
		}
		wantValuesLen := numIn - inOffset
		if len(values) < wantValuesLen {
			return false, newError(MissingArgument, "")
		}
		if len(values) > wantValuesLen {
			return false, newError(TooManyArguments, "")
		}
		for i, val := range values {
			goVal, err := goValueFromValue(val, fnType.In(i+inOffset))
			if err != nil {
				return false, err
			}
			reflectVals = append(reflectVals, reflect.ValueOf(goVal))
		}
		ret := fnVal.Call(reflectVals)
		switch len(ret) {
		case 1:
			return ret[0].Interface().(bool), nil
		case 2:
			return ret[0].Interface().(bool), ret[1].Interface().(error)
		}
		panic("unreachable")
	}
}

type tuple1[A any] struct {
	a A
}
type tuple2[A any, B any] struct {
	a A
	b B
}
type tuple3[A any, B any, C any] struct {
	a A
	b B
	c C
}
type tuple4[A any, B any, C any, D any] struct {
	a A
	b B
	c C
	d D
}
type tuple5[A any, B any, C any, D any, E any] struct {
	a A
	b B
	c C
	d D
	e E
}

type argConvertResult[O any] struct {
	output   O
	consumed uint
}

type argType[O any] interface {
	fromValue(val option.Option[Value]) (O, error)
	fromStateAndValue(state *vmState, val option.Option[Value]) (argConvertResult[O], error)
	fromStateAndValues(state *vmState, values []Value, offset uint) (argConvertResult[O], error)
	isTrailing() bool
}

var _ = (argType[Value])(valueArgType{})

type valueArgType struct{}

func (valueArgType) fromValue(val option.Option[Value]) (Value, error) {
	if val.IsSome() {
		return val.Unwrap(), nil
	}
	return nil, newError(MissingArgument, "")
}

func (valueArgType) fromStateAndValue(state *vmState, val option.Option[Value]) (argConvertResult[Value], error) {
	return fromStateAndValue(state, val)
}

func (valueArgType) fromStateAndValues(state *vmState, values []Value, offset uint) (argConvertResult[Value], error) {
	return fromStateAndValues(state, values, offset)
}

func (valueArgType) isTrailing() bool { return false }

func fromStateAndValue(state *vmState, val option.Option[Value]) (argConvertResult[Value], error) {
	var zero argConvertResult[Value]
	if option.MapOr(val, false, isUndefined) && state != nil && state.undefinedBehavior() == UndefinedBehaviorStrict {
		return zero, newError(UndefinedError, "")
	}
	var o valueArgType
	out, err := o.fromValue(val)
	if err != nil {
		return zero, err
	}
	return argConvertResult[Value]{output: out, consumed: 1}, nil
}

func fromStateAndValues(state *vmState, values []Value, offset uint) (argConvertResult[Value], error) {
	var o valueArgType
	val := option.Option[Value]{}
	if offset < uint(len(values)) {
		val = option.Some(values[offset])
	}
	return o.fromStateAndValue(state, val)
}

func tuple1FromValues(state *vmState, values []Value) (tuple1[Value], error) {
	var zero tuple1[Value]
	var ao Value
	var at valueArgType
	idx := uint(0)
	restFirst := at.isTrailing() && len(values) != 0
	if restFirst {
		avo, err := at.fromStateAndValues(state, values, uint(len(values)-1))
		if err != nil {
			return zero, err
		}
		ao = avo.output
		values = values[:len(values)-int(avo.consumed)]
	}
	if !restFirst {
		avo, err := at.fromStateAndValues(state, values, idx)
		if err != nil {
			return zero, err
		}
		ao = avo.output
		idx += avo.consumed
	}
	if idx < uint(len(values)) {
		return zero, newError(TooManyArguments, "")
	}
	return tuple1[Value]{a: ao}, nil
}

func tuple2FromValues(state *vmState, values []Value) (tuple2[Value, Value], error) {
	var zero tuple2[Value, Value]
	var ao Value
	var bo Value
	var at valueArgType
	var bt valueArgType
	idx := uint(0)
	restFirst := bt.isTrailing() && len(values) != 0
	if restFirst {
		bvo, err := bt.fromStateAndValues(state, values, uint(len(values)-1))
		if err != nil {
			return zero, err
		}
		bo = bvo.output
		values = values[:len(values)-int(bvo.consumed)]
	}
	avo, err := at.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	ao = avo.output
	idx += avo.consumed
	if !restFirst {
		bvo, err := bt.fromStateAndValues(state, values, idx)
		if err != nil {
			return zero, err
		}
		bo = bvo.output
		idx += bvo.consumed
	}
	if idx < uint(len(values)) {
		return zero, newError(TooManyArguments, "")
	}
	return tuple2[Value, Value]{a: ao, b: bo}, nil
}

func tuple3FromValues(state *vmState, values []Value) (tuple3[Value, Value, Value], error) {
	var zero tuple3[Value, Value, Value]
	var ao Value
	var bo Value
	var co Value
	var at valueArgType
	var bt valueArgType
	var ct valueArgType
	idx := uint(0)
	restFirst := bt.isTrailing() && len(values) != 0
	if restFirst {
		cvo, err := ct.fromStateAndValues(state, values, uint(len(values)-1))
		if err != nil {
			return zero, err
		}
		co = cvo.output
		values = values[:len(values)-int(cvo.consumed)]
	}
	avo, err := at.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	ao = avo.output
	idx += avo.consumed
	bvo, err := bt.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	bo = bvo.output
	idx += bvo.consumed
	if !restFirst {
		cvo, err := ct.fromStateAndValues(state, values, idx)
		if err != nil {
			return zero, err
		}
		co = cvo.output
		idx += cvo.consumed
	}
	if idx < uint(len(values)) {
		return zero, newError(TooManyArguments, "")
	}
	return tuple3[Value, Value, Value]{a: ao, b: bo, c: co}, nil
}

func tuple4FromValues(state *vmState, values []Value) (tuple4[Value, Value, Value, Value], error) {
	var zero tuple4[Value, Value, Value, Value]
	var ao Value
	var bo Value
	var co Value
	var do Value
	var at valueArgType
	var bt valueArgType
	var ct valueArgType
	var dt valueArgType
	idx := uint(0)
	restFirst := bt.isTrailing() && len(values) != 0
	if restFirst {
		dvo, err := dt.fromStateAndValues(state, values, uint(len(values)-1))
		if err != nil {
			return zero, err
		}
		do = dvo.output
		values = values[:len(values)-int(dvo.consumed)]
	}
	avo, err := at.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	ao = avo.output
	idx += avo.consumed
	bvo, err := bt.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	bo = bvo.output
	idx += bvo.consumed
	cvo, err := ct.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	co = cvo.output
	idx += cvo.consumed
	if !restFirst {
		dvo, err := dt.fromStateAndValues(state, values, idx)
		if err != nil {
			return zero, err
		}
		do = dvo.output
		idx += dvo.consumed
	}
	if idx < uint(len(values)) {
		return zero, newError(TooManyArguments, "")
	}
	return tuple4[Value, Value, Value, Value]{a: ao, b: bo, c: co, d: do}, nil
}

func tuple5FromValues(state *vmState, values []Value) (tuple5[Value, Value, Value, Value, Value], error) {
	var zero tuple5[Value, Value, Value, Value, Value]
	var ao Value
	var bo Value
	var co Value
	var do Value
	var eo Value
	var at valueArgType
	var bt valueArgType
	var ct valueArgType
	var dt valueArgType
	var et valueArgType
	idx := uint(0)
	restFirst := bt.isTrailing() && len(values) != 0
	if restFirst {
		evo, err := et.fromStateAndValues(state, values, uint(len(values)-1))
		if err != nil {
			return zero, err
		}
		eo = evo.output
		values = values[:len(values)-int(evo.consumed)]
	}
	avo, err := at.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	ao = avo.output
	idx += avo.consumed
	bvo, err := bt.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	bo = bvo.output
	idx += bvo.consumed
	cvo, err := ct.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	co = cvo.output
	idx += cvo.consumed
	dvo, err := dt.fromStateAndValues(state, values, idx)
	if err != nil {
		return zero, err
	}
	do = dvo.output
	idx += cvo.consumed
	if !restFirst {
		evo, err := et.fromStateAndValues(state, values, idx)
		if err != nil {
			return zero, err
		}
		eo = evo.output
		idx += evo.consumed
	}
	if idx < uint(len(values)) {
		return zero, newError(TooManyArguments, "")
	}
	return tuple5[Value, Value, Value, Value, Value]{a: ao, b: bo, c: co, d: do, e: eo}, nil
}

func isUndefined(val Value) bool {
	return val.isUndefined()
}

func isDefined(val Value) bool {
	return !val.isUndefined()
}

func isNone(val Value) bool {
	return val.isNone()
}

func isSafe(val Value) bool {
	return val.isSafe()
}

// Checks if a value is odd.
//
// ```jinja
// {{ 41 is odd }} -> true
// ```
func isOdd(val Value) bool {
	x, err := val.tryToI64()
	if err != nil {
		return false
	}
	return x%2 != 0
}

// Checks if a value is even.
//
// ```jinja
// {{ 42 is even }} -> true
// ```
func isEven(val Value) bool {
	x, err := val.tryToI64()
	if err != nil {
		return false
	}
	return x%2 == 0
}

// Checks if this value is a number.
//
// ```jinja
// {{ 42 is number }} -> true
// {{ "42" is number }} -> false
// ```
func isNumber(val Value) bool { return val.kind() == valueKindNumber }

// Checks if this value is a string.
//
// ```jinja
// {{ "42" is string }} -> true
// {{ 42 is string }} -> false
// ```
func isString(val Value) bool { return val.kind() == valueKindString }

// Checks if this value is a sequence
//
// ```jinja
// {{ [1, 2, 3] is sequence }} -> true
// {{ 42 is sequence }} -> false
// ```
func isSequence(val Value) bool { return val.kind() == valueKindSeq }

// Checks if this value is a mapping
//
// ```jinja
// {{ {"foo": "bar"} is mapping }} -> true
// {{ [1, 2, 3] is mapping }} -> false
// ```
func isMapping(val Value) bool { return val.kind() == valueKindMap }

// Checks if the value is starting with a string.
//
// ```jinja
// {{ "foobar" is startingwith("foo") }} -> true
// {{ "foobar" is startingwith("bar") }} -> false
// ```
func isStartingWith(v, other string) bool { return strings.HasPrefix(v, other) }

// Checks if the value is ending with a string.
//
// ```jinja
// {{ "foobar" is endingwith("bar") }} -> true
// {{ "foobar" is endingwith("foo") }} -> false
// ```
func isEndingWith(v, other string) bool { return strings.HasSuffix(v, other) }

func isEq(val, other Value) bool { return valueEqual(val, other) }
func isNe(val, other Value) bool { return !valueEqual(val, other) }
func isLt(val, other Value) bool { return valueCmp(val, other) < 0 }
func isLe(val, other Value) bool { return valueCmp(val, other) <= 0 }
func isGt(val, other Value) bool { return valueCmp(val, other) > 0 }
func isGe(val, other Value) bool { return valueCmp(val, other) >= 0 }

func isIn(val, other Value) bool {
	b, err := opContains(other, val)
	if err != nil {
		return false
	}
	return b.isTrue()
}

func isTrue(val Value) bool {
	boolVal, ok := val.(boolValue)
	return ok && boolVal.B
}

func isFalse(val Value) bool {
	boolVal, ok := val.(boolValue)
	return ok && !boolVal.B
}

func isFilter(state *vmState, name string) bool {
	return state.env.getFilter(name).IsSome()
}

func isTest(state *vmState, name string) bool {
	return state.env.getTest(name).IsSome()
}
