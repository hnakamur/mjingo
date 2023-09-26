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

func getDefaultBuiltinTests() map[string]testObject {
	if useReflect {
		return getDefaultBuiltinTestsReflect()
	}

	rv := make(map[string]testObject)
	addTest(rv, "undefined", BoxedTestFromFixedArity1ArgNoErrFunc(isUndefined))
	addTest(rv, "defined", BoxedTestFromFixedArity1ArgNoErrFunc(isDefined))
	addTest(rv, "none", BoxedTestFromFixedArity1ArgNoErrFunc(isNone))
	addTest(rv, "safe", BoxedTestFromFixedArity1ArgNoErrFunc(isSafe))
	addTest(rv, "escaped", BoxedTestFromFixedArity1ArgNoErrFunc(isSafe))

	addTest(rv, "odd", BoxedTestFromFixedArity1ArgNoErrFunc(isOdd))
	addTest(rv, "even", BoxedTestFromFixedArity1ArgNoErrFunc(isEven))
	addTest(rv, "number", BoxedTestFromFixedArity1ArgNoErrFunc(isNumber))
	addTest(rv, "string", BoxedTestFromFixedArity1ArgNoErrFunc(isString))
	addTest(rv, "sequence", BoxedTestFromFixedArity1ArgNoErrFunc(isSequence))
	addTest(rv, "mapping", BoxedTestFromFixedArity1ArgNoErrFunc(isMapping))
	addTest(rv, "startingwith", BoxedTestFromFixedArity2ArgNoErrFunc(isStartingWith))
	addTest(rv, "endingwith", BoxedTestFromFixedArity2ArgNoErrFunc(isEndingWith))

	// operators
	addTest(rv, "eq", BoxedTestFromFixedArity2ArgNoErrFunc(isEq))
	addTest(rv, "equalto", BoxedTestFromFixedArity2ArgNoErrFunc(isEq))
	addTest(rv, "==", BoxedTestFromFixedArity2ArgNoErrFunc(isEq))
	addTest(rv, "ne", BoxedTestFromFixedArity2ArgNoErrFunc(isNe))
	addTest(rv, "!=", BoxedTestFromFixedArity2ArgNoErrFunc(isNe))
	addTest(rv, "lt", BoxedTestFromFixedArity2ArgNoErrFunc(isLt))
	addTest(rv, "lessthan", BoxedTestFromFixedArity2ArgNoErrFunc(isLt))
	addTest(rv, "<", BoxedTestFromFixedArity2ArgNoErrFunc(isLt))
	addTest(rv, "le", BoxedTestFromFixedArity2ArgNoErrFunc(isLe))
	addTest(rv, "<=", BoxedTestFromFixedArity2ArgNoErrFunc(isLe))
	addTest(rv, "gt", BoxedTestFromFixedArity2ArgNoErrFunc(isGt))
	addTest(rv, "greaterthan", BoxedTestFromFixedArity2ArgNoErrFunc(isGt))
	addTest(rv, ">", BoxedTestFromFixedArity2ArgNoErrFunc(isGt))
	addTest(rv, "ge", BoxedTestFromFixedArity2ArgNoErrFunc(isGe))
	addTest(rv, ">=", BoxedTestFromFixedArity2ArgNoErrFunc(isGe))
	addTest(rv, "in", BoxedTestFromFixedArity2ArgNoErrFunc(isIn))
	addTest(rv, "true", BoxedTestFromFixedArity1ArgNoErrFunc(isTrue))
	addTest(rv, "false", BoxedTestFromFixedArity1ArgNoErrFunc(isFalse))
	addTest(rv, "filter", BoxedTestFromFixedArity2ArgNoErrFunc(isFilter))
	addTest(rv, "test", BoxedTestFromFixedArity2ArgNoErrFunc(isTest))

	return rv
}

func getDefaultBuiltinTestsReflect() map[string]testObject {
	rv := make(map[string]testObject)
	addTest(rv, "undefined", BoxedTestFromFuncReflect(isUndefined))
	addTest(rv, "defined", BoxedTestFromFuncReflect(isDefined))
	addTest(rv, "none", BoxedTestFromFuncReflect(isNone))
	addTest(rv, "safe", BoxedTestFromFuncReflect(isSafe))
	addTest(rv, "escaped", BoxedTestFromFuncReflect(isSafe))

	addTest(rv, "odd", BoxedTestFromFuncReflect(isOdd))
	addTest(rv, "even", BoxedTestFromFuncReflect(isEven))
	addTest(rv, "number", BoxedTestFromFuncReflect(isNumber))
	addTest(rv, "string", BoxedTestFromFuncReflect(isString))
	addTest(rv, "sequence", BoxedTestFromFuncReflect(isSequence))
	addTest(rv, "mapping", BoxedTestFromFuncReflect(isMapping))
	addTest(rv, "startingwith", BoxedTestFromFuncReflect(isStartingWith))
	addTest(rv, "endingwith", BoxedTestFromFuncReflect(isEndingWith))

	// operators
	addTest(rv, "eq", BoxedTestFromFuncReflect(isEq))
	addTest(rv, "equalto", BoxedTestFromFuncReflect(isEq))
	addTest(rv, "==", BoxedTestFromFuncReflect(isEq))
	addTest(rv, "ne", BoxedTestFromFuncReflect(isNe))
	addTest(rv, "!=", BoxedTestFromFuncReflect(isNe))
	addTest(rv, "lt", BoxedTestFromFuncReflect(isLt))
	addTest(rv, "lessthan", BoxedTestFromFuncReflect(isLt))
	addTest(rv, "<", BoxedTestFromFuncReflect(isLt))
	addTest(rv, "le", BoxedTestFromFuncReflect(isLe))
	addTest(rv, "<=", BoxedTestFromFuncReflect(isLe))
	addTest(rv, "gt", BoxedTestFromFuncReflect(isGt))
	addTest(rv, "greaterthan", BoxedTestFromFuncReflect(isGt))
	addTest(rv, ">", BoxedTestFromFuncReflect(isGt))
	addTest(rv, "ge", BoxedTestFromFuncReflect(isGe))
	addTest(rv, ">=", BoxedTestFromFuncReflect(isGe))
	addTest(rv, "in", BoxedTestFromFuncReflect(isIn))
	addTest(rv, "true", BoxedTestFromFuncReflect(isTrue))
	addTest(rv, "false", BoxedTestFromFuncReflect(isFalse))
	addTest(rv, "filter", BoxedTestFromFuncReflect(isFilter))
	addTest(rv, "test", BoxedTestFromFuncReflect(isTest))

	return rv
}

func getDefaultGlobals() map[string]Value {
	if useReflect {
		return getDefaultGlobalsReflect()
	}

	rv := make(map[string]Value)
	addFunction(rv, "range", BoxedFuncFromFixedArity3ArgWithErrFunc(rangeFunc))
	addFunction(rv, "dict", BoxedFuncFromFixedArity1ArgWithErrFunc(dictFunc))
	return rv
}

func getDefaultGlobalsReflect() map[string]Value {
	rv := make(map[string]Value)
	addFunction(rv, "range", BoxedFuncFromFuncReflect(rangeFunc))
	addFunction(rv, "dict", BoxedFuncFromFuncReflect(dictFunc))
	return rv
}
