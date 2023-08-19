// Package stacks defines various functions useful with stacks of any type.
package stacks

// Push appends a value to the stack represented with a slice. Push modifies
// the stack if the capacity of the stack is not enough to fit the additional
// value.
func Push[S ~[]E, E any](stack *S, v E) {
	*stack = append(*stack, v)
}

// Pop takes the last eleent from the stack and shrinks the size of the
// stack by one. Pop returns the last element if the stack is not empty,
// or panics if the stack is empty.
// Pop zeroes the last element before shrinking the stack, so it is safely
// garbage collected even if the element contains a pointer.
func Pop[S ~[]E, E any](stack *S) E {
	i := len(*stack) - 1
	v := (*stack)[i]
	clear((*stack)[i:])
	*stack = (*stack)[:i]
	return v
}

// TryPop takes the last eleent from the stack and shrinks the size of the
// stack by one. TryPop returns the last element and true if the stack is not empty,
// or the zero value and false if the stack is empty.
// TryPop zeroes the last element before shrinking the stack, so it is safely
// garbage collected even if the element contains a pointer.
func TryPop[S ~[]E, E any](stack *S) (v E, ok bool) {
	if len(*stack) == 0 {
		return
	}
	i := len(*stack) - 1
	v = (*stack)[i]
	clear((*stack)[i:])
	*stack = (*stack)[:i]
	return v, true
}

// Peek returns the last element and true if the stack is not empty, or the zero
// value and false if the stack is empty.
func Peek[S ~[]E, E any](stack S) (v E, ok bool) {
	if len(stack) == 0 {
		return
	}
	i := len(stack) - 1
	return stack[i], true
}

// SliceTop returns the slice of top n elements in the stack.
// SliceTop panics if n > len(stack).
func SliceTop[S ~[]E, E any](stack S, n uint) []E {
	return stack[uint(len(stack))-n:]
}

// DropTop drops the slice of elements from the top of the stack
// and shrinks the stack after zeroing the elements in the stack.
func DropTop[S ~[]E, E any](stack *S, n uint) {
	l := uint(len(*stack)) - n
	clear((*stack)[l:])
	*stack = (*stack)[:l]
}
