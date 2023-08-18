// Package slicex defines various functions useful with slices of any type.
// This is an extension to the slices package in the Go's standard library.
package slicex

// Push appends a value to the stack represented with a slice. Push modifies
// the stack if the capacity of the stack is not enough to fit the additional
// value.
func Push[S ~[]E, E any](stack *S, v E) {
	*stack = append(*stack, v)
}

// Pop takes the last eleent from the stack and shrinks the size of the
// stack by one. Pop returns the last element and true if the stack is not empty,
// or the zero value and false if the stack is empty.
// Pop zeroes the last element before shrinking the stack, so it is safely
// garbage collected even if the element contains a pointer.
func Pop[S ~[]E, E any](stack *S) (v E, ok bool) {
	if len(*stack) == 0 {
		var zero E
		return zero, false
	}
	i := len(*stack) - 1
	v = (*stack)[i]
	clear((*stack)[i:])
	*stack = (*stack)[:i]
	return v, true
}

// Delete removes the elements s[i:j] from s, modifies s in place.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete is O(len(s)-j), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete might not modify the elements s[len(s)-(j-i):len(s)]. If those
// elements contain pointers Delete is zeroing those elements so that
// objects they reference will be garbage collected.
func Delete[S ~[]E, E any](s *S, i, j uint) {
	clear((*s)[i:j])
	*s = append((*s)[:i], (*s)[j:]...)
}
