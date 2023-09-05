package mjingo

import "reflect"

func assertType(ty reflect.Type, nilPtrOfType any, errMsg string) {
	if got, want := ty, reflect.TypeOf(nilPtrOfType).Elem(); got != want {
		panic(errMsg)
	}
}
