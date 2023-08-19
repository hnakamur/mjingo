package vm

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

	return rv
}
