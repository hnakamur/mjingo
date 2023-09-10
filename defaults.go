package mjingo

import "strings"

func noAutoEscape(_ string) AutoEscape { return autoEscapeNone{} }

// DefaultAutoEscapeCallback is the default logic for auto escaping based on file extension.
//
//   - AutoEscapeHTML: `.html`, `.htm`, `.xml`
//   - AutoEscapeJSON: `.json`, `.json5`, `.js`, `.yaml`, `.yml`
//   - AutoEscapeNone: all others
func DefaultAutoEscapeCallback(name string) AutoEscape {
	_, suffix, found := strings.Cut(name, ".")
	if found {
		switch suffix {
		case "html", "htm", "xml":
			return AutoEscapeHTML
		case "json", "json5", "js", "yaml", "yml":
			return AutoEscapeJSON
		}
	}
	return AutoEscapeNone
}

func escapeFormatter(out *output, state *vmState, val Value) error {
	return writeEscaped(out, state.autoEscape, val)
}

func getDefaultBuiltinFilters() map[string]BoxedFilter {
	rv := make(map[string]BoxedFilter)
	rv["safe"] = BoxedFilterFromFunc(safe)
	rv["escape"] = BoxedFilterFromFunc(escape)
	rv["e"] = BoxedFilterFromFunc(escape)

	rv["lower"] = BoxedFilterFromFunc(lower)
	rv["upper"] = BoxedFilterFromFunc(upper)
	rv["title"] = BoxedFilterFromFunc(title)
	rv["capitalize"] = BoxedFilterFromFunc(capitalize)
	rv["replace"] = BoxedFilterFromFunc(replace)
	rv["length"] = BoxedFilterFromFunc(length)
	rv["count"] = BoxedFilterFromFunc(length)
	rv["dictsort"] = BoxedFilterFromFunc(dictsort)
	rv["items"] = BoxedFilterFromFunc(items)
	rv["reverse"] = BoxedFilterFromFunc(reverse)
	rv["trim"] = BoxedFilterFromFunc(trim)
	rv["join"] = BoxedFilterFromFunc(join)
	rv["default"] = BoxedFilterFromFunc(defaultFilter)
	rv["d"] = BoxedFilterFromFunc(defaultFilter)
	rv["round"] = BoxedFilterFromFunc(round)
	rv["abs"] = BoxedFilterFromFunc(abs)
	rv["attr"] = BoxedFilterFromFunc(attr)
	rv["first"] = BoxedFilterFromFunc(first)
	rv["last"] = BoxedFilterFromFunc(last)
	rv["min"] = BoxedFilterFromFunc(minFilter)
	rv["max"] = BoxedFilterFromFunc(maxFilter)
	rv["sort"] = BoxedFilterFromFunc(sortFilter)
	rv["list"] = BoxedFilterFromFunc(listFilter)
	rv["bool"] = BoxedFilterFromFunc(boolFilter)
	rv["batch"] = BoxedFilterFromFunc(batchFilter)
	rv["slice"] = BoxedFilterFromFunc(sliceFilter)
	rv["indent"] = BoxedFilterFromFunc(indentFilter)
	rv["select"] = BoxedFilterFromFunc(selectFilter)
	rv["reject"] = BoxedFilterFromFunc(rejectFilter)
	rv["selectattr"] = BoxedFilterFromFunc(selectAttrFilter)
	rv["rejectattr"] = BoxedFilterFromFunc(rejectAttrFilter)
	rv["map"] = BoxedFilterFromFunc(mapFilter)
	rv["unique"] = BoxedFilterFromFunc(uniqueFilter)

	return rv
}

func getDefaultBuiltinTests() map[string]BoxedTest {
	rv := make(map[string]BoxedTest)
	rv["undefined"] = BoxedTestFromFunc(isUndefined)
	rv["defined"] = BoxedTestFromFunc(isDefined)
	rv["none"] = BoxedTestFromFunc(isNone)
	rv["safe"] = BoxedTestFromFunc(isSafe)
	rv["escaped"] = BoxedTestFromFunc(isSafe)

	rv["odd"] = BoxedTestFromFunc(isOdd)
	rv["even"] = BoxedTestFromFunc(isEven)
	rv["number"] = BoxedTestFromFunc(isNumber)
	rv["string"] = BoxedTestFromFunc(isString)
	rv["sequence"] = BoxedTestFromFunc(isSequence)
	rv["mapping"] = BoxedTestFromFunc(isMapping)
	rv["startingwith"] = BoxedTestFromFunc(isStartingWith)
	rv["endingwith"] = BoxedTestFromFunc(isEndingWith)

	// operators
	rv["eq"] = BoxedTestFromFunc(isEq)
	rv["equalto"] = BoxedTestFromFunc(isEq)
	rv["=="] = BoxedTestFromFunc(isEq)
	rv["ne"] = BoxedTestFromFunc(isNe)
	rv["!="] = BoxedTestFromFunc(isNe)
	rv["lt"] = BoxedTestFromFunc(isLt)
	rv["lessthan"] = BoxedTestFromFunc(isLt)
	rv["<"] = BoxedTestFromFunc(isLt)
	rv["le"] = BoxedTestFromFunc(isLe)
	rv["<="] = BoxedTestFromFunc(isLe)
	rv["gt"] = BoxedTestFromFunc(isGt)
	rv["greaterthan"] = BoxedTestFromFunc(isGt)
	rv[">"] = BoxedTestFromFunc(isGt)
	rv["ge"] = BoxedTestFromFunc(isGe)
	rv[">="] = BoxedTestFromFunc(isGe)
	rv["in"] = BoxedTestFromFunc(isIn)
	rv["true"] = BoxedTestFromFunc(isTrue)
	rv["false"] = BoxedTestFromFunc(isFalse)
	rv["filter"] = BoxedTestFromFunc(isFilter)
	rv["test"] = BoxedTestFromFunc(isTest)

	return rv
}

func getDefaultGlobals() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = valueFromBoxedFunc(BoxedFuncFromFunc(rangeFunc))
	rv["dict"] = valueFromBoxedFunc(BoxedFuncFromFunc(dictFunc))
	return rv
}
