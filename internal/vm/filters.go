package vm

import (
	"strings"

	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/value"
)

type FilterFunc = func(*State, []value.Value) (value.Value, error)

func filterFuncFromFilterWithStringArgValueRet(f func(val string) value.Value) func(*State, []value.Value) (value.Value, error) {
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

func filterFuncFromWithStateValueArgValueErrRet(f func(*State, value.Value) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(state, tpl.a)
	}
}

func filterFuncFromFilterWithStringArgStringRet(f func(val string) string) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := value.StringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		return value.FromString(f(a)), nil
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

func lower(s string) string {
	return strings.ToLower(s)
}

func upper(s string) string {
	return strings.ToUpper(s)
}

func title(s string) string {
	return strings.ToTitle(s)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToTitle(s[:1]) + strings.ToLower(s[1:])
}
