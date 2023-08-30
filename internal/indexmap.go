package internal

import (
	"hash/maphash"
	"slices"
)

type IndexMap struct {
	indexes map[uint64][]uint
	keys    []KeyRef
	values  []Value
}

type IndexMapEntry struct {
	Key   KeyRef
	Value Value
}

func NewIndexMap() *IndexMap {
	return &IndexMap{indexes: make(map[uint64][]uint)}
}

func NewIndexMapWithCapacity(capacity uint) *IndexMap {
	return &IndexMap{
		indexes: make(map[uint64][]uint, capacity),
		keys:    make([]KeyRef, 0, capacity),
		values:  make([]Value, 0, capacity),
	}
}

func NewIndexMapFromEntries(entries []IndexMapEntry) *IndexMap {
	m := NewIndexMapWithCapacity(uint(len(entries)))
	for _, entry := range entries {
		m.Set(entry.Key, entry.Value)
	}
	return m
}

func (m *IndexMap) Set(key KeyRef, val Value) {
	keyHash := keyRefHashSum(key)
	ii, ok := m.indexes[keyHash]
	// log.Printf("IndexMap.Set key=%+v, keyHash=%x, ii=%+v, ok=%v", key, keyHash, ii, ok)
	if ok {
		for _, i := range ii {
			if KeyRefEqual(key, m.keys[i]) {
				// log.Printf("IndexMap.Set overwrite value at index=%d", i)
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
	// log.Printf("IndexMap.Set added value at index=%d", i)
}

func (m *IndexMap) Get(key KeyRef) (v Value, ok bool) {
	keyHash := keyRefHashSum(key)
	ii, ok := m.indexes[keyHash]
	// log.Printf("IndexMap.Get key=%+v, keyHash=%x, ii=%+v, ok=%v", key, keyHash, ii, ok)
	if ok {
		for _, i := range ii {
			if KeyRefEqual(key, m.keys[i]) {
				// log.Printf("IndexMap.Get found key at index=%d", i)
				return m.values[i], true
			}
		}
	}
	return
}

func (m *IndexMap) Delete(key KeyRef) (v Value, ok bool) {
	keyHash := keyRefHashSum(key)
	ii, ok := m.indexes[keyHash]
	if ok {
		for j, i := range ii {
			if KeyRefEqual(key, m.keys[i]) {
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

func (m *IndexMap) updateIndex(key KeyRef, newIndex uint) {
	keyHash := keyRefHashSum(key)
	ii, ok := m.indexes[keyHash]
	if ok {
		for j, i := range ii {
			if KeyRefEqual(key, m.keys[i]) {
				ii[j] = newIndex
				m.indexes[keyHash] = ii
				return
			}
		}
	}
}

func (m *IndexMap) Len() uint { return uint(len(m.keys)) }

func (m *IndexMap) EntryAt(i uint) (e IndexMapEntry, ok bool) {
	if i >= m.Len() {
		return
	}
	return IndexMapEntry{Key: m.keys[i], Value: m.values[i]}, true
}

func (m *IndexMap) Clone() *IndexMap {
	l := m.Len()
	rv := NewIndexMapWithCapacity(l)
	for i := uint(0); i < l; i++ {
		rv.Set(m.keys[i], m.values[i])
	}
	return rv
}

func (m *IndexMap) Keys() []KeyRef { return m.keys }

var seed = maphash.MakeSeed()

func keyRefHashSum(key KeyRef) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	key.Hash(&h)
	return h.Sum64()
}
