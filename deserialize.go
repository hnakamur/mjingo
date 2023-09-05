package mjingo

import (
	"fmt"
	"reflect"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

func assertFuncArgTypes(fnType reflect.Type) {
	numIn := fnType.NumIn()
	for i := 0; i < numIn; i++ {
		if !isSupportedFuncArgType(fnType.In(i), i) {
			panic(fmt.Sprintf("unsupported type for argument %d of function", i))
		}
	}
}

func isSupportedFuncArgType(argType reflect.Type, argPos int) bool {
	switch argType {
	case typeFromPtr((**vmState)(nil)):
		return argPos == 0
	case typeFromPtr((*Value)(nil)), typeFromPtr((*string)(nil)):
		return true
	}
	return false
}

func goValueFromValue(val Value, destType reflect.Type) (any, error) {
	switch destType {
	case typeFromPtr((*Value)(nil)):
		return val, nil
	case typeFromPtr((*string)(nil)):
		return stringFromValue(option.Some(val))
	}
	panic("unsupported destination type")
}
