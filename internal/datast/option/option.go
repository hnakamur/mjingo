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

func (o Option[T]) IsSome() bool {
	return o.valid
}

func (o Option[T]) IsNone() bool {
	return !o.valid
}

func (o Option[T]) UnwrapOr(defaultVal T) T {
	if o.IsSome() {
		return o.Unwrap()
	}
	return defaultVal
}

func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.IsSome() {
		return o.Unwrap()
	}
	return f()
}

func MapOr[T any, E any](o Option[T], defaultVal E, f func(v T) E) E {
	if o.IsSome() {
		return f(o.Unwrap())
	}
	return defaultVal
}

func AndThen[T any, U any](o Option[T], f func(v T) U) Option[U] {
	if o.IsSome() {
		return Some(f(o.Unwrap()))
	}
	return None[U]()
}

func (o *Option[T]) AsPtr() *T {
	if (*o).IsSome() {
		return &o.data
	}
	return nil
}

func (o Option[T]) Compare(other Option[T], cmpData func(a, b T) int) int {
	if o.valid == other.valid {
		if o.valid {
			return cmpData(o.data, other.data)
		}
	} else if o.valid {
		return 1
	} else if other.valid {
		return -1
	}
	return 0
}
