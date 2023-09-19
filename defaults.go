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

var useReflect bool

func getDefaultBuiltinFilters() map[string]BoxedFilter {
	if useReflect {
		return getDefaultBuiltinFiltersReflect()
	}

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
	rv["pprint"] = BoxedFilterFromFixedArity1ArgNoErrFunc(pprint)
	rv["urlencode"] = BoxedFilterFromFixedArity1ArgWithErrFunc(urlencodeFilter)
	rv["tojson"] = BoxedFilterFromFixedArity2ArgWithErrFunc(tojson)
	return rv
}

func getDefaultBuiltinFiltersReflect() map[string]BoxedFilter {
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
	rv["pprint"] = BoxedFilterFromFuncReflect(pprint)
	rv["urlencode"] = BoxedFilterFromFuncReflect(urlencodeFilter)
	rv["tojson"] = BoxedFilterFromFuncReflect(tojson)
	return rv
}

func getDefaultBuiltinTests() map[string]BoxedTest {
	if useReflect {
		return getDefaultBuiltinTestsReflect()
	}

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

func getDefaultBuiltinTestsReflect() map[string]BoxedTest {
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
	if useReflect {
		return getDefaultGlobalsReflect()
	}

	rv := make(map[string]Value)
	rv["range"] = valueFromBoxedFunc(BoxedFuncFromFixedArity3ArgWithErrFunc(rangeFunc))
	rv["dict"] = valueFromBoxedFunc(BoxedFuncFromFixedArity1ArgWithErrFunc(dictFunc))
	return rv
}

func getDefaultGlobalsReflect() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = valueFromBoxedFunc(BoxedFuncFromFuncReflect(rangeFunc))
	rv["dict"] = valueFromBoxedFunc(BoxedFuncFromFuncReflect(dictFunc))
	return rv
}
