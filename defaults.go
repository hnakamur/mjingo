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

func getDefaultBuiltinFilters() map[string]boxedFilter {
	rv := make(map[string]boxedFilter)
	rv["safe"] = boxedFilterFromFunc(safe)
	rv["escape"] = boxedFilterFromFunc(escape)
	rv["e"] = boxedFilterFromFunc(escape)

	rv["lower"] = boxedFilterFromFunc(lower)
	rv["upper"] = boxedFilterFromFunc(upper)
	rv["title"] = boxedFilterFromFunc(title)
	rv["capitalize"] = boxedFilterFromFunc(capitalize)
	rv["replace"] = boxedFilterFromFunc(replace)
	rv["length"] = boxedFilterFromFunc(length)
	rv["count"] = boxedFilterFromFunc(length)
	rv["dictsort"] = boxedFilterFromFunc(dictsort)
	rv["items"] = boxedFilterFromFunc(items)
	rv["reverse"] = boxedFilterFromFunc(reverse)
	rv["trim"] = boxedFilterFromFunc(trim)
	rv["join"] = boxedFilterFromFunc(join)
	rv["default"] = boxedFilterFromFunc(defaultFilter)
	rv["d"] = boxedFilterFromFunc(defaultFilter)
	rv["round"] = boxedFilterFromFunc(round)
	rv["abs"] = boxedFilterFromFunc(abs)
	rv["attr"] = boxedFilterFromFunc(attr)
	rv["first"] = boxedFilterFromFunc(first)
	rv["last"] = boxedFilterFromFunc(last)
	rv["min"] = boxedFilterFromFunc(minFilter)
	rv["max"] = boxedFilterFromFunc(maxFilter)
	rv["sort"] = boxedFilterFromFunc(sortFilter)
	rv["list"] = boxedFilterFromFunc(listFilter)
	rv["bool"] = boxedFilterFromFunc(boolFilter)
	rv["batch"] = boxedFilterFromFunc(batchFilter)
	rv["slice"] = boxedFilterFromFunc(sliceFilter)
	rv["indent"] = boxedFilterFromFunc(indentFilter)
	rv["select"] = boxedFilterFromFunc(selectFilter)
	rv["reject"] = boxedFilterFromFunc(rejectFilter)
	rv["selectattr"] = boxedFilterFromFunc(selectAttrFilter)
	rv["rejectattr"] = boxedFilterFromFunc(rejectAttrFilter)
	rv["map"] = boxedFilterFromFunc(mapFilter)
	rv["unique"] = boxedFilterFromFunc(uniqueFilter)

	return rv
}

func getDefaultBuiltinTests() map[string]boxedTest {
	rv := make(map[string]boxedTest)
	rv["undefined"] = boxedTestFromFunc(isUndefined)
	rv["defined"] = boxedTestFromFunc(isDefined)
	rv["none"] = boxedTestFromFunc(isNone)
	rv["safe"] = boxedTestFromFunc(isSafe)
	rv["escaped"] = boxedTestFromFunc(isSafe)

	rv["odd"] = boxedTestFromFunc(isOdd)
	rv["even"] = boxedTestFromFunc(isEven)
	rv["number"] = boxedTestFromFunc(isNumber)
	rv["string"] = boxedTestFromFunc(isString)
	rv["sequence"] = boxedTestFromFunc(isSequence)
	rv["mapping"] = boxedTestFromFunc(isMapping)
	rv["startingwith"] = boxedTestFromFunc(isStartingWith)
	rv["endingwith"] = boxedTestFromFunc(isEndingWith)

	// operators
	rv["eq"] = boxedTestFromFunc(isEq)
	rv["equalto"] = boxedTestFromFunc(isEq)
	rv["=="] = boxedTestFromFunc(isEq)
	rv["ne"] = boxedTestFromFunc(isNe)
	rv["!="] = boxedTestFromFunc(isNe)
	rv["lt"] = boxedTestFromFunc(isLt)
	rv["lessthan"] = boxedTestFromFunc(isLt)
	rv["<"] = boxedTestFromFunc(isLt)
	rv["le"] = boxedTestFromFunc(isLe)
	rv["<="] = boxedTestFromFunc(isLe)
	rv["gt"] = boxedTestFromFunc(isGt)
	rv["greaterthan"] = boxedTestFromFunc(isGt)
	rv[">"] = boxedTestFromFunc(isGt)
	rv["ge"] = boxedTestFromFunc(isGe)
	rv[">="] = boxedTestFromFunc(isGe)
	rv["in"] = boxedTestFromFunc(isIn)
	rv["true"] = boxedTestFromFunc(isTrue)
	rv["false"] = boxedTestFromFunc(isFalse)
	rv["filter"] = boxedTestFromFunc(isFilter)
	rv["test"] = boxedTestFromFunc(isTest)

	return rv
}

func getDefaultGlobals() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = valueFromBoxedFunc(boxedFuncFromFunc(rangeFunc))
	rv["dict"] = valueFromBoxedFunc(boxedFuncFromFunc(dictFunc))
	return rv
}
