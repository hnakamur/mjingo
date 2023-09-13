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

func getDefaultBuiltinFilters() map[string]BoxedFilter {
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

	return rv
}

func getDefaultBuiltinTests() map[string]BoxedTest {
	rv := make(map[string]BoxedTest)
	rv["undefined"] = BoxTestFromFixedArity1ArgNoErrFunc(isUndefined)
	rv["defined"] = BoxTestFromFixedArity1ArgNoErrFunc(isDefined)
	rv["none"] = BoxTestFromFixedArity1ArgNoErrFunc(isNone)
	rv["safe"] = BoxTestFromFixedArity1ArgNoErrFunc(isSafe)
	rv["escaped"] = BoxTestFromFixedArity1ArgNoErrFunc(isSafe)

	rv["odd"] = BoxTestFromFixedArity1ArgNoErrFunc(isOdd)
	rv["even"] = BoxTestFromFixedArity1ArgNoErrFunc(isEven)
	rv["number"] = BoxTestFromFixedArity1ArgNoErrFunc(isNumber)
	rv["string"] = BoxTestFromFixedArity1ArgNoErrFunc(isString)
	rv["sequence"] = BoxTestFromFixedArity1ArgNoErrFunc(isSequence)
	rv["mapping"] = BoxTestFromFixedArity1ArgNoErrFunc(isMapping)
	rv["startingwith"] = BoxTestFromFixedArity2ArgNoErrFunc(isStartingWith)
	rv["endingwith"] = BoxTestFromFixedArity2ArgNoErrFunc(isEndingWith)

	// operators
	rv["eq"] = BoxTestFromFixedArity2ArgNoErrFunc(isEq)
	rv["equalto"] = BoxTestFromFixedArity2ArgNoErrFunc(isEq)
	rv["=="] = BoxTestFromFixedArity2ArgNoErrFunc(isEq)
	rv["ne"] = BoxTestFromFixedArity2ArgNoErrFunc(isNe)
	rv["!="] = BoxTestFromFixedArity2ArgNoErrFunc(isNe)
	rv["lt"] = BoxTestFromFixedArity2ArgNoErrFunc(isLt)
	rv["lessthan"] = BoxTestFromFixedArity2ArgNoErrFunc(isLt)
	rv["<"] = BoxTestFromFixedArity2ArgNoErrFunc(isLt)
	rv["le"] = BoxTestFromFixedArity2ArgNoErrFunc(isLe)
	rv["<="] = BoxTestFromFixedArity2ArgNoErrFunc(isLe)
	rv["gt"] = BoxTestFromFixedArity2ArgNoErrFunc(isGt)
	rv["greaterthan"] = BoxTestFromFixedArity2ArgNoErrFunc(isGt)
	rv[">"] = BoxTestFromFixedArity2ArgNoErrFunc(isGt)
	rv["ge"] = BoxTestFromFixedArity2ArgNoErrFunc(isGe)
	rv[">="] = BoxTestFromFixedArity2ArgNoErrFunc(isGe)
	rv["in"] = BoxTestFromFixedArity2ArgNoErrFunc(isIn)
	rv["true"] = BoxTestFromFixedArity1ArgNoErrFunc(isTrue)
	rv["false"] = BoxTestFromFixedArity1ArgNoErrFunc(isFalse)
	rv["filter"] = BoxTestFromFixedArity2ArgNoErrFunc(isFilter)
	rv["test"] = BoxTestFromFixedArity2ArgNoErrFunc(isTest)

	return rv
}

func getDefaultGlobals() map[string]Value {
	rv := make(map[string]Value)
	rv["range"] = valueFromBoxedFunc(BoxedFuncFromFuncReflect(rangeFunc))
	rv["dict"] = valueFromBoxedFunc(BoxedFuncFromFuncReflect(dictFunc))
	return rv
}
