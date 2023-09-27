// Package Map defines various functions useful for Map.
package indexmap

import (
	"hash"
	"hash/maphash"
	"slices"
)

type HashEqualer interface {
	comparable
	Hash(h hash.Hash)
}

// Map represents a map which preserves the insertion order of keys.
type Map[K HashEqualer, V any] struct {
	indexes map[uint64][]uint
	keys    []K
	values  []V
}

// Entry represents an entry in an Map.
type Entry[K HashEqualer, V any] struct {
	Key   K
	Value V
}

// New creates an Map.
func New[K HashEqualer, V any]() *Map[K, V] {
	return &Map[K, V]{indexes: make(map[uint64][]uint)}
}

// WithCapacity creates an Map with the specified capacity.
func WithCapacity[K HashEqualer, V any](capacity uint) *Map[K, V] {
	return &Map[K, V]{
		indexes: make(map[uint64][]uint, capacity),
		keys:    make([]K, 0, capacity),
		values:  make([]V, 0, capacity),
	}
}

// FromEntries creates an Map and inserts specified entries.
func FromEntries[K HashEqualer, V any](entries []Entry[K, V]) *Map[K, V] {
	m := WithCapacity[K, V](uint(len(entries)))
	for _, entry := range entries {
		m.Set(entry.Key, entry.Value)
	}
	return m
}

// Set sets the value for the key in m. If m contains the key, the just
// value is updated without modifying the insertion order of the keys.
func (m *Map[K, V]) Set(key K, val V) {
	keyHash := m.keyHashSum(key)
	ii, ok := m.indexes[keyHash]
	// log.Printf("Map.Set key=%+v, keyHash=%x, ii=%+v, ok=%v", key, keyHash, ii, ok)
	if ok {
		for _, i := range ii {
			if key == m.keys[i] {
				// log.Printf("Map.Set overwrite value at index=%d", i)
				m.values[i] = val
				return
			}
		}
	}
	i := uint(len(m.keys))
	ii = append(ii, i)
	m.indexes[keyHash] = ii
	m.keys = append(m.keys, key)
	m.values = append(m.values, val)
	// log.Printf("Map.Set added value at index=%d", i)
}

// Get gets the value for the key in m. Get returns the value for the key and
// true if the key is found, or zero value of V and false otherwise.
func (m *Map[K, V]) Get(key K) (v V, ok bool) {
	keyHash := m.keyHashSum(key)
	ii, ok := m.indexes[keyHash]
	// log.Printf("Map.Get key=%+v, keyHash=%x, ii=%+v, ok=%v", key, keyHash, ii, ok)
	if ok {
		for _, i := range ii {
			if key == m.keys[i] {
				// log.Printf("Map.Get found key at index=%d", i)
				return m.values[i], true
			}
		}
	}
	return
}

// Delete deletes the value for the key in m. Delete returns the value for the key and
// true if the key is found, or zero value of V and false otherwise.
func (m *Map[K, V]) Delete(key K) (v V, ok bool) {
	keyHash := m.keyHashSum(key)
	ii, ok := m.indexes[keyHash]
	if ok {
		for j, i := range ii {
			if key == m.keys[i] {
				v, ok = m.values[i], true
				if len(ii) > 1 {
					m.indexes[keyHash] = slices.Delete(ii, j, j+1)
				} else {
					delete(m.indexes, keyHash)
				}

				for j, key := range m.keys[i+1:] {
					m.updateIndex(key, i+uint(j))
				}
				clear(m.keys[i : i+1])
				m.keys = append(m.keys[:i], m.keys[i+1:]...)

				clear(m.values[i : i+1])
				m.values = append(m.values[:i], m.values[i+1:]...)
			}
		}
	}
	return
}

func (m *Map[K, V]) updateIndex(key K, newIndex uint) {
	keyHash := m.keyHashSum(key)
	ii, ok := m.indexes[keyHash]
	if ok {
		for j, i := range ii {
			if key == m.keys[i] {
				ii[j] = newIndex
				return
			}
		}
	}
}

// Len returns the size of m.
func (m *Map[K, V]) Len() uint { return uint(len(m.keys)) }

// EntryAt returns the entry at i in m. EntryAt returns the key, the value,
// and whether the i was in the range or not.
func (m *Map[K, V]) EntryAt(i uint) (e Entry[K, V], ok bool) {
	if i >= m.Len() {
		return
	}
	return Entry[K, V]{Key: m.keys[i], Value: m.values[i]}, true
}

// Clone returns the shallow copy of m.
func (m *Map[K, V]) Clone() *Map[K, V] {
	l := m.Len()
	rv := WithCapacity[K, V](l)
	for i := uint(0); i < l; i++ {
		e, _ := m.EntryAt(i)
		rv.Set(e.Key, e.Value)
	}
	return rv
}

// Keys returns keys in preserved order.
func (m *Map[K, V]) Keys() []K {
	return m.keys
}

var seed = maphash.MakeSeed()

func (m *Map[K, V]) keyHashSum(key K) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	key.Hash(&h)
	return h.Sum64()
}
