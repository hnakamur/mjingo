package vm

import "github.com/hnakamur/mjingo/value"

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
