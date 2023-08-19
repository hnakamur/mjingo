package vm

func getDefaultBuiltinFilters() map[string]FilterFunc {
	rv := make(map[string]FilterFunc)
	rv["safe"] = filterFuncFromFilterWithStringArgValueRet(safe)
	rv["escape"] = filterFuncFromWithStateValueArgValueErrRet(escape)
	rv["e"] = filterFuncFromWithStateValueArgValueErrRet(escape)

	rv["lower"] = filterFuncFromFilterWithStringArgStringRet(lower)
	rv["upper"] = filterFuncFromFilterWithStringArgStringRet(upper)
	rv["title"] = filterFuncFromFilterWithStringArgStringRet(title)
	rv["capitalize"] = filterFuncFromFilterWithStringArgStringRet(capitalize)

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
