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

func getDefaultBuiltinFilters() map[string]filterObject {
	if useReflect {
		return getDefaultBuiltinFiltersReflect()
	}

	rv := make(map[string]filterObject)
	addFilter(rv, "safe", BoxedFilterFromFixedArity1ArgNoErrFunc(safe))
	addFilter(rv, "escape", BoxedFilterFromFixedArity2ArgWithErrFunc(escape), "e")

	addFilter(rv, "lower", BoxedFilterFromFixedArity1ArgNoErrFunc(lower))
	addFilter(rv, "upper", BoxedFilterFromFixedArity1ArgNoErrFunc(upper))
	addFilter(rv, "title", BoxedFilterFromFixedArity1ArgNoErrFunc(title))
	addFilter(rv, "capitalize", BoxedFilterFromFixedArity1ArgNoErrFunc(capitalize))
	addFilter(rv, "replace", BoxedFilterFromFixedArity4ArgNoErrFunc(replace))
	addFilter(rv, "length", BoxedFilterFromFixedArity1ArgWithErrFunc(length), "count")
	addFilter(rv, "dictsort", BoxedFilterFromFixedArity2ArgWithErrFunc(dictsort))
	addFilter(rv, "items", BoxedFilterFromFixedArity1ArgWithErrFunc(items))
	addFilter(rv, "reverse", BoxedFilterFromFixedArity1ArgWithErrFunc(reverse))
	addFilter(rv, "trim", BoxedFilterFromFixedArity2ArgNoErrFunc(trim))
	addFilter(rv, "join", BoxedFilterFromFixedArity2ArgWithErrFunc(join))
	addFilter(rv, "default", BoxedFilterFromFixedArity2ArgNoErrFunc(defaultFilter), "d")
	addFilter(rv, "round", BoxedFilterFromFixedArity2ArgWithErrFunc(round))
	addFilter(rv, "abs", BoxedFilterFromFixedArity1ArgWithErrFunc(abs))
	addFilter(rv, "attr", BoxedFilterFromFixedArity2ArgWithErrFunc(attr))
	addFilter(rv, "first", BoxedFilterFromFixedArity1ArgWithErrFunc(first))
	addFilter(rv, "last", BoxedFilterFromFixedArity1ArgWithErrFunc(last))
	addFilter(rv, "min", BoxedFilterFromFixedArity2ArgWithErrFunc(minFilter))
	addFilter(rv, "max", BoxedFilterFromFixedArity2ArgWithErrFunc(maxFilter))
	addFilter(rv, "sort", BoxedFilterFromFixedArity3ArgWithErrFunc(sortFilter))
	addFilter(rv, "list", BoxedFilterFromFixedArity2ArgWithErrFunc(listFilter))
	addFilter(rv, "bool", BoxedFilterFromFixedArity1ArgNoErrFunc(boolFilter))
	addFilter(rv, "batch", BoxedFilterFromFixedArity4ArgWithErrFunc(batchFilter))
	addFilter(rv, "slice", BoxedFilterFromFixedArity4ArgWithErrFunc(sliceFilter))
	addFilter(rv, "indent", BoxedFilterFromFixedArity4ArgNoErrFunc(indentFilter))
	addFilter(rv, "select", BoxedFilterFromVariadic4ArgWithErrFunc(selectFilter))
	addFilter(rv, "reject", BoxedFilterFromVariadic4ArgWithErrFunc(rejectFilter))
	addFilter(rv, "selectattr", BoxedFilterFromVariadic5ArgWithErrFunc(selectAttrFilter))
	addFilter(rv, "rejectattr", BoxedFilterFromVariadic5ArgWithErrFunc(rejectAttrFilter))
	addFilter(rv, "map", BoxedFilterFromVariadic3ArgWithErrFunc(mapFilter))
	addFilter(rv, "unique", BoxedFilterFromFixedArity1ArgNoErrFunc(uniqueFilter))
	addFilter(rv, "pprint", BoxedFilterFromFixedArity1ArgNoErrFunc(pprint))
	addFilter(rv, "urlencode", BoxedFilterFromFixedArity1ArgWithErrFunc(urlencodeFilter))
	addFilter(rv, "tojson", BoxedFilterFromFixedArity2ArgWithErrFunc(tojson))
	return rv
}

func getDefaultBuiltinFiltersReflect() map[string]filterObject {
	rv := make(map[string]filterObject)
	addFilter(rv, "safe", BoxedFilterFromFuncReflect(safe))
	addFilter(rv, "escape", BoxedFilterFromFuncReflect(escape), "e")

	addFilter(rv, "lower", BoxedFilterFromFuncReflect(lower))
	addFilter(rv, "upper", BoxedFilterFromFuncReflect(upper))
	addFilter(rv, "title", BoxedFilterFromFuncReflect(title))
	addFilter(rv, "capitalize", BoxedFilterFromFuncReflect(capitalize))
	addFilter(rv, "replace", BoxedFilterFromFuncReflect(replace))
	addFilter(rv, "length", BoxedFilterFromFuncReflect(length), "count")
	addFilter(rv, "dictsort", BoxedFilterFromFuncReflect(dictsort))
	addFilter(rv, "items", BoxedFilterFromFuncReflect(items))
	addFilter(rv, "reverse", BoxedFilterFromFuncReflect(reverse))
	addFilter(rv, "trim", BoxedFilterFromFuncReflect(trim))
	addFilter(rv, "join", BoxedFilterFromFuncReflect(join))
	addFilter(rv, "default", BoxedFilterFromFuncReflect(defaultFilter), "d")
	addFilter(rv, "round", BoxedFilterFromFuncReflect(round))
	addFilter(rv, "abs", BoxedFilterFromFuncReflect(abs))
	addFilter(rv, "attr", BoxedFilterFromFuncReflect(attr))
	addFilter(rv, "first", BoxedFilterFromFuncReflect(first))
	addFilter(rv, "last", BoxedFilterFromFuncReflect(last))
	addFilter(rv, "min", BoxedFilterFromFuncReflect(minFilter))
	addFilter(rv, "max", BoxedFilterFromFuncReflect(maxFilter))
	addFilter(rv, "sort", BoxedFilterFromFuncReflect(sortFilter))
	addFilter(rv, "list", BoxedFilterFromFuncReflect(listFilter))
	addFilter(rv, "bool", BoxedFilterFromFuncReflect(boolFilter))
	addFilter(rv, "batch", BoxedFilterFromFuncReflect(batchFilter))
	addFilter(rv, "slice", BoxedFilterFromFuncReflect(sliceFilter))
	addFilter(rv, "indent", BoxedFilterFromFuncReflect(indentFilter))
	addFilter(rv, "select", BoxedFilterFromFuncReflect(selectFilter))
	addFilter(rv, "reject", BoxedFilterFromFuncReflect(rejectFilter))
	addFilter(rv, "selectattr", BoxedFilterFromFuncReflect(selectAttrFilter))
	addFilter(rv, "rejectattr", BoxedFilterFromFuncReflect(rejectAttrFilter))
	addFilter(rv, "map", BoxedFilterFromFuncReflect(mapFilter))
	addFilter(rv, "unique", BoxedFilterFromFuncReflect(uniqueFilter))
	addFilter(rv, "pprint", BoxedFilterFromFuncReflect(pprint))
	addFilter(rv, "urlencode", BoxedFilterFromFuncReflect(urlencodeFilter))
	addFilter(rv, "tojson", BoxedFilterFromFuncReflect(tojson))
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
