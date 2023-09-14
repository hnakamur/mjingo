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
	rv["safe"] = BoxedFilterFromFixedArity1ArgNoErrFunc(safe)
	rv["escape"] = BoxedFilterFromFixedArity2ArgWithErrFunc(escape)
	rv["e"] = BoxedFilterFromFixedArity2ArgWithErrFunc(escape)

	rv["lower"] = BoxedFilterFromFixedArity1ArgNoErrFunc(lower)
	rv["upper"] = BoxedFilterFromFixedArity1ArgNoErrFunc(upper)
	rv["title"] = BoxedFilterFromFixedArity1ArgNoErrFunc(title)
	rv["capitalize"] = BoxedFilterFromFixedArity1ArgNoErrFunc(capitalize)
	rv["replace"] = BoxedFilterFromFixedArity4ArgNoErrFunc(replace)
	rv["length"] = BoxedFilterFromFixedArity1ArgWithErrFunc(length)
	rv["count"] = BoxedFilterFromFixedArity1ArgWithErrFunc(length)
	rv["dictsort"] = BoxedFilterFromFixedArity2ArgWithErrFunc(dictsort)
	rv["items"] = BoxedFilterFromFixedArity1ArgWithErrFunc(items)
	rv["reverse"] = BoxedFilterFromFixedArity1ArgWithErrFunc(reverse)
	rv["trim"] = BoxedFilterFromFixedArity2ArgNoErrFunc(trim)
	rv["join"] = BoxedFilterFromFixedArity2ArgWithErrFunc(join)
	rv["default"] = BoxedFilterFromFixedArity2ArgNoErrFunc(defaultFilter)
	rv["d"] = BoxedFilterFromFixedArity2ArgNoErrFunc(defaultFilter)
	rv["round"] = BoxedFilterFromFixedArity2ArgWithErrFunc(round)
	rv["abs"] = BoxedFilterFromFixedArity1ArgWithErrFunc(abs)
	rv["attr"] = BoxedFilterFromFixedArity2ArgWithErrFunc(attr)
	rv["first"] = BoxedFilterFromFixedArity1ArgWithErrFunc(first)
	rv["last"] = BoxedFilterFromFixedArity1ArgWithErrFunc(last)
	rv["min"] = BoxedFilterFromFixedArity2ArgWithErrFunc(minFilter)
	rv["max"] = BoxedFilterFromFixedArity2ArgWithErrFunc(maxFilter)
	rv["sort"] = BoxedFilterFromFixedArity3ArgWithErrFunc(sortFilter)
	rv["list"] = BoxedFilterFromFixedArity2ArgWithErrFunc(listFilter)
	rv["bool"] = BoxedFilterFromFixedArity1ArgNoErrFunc(boolFilter)
	rv["batch"] = BoxedFilterFromFixedArity4ArgWithErrFunc(batchFilter)
	rv["slice"] = BoxedFilterFromFixedArity4ArgWithErrFunc(sliceFilter)
	rv["indent"] = BoxedFilterFromFixedArity4ArgNoErrFunc(indentFilter)
	rv["select"] = BoxedFilterFromVariadic4ArgWithErrFunc(selectFilter)
	rv["reject"] = BoxedFilterFromVariadic4ArgWithErrFunc(rejectFilter)
	rv["selectattr"] = BoxedFilterFromVariadic5ArgWithErrFunc(selectAttrFilter)
	rv["rejectattr"] = BoxedFilterFromVariadic5ArgWithErrFunc(rejectAttrFilter)
	rv["map"] = BoxedFilterFromVariadic3ArgWithErrFunc(mapFilter)
	rv["unique"] = BoxedFilterFromFixedArity1ArgNoErrFunc(uniqueFilter)

	return rv
}

func getDefaultBuiltinTests() map[string]BoxedTest {
	rv := make(map[string]BoxedTest)
	rv["undefined"] = BoxedTestFromFixedArity1ArgNoErrFunc(isUndefined)
	rv["defined"] = BoxedTestFromFixedArity1ArgNoErrFunc(isDefined)
	rv["none"] = BoxedTestFromFixedArity1ArgNoErrFunc(isNone)
	rv["safe"] = BoxedTestFromFixedArity1ArgNoErrFunc(isSafe)
	rv["escaped"] = BoxedTestFromFixedArity1ArgNoErrFunc(isSafe)

	rv["odd"] = BoxedTestFromFixedArity1ArgNoErrFunc(isOdd)
	rv["even"] = BoxedTestFromFixedArity1ArgNoErrFunc(isEven)
	rv["number"] = BoxedTestFromFixedArity1ArgNoErrFunc(isNumber)
	rv["string"] = BoxedTestFromFixedArity1ArgNoErrFunc(isString)
	rv["sequence"] = BoxedTestFromFixedArity1ArgNoErrFunc(isSequence)
	rv["mapping"] = BoxedTestFromFixedArity1ArgNoErrFunc(isMapping)
	rv["startingwith"] = BoxedTestFromFixedArity2ArgNoErrFunc(isStartingWith)
	rv["endingwith"] = BoxedTestFromFixedArity2ArgNoErrFunc(isEndingWith)

	// operators
	rv["eq"] = BoxedTestFromFixedArity2ArgNoErrFunc(isEq)
	rv["equalto"] = BoxedTestFromFixedArity2ArgNoErrFunc(isEq)
	rv["=="] = BoxedTestFromFixedArity2ArgNoErrFunc(isEq)
	rv["ne"] = BoxedTestFromFixedArity2ArgNoErrFunc(isNe)
	rv["!="] = BoxedTestFromFixedArity2ArgNoErrFunc(isNe)
	rv["lt"] = BoxedTestFromFixedArity2ArgNoErrFunc(isLt)
	rv["lessthan"] = BoxedTestFromFixedArity2ArgNoErrFunc(isLt)
	rv["<"] = BoxedTestFromFixedArity2ArgNoErrFunc(isLt)
	rv["le"] = BoxedTestFromFixedArity2ArgNoErrFunc(isLe)
	rv["<="] = BoxedTestFromFixedArity2ArgNoErrFunc(isLe)
	rv["gt"] = BoxedTestFromFixedArity2ArgNoErrFunc(isGt)
	rv["greaterthan"] = BoxedTestFromFixedArity2ArgNoErrFunc(isGt)
	rv[">"] = BoxedTestFromFixedArity2ArgNoErrFunc(isGt)
	rv["ge"] = BoxedTestFromFixedArity2ArgNoErrFunc(isGe)
	rv[">="] = BoxedTestFromFixedArity2ArgNoErrFunc(isGe)
	rv["in"] = BoxedTestFromFixedArity2ArgNoErrFunc(isIn)
	rv["true"] = BoxedTestFromFixedArity1ArgNoErrFunc(isTrue)
	rv["false"] = BoxedTestFromFixedArity1ArgNoErrFunc(isFalse)
	rv["filter"] = BoxedTestFromFixedArity2ArgNoErrFunc(isFilter)
	rv["test"] = BoxedTestFromFixedArity2ArgNoErrFunc(isTest)

	return rv
}

func getDefaultGlobals() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = valueFromBoxedFunc(BoxedFuncFromFixedArity3ArgWithErrFunc(rangeFunc))
	rv["dict"] = valueFromBoxedFunc(BoxedFuncFromFixedArity1ArgWithErrFunc(dictFunc))
	return rv
}
