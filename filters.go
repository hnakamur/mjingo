package mjingo

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hnakamur/mjingo/option"
)

type BoxedFilter = func(State, []Value) (Value, error)

func BoxedFilterFromFunc(fn any) BoxedFilter {
	if bf, ok := fn.(BoxedFilter); ok {
		return bf
	}

	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("argument must be a function")
	}

	numOut := fnType.NumOut()
	if numOut != 1 && numOut != 2 {
		panic("return value count must be 1 or 2")
	}
	if !canConvertibleToValue(fnType.Out(0)) {
		panic("first return value type is unsupported")
	}
	if numOut == 2 {
		assertType[error](fnType.Out(1), "type of second return value must be error")
	}

	numIn := fnType.NumIn()
	if numIn < 1 && numIn > 5 {
		panic("only functions with argument count between 1 and 5 are supported")
	}
	optCount := checkFuncArgTypes(fnType)

	fnVal := reflect.ValueOf(fn)
	return func(state State, values []Value) (Value, error) {
		reflectVals := make([]reflect.Value, 0, numIn)
		inOffset := 0
		if fnType.In(0) == reflectType[State]() {
			reflectVals = append(reflectVals, reflect.ValueOf(state))
			inOffset++
		}
		wantValuesLen := numIn - inOffset
		if fnType.IsVariadic() {
			wantValuesLen--
		}
		if len(values) < wantValuesLen-optCount {
			return nil, NewError(MissingArgument, "")
		}
		if len(values) > wantValuesLen && !fnType.IsVariadic() {
			return nil, NewError(TooManyArguments, "")
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
			goVal, err := valueTryToGoValueReflect(val, argType)
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

func safe(v string) Value {
	return ValueFromSafeString(v)
}

func escape(state State, v Value) (Value, error) {
	if v.isSafe() {
		return v, nil
	}

	// this tries to use the escaping flag of the current scope, then
	// of the initial state and if that is also not set it falls back
	// to HTML.
	autoEscape := state.AutoEscape()
	if _, ok := state.AutoEscape().(autoEscapeNone); ok {
		if _, ok := state.Env().initialAutoEscape(state.Name()).(autoEscapeNone); ok {
			autoEscape = autoEscapeHTML{}
		}
	}
	var b strings.Builder
	if s, ok := valueAsGoString(v); ok {
		b.Grow(len(s))
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
func replace(_ State, v, from, to string) string {
	r := strings.NewReplacer(from, to)
	return r.Replace(v)
}

func length(val Value) (uint, error) {
	if optLen := val.len(); optLen.IsSome() {
		return optLen.Unwrap(), nil
	}
	return 0, NewError(InvalidOperation,
		fmt.Sprintf("cannot calculate length of value of type %s", val.kind()))
}

func compareValuesCaseInsensitive(a, b Value) int {
	strA, okA := valueAsGoString(a)
	strB, okB := valueAsGoString(b)
	if okA && okB {
		return strings.Compare(strings.ToLower(strA), strings.ToLower(strB))
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
		return nil, NewError(InvalidOperation, "cannot convert value into pair list")
	}
	entries := make([]keyAndValue, 0, v.len().UnwrapOr(0))
	iter, err := v.tryIter()
	if err != nil {
		return nil, err
	}
	for key := Value(nil); iter.Next().UnwrapTo(&key); {
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
				return nil, NewError(InvalidOperation,
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

func sortFilter(state State, val Value, kwargs kwArgs) (Value, error) {
	iter, err := state.UndefinedBehavior().tryIter(val)
	if err != nil {
		return nil, NewError(InvalidOperation, "cannot convert value to list").withSource(err)
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
		return nil, NewError(InvalidOperation, "cannot convert value into pair list")
	}
	items := make([]Value, 0, v.len().UnwrapOr(0))
	iter, err := v.tryIter()
	if err != nil {
		return nil, err
	}
	for key := Value(nil); iter.Next().UnwrapTo(&key); {
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
	if rest, err := valueTryToGoString(val); err == nil {
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
			if itemStr, err := valueTryToGoString(item); err == nil {
				b.WriteString(itemStr)
			} else {
				fmt.Fprintf(&b, "%s", item)
			}
		}
		return b.String(), nil
	}
	return "", NewError(InvalidOperation,
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
	if rest, err := valueTryToGoString(val); err == nil {
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
	return nil, NewError(InvalidOperation,
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
		return nil, NewError(InvalidOperation, "cannot round value")
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
		return nil, NewError(InvalidOperation, "cannot round value")
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
	if rest, err := valueTryToGoString(val); err == nil {
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
	return nil, NewError(InvalidOperation, "cannot get first item from value")
}

func last(val Value) (Value, error) {
	if rest, err := valueTryToGoString(val); err == nil {
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
	return nil, NewError(InvalidOperation, "cannot get last item from value")
}

func minFilter(state State, val Value) (Value, error) {
	iter, err := state.UndefinedBehavior().tryIter(val)
	if err != nil {
		return nil, NewError(InvalidDelimiter, "cannot convert value to list").withSource(err)
	}
	return iter.Min().UnwrapOr(Undefined), nil
}

func maxFilter(state State, val Value) (Value, error) {
	iter, err := state.UndefinedBehavior().tryIter(val)
	if err != nil {
		return nil, NewError(InvalidDelimiter, "cannot convert value to list").withSource(err)
	}
	return iter.Max().UnwrapOr(Undefined), nil
}

func listFilter(state State, val Value) (Value, error) {
	iter, err := state.UndefinedBehavior().tryIter(val)
	if err != nil {
		return nil, NewError(InvalidDelimiter, "cannot convert value to list").withSource(err)
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
func batchFilter(state State, val Value, count uint, fillWith option.Option[Value]) (Value, error) {
	if count == 0 {
		return nil, NewError(InvalidOperation, "count cannot be 0")
	}

	rv := make([]Value, 0, val.len().UnwrapOr(0)/count)
	tmp := make([]Value, 0, count)

	iter, err := state.UndefinedBehavior().tryIter(val)
	if err != nil {
		return nil, err
	}
	for item := Value(nil); iter.Next().UnwrapTo(&item); {
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
func sliceFilter(state State, val Value, count uint, fillWith option.Option[Value]) (Value, error) {
	if count == 0 {
		return nil, NewError(InvalidOperation, "count cannot be 0")
	}

	iter, err := state.UndefinedBehavior().tryIter(val)
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

func selectOrReject(state State, invert bool, val Value, attr, testName option.Option[string], args ...Value) ([]Value, error) {
	var rv []Value
	test := option.None[BoxedTest]()
	if testName.IsSome() {
		test = state.Env().getTest(testName.Unwrap())
		if test.IsNone() {
			return nil, NewError(UnknownTest, "")
		}
	}
	iter, err := state.UndefinedBehavior().tryIter(val)
	if err != nil {
		return nil, err
	}
	for item := Value(nil); iter.Next().UnwrapTo(&item); {
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

func selectFilter(state State, val Value, testName option.Option[string], args ...Value) ([]Value, error) {
	return selectOrReject(state, false, val, option.None[string](), testName, args...)
}

func selectAttrFilter(state State, val Value, attr string, testName option.Option[string], args ...Value) ([]Value, error) {
	return selectOrReject(state, false, val, option.Some(attr), testName, args...)
}

func rejectFilter(state State, val Value, testName option.Option[string], args ...Value) ([]Value, error) {
	return selectOrReject(state, true, val, option.None[string](), testName, args...)
}

func rejectAttrFilter(state State, val Value, attr string, testName option.Option[string], args ...Value) ([]Value, error) {
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
func mapFilter(state State, val Value, args ...Value) ([]Value, error) {
	rv := make([]Value, 0, val.len().UnwrapOr(0))
	var kwargs kwArgs
	var err error
	if len(args) == 0 {
		kwargs = newKwArgs(*newValueMap())
	} else {
		kwargs, err = valueTryToKwArgs(args[len(args)-1])
		if err != nil {
			kwargs = newKwArgs(*newValueMap())
		} else {
			args = args[:len(args)-1]
		}
	}

	if optAttr := kwargs.GetValue("attribute"); optAttr.IsSome() {
		attrVal := optAttr.Unwrap()
		if len(args) != 0 {
			return nil, NewError(TooManyArguments, "")
		}
		defVal := kwargs.GetValue("default").UnwrapOr(Undefined)
		iter, err := state.UndefinedBehavior().tryIter(val)
		if err != nil {
			return nil, err
		}
		for item := Value(nil); iter.Next().UnwrapTo(&item); {
			var subVal Value
			if path, err := valueTryToGoString(attrVal); err == nil {
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
		return nil, NewError(InvalidOperation, "filter name is required")
	}
	filterNameVal := args[0]
	filterName, err := valueTryToGoString(filterNameVal)
	if err != nil {
		return nil, NewError(InvalidOperation, "filter name must be a string")
	}
	optFilter := state.Env().getFilter(filterName)
	if optFilter.IsNone() {
		return nil, NewError(UnknownFilter, "")
	}
	filter := optFilter.Unwrap()
	iter, err := state.UndefinedBehavior().tryIter(val)
	if err != nil {
		return nil, err
	}
	for item := Value(nil); iter.Next().UnwrapTo(&item); {
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
