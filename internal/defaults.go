package internal

func escapeFormatter(out *Output, state *State, val Value) error {
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

func getDefaultGlobals() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = ValueFromFunc(funcFuncFromU32OptU32OptU32ArgU32SliceAndErrRet(fnRange))
	return rv
}
