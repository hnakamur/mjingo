package hashset

type HashSet[T comparable] struct {
	s map[T]struct{}
}

func New[T comparable]() *HashSet[T] {
	return &HashSet[T]{
		s: make(map[T]struct{}),
	}
}

func (s *HashSet[T]) Add(v T) bool {
	_, ok := s.s[v]
	if !ok {
		s.s[v] = struct{}{}
	}
	return !ok
}

func (s *HashSet[T]) Delete(v T) {
	delete(s.s, v)
}

func (s *HashSet[T]) Contains(v T) bool {
	_, ok := s.s[v]
	return ok
}

func (s *HashSet[T]) Keys() []T {
	keys := make([]T, 0, len(s.s))
	for key := range s.s {
		keys = append(keys, key)
	}
	return keys
}
