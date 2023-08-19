package option

type Option[T any] struct {
	data  T
	valid bool
}

func Some[T any](data T) Option[T] {
	return Option[T]{data: data, valid: true}
}

func None[T any]() (none Option[T]) {
	return
}

func Unwrap[T any](o Option[T]) T {
	if !o.valid {
		panic("option must be valid to unwrap")
	}
	return o.data
}

func IsSome[T any](o Option[T]) bool {
	return o.valid
}

func IsNone[T any](o Option[T]) bool {
	return !o.valid
}

func MapOr[T any, E any](o Option[T], defaultVal E, f func(v T) E) E {
	if IsSome(o) {
		return f(Unwrap(o))
	}
	return defaultVal
}

func AsPtr[T any](o *Option[T]) *T {
	if IsSome(*o) {
		return &o.data
	}
	return nil
}
