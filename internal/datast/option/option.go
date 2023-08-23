package option

type Option[T any] struct {
	valid bool
	data  T
}

func Some[T any](data T) Option[T] {
	return Option[T]{data: data, valid: true}
}

func None[T any]() (none Option[T]) {
	return
}

func (o Option[T]) Unwrap() T {
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

func UnwrapOr[T any](o Option[T], defaultVal T) T {
	if IsSome(o) {
		return o.Unwrap()
	}
	return defaultVal
}

func MapOr[T any, E any](o Option[T], defaultVal E, f func(v T) E) E {
	if IsSome(o) {
		return f(o.Unwrap())
	}
	return defaultVal
}

func AndThen[T any, U any](o Option[T], f func(v T) U) Option[U] {
	if IsSome(o) {
		return Some(f(o.Unwrap()))
	}
	return None[U]()
}

func AsPtr[T any](o *Option[T]) *T {
	if IsSome(*o) {
		return &o.data
	}
	return nil
}

func Compare[T any](a, b Option[T], cmpData func(a, b T) int) int {
	if a.valid == b.valid {
		if a.valid {
			return cmpData(a.data, b.data)
		}
	} else if a.valid {
		return 1
	} else if b.valid {
		return -1
	}
	return 0
}
