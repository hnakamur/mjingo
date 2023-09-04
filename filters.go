package mjingo

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type BoxedFilter = func(*vmState, []Value) (Value, error)

func boxedFilterFromFilterWithStrArgValRet(f func(val string) Value) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := stringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		return f(a), nil
	}
}

func boxedFilterFromWithStateValArgValErrRet(f func(*vmState, Value) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(state, tpl.a)
	}
}

func boxedFilterFromWithValArgValErrRet(f func(Value) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(tpl.a)
	}
}

func boxedFilterFromWithValValArgValErrRet(f func(Value, Value) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple2FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return f(tpl.a, tpl.b)
	}
}
func boxedFilterFromFilterWithStrArgStrRet(f func(val string) string) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := stringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		return valueFromString(f(a)), nil
	}
}

func boxedFilterFromFilterWithStateStrStrStrArgStrRet(f func(state *vmState, v1, v2, v3 string) string) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple3FromValues(state, values)
		if err != nil {
			return nil, err
		}
		a, err := stringFromValue(option.Some(tpl.a))
		if err != nil {
			return nil, err
		}
		b, err := stringFromValue(option.Some(tpl.b))
		if err != nil {
			return nil, err
		}
		c, err := stringFromValue(option.Some(tpl.c))
		if err != nil {
			return nil, err
		}
		return valueFromString(f(state, a, b, c)), nil
	}
}

func boxedFilterFromFilterWithValArgUintErrRet(f func(val Value) (uint, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		l, err := f(tpl.a)
		if err != nil {
			return nil, err
		}
		return valueFromI64(int64(l)), nil
	}
}

func boxedFilterFromFilterWithValOptStrArgStrErrRet(f func(val Value, optStr option.Option[string]) (string, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
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
		return valueFromString(rv), nil
	}
}

func boxedFilterFromFilterWithValOptI32ArgValErrRet(f func(val Value, optI32 option.Option[int32]) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var val Value
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
			n, err := i32TryFromValue(tpl2.b)
			if err != nil {
				return nil, err
			}
			optI32 = option.Some(n)
		}
		return f(val, optI32)
	}
}

func boxedFilterFromFilterWithValOptValArgValRet(f func(a Value, optB option.Option[Value]) Value) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var a Value
		optB := option.None[Value]()
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

func boxedFilterFromFilterWithStrOptStrArgStrRet(f func(s string, optStr option.Option[string]) string) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
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
		s, err := stringFromValue(option.Some(val))
		if err != nil {
			return nil, err
		}
		return valueFromString(f(s, optStr)), nil
	}
}

func boxedFilterFromFilterWithValArgBoolRet(f func(val Value) bool) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		return valueFromBool(f(tpl.a)), nil
	}
}

func boxedFilterFromFilterWithStateValUintOptValArgValErrRet(f func(*vmState, Value, uint, option.Option[Value]) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var val, countVal Value
		fillWith := option.None[Value]()
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
		count, err := countVal.tryToUint()
		if err != nil {
			return nil, err
		}
		return f(state, val, count, fillWith)
	}
}

func boxedFilterFromFilterWithStrUintOptBoolOptBoolArgStrrRet(f func(string, uint, option.Option[bool], option.Option[bool]) string) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var strVal, uintVal Value
		optBoolVal1 := option.None[Value]()
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
			b, err := boolTryFromValue(tpl4.d)
			if err != nil {
				return nil, err
			}
			optBool2 = option.Some(b)
		}
		s, err := stringFromValue(option.Some(strVal))
		if err != nil {
			return nil, err
		}
		n, err := uintVal.tryToUint()
		if err != nil {
			return nil, err
		}
		optBool1 := option.None[bool]()
		if optBoolVal1.IsSome() {
			b, err := boolTryFromValue(optBoolVal1.Unwrap())
			if err != nil {
				return nil, err
			}
			optBool1 = option.Some(b)
		}
		return valueFromString(f(s, n, optBool1, optBool2)), nil
	}
}

func boxedFilterFromFilterWithStateValOptStrValVarArgValSliceErrRet(f func(*vmState, Value, option.Option[string], ...Value) ([]Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var val Value
		optStr := option.None[string]()
		var args []Value
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
			s, err := stringFromValue(option.Some(tpl2.b))
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
		return valueFromSlice(rv), nil
	}
}

func boxedFilterFromFilterWithStateValStrOptStrValVarArgValSliceErrRet(f func(*vmState, Value, string, option.Option[string], ...Value) ([]Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var val, strVal Value
		optStr := option.None[string]()
		var args []Value
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
			s, err := stringFromValue(option.Some(tpl3.c))
			if err != nil {
				return nil, err
			}
			optStr = option.Some(s)
			args = values[3:]
		}
		s, err := stringFromValue(option.Some(strVal))
		if err != nil {
			return nil, err
		}
		rv, err := f(state, val, s, optStr, args...)
		if err != nil {
			return nil, err
		}
		return valueFromSlice(rv), nil
	}
}

func boxedFilterFromFilterWithValSliceArgValRet(f func([]Value) Value) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		tpl, err := tuple1FromValues(state, values)
		if err != nil {
			return nil, err
		}
		iter, err := state.undefinedBehavior().tryIter(tpl.a)
		if err != nil {
			return nil, err
		}
		return f(iter.Collect()), nil
	}
}

func boxedFilterFromWithValKwargsArgValErrRet(f func(Value, kwArgs) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var val Value
		var kwargs kwArgs
		switch {
		case len(values) <= 1:
			tpl, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs = newKwArgs(*newValueMap())
		case len(values) >= 2:
			tpl, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs, err = kwArgsTryFromValue(tpl.b)
			if err != nil {
				return nil, err
			}
		}
		return f(val, kwargs)
	}
}

func boxedFilterFromWithStateValKwargsArgValErrRet(f func(*vmState, Value, kwArgs) (Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		var val Value
		var kwargs kwArgs
		switch {
		case len(values) <= 1:
			tpl, err := tuple1FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs = newKwArgs(*newValueMap())
		case len(values) >= 2:
			tpl, err := tuple2FromValues(state, values)
			if err != nil {
				return nil, err
			}
			val = tpl.a
			kwargs, err = kwArgsTryFromValue(tpl.b)
			if err != nil {
				return nil, err
			}
		}
		return f(state, val, kwargs)
	}
}

func boxedFilterFromFilterWithStateValValVarArgValSliceErrRet(f func(*vmState, Value, ...Value) ([]Value, error)) func(*vmState, []Value) (Value, error) {
	return func(state *vmState, values []Value) (Value, error) {
		if len(values) == 0 {
			return nil, newError(MissingArgument, "")
		}
		val := values[0]
		args := values[1:]
		rv, err := f(state, val, args...)
		if err != nil {
			return nil, err
		}
		return valueFromSlice(rv), nil
	}
}

func safe(v string) Value {
	return ValueFromSafeString(v)
}

func escape(state *vmState, v Value) (Value, error) {
	if v.isSafe() {
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
	if optStr := v.asStr(); optStr.IsSome() {
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
func replace(_ *vmState, v, from, to string) string {
	r := strings.NewReplacer(from, to)
	return r.Replace(v)
}

func length(val Value) (uint, error) {
	if optLen := val.len(); optLen.IsSome() {
		return optLen.Unwrap(), nil
	}
	return 0, newError(InvalidOperation,
		fmt.Sprintf("cannot calculate length of value of type %s", val.kind()))
}

func compareValuesCaseInsensitive(a, b Value) int {
	if optA, optB := a.asStr(), b.asStr(); optA.IsSome() && optB.IsSome() {
		return strings.Compare(strings.ToLower(optA.Unwrap()), strings.ToLower(optB.Unwrap()))
	}
	return valueCmp(a, b)
}

type keyAndValue struct {
	Key   Value
	Value Value
}

// Dict sorting functionality.
//
// This filter works like `|items` but sorts the pairs by key first.
//
// The filter accepts a few keyword arguments:
//
// * `case_sensitive`: set to `true` to make the sorting of strings case sensitive.
// * `by`: set to `"value"` to sort by  Defaults to `"key"`.
// * `reverse`: set to `true` to sort in reverse.
func dictsort(v Value, kwargs kwArgs) (Value, error) {
	if v.kind() != valueKindMap {
		return nil, newError(InvalidOperation, "cannot convert value into pair list")
	}
	entries := make([]keyAndValue, 0, v.len().UnwrapOr(0))
	iter, err := v.tryIter()
	if err != nil {
		return nil, err
	}
	var key Value
	for iter.Next().UnwrapTo(&key) {
		val, err := getItem(v, key)
		if err != nil {
			val = Undefined
		}
		entries = append(entries, keyAndValue{Key: key, Value: val})
	}

	byVal := false
	if optBy := kwargs.GetValue("by"); optBy.IsSome() {
		if by, ok := optBy.Unwrap().(stringValue); ok {
			switch by.Str {
			case "key":
				byVal = false
			case "value":
				byVal = true
			default:
				return nil, newError(InvalidOperation,
					fmt.Sprintf("invalid value '%s' for 'by' parameter", by.Str))
			}
		}
	}
	getKeyOrVal := func(entry keyAndValue) Value { return entry.Key }
	if byVal {
		getKeyOrVal = func(entry keyAndValue) Value { return entry.Value }
	}

	caseSensitive := false
	if optCaseSensitive := kwargs.GetValue("case_sensitive"); optCaseSensitive.IsSome() {
		if cs, ok := optCaseSensitive.Unwrap().(boolValue); ok && cs.B {
			caseSensitive = true
		}
	}
	sortFn := valueCmp
	if !caseSensitive {
		sortFn = compareValuesCaseInsensitive
	}

	reverse := false
	if optReverse := kwargs.GetValue("reverse"); optReverse.IsSome() {
		if cs, ok := optReverse.Unwrap().(boolValue); ok && cs.B {
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

	items := make([]Value, 0, len(entries))
	for _, entry := range entries {
		item := valueFromSlice([]Value{entry.Key, entry.Value})
		items = append(items, item)
	}
	return valueFromSlice(items), nil
}

func sortFilter(state *vmState, val Value, kwargs kwArgs) (Value, error) {
	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, newError(InvalidOperation, "cannot convert value to list").withSource(err)
	}
	items := iter.Collect()
	caseSensitive := false
	if optCaseSensitive := kwargs.GetValue("case_sensitive"); optCaseSensitive.IsSome() {
		if cs, ok := optCaseSensitive.Unwrap().(boolValue); ok && cs.B {
			caseSensitive = true
		}
	}
	sortFn := valueCmp
	if !caseSensitive {
		sortFn = compareValuesCaseInsensitive
	}

	var attr string
	if optAttr := kwargs.GetValue("attribute"); optAttr.IsSome() {
		if strVal, ok := optAttr.Unwrap().(stringValue); ok {
			attr = strVal.Str
		}
	}
	reverse := false
	if optReverse := kwargs.GetValue("reverse"); optReverse.IsSome() {
		if cs, ok := optReverse.Unwrap().(boolValue); ok && cs.B {
			reverse = true
		}
	}

	if attr != "" {
		slices.SortFunc(items, func(a, b Value) int {
			aVal, err := getPath(a, attr)
			if err != nil {
				return 0
			}
			bVal, err := getPath(b, attr)
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
		slices.SortFunc(items, func(a, b Value) int {
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
	return valueFromSlice(items), nil
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
	if v.kind() != valueKindMap {
		return nil, newError(InvalidOperation, "cannot convert value into pair list")
	}
	items := make([]Value, 0, v.len().UnwrapOr(0))
	iter, err := v.tryIter()
	if err != nil {
		return nil, err
	}
	var key Value
	for iter.Next().UnwrapTo(&key) {
		val, err := getItem(v, key)
		if err != nil {
			val = Undefined
		}
		item := valueFromSlice([]Value{key, val})
		items = append(items, item)
	}
	return valueFromSlice(items), nil
}

// Joins a sequence by a character
func join(val Value, joiner option.Option[string]) (string, error) {
	if val.isUndefined() || val.isNone() {
		return "", nil
	}

	joinerStr := joiner.UnwrapOr("")
	if optValStr := val.asStr(); optValStr.IsSome() {
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
	if optValSeq := val.asSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		var b strings.Builder
		n := valSeq.ItemCount()
		for i := uint(0); i < n; i++ {
			if b.Len() != 0 {
				b.WriteString(joinerStr)
			}
			item := valSeq.GetItem(i).Unwrap()
			if optItemStr := item.asStr(); optItemStr.IsSome() {
				b.WriteString(optItemStr.Unwrap())
			} else {
				fmt.Fprintf(&b, "%s", item)
			}
		}
		return b.String(), nil
	}
	return "", newError(InvalidOperation,
		fmt.Sprintf("cannot join value of type %s", val.kind()))
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
	if optValStr := val.asStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		var b strings.Builder
		for len(rest) > 0 {
			r, size := utf8.DecodeLastRuneInString(rest)
			b.WriteRune(r)
			rest = rest[:len(rest)-size]
		}
		return valueFromString(b.String()), nil
	}
	if optValSeq := val.asSeq(); optValSeq.IsSome() {
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
		return valueFromSlice(items), nil
	}
	return nil, newError(InvalidOperation,
		fmt.Sprintf("cannot reverse value of type %s", val.kind()))
}

func trim(s string, cutset option.Option[string]) string {
	if cutset.IsSome() {
		return strings.Trim(s, cutset.Unwrap())
	}
	return strings.TrimSpace(s)
}

func defaultFilter(val Value, other option.Option[Value]) Value {
	if val.isUndefined() {
		return other.UnwrapOrElse(func() Value { return valueFromString("") })
	}
	return val
}

func round(val Value, precision option.Option[int32]) (Value, error) {
	switch v := val.(type) {
	case i64Value, i128Value:
		return val, nil
	case f64Value:
		x := math.Pow10(int(precision.UnwrapOr(0)))
		return valueFromF64(math.Round(x*v.F) / x), nil
	default:
		return nil, newError(InvalidOperation, "cannot round value")
	}
}

func abs(val Value) (Value, error) {
	switch v := val.(type) {
	case i64Value:
		n := v.N
		if n < 0 {
			n = -n
		}
		return i64Value{N: n}, nil
	case i128Value:
		var n big.Int
		n.Abs(&v.N)
		return i128Value{N: n}, nil
	case f64Value:
		return f64Value{F: math.Abs(v.F)}, nil
	default:
		// TODO: Verify MiniJinja error message is really intentional.
		return nil, newError(InvalidOperation, "cannot round value")
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
func attr(val, key Value) (Value, error) {
	return getItem(val, key)
}

func first(val Value) (Value, error) {
	if optValStr := val.asStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		if rest == "" {
			return Undefined, nil
		}
		var b strings.Builder
		r, _ := utf8.DecodeRuneInString(rest)
		b.WriteRune(r)
		return valueFromString(b.String()), nil
	}
	if optValSeq := val.asSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		return valSeq.GetItem(0).UnwrapOr(Undefined), nil
	}
	return nil, newError(InvalidOperation, "cannot get first item from value")
}

func last(val Value) (Value, error) {
	if optValStr := val.asStr(); optValStr.IsSome() {
		rest := optValStr.Unwrap()
		if rest == "" {
			return Undefined, nil
		}
		var b strings.Builder
		r, _ := utf8.DecodeLastRuneInString(rest)
		b.WriteRune(r)
		return valueFromString(b.String()), nil
	}
	if optValSeq := val.asSeq(); optValSeq.IsSome() {
		valSeq := optValSeq.Unwrap()
		n := valSeq.ItemCount()
		if n == 0 {
			return Undefined, nil
		}
		return valSeq.GetItem(n - 1).UnwrapOr(Undefined), nil
	}
	return nil, newError(InvalidOperation, "cannot get last item from value")
}

func minFilter(state *vmState, val Value) (Value, error) {
	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, newError(InvalidDelimiter, "cannot convert value to list").withSource(err)
	}
	return iter.Min().UnwrapOr(Undefined), nil
}

func maxFilter(state *vmState, val Value) (Value, error) {
	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, newError(InvalidDelimiter, "cannot convert value to list").withSource(err)
	}
	return iter.Max().UnwrapOr(Undefined), nil
}

func listFilter(state *vmState, val Value) (Value, error) {
	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, newError(InvalidDelimiter, "cannot convert value to list").withSource(err)
	}
	return valueFromSlice(iter.Collect()), nil
}

// Converts the value into a boolean
//
// This behaves the same as the if statement does with regards to
// handling of boolean values.
func boolFilter(val Value) bool {
	return val.isTrue()
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
func batchFilter(state *vmState, val Value, count uint, fillWith option.Option[Value]) (Value, error) {
	if count == 0 {
		return nil, newError(InvalidOperation, "count cannot be 0")
	}

	rv := make([]Value, 0, val.len().UnwrapOr(0)/count)
	tmp := make([]Value, 0, count)

	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, err
	}
	var item Value
	for iter.Next().UnwrapTo(&item) {
		if uint(len(tmp)) == count {
			rv = append(rv, valueFromSlice(tmp))
			tmp = make([]Value, 0, count)
		}
		tmp = append(tmp, item)
	}

	if len(tmp) != 0 {
		if fillWith.IsSome() {
			filler := fillWith.Unwrap()
			for i := uint(0); i < count-uint(len(tmp)); i++ {
				tmp = append(tmp, filler.clone())
			}
		}
		rv = append(rv, valueFromSlice(tmp))
	}

	return valueFromSlice(rv), nil
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
func sliceFilter(state *vmState, val Value, count uint, fillWith option.Option[Value]) (Value, error) {
	if count == 0 {
		return nil, newError(InvalidOperation, "count cannot be 0")
	}

	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, err
	}
	items := iter.Collect()
	l := uint(len(items))
	itemsPerSlice := l / count
	slicesWithExtra := l % count
	offset := uint(0)
	rv := make([]Value, 0, count)
	for slice := uint(0); slice < count; slice++ {
		start := offset + slice*itemsPerSlice
		if slice < slicesWithExtra {
			offset++
		}
		end := offset + (slice+1)*itemsPerSlice
		tmp := items[start:end]
		if fillWith.IsSome() && slice >= slicesWithExtra {
			filler := fillWith.Unwrap()
			tmp = append(tmp, filler.clone())
			rv = append(rv, valueFromSlice(tmp))
			continue
		}
		rv = append(rv, valueFromSlice(tmp))
	}

	return valueFromSlice(rv), nil
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

func selectOrReject(state *vmState, invert bool, val Value, attr, testName option.Option[string], args ...Value) ([]Value, error) {
	var rv []Value
	test := option.None[BoxedTest]()
	if testName.IsSome() {
		test = state.env.getTest(testName.Unwrap())
		if test.IsNone() {
			return nil, newError(UnknownTest, "")
		}
	}
	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, err
	}
	var item Value
	for iter.Next().UnwrapTo(&item) {
		var testVal Value
		if attr.IsSome() {
			testVal, err = getAttr(item, attr.Unwrap())
			if err != nil {
				return nil, err
			}
		} else {
			testVal = item.clone()
		}
		var passed bool
		if test.IsSome() {
			iter, _ := valueFromSlice([]Value{testVal}).tryIter()
			iter2, _ := valueFromSlice(args).tryIter()
			chainedIter := iter.Chain(iter2.Cloned())
			testArgs := chainedIter.Collect()
			passed, err = test.Unwrap()(state, testArgs)
			if err != nil {
				return nil, err
			}
		} else {
			passed = testVal.isTrue()
		}
		if passed != invert {
			rv = append(rv, item)
		}
	}
	return rv, nil
}

func selectFilter(state *vmState, val Value, testName option.Option[string], args ...Value) ([]Value, error) {
	return selectOrReject(state, false, val, option.None[string](), testName, args...)
}

func selectAttrFilter(state *vmState, val Value, attr string, testName option.Option[string], args ...Value) ([]Value, error) {
	return selectOrReject(state, false, val, option.Some(attr), testName, args...)
}

func rejectFilter(state *vmState, val Value, testName option.Option[string], args ...Value) ([]Value, error) {
	return selectOrReject(state, true, val, option.None[string](), testName, args...)
}

func rejectAttrFilter(state *vmState, val Value, attr string, testName option.Option[string], args ...Value) ([]Value, error) {
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
func uniqueFilter(values []Value) Value {
	var rv []Value
	seen := make(map[Value]struct{})
	for _, item := range values {
		if _, ok := seen[item]; !ok {
			rv = append(rv, item)
			seen[item] = struct{}{}
		}
	}
	return valueFromSlice(rv)
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
func mapFilter(state *vmState, val Value, args ...Value) ([]Value, error) {
	rv := make([]Value, 0, val.len().UnwrapOr(0))
	kwargs, err := kwArgsTryFromValue(args[len(args)-1])
	if err != nil {
		kwargs = newKwArgs(*newValueMap())
	} else {
		args = args[:len(args)-1]
	}

	if optAttr := kwargs.GetValue("attribute"); optAttr.IsSome() {
		attrVal := optAttr.Unwrap()
		if len(args) != 0 {
			return nil, newError(TooManyArguments, "")
		}
		defVal := kwargs.GetValue("default").UnwrapOr(Undefined)
		iter, err := state.undefinedBehavior().tryIter(val)
		if err != nil {
			return nil, err
		}
		var item Value
		for iter.Next().UnwrapTo(&item) {
			var subVal Value
			if optAttrStr := attrVal.asStr(); optAttrStr.IsSome() {
				path := optAttrStr.Unwrap()
				subVal, err = getPath(item, path)
			} else {
				subVal, err = getItem(item, attrVal)
			}
			if err != nil {
				if defVal.isUndefined() {
					return nil, err
				}
				subVal = defVal.clone()
			} else if subVal.isUndefined() {
				subVal = defVal.clone()
			}
			rv = append(rv, subVal)
		}
		return rv, nil
	}

	// filter mapping
	if len(args) == 0 {
		return nil, newError(InvalidOperation, "filter name is required")
	}
	filterNameVal := args[0]
	optFilterName := filterNameVal.asStr()
	if optFilterName.IsNone() {
		return nil, newError(InvalidOperation, "filter name must be a string")
	}
	filterName := optFilterName.Unwrap()
	optFilter := state.env.getFilter(filterName)
	if optFilter.IsNone() {
		return nil, newError(UnknownFilter, "")
	}
	filter := optFilter.Unwrap()
	iter, err := state.undefinedBehavior().tryIter(val)
	if err != nil {
		return nil, err
	}
	var item Value
	for iter.Next().UnwrapTo(&item) {
		iter2, _ := valueFromSlice([]Value{item.clone()}).tryIter()
		iter3, _ := valueFromSlice(args[1:]).tryIter()
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