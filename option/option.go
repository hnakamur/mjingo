// Package option provides the Option type which represents an optional value.
package option

import "hash"

// Option represents an optional value.
//
// Type Option represents an optional value: every Option
// is either Some and contains a value, or None, and
// does not. Option types are very common in Rust code, as
// they have a number of uses.
//
// Option in this package is a port of subset of Rust Option type.
//
// One import usage of Options is an optional argument type
// of filters, tests, or functions.
type Option[T any] struct {
	valid bool
	data  T
}

// Some creates a Some value of type T.
func Some[T any](data T) Option[T] {
	return Option[T]{valid: true, data: data}
}

// None creates a None value of type T.
func None[T any]() Option[T] {
	var zero Option[T]
	return zero
}

// Unwrap returns the contained Some value if the options is a Some value.
// It panics if the the option is a None value.
func (o Option[T]) Unwrap() T {
	if !o.valid {
		panic("option must be valid to unwrap")
	}
	return o.data
}

// IsSome returns true if the option is a Some value.
func (o Option[T]) IsSome() bool {
	return o.valid
}

// IsNone returns true if the option is a None value.
func (o Option[T]) IsNone() bool {
	return !o.valid
}

// Take takes the value out of the option, leaving a None in its place.
func (o *Option[T]) Take() Option[T] {
	rv := *o
	*o = None[T]()
	return rv
}

// UnwrapTo sets the contained Some value to dest and returns true
// if the options is a Some value.
// It does nothing and returns false if the option is a None value.
//
// This can be used with alternative of Rust's if let Some(dest) = o { ... }
// or while let Some(dest) = o { ... }
//
// An example with if statement:
//
//	if v := (Value{}); o.UnwrapTo(&v) {
//	  // do something with v
//	}
//
// An example with for statement and an Iterator.
//
//	for v := (Value{}); iter.Next().UnwrapTo(&v); {
//	  // do something with v
//	}
func (o Option[T]) UnwrapTo(dest *T) bool {
	if o.IsSome() {
		*dest = o.Unwrap()
		return true
	}
	return false
}

// UnwrapOr returns the contained Some value or a provided default.
//
// Arguments passed to UnwrapOr are eagerly evaluated; if you are passing
// the result of a function call, it is recommended to use [Option.UnwrapOrElse],
// which is lazily evaluated.
func (o Option[T]) UnwrapOr(defaultVal T) T {
	if o.IsSome() {
		return o.Unwrap()
	}
	return defaultVal
}

// UnwrapOrElse returns the contained Some value or computes it from a closure.
func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.IsSome() {
		return o.Unwrap()
	}
	return f()
}

// Map maps an Option[T] to Option[U] by applying a function to a contained value
// (if Some) or returns None (if None).
func Map[T any, U any](o Option[T], f func(v T) U) Option[U] {
	if o.IsSome() {
		return Some(f(o.Unwrap()))
	}
	return None[U]()
}

// MapOr returns the provided default result (if none),
// or applies a function to the contained value (if any).
//
// Arguments passed to MapOr are eagerly evaluated; if you are passing
// the result of a function call, it is recommended to use [Option.MapOrElse],
// which is lazily evaluated.
func MapOr[T any, E any](o Option[T], defaultVal E, f func(v T) E) E {
	if o.IsSome() {
		return f(o.Unwrap())
	}
	return defaultVal
}

// MapOrElse returns the default function result (if none),
// or applies a function to the contained value (if any).
func MapOrElse[T any, E any](o Option[T], defaultFn func() E, f func(v T) E) E {
	if o.IsSome() {
		return f(o.Unwrap())
	}
	return defaultFn()
}

// AndThen returns None if the option is None, otherwise calls `f` with the
// wrapped value and returns the result.
//
// Some languages call this operation flatmap.
func AndThen[T any, U any](o Option[T], f func(v T) U) Option[U] {
	if o.IsSome() {
		return Some(f(o.Unwrap()))
	}
	return None[U]()
}

// AsPtr returns the pointer to the contained data if the options is a Some value.
// It returns nil if the option is a None value.
func (o *Option[T]) AsPtr() *T {
	if (*o).IsSome() {
		return &o.data
	}
	return nil
}

// Equal returns true if o and other are both Some values and and those values are
// equal or both None. It returns false otherwise.
func (o Option[T]) Equal(other Option[T], eqData func(a, b T) bool) bool {
	if o.valid == other.valid {
		if o.valid {
			return eqData(o.data, other.data)
		}
		return true
	}
	return false
}

// Compare compares o and other and returns:
//
//	-1 if o <  other
//	 0 if o == other
//	+1 if o >  other
//
// where None < Some and for case both options are Some, the contained data are compared with cmpData.
func (o Option[T]) Compare(other Option[T], cmpData func(a, b T) int) int {
	if o.valid == other.valid {
		if o.valid {
			return cmpData(o.data, other.data)
		}
		return 0
	}
	if o.valid {
		return 1
	}
	return -1
}

// Hash returns the hashed value of the option.
//
// If the option is a Some value, the contained data are hashed with hashData.
func (o Option[T]) Hash(h hash.Hash, hashData func(data T, h hash.Hash)) {
	if o.IsSome() {
		hashData(o.data, h)
	}
	h.Write([]byte{0})
}
