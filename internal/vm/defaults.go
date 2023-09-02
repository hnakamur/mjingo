package vm

import "github.com/hnakamur/mjingo/internal/value"

func escapeFormatter(out *Output, state *State, val value.Value) error {
	return writeEscaped(out, state.autoEscape, val)
}

func getDefaultBuiltinFilters() map[string]FilterFunc {
	rv := make(map[string]FilterFunc)
	rv["safe"] = filterFuncFromFilterWithStrArgValRet(safe)
	rv["escape"] = filterFuncFromWithStateValArgValErrRet(escape)
	rv["e"] = filterFuncFromWithStateValArgValErrRet(escape)

	rv["lower"] = filterFuncFromFilterWithStrArgStrRet(lower)
	rv["upper"] = filterFuncFromFilterWithStrArgStrRet(upper)
	rv["title"] = filterFuncFromFilterWithStrArgStrRet(title)
	rv["capitalize"] = filterFuncFromFilterWithStrArgStrRet(capitalize)
	rv["replace"] = filterFuncFromFilterWithStateStrStrStrArgStrRet(replace)
	rv["length"] = filterFuncFromFilterWithValArgUintErrRet(length)
	rv["count"] = filterFuncFromFilterWithValArgUintErrRet(length)
	rv["dictsort"] = filterFuncFromWithValKwargsArgValErrRet(dictsort)
	rv["items"] = filterFuncFromWithValArgValErrRet(items)
	rv["reverse"] = filterFuncFromWithValArgValErrRet(reverse)
	rv["trim"] = filterFuncFromFilterWithStrOptStrArgStrRet(trim)
	rv["join"] = filterFuncFromFilterWithValOptStrArgStrErrRet(join)
	rv["default"] = filterFuncFromFilterWithValOptValArgValRet(defaultFilter)
	rv["d"] = filterFuncFromFilterWithValOptValArgValRet(defaultFilter)
	rv["round"] = filterFuncFromFilterWithValOptI32ArgValErrRet(round)
	rv["abs"] = filterFuncFromWithValArgValErrRet(abs)
	rv["attr"] = filterFuncFromWithValValArgValErrRet(attr)
	rv["first"] = filterFuncFromWithValArgValErrRet(first)
	rv["last"] = filterFuncFromWithValArgValErrRet(last)
	rv["min"] = filterFuncFromWithStateValArgValErrRet(minFilter)
	rv["max"] = filterFuncFromWithStateValArgValErrRet(maxFilter)
	rv["sort"] = filterFuncFromWithStateValKwargsArgValErrRet(sortFilter)
	rv["list"] = filterFuncFromWithStateValArgValErrRet(listFilter)
	rv["bool"] = filterFuncFromFilterWithValArgBoolRet(boolFilter)
	rv["batch"] = filterFuncFromFilterWithStateValUintOptValArgValErrRet(batchFilter)
	rv["slice"] = filterFuncFromFilterWithStateValUintOptValArgValErrRet(sliceFilter)
	rv["indent"] = filterFuncFromFilterWithStrUintOptBoolOptBoolArgStrrRet(indentFilter)
	rv["select"] = filterFuncFromFilterWithStateValOptStrValVarArgValSliceErrRet(selectFilter)
	rv["reject"] = filterFuncFromFilterWithStateValOptStrValVarArgValSliceErrRet(rejectFilter)
	rv["selectattr"] = filterFuncFromFilterWithStateValStrOptStrValVarArgValSliceErrRet(selectAttrFilter)
	rv["rejectattr"] = filterFuncFromFilterWithStateValStrOptStrValVarArgValSliceErrRet(rejectAttrFilter)
	rv["map"] = filterFuncFromFilterWithStateValValVarArgValSliceErrRet(mapFilter)
	rv["unique"] = filterFuncFromFilterWithValSliceArgValRet(uniqueFilter)

	return rv
}

func getDefaultBuiltinTests() map[string]TestFunc {
	rv := make(map[string]TestFunc)
	rv["undefined"] = testFuncFromPredicateWithValueArg(isUndefined)
	rv["defined"] = testFuncFromPredicateWithValueArg(isDefined)
	rv["none"] = testFuncFromPredicateWithValueArg(isNone)
	rv["safe"] = testFuncFromPredicateWithValueArg(isSafe)
	rv["escaped"] = testFuncFromPredicateWithValueArg(isSafe)

	rv["odd"] = testFuncFromPredicateWithValueArg(isOdd)
	rv["even"] = testFuncFromPredicateWithValueArg(isEven)
	rv["number"] = testFuncFromPredicateWithValueArg(isNumber)
	rv["string"] = testFuncFromPredicateWithValueArg(isString)
	rv["sequence"] = testFuncFromPredicateWithValueArg(isSequence)
	rv["mapping"] = testFuncFromPredicateWithValueArg(isMapping)
	rv["startingwith"] = testFuncFromPredicateWithStringStringArgs(isStartingWith)
	rv["endingwith"] = testFuncFromPredicateWithStringStringArgs(isEndingWith)

	// operators
	rv["eq"] = testFuncFromPredicateWithValValArgs(isEq)
	rv["equalto"] = testFuncFromPredicateWithValValArgs(isEq)
	rv["=="] = testFuncFromPredicateWithValValArgs(isEq)
	rv["ne"] = testFuncFromPredicateWithValValArgs(isNe)
	rv["!="] = testFuncFromPredicateWithValValArgs(isNe)
	rv["lt"] = testFuncFromPredicateWithValValArgs(isLt)
	rv["lessthan"] = testFuncFromPredicateWithValValArgs(isLt)
	rv["<"] = testFuncFromPredicateWithValValArgs(isLt)
	rv["le"] = testFuncFromPredicateWithValValArgs(isLe)
	rv["<="] = testFuncFromPredicateWithValValArgs(isLe)
	rv["gt"] = testFuncFromPredicateWithValValArgs(isGt)
	rv["greaterthan"] = testFuncFromPredicateWithValValArgs(isGt)
	rv[">"] = testFuncFromPredicateWithValValArgs(isGt)
	rv["ge"] = testFuncFromPredicateWithValValArgs(isGe)
	rv[">="] = testFuncFromPredicateWithValValArgs(isGe)
	rv["in"] = testFuncFromPredicateWithValValArgs(isIn)
	rv["true"] = testFuncFromPredicateWithValueArg(isTrue)
	rv["false"] = testFuncFromPredicateWithValueArg(isFalse)
	rv["filter"] = testFuncFromPredicateWithStateStrArgs(isFilter)
	rv["test"] = testFuncFromPredicateWithStateStrArgs(isTest)

	return rv
}

func getDefaultGlobals() map[string]value.Value {
	rv := make(map[string]value.Value)
	rv["range"] = ValueFromFunc(funcFuncFromU32OptU32OptU32ArgU32SliceAndErrRet(fnRange))
	rv["dict"] = ValueFromFunc(funcFuncFromValArgValErrRet(dictFunc))
	return rv
}
