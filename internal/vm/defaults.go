package vm

func getDefaultBuiltinTests() map[string]TestFunc {
	rv := make(map[string]TestFunc)
	rv["undefined"] = testFuncFromPredicate(isUndefined)
	rv["defined"] = testFuncFromPredicate(isDefined)
	rv["none"] = testFuncFromPredicate(isNone)
	return rv
}
