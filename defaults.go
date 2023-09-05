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
	rv["safe"] = boxedFilterFromFilterWithStrArgValRet(safe)
	rv["escape"] = boxedFilterFromWithStateValArgValErrRet(escape)
	rv["e"] = boxedFilterFromWithStateValArgValErrRet(escape)

	rv["lower"] = boxedFilterFromFilterWithStrArgStrRet(lower)
	rv["upper"] = boxedFilterFromFilterWithStrArgStrRet(upper)
	rv["title"] = boxedFilterFromFilterWithStrArgStrRet(title)
	rv["capitalize"] = boxedFilterFromFilterWithStrArgStrRet(capitalize)
	rv["replace"] = boxedFilterFromFilterWithStateStrStrStrArgStrRet(replace)
	rv["length"] = boxedFilterFromFilterWithValArgUintErrRet(length)
	rv["count"] = boxedFilterFromFilterWithValArgUintErrRet(length)
	rv["dictsort"] = boxedFilterFromWithValKwargsArgValErrRet(dictsort)
	rv["items"] = boxedFilterFromWithValArgValErrRet(items)
	rv["reverse"] = boxedFilterFromWithValArgValErrRet(reverse)
	rv["trim"] = boxedFilterFromFilterWithStrOptStrArgStrRet(trim)
	rv["join"] = boxedFilterFromFilterWithValOptStrArgStrErrRet(join)
	rv["default"] = boxedFilterFromFilterWithValOptValArgValRet(defaultFilter)
	rv["d"] = boxedFilterFromFilterWithValOptValArgValRet(defaultFilter)
	rv["round"] = boxedFilterFromFilterWithValOptI32ArgValErrRet(round)
	rv["abs"] = boxedFilterFromWithValArgValErrRet(abs)
	rv["attr"] = boxedFilterFromWithValValArgValErrRet(attr)
	rv["first"] = boxedFilterFromWithValArgValErrRet(first)
	rv["last"] = boxedFilterFromWithValArgValErrRet(last)
	rv["min"] = boxedFilterFromWithStateValArgValErrRet(minFilter)
	rv["max"] = boxedFilterFromWithStateValArgValErrRet(maxFilter)
	rv["sort"] = boxedFilterFromWithStateValKwargsArgValErrRet(sortFilter)
	rv["list"] = boxedFilterFromWithStateValArgValErrRet(listFilter)
	rv["bool"] = boxedFilterFromFilterWithValArgBoolRet(boolFilter)
	rv["batch"] = boxedFilterFromFilterWithStateValUintOptValArgValErrRet(batchFilter)
	rv["slice"] = boxedFilterFromFilterWithStateValUintOptValArgValErrRet(sliceFilter)
	rv["indent"] = boxedFilterFromFilterWithStrUintOptBoolOptBoolArgStrrRet(indentFilter)
	rv["select"] = boxedFilterFromFilterWithStateValOptStrValVarArgValSliceErrRet(selectFilter)
	rv["reject"] = boxedFilterFromFilterWithStateValOptStrValVarArgValSliceErrRet(rejectFilter)
	rv["selectattr"] = boxedFilterFromFilterWithStateValStrOptStrValVarArgValSliceErrRet(selectAttrFilter)
	rv["rejectattr"] = boxedFilterFromFilterWithStateValStrOptStrValVarArgValSliceErrRet(rejectAttrFilter)
	rv["map"] = boxedFilterFromFilterWithStateValValVarArgValSliceErrRet(mapFilter)
	rv["unique"] = boxedFilterFromFilterWithValSliceArgValRet(uniqueFilter)

	return rv
}

func getDefaultBuiltinTests() map[string]BoxedTest {
	rv := make(map[string]BoxedTest)
	// rv["undefined"] = boxedTestFromPredicateWithValueArg(isUndefined)
	rv["undefined"] = boxedTestFromFunc(isUndefined)
	rv["defined"] = boxedTestFromPredicateWithValueArg(isDefined)
	rv["none"] = boxedTestFromPredicateWithValueArg(isNone)
	rv["safe"] = boxedTestFromPredicateWithValueArg(isSafe)
	rv["escaped"] = boxedTestFromPredicateWithValueArg(isSafe)

	rv["odd"] = boxedTestFromPredicateWithValueArg(isOdd)
	rv["even"] = boxedTestFromPredicateWithValueArg(isEven)
	rv["number"] = boxedTestFromPredicateWithValueArg(isNumber)
	rv["string"] = boxedTestFromPredicateWithValueArg(isString)
	rv["sequence"] = boxedTestFromPredicateWithValueArg(isSequence)
	rv["mapping"] = boxedTestFromPredicateWithValueArg(isMapping)
	rv["startingwith"] = boxedTestFromPredicateWithStringStringArgs(isStartingWith)
	rv["endingwith"] = boxedTestFromPredicateWithStringStringArgs(isEndingWith)

	// operators
	rv["eq"] = boxedTestFromPredicateWithValValArgs(isEq)
	rv["equalto"] = boxedTestFromPredicateWithValValArgs(isEq)
	rv["=="] = boxedTestFromPredicateWithValValArgs(isEq)
	rv["ne"] = boxedTestFromPredicateWithValValArgs(isNe)
	rv["!="] = boxedTestFromPredicateWithValValArgs(isNe)
	rv["lt"] = boxedTestFromPredicateWithValValArgs(isLt)
	rv["lessthan"] = boxedTestFromPredicateWithValValArgs(isLt)
	rv["<"] = boxedTestFromPredicateWithValValArgs(isLt)
	rv["le"] = boxedTestFromPredicateWithValValArgs(isLe)
	rv["<="] = boxedTestFromPredicateWithValValArgs(isLe)
	rv["gt"] = boxedTestFromPredicateWithValValArgs(isGt)
	rv["greaterthan"] = boxedTestFromPredicateWithValValArgs(isGt)
	rv[">"] = boxedTestFromPredicateWithValValArgs(isGt)
	rv["ge"] = boxedTestFromPredicateWithValValArgs(isGe)
	rv[">="] = boxedTestFromPredicateWithValValArgs(isGe)
	rv["in"] = boxedTestFromPredicateWithValValArgs(isIn)
	rv["true"] = boxedTestFromPredicateWithValueArg(isTrue)
	rv["false"] = boxedTestFromPredicateWithValueArg(isFalse)
	rv["filter"] = boxedTestFromPredicateWithStateStrArgs(isFilter)
	rv["test"] = boxedTestFromPredicateWithStateStrArgs(isTest)

	return rv
}

func getDefaultGlobals() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = valueFromFunc(boxedFuncFromU32OptU32OptU32ArgU32SliceAndErrRet(rangeFunc))
	rv["dict"] = valueFromFunc(boxedFuncFromValArgValErrRet(dictFunc))
	return rv
}
