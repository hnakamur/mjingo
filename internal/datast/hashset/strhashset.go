package hashset

type StrHashSet = HashSet[string]

func NewStrHashSet() *StrHashSet {
	return New[string]()
}
