package vm

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/value"
)

type FilterFunc = func(*State, []value.Value) (value.Value, error)

func filterFuncFromFilterWithStrArgValRet(f func(val string) value.Value) func(*State, []value.Value) (value.Value, error) {
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

func filterFuncFromWithStateValArgValErrRet(f func(*State, value.Value) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(state, tpl.a)
	}
}

func filterFuncFromWithValArgValErrRet(f func(value.Value) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(tpl.a)
	}
}

func filterFuncFromWithValValArgValErrRet(f func(value.Value, value.Value) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple2FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(tpl.a, tpl.b)
	}
}
func filterFuncFromFilterWithStrArgStrRet(f func(val string) string) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := value.StringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		return value.ValueFromString(f(a)), nil
	}
}

func filterFuncFromFilterWithStateStrStrStrArgStrRet(f func(state *State, v1, v2, v3 string) string) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple3FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := value.StringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		b, err := value.StringFromValue(option.Some(tpl.b))
		if err != nil {
			return nil, err
		}
		c, err := value.StringFromValue(option.Some(tpl.c))
		if err != nil {
			return nil, err
		}
		return value.ValueFromString(f(state, a, b, c)), nil
	}
}

func filterFuncFromFilterWithValArgUintErrRet(f func(val value.Value) (uint, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		l, err := f(tpl.a)
		if err != nil {
			return nil, err
		}
		return value.ValueFromI64(int64(l)), nil
	}
}

func filterFuncFromFilterWithValOptStrArgStrErrRet(f func(val value.Value, optStr option.Option[string]) (string, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val value.Value
		optStr := option.None[string]()
		switch {
		case len(values) <= 1:
			tpl1, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl1.a
		case len(values) >= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl2.a
			optStr = option.Some(tpl2.b.String())
		}
		rv, err := f(val, optStr)
		if err != nil {
			return nil, err
		}
		return value.ValueFromString(rv), nil
	}
}

func filterFuncFromFilterWithValOptI32ArgValErrRet(f func(val value.Value, optI32 option.Option[int32]) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val value.Value
		optI32 := option.None[int32]()
		switch {
		case len(values) <= 1:
			tpl1, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl1.a
		case len(values) >= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl2.a
			n, err := value.I32TryFromValue(tpl2.b)
			if err != nil {
				return nil, err
			}
			optI32 = option.Some(n)
		}
		return f(val, optI32)
	}
}

func filterFuncFromFilterWithValOptValArgValRet(f func(a value.Value, optB option.Option[value.Value]) value.Value) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var a value.Value
		optB := option.None[value.Value]()
		switch {
		case len(values) <= 1:
			tpl1, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			a = tpl1.a
		case len(values) >= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			a = tpl2.a
			optB = option.Some(tpl2.b)
		}
		return f(a, optB), nil
	}
}

func filterFuncFromFilterWithStrOptStrArgStrRet(f func(s string, optStr option.Option[string]) string) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val value.Value
		optStr := option.None[string]()
		switch {
		case len(values) <= 1:
			tpl1, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl1.a
		case len(values) >= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl2.a
			optStr = option.Some(tpl2.b.String())
		}
		s, err := value.StringFromValue(option.Some(val))
		if err != nil {
			return nil, err
		}
		return value.ValueFromString(f(s, optStr)), nil
	}
}

func filterFuncFromFilterWithValArgBoolRet(f func(val value.Value) bool) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return value.ValueFromBool(f(tpl.a)), nil
	}
}

func filterFuncFromFilterWithStateValUintOptValArgValErrRet(f func(*State, value.Value, uint, option.Option[value.Value]) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val, countVal value.Value
		fillWith := option.None[value.Value]()
		switch {
		case len(values) <= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl2.a
			countVal = tpl2.b
		case len(values) >= 3:
			tpl3, err := tuple3FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl3.a
			countVal = tpl3.b
			fillWith = option.Some(tpl3.c)
		}
		count, err := countVal.TryToUint()
		if err != nil {
			return nil, err
		}
		return f(state, val, count, fillWith)
	}
}

func filterFuncFromFilterWithStrUintOptBoolOptBoolArgStrrRet(f func(string, uint, option.Option[bool], option.Option[bool]) string) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var strVal, uintVal value.Value
		optBoolVal1 := option.None[value.Value]()
		optBool2 := option.None[bool]()
		switch {
		case len(values) <= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			strVal = tpl2.a
			uintVal = tpl2.b
		case len(values) == 3:
			tpl3, err := tuple3FromValues(state, values)
			if err != nil {
				return nil, err
			}
			strVal = tpl3.a
			uintVal = tpl3.b
			optBoolVal1 = option.Some(tpl3.c)
		case len(values) >= 4:
			tpl4, err := tuple4FromValues(state, values)
			if err != nil {
				return nil, err
			}
			strVal = tpl4.a
			uintVal = tpl4.b
			optBoolVal1 = option.Some(tpl4.c)
			b, err := value.BoolTryFromValue(tpl4.d)
			if err != nil {
				return nil, err
			}
			optBool2 = option.Some(b)
		}
		s, err := value.StringFromValue(option.Some(strVal))
		if err != nil {
			return nil, err
		}
		n, err := uintVal.TryToUint()
		if err != nil {
			return nil, err
		}
		optBool1 := option.None[bool]()
		if optBoolVal1.IsSome() {
			b, err := value.BoolTryFromValue(optBoolVal1.Unwrap())
			if err != nil {
				return nil, err
			}
			optBool1 = option.Some(b)
		}
		return value.ValueFromString(f(s, n, optBool1, optBool2)), nil
	}
}

func filterFuncFromFilterWithStateValOptStrValVarArgValSliceErrRet(f func(*State, value.Value, option.Option[string], ...value.Value) ([]value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val value.Value
		optStr := option.None[string]()
		var args []value.Value
		switch {
		case len(values) <= 1:
			tpl1, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl1.a
			args = values[1:]
		case len(values) >= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl2.a
			s, err := value.StringFromValue(option.Some(tpl2.b))
			if err != nil {
				return nil, err
			}
			optStr = option.Some(s)
			args = values[2:]
		}
		rv, err := f(state, val, optStr, args...)
		if err != nil {
			return nil, err
		}
		return value.ValueFromSlice(rv), nil
	}
}

func filterFuncFromFilterWithStateValStrOptStrValVarArgValSliceErrRet(f func(*State, value.Value, string, option.Option[string], ...value.Value) ([]value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val, strVal value.Value
		optStr := option.None[string]()
		var args []value.Value
		switch {
		case len(values) <= 2:
			tpl2, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl2.a
			strVal = tpl2.b
			args = values[2:]
		case len(values) >= 3:
			tpl3, err := tuple3FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl3.a
			strVal = tpl3.b
			s, err := value.StringFromValue(option.Some(tpl3.c))
			if err != nil {
				return nil, err
			}
			optStr = option.Some(s)
			args = values[3:]
		}
		s, err := value.StringFromValue(option.Some(strVal))
		if err != nil {
			return nil, err
		}
		rv, err := f(state, val, s, optStr, args...)
		if err != nil {
			return nil, err
		}
		return value.ValueFromSlice(rv), nil
	}
}

func filterFuncFromFilterWithValSliceArgValRet(f func([]value.Value) value.Value) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		iter, err := state.undefinedBehavior().TryIter(tpl.a)
		if err != nil {
			return nil, err
		}
		return f(iter.Collect()), nil
	}
}

func filterFuncFromWithValKwargsArgValErrRet(f func(value.Value, value.Kwargs) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val value.Value
		var kwargs value.Kwargs
		switch {
		case len(values) <= 1:
			tpl, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs = value.NewKwargs(*value.NewValueMap())
		case len(values) >= 2:
			tpl, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs, err = value.KwargsTryFromValue(tpl.b)
			if err != nil {
				return nil, err
			}
		}
		return f(val, kwargs)
	}
}

func filterFuncFromWithStateValKwargsArgValErrRet(f func(*State, value.Value, value.Kwargs) (value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		var val value.Value
		var kwargs value.Kwargs
		switch {
		case len(values) <= 1:
			tpl, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs = value.NewKwargs(*value.NewValueMap())
		case len(values) >= 2:
			tpl, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs, err = value.KwargsTryFromValue(tpl.b)
			if err != nil {
				return nil, err
			}
		}
		return f(state, val, kwargs)
	}
}

func filterFuncFromFilterWithStateValValVarArgValSliceErrRet(f func(*State, value.Value, ...value.Value) ([]value.Value, error)) func(*State, []value.Value) (value.Value, error) {
	return func(state *State, values []value.Value) (value.Value, error) {
		if len(values) == 0 {
			return nil, common.NewError(common.MissingArgument, "")
		}
		val := values[0]
		args := values[1:]
		rv, err := f(state, val, args...)
		if err != nil {
			return nil, err
		}
		return value.ValueFromSlice(rv), nil
	}
}

func safe(v string) value.Value {
	return value.ValueFromSafeString(v)
}

func escape(state *State, v value.Value) (value.Value, error) {
	if v.IsSafe() {
		return v, nil
	}

	// this tries to use the escaping flag of the current scope, then
	// of the initial state and if that is also not set it falls back
	// to HTML.
	autoEscape := state.autoEscape
	if _, ok := state.autoEscape.(AutoEscapeNone); ok {
		if _, ok := state.env.initialAutoEscape(state.name()).(AutoEscapeNone); ok {
			autoEscape = AutoEscapeHTML{}
		}
	}
	var b strings.Builder
	if optStr := v.AsStr(); optStr.IsSome() {
		b.Grow(len(optStr.Unwrap()))
	}
	out := newOutput(&b)
	if err := writeEscaped(out, autoEscape, v); err != nil {
		return nil, err
	}
	return value.ValueFromSafeString(b.String()), nil
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

// Does a string replace.
//
// It replaces all occurrences of the first parameter with the second.
//
// ```jinja
// {{ "Hello World"|replace("Hello", "Goodbye") }}
//
//	-> Goodbye World
//
// ```
func replace(_ *State, v, from, to string) string {
	r := strings.NewReplacer(from, to)
	return r.Replace(v)
}

func length(val value.Value) (uint, error) {
	if optLen := val.Len(); optLen.IsSome() {
		return optLen.Unwrap(), nil
	}
	return 0, common.NewError(common.InvalidOperation,
		fmt.Sprintf("cannot calculate length of value of type %s", val.Kind()))
}

func compareValuesCaseInsensitive(a, b value.Value) int {
	if optA, optB := a.AsStr(), b.AsStr(); optA.IsSome() && optB.IsSome() {
		return strings.Compare(strings.ToLower(optA.Unwrap()), strings.ToLower(optB.Unwrap()))
	}
	return value.Cmp(a, b)
}

type keyAndValue struct {
	Key   value.Value
	Value value.Value
}

// Dict sorting functionality.
//
// This filter works like `|items` but sorts the pairs by key first.
//
// The filter accepts a few keyword arguments:
//
// * `case_sensitive`: set to `true` to make the sorting of strings case sensitive.
// * `by`: set to `"value"` to sort by value. Defaults to `"key"`.
// * `reverse`: set to `true` to sort in reverse.
func dictsort(v value.Value, kwargs value.Kwargs) (value.Value, error) {
	if v.Kind() != value.ValueKindMap {
		return nil, common.NewError(common.InvalidOperation, "cannot convert value into pair list")
	}
	entries := make([]keyAndValue, 0, v.Len().UnwrapOr(0))
	iter, err := v.TryIter()
	if err != nil {
		return nil, err
	}
	var key value.Value
	for iter.Next().UnwrapTo(&key) {
		val, err := value.GetItem(v, key)
		if err != nil {
			val = value.Undefined
		}
		entries = append(entries, keyAndValue{Key: key, Value: val})
	}

	byVal := false
	if optBy := kwargs.GetValue("by"); optBy.IsSome() {
		if by, ok := optBy.Unwrap().(value.StringValue); ok {
			switch by.Str {
			case "key":
				byVal = false
			case "value":
				byVal = true
			default:
				return nil, common.NewError(common.InvalidOperation,
					fmt.Sprintf("invalid value '%s' for 'by' parameter", by.Str))
			}
		}
	}
	getKeyOrVal := func(entry keyAndValue) value.Value { return entry.Key }
	if byVal {
		getKeyOrVal = func(entry keyAndValue) value.Value { return entry.Value }
	}

	caseSensitive := false
	if optCaseSensitive := kwargs.GetValue("case_sensitive"); optCaseSensitive.IsSome() {
		if cs, ok := optCaseSensitive.Unwrap().(value.BoolValue); ok && cs.B {
			caseSensitive = true
		}
	}
	sortFn := value.Cmp
	if !caseSensitive {
		sortFn = compareValuesCaseInsensitive
	}

	reverse := false
	if optReverse := kwargs.GetValue("reverse"); optReverse.IsSome() {
		if cs, ok := optReverse.Unwrap().(value.BoolValue); ok && cs.B {
			reverse = true
		}
	}

	slices.SortFunc(entries, func(a, b keyAndValue) int {
		ret := sortFn(getKeyOrVal(a), getKeyOrVal(b))
		if reverse {
			return -ret
		}
		return ret
	})

	if err := kwargs.AssertAllUsed(); err != nil {
		return nil, err
	}

	items := make([]value.Value, 0, len(entries))
	for _, entry := range entries {
		item := value.ValueFromSlice([]value.Value{entry.Key, entry.Value})
		items = append(items, item)
	}
	return value.ValueFromSlice(items), nil
}

func sortFilter(state *State, val value.Value, kwargs value.Kwargs) (value.Value, error) {
	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, common.NewError(common.InvalidOperation, "cannot convert value to list").WithSource(err)
	}
	items := iter.Collect()
	caseSensitive := false
	if optCaseSensitive := kwargs.GetValue("case_sensitive"); optCaseSensitive.IsSome() {
		if cs, ok := optCaseSensitive.Unwrap().(value.BoolValue); ok && cs.B {
			caseSensitive = true
		}
	}
	sortFn := value.Cmp
	if !caseSensitive {
		sortFn = compareValuesCaseInsensitive
	}

	var attr string
	if optAttr := kwargs.GetValue("attribute"); optAttr.IsSome() {
		if strVal, ok := optAttr.Unwrap().(value.StringValue); ok {
			attr = strVal.Str
		}
	}
	reverse := false
	if optReverse := kwargs.GetValue("reverse"); optReverse.IsSome() {
		if cs, ok := optReverse.Unwrap().(value.BoolValue); ok && cs.B {
			reverse = true
		}
	}

	if attr != "" {
		slices.SortFunc(items, func(a, b value.Value) int {
			aVal, err := value.GetPath(a, attr)
			if err != nil {
				return 0
			}
			bVal, err := value.GetPath(b, attr)
			if err != nil {
				return 0
			}
			ret := sortFn(aVal, bVal)
			if reverse {
				return -ret
			}
			return ret
		})
	} else {
		slices.SortFunc(items, func(a, b value.Value) int {
			ret := sortFn(a, b)
			if reverse {
				return -ret
			}
			return ret
		})
	}

	if err := kwargs.AssertAllUsed(); err != nil {
		return nil, err
	}
	return value.ValueFromSlice(items), nil
}

// Returns a list of pairs (items) from a mapping.
//
// This can be used to iterate over keys and values of a mapping
// at once.  Note that this will use the original order of the map
// which is typically arbitrary unless the `preserve_order` feature
// is used in which case the original order of the map is retained.
// It's generally better to use `|dictsort` which sorts the map by
// key before iterating.
//
// ```jinja
// <dl>
// {% for key, value in my_dict|items %}
//
//	<dt>{{ key }}
//	<dd>{{ value }}
//
// {% endfor %}
// </dl>
// ```
func items(v value.Value) (value.Value, error) {
	if v.Kind() != value.ValueKindMap {
		return nil, common.NewError(common.InvalidOperation, "cannot convert value into pair list")
	}
	items := make([]value.Value, 0, v.Len().UnwrapOr(0))
	iter, err := v.TryIter()
	if err != nil {
		return nil, err
	}
	var key value.Value
	for iter.Next().UnwrapTo(&key) {
		val, err := value.GetItem(v, key)
		if err != nil {
			val = value.Undefined
		}
		item := value.ValueFromSlice([]value.Value{key, val})
		items = append(items, item)
	}
	return value.ValueFromSlice(items), nil
}

// Joins a sequence by a character
func join(val value.Value, joiner option.Option[string]) (string, error) {
	if val.IsUndefined() || val.IsNone() {
		return "", nil
	}

	joinerStr := joiner.UnwrapOr("")
	if optValStr := val.AsStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		var b strings.Builder
		for len(rest) > 0 {
			if b.Len() != 0 {
				b.WriteString(joinerStr)
			}
			r, size := utf8.DecodeRuneInString(rest)
			b.WriteRune(r)
			rest = rest[size:]
		}
		return b.String(), nil
	}
	if optValSeq := val.AsSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		var b strings.Builder
		n := valSeq.ItemCount()
		for i := uint(0); i < n; i++ {
			if b.Len() != 0 {
				b.WriteString(joinerStr)
			}
			item := valSeq.GetItem(i).Unwrap()
			if optItemStr := item.AsStr(); optItemStr.IsSome() {
				b.WriteString(optItemStr.Unwrap())
			} else {
				fmt.Fprintf(&b, "%s", item)
			}
		}
		return b.String(), nil
	}
	return "", common.NewError(common.InvalidOperation,
		fmt.Sprintf("cannot join value of type %s", val.Kind()))
}

// Reverses a list or string
//
// ```jinja
// {% for user in users|reverse %}
//
//	<li>{{ user.name }}
//
// {% endfor %}
// ```
func reverse(val value.Value) (value.Value, error) {
	if optValStr := val.AsStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		var b strings.Builder
		for len(rest) > 0 {
			r, size := utf8.DecodeLastRuneInString(rest)
			b.WriteRune(r)
			rest = rest[:len(rest)-size]
		}
		return value.ValueFromString(b.String()), nil
	}
	if optValSeq := val.AsSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		n := valSeq.ItemCount()
		items := make([]value.Value, 0, n)
		for i := n - 1; ; i-- {
			item := valSeq.GetItem(i).Unwrap()
			items = append(items, item)
			if i == 0 {
				break
			}
		}
		return value.ValueFromSlice(items), nil
	}
	return nil, common.NewError(common.InvalidOperation,
		fmt.Sprintf("cannot reverse value of type %s", val.Kind()))
}

func trim(s string, cutset option.Option[string]) string {
	if cutset.IsSome() {
		return strings.Trim(s, cutset.Unwrap())
	}
	return strings.TrimSpace(s)
}

func defaultFilter(val value.Value, other option.Option[value.Value]) value.Value {
	if val.IsUndefined() {
		return other.UnwrapOrElse(func() value.Value { return value.ValueFromString("") })
	}
	return val
}

func round(val value.Value, precision option.Option[int32]) (value.Value, error) {
	switch v := val.(type) {
	case value.I64Value, value.I128Value:
		return val, nil
	case value.F64Value:
		x := math.Pow10(int(precision.UnwrapOr(0)))
		return value.ValueFromF64(math.Round(x*v.F) / x), nil
	default:
		return nil, common.NewError(common.InvalidOperation, "cannot round value")
	}
}

func abs(val value.Value) (value.Value, error) {
	switch v := val.(type) {
	case value.I64Value:
		n := v.N
		if n < 0 {
			n = -n
		}
		return value.I64Value{N: n}, nil
	case value.I128Value:
		var n big.Int
		n.Abs(&v.N)
		return value.I128Value{N: n}, nil
	case value.F64Value:
		return value.F64Value{F: math.Abs(v.F)}, nil
	default:
		// TODO: Verify MiniJinja error message is really intentional.
		return nil, common.NewError(common.InvalidOperation, "cannot round value")
	}
}

// Looks up an attribute.
//
// In MiniJinja this is the same as the `[]` operator.  In Jinja2 there is a
// small difference which is why this filter is sometimes used in Jinja2
// templates.  For compatibility it's provided here as well.
//
// ```jinja
// {{ value['key'] == value|attr('key') }} -> true
// ```
func attr(val, key value.Value) (value.Value, error) {
	return value.GetItem(val, key)
}

func first(val value.Value) (value.Value, error) {
	if optValStr := val.AsStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		if rest == "" {
			return value.Undefined, nil
		}
		var b strings.Builder
		r, _ := utf8.DecodeRuneInString(rest)
		b.WriteRune(r)
		return value.ValueFromString(b.String()), nil
	}
	if optValSeq := val.AsSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		return valSeq.GetItem(0).UnwrapOr(value.Undefined), nil
	}
	return nil, common.NewError(common.InvalidOperation, "cannot get first item from value")
}

func last(val value.Value) (value.Value, error) {
	if optValStr := val.AsStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		if rest == "" {
			return value.Undefined, nil
		}
		var b strings.Builder
		r, _ := utf8.DecodeLastRuneInString(rest)
		b.WriteRune(r)
		return value.ValueFromString(b.String()), nil
	}
	if optValSeq := val.AsSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		n := valSeq.ItemCount()
		if n == 0 {
			return value.Undefined, nil
		}
		return valSeq.GetItem(n - 1).UnwrapOr(value.Undefined), nil
	}
	return nil, common.NewError(common.InvalidOperation, "cannot get last item from value")
}

func minFilter(state *State, val value.Value) (value.Value, error) {
	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, common.NewError(common.InvalidDelimiter, "cannot convert value to list").WithSource(err)
	}
	return iter.Min().UnwrapOr(value.Undefined), nil
}

func maxFilter(state *State, val value.Value) (value.Value, error) {
	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, common.NewError(common.InvalidDelimiter, "cannot convert value to list").WithSource(err)
	}
	return iter.Max().UnwrapOr(value.Undefined), nil
}

func listFilter(state *State, val value.Value) (value.Value, error) {
	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, common.NewError(common.InvalidDelimiter, "cannot convert value to list").WithSource(err)
	}
	return value.ValueFromSlice(iter.Collect()), nil
}

// Converts the value into a boolean value.
//
// This behaves the same as the if statement does with regards to
// handling of boolean values.
func boolFilter(val value.Value) bool {
	return val.IsTrue()
}

// Batch items.
//
// This filter works pretty much like `slice` just the other way round. It
// returns a list of lists with the given number of items. If you provide a
// second parameter this is used to fill up missing items.
//
// ```jinja
// <table>
//
//	{% for row in items|batch(3, '&nbsp;') %}
//	<tr>
//	{% for column in row %}
//	  <td>{{ column }}</td>
//	{% endfor %}
//	</tr>
//	{% endfor %}
//
// </table>
// ```
func batchFilter(state *State, val value.Value, count uint, fillWith option.Option[value.Value]) (value.Value, error) {
	if count == 0 {
		return nil, common.NewError(common.InvalidOperation, "count cannot be 0")
	}

	rv := make([]value.Value, 0, val.Len().UnwrapOr(0)/count)
	tmp := make([]value.Value, 0, count)

	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, err
	}
	var item value.Value
	for iter.Next().UnwrapTo(&item) {
		if uint(len(tmp)) == count {
			rv = append(rv, value.ValueFromSlice(tmp))
			tmp = make([]value.Value, 0, count)
		}
		tmp = append(tmp, item)
	}

	if len(tmp) != 0 {
		if fillWith.IsSome() {
			filler := fillWith.Unwrap()
			for i := uint(0); i < count-uint(len(tmp)); i++ {
				tmp = append(tmp, filler.Clone())
			}
		}
		rv = append(rv, value.ValueFromSlice(tmp))
	}

	return value.ValueFromSlice(rv), nil
}

// Slice an iterable and return a list of lists containing
// those items.
//
// Useful if you want to create a div containing three ul tags that
// represent columns:
//
// ```jinja
// <div class="columnwrapper">
// {% for column in items|slice(3) %}
//
//	<ul class="column-{{ loop.index }}">
//	{% for item in column %}
//	  <li>{{ item }}</li>
//	{% endfor %}
//	</ul>
//
// {% endfor %}
// </div>
// ```
//
// If you pass it a second argument it’s used to fill missing values on the
// last iteration.
func sliceFilter(state *State, val value.Value, count uint, fillWith option.Option[value.Value]) (value.Value, error) {
	if count == 0 {
		return nil, common.NewError(common.InvalidOperation, "count cannot be 0")
	}

	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, err
	}
	items := iter.Collect()
	l := uint(len(items))
	itemsPerSlice := l / count
	slicesWithExtra := l % count
	offset := uint(0)
	rv := make([]value.Value, 0, count)
	for slice := uint(0); slice < count; slice++ {
		start := offset + slice*itemsPerSlice
		if slice < slicesWithExtra {
			offset++
		}
		end := offset + (slice+1)*itemsPerSlice
		tmp := items[start:end]
		if fillWith.IsSome() && slice >= slicesWithExtra {
			filler := fillWith.Unwrap()
			tmp = append(tmp, filler.Clone())
			rv = append(rv, value.ValueFromSlice(tmp))
			continue
		}
		rv = append(rv, value.ValueFromSlice(tmp))
	}

	return value.ValueFromSlice(rv), nil
}

func indentFilter(val string, width uint, indentFirstLine, indentBlankLines option.Option[bool]) string {
	stripTrailingNewline := func(s *string) {
		if strings.HasSuffix(*s, "\n") {
			*s = (*s)[:len(*s)-1]
		}
		if strings.HasSuffix(*s, "\r") {
			*s = (*s)[:len(*s)-1]
		}
	}

	v := val
	stripTrailingNewline(&v)
	indentWith := strings.Repeat(" ", int(width))
	var output strings.Builder
	lines := strings.Split(v, "\n")
	i := 0
	if !indentFirstLine.UnwrapOr(false) && i < len(lines) {
		line := lines[i]
		i++
		output.WriteString(line)
		output.WriteRune('\n')
	}
	for ; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			if indentBlankLines.UnwrapOr(false) {
				output.WriteString(indentWith)
			}
		} else {
			output.WriteString(indentWith)
			output.WriteString(line)
		}
		output.WriteRune('\n')
	}
	rv := output.String()
	stripTrailingNewline(&rv)
	return rv
}

func selectOrReject(state *State, invert bool, val value.Value, attr, testName option.Option[string], args ...value.Value) ([]value.Value, error) {
	var rv []value.Value
	test := option.None[TestFunc]()
	if testName.IsSome() {
		test = state.env.getTest(testName.Unwrap())
		if test.IsNone() {
			return nil, common.NewError(common.UnknownTest, "")
		}
	}
	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, err
	}
	var item value.Value
	for iter.Next().UnwrapTo(&item) {
		var testVal value.Value
		if attr.IsSome() {
			testVal, err = value.GetAttr(item, attr.Unwrap())
			if err != nil {
				return nil, err
			}
		} else {
			testVal = item.Clone()
		}
		var passed bool
		if test.IsSome() {
			iter, _ := value.ValueFromSlice([]value.Value{testVal}).TryIter()
			iter2, _ := value.ValueFromSlice(args).TryIter()
			chainedIter := iter.Chain(iter2.Cloned())
			testArgs := chainedIter.Collect()
			passed, err = test.Unwrap()(state, testArgs)
			if err != nil {
				return nil, err
			}
		} else {
			passed = testVal.IsTrue()
		}
		if passed != invert {
			rv = append(rv, item)
		}
	}
	return rv, nil
}

func selectFilter(state *State, val value.Value, testName option.Option[string], args ...value.Value) ([]value.Value, error) {
	return selectOrReject(state, false, val, option.None[string](), testName, args...)
}

func selectAttrFilter(state *State, val value.Value, attr string, testName option.Option[string], args ...value.Value) ([]value.Value, error) {
	return selectOrReject(state, false, val, option.Some(attr), testName, args...)
}

func rejectFilter(state *State, val value.Value, testName option.Option[string], args ...value.Value) ([]value.Value, error) {
	return selectOrReject(state, true, val, option.None[string](), testName, args...)
}

func rejectAttrFilter(state *State, val value.Value, attr string, testName option.Option[string], args ...value.Value) ([]value.Value, error) {
	return selectOrReject(state, true, val, option.Some(attr), testName, args...)
}

// Returns a list of unique items from the given iterable.
//
// ```jinja
// {{ ['foo', 'bar', 'foobar', 'foobar']|unique|list }}
//
//	-> ['foo', 'bar', 'foobar']
//
// ```
//
// The unique items are yielded in the same order as their first occurrence
// in the iterable passed to the filter.  The filter will not detect
// duplicate objects or arrays, only primitives such as strings or numbers.
func uniqueFilter(values []value.Value) value.Value {
	var rv []value.Value
	seen := make(map[value.Value]struct{})
	for _, item := range values {
		if _, ok := seen[item]; !ok {
			rv = append(rv, item)
			seen[item] = struct{}{}
		}
	}
	return value.ValueFromSlice(rv)
}

// Applies a filter to a sequence of objects or looks up an attribute.
//
// This is useful when dealing with lists of objects but you are really
// only interested in a certain value of it.
//
// The basic usage is mapping on an attribute. Given a list of users
// you can for instance quickly select the username and join on it:
//
// ```jinja
// {{ users|map(attribute='username')|join(', ') }}
// ```
//
// You can specify a `default` value to use if an object in the list does
// not have the given attribute.
//
// ```jinja
// {{ users|map(attribute="username", default="Anonymous")|join(", ") }}
// ```
//
// Alternatively you can have `map` invoke a filter by passing the name of the
// filter and the arguments afterwards. A good example would be applying a
// text conversion filter on a sequence:
//
// ```jinja
// Users on this page: {{ titles|map('lower')|join(', ') }}
// ```
func mapFilter(state *State, val value.Value, args ...value.Value) ([]value.Value, error) {
	rv := make([]value.Value, 0, val.Len().UnwrapOr(0))
	kwargs, err := value.KwargsTryFromValue(args[len(args)-1])
	if err != nil {
		kwargs = value.NewKwargs(*value.NewValueMap())
	} else {
		args = args[:len(args)-1]
	}

	if optAttr := kwargs.GetValue("attribute"); optAttr.IsSome() {
		attrVal := optAttr.Unwrap()
		if len(args) != 0 {
			return nil, common.NewError(common.TooManyArguments, "")
		}
		defVal := kwargs.GetValue("default").UnwrapOr(value.Undefined)
		iter, err := state.undefinedBehavior().TryIter(val)
		if err != nil {
			return nil, err
		}
		var item value.Value
		for iter.Next().UnwrapTo(&item) {
			var subVal value.Value
			if optAttrStr := attrVal.AsStr(); optAttrStr.IsSome() {
				path := optAttrStr.Unwrap()
				subVal, err = value.GetPath(item, path)
			} else {
				subVal, err = value.GetItem(item, attrVal)
			}
			if err != nil {
				if defVal.IsUndefined() {
					return nil, err
				}
				subVal = defVal.Clone()
			} else if subVal.IsUndefined() {
				subVal = defVal.Clone()
			}
			rv = append(rv, subVal)
		}
		return rv, nil
	}

	// filter mapping
	if len(args) == 0 {
		return nil, common.NewError(common.InvalidOperation, "filter name is required")
	}
	filterNameVal := args[0]
	optFilterName := filterNameVal.AsStr()
	if optFilterName.IsNone() {
		return nil, common.NewError(common.InvalidOperation, "filter name must be a string")
	}
	filterName := optFilterName.Unwrap()
	optFilter := state.env.getFilter(filterName)
	if optFilter.IsNone() {
		return nil, common.NewError(common.UnknownFilter, "")
	}
	filter := optFilter.Unwrap()
	iter, err := state.undefinedBehavior().TryIter(val)
	if err != nil {
		return nil, err
	}
	var item value.Value
	for iter.Next().UnwrapTo(&item) {
		iter2, _ := value.ValueFromSlice([]value.Value{item.Clone()}).TryIter()
		iter3, _ := value.ValueFromSlice(args[1:]).TryIter()
		iter4 := iter2.Chain(iter3.Cloned())
		newArgs := iter4.Collect()
		rvItem, err := filter(state, newArgs)
		if err != nil {
			return nil, err
		}
		rv = append(rv, rvItem)
	}

	return rv, nil
}
