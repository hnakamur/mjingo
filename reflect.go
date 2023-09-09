package mjingo

import "reflect"

func assertType[T any](ty reflect.Type, errMsg string) {
	if got, want := ty, reflectType[T](); got != want {
		panic(errMsg)
	}
}

func reflectType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
