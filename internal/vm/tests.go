package vm

import (
	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/value"
)

type TestFunc = func(*State, []value.Value) (bool, error)

func testFuncFromPredicate(f func(val value.Value) bool) func(*State, []value.Value) (bool, error) {
	return func(state *State, values []value.Value) (bool, error) {
		// tpl, err := tuple1FromValues[valu.Value, argType[valu.Value]](state, values)
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return false, err
		}
		return f(tpl.a), nil
	}
}

type TestPerformFunc[A any] func(args A) testResult

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
	fromValue(val option.Option[value.Value]) (O, error)
	fromStateAndValue(state *State, val option.Option[value.Value]) (argConvertResult[O], error)
	fromStateAndValues(state *State, values []value.Value, offset uint) (argConvertResult[O], error)
	isTrailing() bool
}

var _ = (argType[value.Value])(valueArgType{})

type valueArgType struct{}

func (valueArgType) fromValue(val option.Option[value.Value]) (value.Value, error) {
	if option.IsSome(val) {
		return option.Unwrap(val), nil
	}
	return nil, internal.NewError(internal.MissingArgument, "")
}

func (valueArgType) fromStateAndValue(state *State, val option.Option[value.Value]) (argConvertResult[value.Value], error) {
	return fromStateAndValue(state, val)
}

func (valueArgType) fromStateAndValues(state *State, values []value.Value, offset uint) (argConvertResult[value.Value], error) {
	return fromStateAndValues(state, values, offset)
}

func (valueArgType) isTrailing() bool { return false }

func fromStateAndValue(state *State, val option.Option[value.Value]) (argConvertResult[value.Value], error) {
	var zero argConvertResult[value.Value]
	if option.MapOr(val, false, isUndefined) && state != nil && state.undefinedBehavior() == compiler.UndefinedBehaviorStrict {
		return zero, internal.NewError(internal.UndefinedError, "")
	}
	var o valueArgType
	out, err := o.fromValue(val)
	if err != nil {
		return zero, err
	}
	return argConvertResult[value.Value]{output: out, consumed: 1}, nil
}

func fromStateAndValues(state *State, values []value.Value, offset uint) (argConvertResult[value.Value], error) {
	var o valueArgType
	val := option.Option[value.Value]{}
	if offset < uint(len(values)) {
		val = option.Some(values[offset])
	}
	return o.fromStateAndValue(state, val)
}

// func fromStateAndValue[O any](state *State, val option.Option[valu.Value]) (argConvertResult[O], error) {
// 	var zero argConvertResult[O]
// 	if optionMapOr(val, false, isUndefined) && state != nil && state.undefinedBehavior() == UndefinedBehaviorStrict {
// 		return zero, &Error{typ: UndefinedError}
// 	}
// 	var o argType[O]
// 	out, err := o.fromValue(val)
// 	if err != nil {
// 		return zero, err
// 	}
// 	return argConvertResult[O]{output: out, consumed: 1}, nil
// }

// func fromStateAndValues[O any](state *State, values []valu.Value, offset uint) (argConvertResult[O], error) {
// 	var o argType[O]
// 	val := option.Option[valu.Value]{}
// 	if offset < uint(len(values)) {
// 		val = option.Option[valu.Value]{valid: true, data: values[offset]}
// 	}
// 	return o.fromStateAndValue(state, val)
// }

func unitFromValues(_ *State, values []value.Value) (value.Unit, error) {
	if len(values) == 0 {
		return value.Unit{}, nil
	}
	return value.Unit{}, internal.NewError(internal.TooManyArguments, "")
}

func tuple1FromValues(state *State, values []value.Value) (tuple1[value.Value], error) {
	var zero tuple1[value.Value]
	var ao value.Value
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
		return zero, internal.NewError(internal.TooManyArguments, "")
	}
	return tuple1[value.Value]{a: ao}, nil
}

// func tuple1FromValues[AO any, A argType[AO]](state *State, values []valu.Value) (tuple1[AO], error) {
// 	var zero tuple1[AO]
// 	var ao AO
// 	var at argType[AO]
// 	idx := uint(0)
// 	restFirst := at.isTrailing() && len(values) != 0
// 	if restFirst {
// 		avo, err := at.fromStateAndValues(state, values, uint(len(values)-1))
// 		if err != nil {
// 			return zero, err
// 		}
// 		ao = avo.output
// 		values = values[:len(values)-int(avo.consumed)]
// 	}
// 	if !restFirst {
// 		avo, err := at.fromStateAndValues(state, values, idx)
// 		if err != nil {
// 			return zero, err
// 		}
// 		ao = avo.output
// 		idx += avo.consumed
// 	}
// 	if idx < uint(len(values)) {
// 		return zero, internal.NewError(internal.TooManyArguments, "")
// 	}
// 	return tuple1[AO]{a: ao}, nil
// }

func tuple2FromValues[AO any, BO any, A argType[AO], B argType[BO]](state *State, values []value.Value) (tuple2[AO, BO], error) {
	var zero tuple2[AO, BO]
	var ao AO
	var bo BO
	var at argType[AO]
	var bt argType[BO]
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
		return zero, internal.NewError(internal.TooManyArguments, "")
	}
	return tuple2[AO, BO]{a: ao, b: bo}, nil
}

func tuple3FromValues[AO any, BO any, CO any, A argType[AO], B argType[BO], C argType[CO]](state *State, values []value.Value) (tuple3[AO, BO, CO], error) {
	var zero tuple3[AO, BO, CO]
	var ao AO
	var bo BO
	var co CO
	var at argType[AO]
	var bt argType[BO]
	var ct argType[CO]
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
		return zero, internal.NewError(internal.TooManyArguments, "")
	}
	return tuple3[AO, BO, CO]{a: ao, b: bo, c: co}, nil
}

func tuple4FromValues[AO any, BO any, CO any, DO any, A argType[AO], B argType[BO], C argType[CO], D argType[DO]](state *State, values []value.Value) (tuple4[AO, BO, CO, DO], error) {
	var zero tuple4[AO, BO, CO, DO]
	var ao AO
	var bo BO
	var co CO
	var do DO
	var at argType[AO]
	var bt argType[BO]
	var ct argType[CO]
	var dt argType[DO]
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
		return zero, internal.NewError(internal.TooManyArguments, "")
	}
	return tuple4[AO, BO, CO, DO]{a: ao, b: bo, c: co, d: do}, nil
}

func tuple5FromValues[AO any, BO any, CO any, DO any, EO any,
	A argType[AO], B argType[BO], C argType[CO], D argType[DO],
	E argType[EO]](state *State, values []value.Value) (tuple5[AO, BO, CO, DO, EO], error) {
	var zero tuple5[AO, BO, CO, DO, EO]
	var ao AO
	var bo BO
	var co CO
	var do DO
	var eo EO
	var at argType[AO]
	var bt argType[BO]
	var ct argType[CO]
	var dt argType[DO]
	var et argType[EO]
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
		return zero, internal.NewError(internal.TooManyArguments, "")
	}
	return tuple5[AO, BO, CO, DO, EO]{a: ao, b: bo, c: co, d: do, e: eo}, nil
}

type testResult interface {
	intoResult() (bool, error)
}

var _ = (testResult)(resultTestResult{})
var _ = (testResult)(boolTestResult{})

type resultTestResult struct {
	ret bool
	err error
}

func (r resultTestResult) intoResult() (bool, error) {
	return r.ret, r.err
}

type boolTestResult struct {
	ret bool
}

func (r boolTestResult) intoResult() (bool, error) {
	return r.ret, nil
}

func arg0TestToPerform(f func() testResult) TestPerformFunc[value.Unit] {
	return func(_ value.Unit) testResult {
		return f()
	}
}

func arg1TestToPerform[A any](f func(A) testResult) TestPerformFunc[tuple1[A]] {
	return func(args tuple1[A]) testResult {
		return f(args.a)
	}
}

func arg2TestToPerform[A any, B any](f func(A, B) testResult) TestPerformFunc[tuple2[A, B]] {
	return func(args tuple2[A, B]) testResult {
		return f(args.a, args.b)
	}
}

func arg3TestToPerform[A any, B any, C any](f func(A, B, C) testResult) TestPerformFunc[tuple3[A, B, C]] {
	return func(args tuple3[A, B, C]) testResult {
		return f(args.a, args.b, args.c)
	}
}

func arg4TestToPerform[A any, B any, C any, D any](f func(A, B, C, D) testResult) TestPerformFunc[tuple4[A, B, C, D]] {
	return func(args tuple4[A, B, C, D]) testResult {
		return f(args.a, args.b, args.c, args.d)
	}
}

func arg5TestToPerform[A any, B any, C any, D any, E any](f func(A, B, C, D, E) testResult) TestPerformFunc[tuple5[A, B, C, D, E]] {
	return func(args tuple5[A, B, C, D, E]) testResult {
		return f(args.a, args.b, args.c, args.d, args.e)
	}
}

func isUndefined(val value.Value) bool {
	return val.IsUndefined()
}

func isDefined(val value.Value) bool {
	return !val.IsUndefined()
}
