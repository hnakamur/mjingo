package stack

type Stack[T any] []T

func NewStackWithCapacity[T any](capacity uint) Stack[T] {
	return make([]T, 0, capacity)
}

func (s *Stack[T]) Push(elem T) {
	*s = append(*s, elem)
}

func (s *Stack[T]) Empty() bool {
	return len(*s) == 0
}

func (s *Stack[T]) Pop() T {
	st := (*s)[len(*s)-1]
	clear((*s)[len(*s)-1:])
	*s = (*s)[:len(*s)-1]
	return st
}

func (s *Stack[T]) TryPop() (v T, ok bool) {
	if s.Empty() {
		return
	}
	st := (*s)[len(*s)-1]
	clear((*s)[len(*s)-1:])
	*s = (*s)[:len(*s)-1]
	return st, true
}

func (s *Stack[T]) Peek() (v T, ok bool) {
	if s.Empty() {
		return
	}
	return (*s)[len(*s)-1], true
}

func (s *Stack[T]) SliceTop(n uint) []T {
	return (*s)[uint(len(*s))-n:]
}

func (s *Stack[T]) DropTop(n uint) []T {
	l := uint(len(*s)) - n
	clear((*s)[l:])
	*s = (*s)[:l]
	return (*s)[l:]
}
