package mjingo

import (
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type TestFunc = func(*State, []Value) (bool, error)

func testFuncFromPredicateWithValueArg(f func(val Value) bool) func(*State, []Value) (bool, error) {
	return func(state *State, values []Value) (bool, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return false, err
		}
		return f(tpl.a), nil
	}
}

func testFuncFromPredicateWithValValArgs(f func(val, other Value) bool) func(*State, []Value) (bool, error) {
	return func(state *State, values []Value) (bool, error) {
		tpl, err := tuple2FromValues(state, values)
		if err != nil {
			return false, err
		}
		return f(tpl.a, tpl.b), nil
	}
}

func testFuncFromPredicateWithStateStrArgs(f func(state *State, name string) bool) func(*State, []Value) (bool, error) {
	return func(state *State, values []Value) (bool, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return false, err
		}
		a, err := StringFromValue(option.Some(tpl.a))
		if err != nil {
			return false, err
		}
		return f(state, a), nil
	}
}

func testFuncFromPredicateWithStringStringArgs(f func(a, b string) bool) func(*State, []Value) (bool, error) {
	return func(state *State, values []Value) (bool, error) {
		tpl, err := tuple2FromValues(state, values)
		if err != nil {
			return false, err
		}
		a, err := StringFromValue(option.Some(tpl.a))
		if err != nil {
			return false, err
		}
		b, err := StringFromValue(option.Some(tpl.b))
		if err != nil {
			return false, err
		}
		return f(a, b), nil
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
	fromStateAndValue(state *State, val option.Option[Value]) (argConvertResult[O], error)
	fromStateAndValues(state *State, values []Value, offset uint) (argConvertResult[O], error)
	isTrailing() bool
}

var _ = (argType[Value])(valueArgType{})

type valueArgType struct{}

func (valueArgType) fromValue(val option.Option[Value]) (Value, error) {
	if val.IsSome() {
		return val.Unwrap(), nil
	}
	return nil, NewError(MissingArgument, "")
}

func (valueArgType) fromStateAndValue(state *State, val option.Option[Value]) (argConvertResult[Value], error) {
	return fromStateAndValue(state, val)
}

func (valueArgType) fromStateAndValues(state *State, values []Value, offset uint) (argConvertResult[Value], error) {
	return fromStateAndValues(state, values, offset)
}

func (valueArgType) isTrailing() bool { return false }

func fromStateAndValue(state *State, val option.Option[Value]) (argConvertResult[Value], error) {
	var zero argConvertResult[Value]
	if option.MapOr(val, false, isUndefined) && state != nil && state.undefinedBehavior() == UndefinedBehaviorStrict {
		return zero, NewError(UndefinedError, "")
	}
	var o valueArgType
	out, err := o.fromValue(val)
	if err != nil {
		return zero, err
	}
	return argConvertResult[Value]{output: out, consumed: 1}, nil
}

func fromStateAndValues(state *State, values []Value, offset uint) (argConvertResult[Value], error) {
	var o valueArgType
	val := option.Option[Value]{}
	if offset < uint(len(values)) {
		val = option.Some(values[offset])
	}
	return o.fromStateAndValue(state, val)
}

func tuple1FromValues(state *State, values []Value) (tuple1[Value], error) {
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
		return zero, NewError(TooManyArguments, "")
	}
	return tuple1[Value]{a: ao}, nil
}

func tuple2FromValues(state *State, values []Value) (tuple2[Value, Value], error) {
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
		return zero, NewError(TooManyArguments, "")
	}
	return tuple2[Value, Value]{a: ao, b: bo}, nil
}

func tuple3FromValues(state *State, values []Value) (tuple3[Value, Value, Value], error) {
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
		return zero, NewError(TooManyArguments, "")
	}
	return tuple3[Value, Value, Value]{a: ao, b: bo, c: co}, nil
}

func tuple4FromValues(state *State, values []Value) (tuple4[Value, Value, Value, Value], error) {
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
		return zero, NewError(TooManyArguments, "")
	}
	return tuple4[Value, Value, Value, Value]{a: ao, b: bo, c: co, d: do}, nil
}

func tuple5FromValues(state *State, values []Value) (tuple5[Value, Value, Value, Value, Value], error) {
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
		return zero, NewError(TooManyArguments, "")
	}
	return tuple5[Value, Value, Value, Value, Value]{a: ao, b: bo, c: co, d: do, e: eo}, nil
}

func isUndefined(val Value) bool {
	return val.IsUndefined()
}

func isDefined(val Value) bool {
	return !val.IsUndefined()
}

func isNone(val Value) bool {
	return val.IsNone()
}

func isSafe(val Value) bool {
	return val.IsSafe()
}

// Checks if a value is odd.
//
// ```jinja
// {{ 41 is odd }} -> true
// ```
func isOdd(val Value) bool {
	x, err := val.TryToI64()
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
	x, err := val.TryToI64()
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
func isNumber(val Value) bool { return val.Kind() == ValueKindNumber }

// Checks if this value is a string.
//
// ```jinja
// {{ "42" is string }} -> true
// {{ 42 is string }} -> false
// ```
func isString(val Value) bool { return val.Kind() == ValueKindString }

// Checks if this value is a sequence
//
// ```jinja
// {{ [1, 2, 3] is sequence }} -> true
// {{ 42 is sequence }} -> false
// ```
func isSequence(val Value) bool { return val.Kind() == ValueKindSeq }

// Checks if this value is a mapping
//
// ```jinja
// {{ {"foo": "bar"} is mapping }} -> true
// {{ [1, 2, 3] is mapping }} -> false
// ```
func isMapping(val Value) bool { return val.Kind() == ValueKindMap }

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

func isEq(val, other Value) bool { return Equal(val, other) }
func isNe(val, other Value) bool { return !Equal(val, other) }
func isLt(val, other Value) bool { return Cmp(val, other) < 0 }
func isLe(val, other Value) bool { return Cmp(val, other) <= 0 }
func isGt(val, other Value) bool { return Cmp(val, other) > 0 }
func isGe(val, other Value) bool { return Cmp(val, other) >= 0 }

func isIn(val, other Value) bool {
	b, err := Contains(other, val)
	if err != nil {
		return false
	}
	return b.IsTrue()
}

func isTrue(val Value) bool {
	boolVal, ok := val.(BoolValue)
	return ok && boolVal.B
}

func isFalse(val Value) bool {
	boolVal, ok := val.(BoolValue)
	return ok && !boolVal.B
}

func isFilter(state *State, name string) bool {
	return state.env.getFilter(name).IsSome()
}

func isTest(state *State, name string) bool {
	return state.env.getTest(name).IsSome()
}
