// Package indexmap defines various functions useful for indexmap.
package indexmap

// IndexMap represents a map which preserves the insertion order of keys.
type IndexMap[K comparable, V any] struct {
	indexes map[K]uint
	keys    []K
	values  []V
}

// New creates an IndexMap.
func New[K comparable, V any]() *IndexMap[K, V] {
	return &IndexMap[K, V]{indexes: make(map[K]uint)}
}

// WithCapacity creates an IndexMap with the specified capacity.
func WithCapacity[K comparable, V any](capacity uint) *IndexMap[K, V] {
	return &IndexMap[K, V]{
		indexes: make(map[K]uint, capacity),
		keys:    make([]K, 0, capacity),
		values:  make([]V, 0, capacity),
	}
}

// Set sets the value for the key in m. If m contains the key, the just
// value is updated without modifying the insertion order of the keys.
func Set[K comparable, V any](m *IndexMap[K, V], key K, value V) {
	i, ok := m.indexes[key]
	if ok {
		m.values[i] = value
	} else {
		i := uint(len(m.keys))
		m.indexes[key] = i
		m.keys = append(m.keys, key)
		m.values = append(m.values, value)
	}
}

// Get gets the value for the key in m. Get returns the value for the key and
// true if the key is found, or zero value of V and false otherwise.
func Get[K comparable, V any](m *IndexMap[K, V], key K) (v V, ok bool) {
	var i uint
	i, ok = m.indexes[key]
	if ok {
		return m.values[i], true
	}
	return
}

// Delete deletes the value for the key in m. Delete returns the value for the key and
// true if the key is found, or zero value of V and false otherwise.
func Delete[K comparable, V any](m *IndexMap[K, V], key K) (v V, ok bool) {
	var i uint
	i, ok = m.indexes[key]
	if ok {
		v = m.values[i]
		delete(m.indexes, key)

		clear(m.keys[i : i+1])
		m.keys = append(m.keys[:i], m.keys[i+1:]...)

		clear(m.values[i : i+1])
		m.values = append(m.values[:i], m.values[i+1:]...)
	}
	return
}

// Len returns the size of m.
func Len[K comparable, V any](m *IndexMap[K, V]) uint {
	return uint(len(m.keys))
}

// GetEntryAt returns the entry at i in m. GetEntryAt returns the key, the value,
// and whether the i was in the range or not.
func GetEntryAt[K comparable, V any](m *IndexMap[K, V], i uint) (k K, v V, ok bool) {
	if i >= Len(m) {
		return
	}
	return m.keys[i], m.values[i], true
}
