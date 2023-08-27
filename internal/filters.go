package internal

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/internal/datast/indexmap"
	"github.com/hnakamur/mjingo/internal/datast/option"
)

type FilterFunc = func(*State, []Value) (Value, error)

func filterFuncFromFilterWithStrArgValRet(f func(val string) Value) func(*State, []Value) (Value, error) {
	return func(state *State, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := StringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		return f(a), nil
	}
}

func filterFuncFromWithStateValArgValErrRet(f func(*State, Value) (Value, error)) func(*State, []Value) (Value, error) {
	return func(state *State, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(state, tpl.a)
	}
}

func filterFuncFromWithValArgValErrRet(f func(Value) (Value, error)) func(*State, []Value) (Value, error) {
	return func(state *State, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(tpl.a)
	}
}

func filterFuncFromFilterWithStrArgStrRet(f func(val string) string) func(*State, []Value) (Value, error) {
	return func(state *State, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := StringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		return ValueFromString(f(a)), nil
	}
}

func filterFuncFromFilterWithStateStrStrStrArgStrRet(f func(state *State, v1, v2, v3 string) string) func(*State, []Value) (Value, error) {
	return func(state *State, values []Value) (Value, error) {
		tpl, err := tuple3FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := StringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		b, err := StringFromValue(option.Some(tpl.b))
		if err != nil {
			return nil, err
		}
		c, err := StringFromValue(option.Some(tpl.c))
		if err != nil {
			return nil, err
		}
		return ValueFromString(f(state, a, b, c)), nil
	}
}

func filterFuncFromFilterWithValArgUintErrRet(f func(val Value) (uint, error)) func(*State, []Value) (Value, error) {
	return func(state *State, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		l, err := f(tpl.a)
		if err != nil {
			return nil, err
		}
		return ValueFromI64(int64(l)), nil
	}
}

func filterFuncFromFilterWithValOptStrArgStrErrRet(f func(val Value, optStr option.Option[string]) (string, error)) func(*State, []Value) (Value, error) {
	return func(state *State, values []Value) (Value, error) {
		var val Value
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
		return ValueFromString(rv), nil
	}
}

func safe(v string) Value {
	return ValueFromSafeString(v)
}

func escape(state *State, v Value) (Value, error) {
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
	return ValueFromSafeString(b.String()), nil
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

func length(val Value) (uint, error) {
	if optLen := val.Len(); optLen.IsSome() {
		return optLen.Unwrap(), nil
	}
	return 0, NewError(InvalidOperation,
		fmt.Sprintf("cannot calculate length of value of type %s", val.Kind()))
}

func compareValuesCaseInsensitive(a, b Value) int {
	if optA, optB := a.AsStr(), b.AsStr(); optA.IsSome() && optB.IsSome() {
		return strings.Compare(optA.Unwrap(), optB.Unwrap())
	}
	return Cmp(a, b)
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
func dictsort(v Value, kwargs Kwargs) (Value, error) {
	if v.Kind() != ValueKindMap {
		return nil, NewError(InvalidOperation, "cannot convert value into pair list")
	}
	entries := make([]indexmap.Entry[Value, Value], 0, v.Len().UnwrapOr(0))
	iter, err := v.TryIter()
	if err != nil {
		return nil, err
	}
	for {
		optKey := iter.Next()
		if optKey.IsNone() {
			break
		}
		key := optKey.Unwrap()
		val, err := getItem(v, key)
		if err != nil {
			val = Undefined
		}
		entries = append(entries, indexmap.Entry[Value, Value]{Key: key, Value: val})
	}

	byVal := false
	if optBy := kwargs.getValue("by"); optBy.IsSome() {
		if by, ok := optBy.Unwrap().(stringValue); ok {
			switch by.str {
			case "key":
				byVal = false
			case "value":
				byVal = true
			default:
				return nil, NewError(InvalidOperation,
					fmt.Sprintf("invalid value '%s' for 'by' parameter", by.str))
			}
		}
	}
	getKeyOrVal := func(entry indexmap.Entry[Value, Value]) Value { return entry.Key }
	if byVal {
		getKeyOrVal = func(entry indexmap.Entry[Value, Value]) Value { return entry.Value }
	}

	caseSensitive := false
	if optCaseSensitive := kwargs.getValue("case_sensitive"); optCaseSensitive.IsSome() {
		if cs, ok := optCaseSensitive.Unwrap().(BoolValue); ok && cs.B {
			caseSensitive = true
		}
	}
	sortFn := Cmp
	if caseSensitive {
		sortFn = compareValuesCaseInsensitive
	}

	reverse := false
	if optReverse := kwargs.getValue("reverse"); optReverse.IsSome() {
		if cs, ok := optReverse.Unwrap().(BoolValue); ok && cs.B {
			reverse = true
		}
	}

	slices.SortFunc(entries, func(a, b indexmap.Entry[Value, Value]) int {
		ret := sortFn(getKeyOrVal(a), getKeyOrVal(b))
		if reverse {
			return -ret
		}
		return ret
	})

	if err := kwargs.assertAllUsed(); err != nil {
		return nil, err
	}

	items := make([]Value, 0, len(entries))
	for _, entry := range entries {
		item := ValueFromSlice([]Value{entry.Key, entry.Value})
		items = append(items, item)
	}
	return ValueFromSlice(items), nil
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
func items(v Value) (Value, error) {
	if v.Kind() != ValueKindMap {
		return nil, NewError(InvalidOperation, "cannot convert value into pair list")
	}
	items := make([]Value, 0, v.Len().UnwrapOr(0))
	iter, err := v.TryIter()
	if err != nil {
		return nil, err
	}
	for {
		optKey := iter.Next()
		if optKey.IsNone() {
			break
		}
		key := optKey.Unwrap()
		val, err := getItem(v, key)
		if err != nil {
			val = Undefined
		}
		item := ValueFromSlice([]Value{key, val})
		items = append(items, item)
	}
	return ValueFromSlice(items), nil
}

// Joins a sequence by a character
func join(val Value, joiner option.Option[string]) (string, error) {
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
	return "", NewError(InvalidOperation,
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
func reverse(val Value) (Value, error) {
	if optValStr := val.AsStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		var b strings.Builder
		for len(rest) > 0 {
			r, size := utf8.DecodeLastRuneInString(rest)
			b.WriteRune(r)
			rest = rest[:len(rest)-size]
		}
		return ValueFromString(b.String()), nil
	}
	if optValSeq := val.AsSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		n := valSeq.ItemCount()
		items := make([]Value, 0, n)
		for i := n - 1; ; i-- {
			item := valSeq.GetItem(i).Unwrap()
			items = append(items, item)
			if i == 0 {
				break
			}
		}
		return ValueFromSlice(items), nil
	}
	return nil, NewError(InvalidOperation,
		fmt.Sprintf("cannot reverse value of type %s", val.Kind()))
}
