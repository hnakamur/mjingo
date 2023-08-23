package hashset

type StrHashSet HashSet[string]

func NewStrHashSet() *StrHashSet {
	return (*StrHashSet)(New[string]())
}

func (s *StrHashSet) Add(v string) bool {
	return Add[string]((*HashSet[string])(s), v)
}

func (s *StrHashSet) Delete(v string) {
	Delete[string]((*HashSet[string])(s), v)
}

func (s *StrHashSet) Contains(v string) bool {
	return Contains[string]((*HashSet[string])(s), v)
}

func (s *StrHashSet) Keys() []string {
	return Keys[string]((*HashSet[string])(s))
}
