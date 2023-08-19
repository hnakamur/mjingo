package vm

import (
	"strings"

	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/value"
)

type FilterFunc = func(*State, []value.Value) (value.Value, error)

func filterFuncFromFilterWithStringArg(f func(val string) value.Value) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := value.StringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		return f(a), nil
	}
}

func filterFuncFromWithStateValueArgErr(f func(*State, value.Value) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(state, tpl.a)
	}
}

func safe(v string) value.Value {
	return value.FromSafeString(v)
}

func escape(state *State, v value.Value) (value.Value, error) {
	if v.IsSafe() {
		return v, nil
	}

	// this tries to use the escaping flag of the current scope, then
	// of the initial state and if that is also not set it falls back
	// to HTML.
	autoEscape := state.autoEscape
	if _, ok := state.autoEscape.(compiler.AutoEscapeNone); ok {
		if _, ok := state.env.initialAutoEscape(state.name()).(compiler.AutoEscapeNone); ok {
			autoEscape = compiler.AutoEscapeHTML{}
		}
	}
	var b strings.Builder
	if optStr := v.AsStr(); option.IsSome(optStr) {
		b.Grow(len(option.Unwrap(optStr)))
	}
	out := newOutput(&b)
	if err := writeEscaped(out, autoEscape, v); err != nil {
		return nil, err
	}
	return value.FromSafeString(b.String()), nil
}
