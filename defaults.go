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

func escapeFormatter(out *output, state *State, val Value) error {
	return writeEscaped(out, state.autoEscape, val)
}

func getDefaultBuiltinFilters() map[string]BoxedFilter {
	rv := make(map[string]BoxedFilter)
	rv["safe"] = BoxedFilterFromFuncReflect(safe)
	rv["escape"] = BoxedFilterFromFuncReflect(escape)
	rv["e"] = BoxedFilterFromFuncReflect(escape)

	rv["lower"] = BoxedFilterFromFuncReflect(lower)
	rv["upper"] = BoxedFilterFromFuncReflect(upper)
	rv["title"] = BoxedFilterFromFuncReflect(title)
	rv["capitalize"] = BoxedFilterFromFuncReflect(capitalize)
	rv["replace"] = BoxedFilterFromFuncReflect(replace)
	rv["length"] = BoxedFilterFromFuncReflect(length)
	rv["count"] = BoxedFilterFromFuncReflect(length)
	rv["dictsort"] = BoxedFilterFromFuncReflect(dictsort)
	rv["items"] = BoxedFilterFromFuncReflect(items)
	rv["reverse"] = BoxedFilterFromFuncReflect(reverse)
	rv["trim"] = BoxedFilterFromFuncReflect(trim)
	rv["join"] = BoxedFilterFromFuncReflect(join)
	rv["default"] = BoxedFilterFromFuncReflect(defaultFilter)
	rv["d"] = BoxedFilterFromFuncReflect(defaultFilter)
	rv["round"] = BoxedFilterFromFuncReflect(round)
	rv["abs"] = BoxedFilterFromFuncReflect(abs)
	rv["attr"] = BoxedFilterFromFuncReflect(attr)
	rv["first"] = BoxedFilterFromFuncReflect(first)
	rv["last"] = BoxedFilterFromFuncReflect(last)
	rv["min"] = BoxedFilterFromFuncReflect(minFilter)
	rv["max"] = BoxedFilterFromFuncReflect(maxFilter)
	rv["sort"] = BoxedFilterFromFuncReflect(sortFilter)
	rv["list"] = BoxedFilterFromFuncReflect(listFilter)
	rv["bool"] = BoxedFilterFromFuncReflect(boolFilter)
	rv["batch"] = BoxedFilterFromFuncReflect(batchFilter)
	rv["slice"] = BoxedFilterFromFuncReflect(sliceFilter)
	rv["indent"] = BoxedFilterFromFuncReflect(indentFilter)
	rv["select"] = BoxedFilterFromFuncReflect(selectFilter)
	rv["reject"] = BoxedFilterFromFuncReflect(rejectFilter)
	rv["selectattr"] = BoxedFilterFromFuncReflect(selectAttrFilter)
	rv["rejectattr"] = BoxedFilterFromFuncReflect(rejectAttrFilter)
	rv["map"] = BoxedFilterFromFuncReflect(mapFilter)
	rv["unique"] = BoxedFilterFromFuncReflect(uniqueFilter)

	return rv
}

func getDefaultBuiltinTests() map[string]BoxedTest {
	rv := make(map[string]BoxedTest)
	rv["undefined"] = BoxedTestFromFuncReflect(isUndefined)
	rv["defined"] = BoxedTestFromFuncReflect(isDefined)
	rv["none"] = BoxedTestFromFuncReflect(isNone)
	rv["safe"] = BoxedTestFromFuncReflect(isSafe)
	rv["escaped"] = BoxedTestFromFuncReflect(isSafe)

	rv["odd"] = BoxedTestFromFuncReflect(isOdd)
	rv["even"] = BoxedTestFromFuncReflect(isEven)
	rv["number"] = BoxedTestFromFuncReflect(isNumber)
	rv["string"] = BoxedTestFromFuncReflect(isString)
	rv["sequence"] = BoxedTestFromFuncReflect(isSequence)
	rv["mapping"] = BoxedTestFromFuncReflect(isMapping)
	rv["startingwith"] = BoxedTestFromFuncReflect(isStartingWith)
	rv["endingwith"] = BoxedTestFromFuncReflect(isEndingWith)

	// operators
	rv["eq"] = BoxedTestFromFuncReflect(isEq)
	rv["equalto"] = BoxedTestFromFuncReflect(isEq)
	rv["=="] = BoxedTestFromFuncReflect(isEq)
	rv["ne"] = BoxedTestFromFuncReflect(isNe)
	rv["!="] = BoxedTestFromFuncReflect(isNe)
	rv["lt"] = BoxedTestFromFuncReflect(isLt)
	rv["lessthan"] = BoxedTestFromFuncReflect(isLt)
	rv["<"] = BoxedTestFromFuncReflect(isLt)
	rv["le"] = BoxedTestFromFuncReflect(isLe)
	rv["<="] = BoxedTestFromFuncReflect(isLe)
	rv["gt"] = BoxedTestFromFuncReflect(isGt)
	rv["greaterthan"] = BoxedTestFromFuncReflect(isGt)
	rv[">"] = BoxedTestFromFuncReflect(isGt)
	rv["ge"] = BoxedTestFromFuncReflect(isGe)
	rv[">="] = BoxedTestFromFuncReflect(isGe)
	rv["in"] = BoxedTestFromFuncReflect(isIn)
	rv["true"] = BoxedTestFromFuncReflect(isTrue)
	rv["false"] = BoxedTestFromFuncReflect(isFalse)
	rv["filter"] = BoxedTestFromFuncReflect(isFilter)
	rv["test"] = BoxedTestFromFuncReflect(isTest)

	return rv
}

func getDefaultGlobals() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = valueFromBoxedFunc(BoxedFuncFromFuncReflect(rangeFunc))
	rv["dict"] = valueFromBoxedFunc(BoxedFuncFromFuncReflect(dictFunc))
	return rv
}
