package mjingo

import (
	"errors"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type boxedFunc = func(*vmState, []Value) (Value, error)

type funcObject struct{ f boxedFunc }

func valueFromFunc(f boxedFunc) Value {
	return valueFromObject(funcObject{f: f})
}

var _ = (object)(funcObject{})
var _ = (caller)(funcObject{})

func (funcObject) Kind() objectKind { return objectKindPlain }

func (f funcObject) Call(state *vmState, args []Value) (Value, error) {
	return f.f(state, args)
}

func boxedFuncFromU32OptU32OptU32ArgU32SliceAndErrRet(f func(lower uint32, upper, step option.Option[uint32]) ([]uint32, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var a, b, c Value
		switch {
		case len(values) <= 1:
			tpl1, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			a = tpl1.a
		case len(values) <= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			a = tpl2.a
			b = tpl2.b
		case len(values) >= 3:
			tpl3, err := tuple3FromValues(state, values)
			if err != nil {
				return nil, err
			}
			a = tpl3.a
			b = tpl3.b
			c = tpl3.c
		}
		l, err := a.tryToI64()
		if err != nil {
			return nil, err
		}
		lower := uint32(l)
		upper := option.None[uint32]()
		if b != nil {
			u, err := b.tryToI64()
			if err != nil {
				return nil, err
			}
			upper = option.Some(uint32(u))
		}
		step := option.None[uint32]()
		if c != nil {
			s, err := c.tryToI64()
			if err != nil {
				return nil, err
			}
			step = option.Some(uint32(s))
		}
		rng, err := f(lower, upper, step)
		if err != nil {
			return nil, err
		}
		rv := make([]Value, 0, len(rng))
		for _, r := range rng {
			rv = append(rv, valueFromI64(int64(r)))
		}
		return valueFromSlice(rv), nil
	}
}

func boxedFuncFromValArgValErrRet(f func(Value) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl1, err := tuple1FromValues(state, values)
		if err != nil {
			var err2 *Error
			if errors.As(err, &err2) && err2.Type() == MissingArgument {
				tpl1.a = Undefined
			} else {
				return nil, err
			}
		}
		return f(tpl1.a)
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
