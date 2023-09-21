package mjingo

import (
	"cmp"
	"slices"
)

func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func mapSortedKeys[K interface {
	comparable
	cmp.Ordered
}, V any](m map[K]V) []K {
	keys := mapKeys(m)
	slices.Sort(keys)
	return keys
}
