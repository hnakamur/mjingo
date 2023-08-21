package hashset

type HashSet[T comparable] struct {
	s map[T]struct{}
}

func New[T comparable]() *HashSet[T] {
	return &HashSet[T]{
		s: make(map[T]struct{}),
	}
}

func Add[T comparable](s *HashSet[T], v T) {
	s.s[v] = struct{}{}
}

func Delete[T comparable](s *HashSet[T], v T) {
	delete(s.s, v)
}

func Contains[T comparable](s *HashSet[T], v T) bool {
	_, ok := s.s[v]
	return ok
}

func Keys[T comparable](s *HashSet[T]) []T {
	keys := make([]T, 0, len(s.s))
	for key := range s.s {
		keys = append(keys, key)
	}
	return keys
}
