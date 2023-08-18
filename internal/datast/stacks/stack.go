// Package stacks defines various functions useful with stacks of any type.
package stacks

// Push appends a value to the stack. Push modifies the stack if the capacity
// of the stack is not enough to fit the additional value.
func Push[S ~[]E, E any](stack *S, v E) {
	*stack = append(*stack, v)
}

// Pop takes the last eleent from the stack and shrinks the size of the
// stack by one. Pop returns the last element and true if the stack is not empty,
// or the zero value and false if the stack is empty.
func Pop[S ~[]E, E any](stack *S) (v E, ok bool) {
	var zero E
	if len(*stack) == 0 {
		return zero, false
	}
	v = (*stack)[len(*stack)-1]
	(*stack)[len(*stack)-1] = zero
	*stack = (*stack)[:len(*stack)-1]
	return v, true
}
