// Package indexmap defines various functions useful for indexmap.
package indexmap

// Map represents a map which preserves the insertion order of keys.
type Map[K comparable, V any] struct {
	indexes map[K]uint
	keys    []K
	values  []V
}

// Entry represents an entry in an IndexMap.
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// New creates an IndexMap.
func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{indexes: make(map[K]uint)}
}

// WithCapacity creates an IndexMap with the specified capacity.
func WithCapacity[K comparable, V any](capacity uint) *Map[K, V] {
	return &Map[K, V]{
		indexes: make(map[K]uint, capacity),
		keys:    make([]K, 0, capacity),
		values:  make([]V, 0, capacity),
	}
}

// FromEntries creates an IndexMap and inserts specified entries.
func FromEntries[K comparable, V any](entries []Entry[K, V]) *Map[K, V] {
	m := WithCapacity[K, V](uint(len(entries)))
	for _, entry := range entries {
		Set(m, entry.Key, entry.Value)
	}
	return m
}

// Set sets the value for the key in m. If m contains the key, the just
// value is updated without modifying the insertion order of the keys.
func Set[K comparable, V any](m *Map[K, V], key K, value V) {
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
func Get[K comparable, V any](m *Map[K, V], key K) (v V, ok bool) {
	var i uint
	i, ok = m.indexes[key]
	if ok {
		return m.values[i], true
	}
	return
}

// Delete deletes the value for the key in m. Delete returns the value for the key and
// true if the key is found, or zero value of V and false otherwise.
func Delete[K comparable, V any](m *Map[K, V], key K) (v V, ok bool) {
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
func Len[K comparable, V any](m *Map[K, V]) uint {
	return uint(len(m.keys))
}

// EntryAt returns the entry at i in m. EntryAt returns the key, the value,
// and whether the i was in the range or not.
func EntryAt[K comparable, V any](m *Map[K, V], i uint) (e Entry[K, V], ok bool) {
	if i >= Len(m) {
		return
	}
	return Entry[K, V]{Key: m.keys[i], Value: m.values[i]}, true
}

// Clone returns the shallow copy of m.
func Clone[K comparable, V any](m *Map[K, V]) *Map[K, V] {
	l := Len(m)
	rv := WithCapacity[K, V](l)
	for i := uint(0); i < l; i++ {
		e, _ := EntryAt(m, l)
		Set(rv, e.Key, e.Value)
	}
	return rv
}
