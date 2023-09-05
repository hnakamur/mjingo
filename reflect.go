package mjingo

import "reflect"

func assertType(ty reflect.Type, nilPtrOfType any, errMsg string) {
	if got, want := ty, typeFromPtr(nilPtrOfType); got != want {
		panic(errMsg)
	}
}

func typeFromPtr(nilPtrOfType any) reflect.Type {
	return reflect.TypeOf(nilPtrOfType).Elem()
}
