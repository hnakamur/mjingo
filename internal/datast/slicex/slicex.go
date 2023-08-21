// Package slicex defines various functions useful with slices of any type.
// This is an extension to the slices package in the Go's standard library.
package slicex

// Delete removes the elements s[i:j] from s, modifies s in place.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete is O(len(s)-j), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete might not modify the elements s[len(s)-(j-i):len(s)]. If those
// elements contain pointers Delete is zeroing those elements so that
// objects they reference will be garbage collected.
func Delete[S ~[]E, E any](s *S, i, j int) {
	clear((*s)[i:j])
	*s = append((*s)[:i], (*s)[j:]...)
}

// All returns true if predicate returns true for all elements in the slice.
// All returns true if the slice is empty.
func All[S ~[]E, E any](s S, predicate func(e E) bool) bool {
	for _, e := range s {
		if !predicate(e) {
			return false
		}
	}
	return true
}

// Any returns true if predicate returns true for any element in the slice.
// Any returns false if the slice is empty.
func Any[S ~[]E, E any](s S, predicate func(e E) bool) bool {
	for _, e := range s {
		if predicate(e) {
			return true
		}
	}
	return false
}

// Map returns a new slice whose elements are converted with f from elements in s.
func Map[S ~[]E, E any, T any](s S, f func(e E) T) []T {
	if s == nil {
		return nil
	}
	ret := make([]T, len(s))
	for i, e := range s {
		ret[i] = f(e)
	}
	return ret
}
