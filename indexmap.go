package mjingo

type indexMap[K comparable, V any] struct {
	indexes map[K]uint
	keys    []K
	values  []V
}

func newIndexMap[K comparable, V any]() *indexMap[K, V] {
	return &indexMap[K, V]{
		indexes: make(map[K]uint),
	}
}

func newIndexMapWithCapacity[K comparable, V any](capacity uint) *indexMap[K, V] {
	return &indexMap[K, V]{
		indexes: make(map[K]uint, capacity),
		keys:    make([]K, 0, capacity),
		values:  make([]V, 0, capacity),
	}
}

func indexMapLoad[K comparable, V any](m *indexMap[K, V], key K) (val V, ok bool) {
	var idx uint
	idx, ok = m.indexes[key]
	if ok {
		val = m.values[idx]
	}
	return
}

func indexMapStore[K comparable, V any](m *indexMap[K, V], key K, val V) {
	idx, ok := m.indexes[key]
	if ok {
		m.values[idx] = val
	} else {
		m.indexes[key] = uint(len(m.keys))
		m.keys = append(m.keys, key)
		m.values = append(m.values, val)
	}
}

func indexMapDelete[K comparable, V any](m *indexMap[K, V], key K) {
	idx, ok := m.indexes[key]
	if ok {
		// https://github.com/golang/go/wiki/SliceTricks#delete
		i := int(idx)

		copy(m.keys[i:], m.keys[i+1:])
		var zeroK K
		m.keys[len(m.keys)-1] = zeroK
		m.keys = m.keys[:len(m.keys)-1]

		copy(m.values[i:], m.values[i+1:])
		var zeroV V
		m.values[len(m.values)-1] = zeroV
		m.values = m.values[:len(m.values)-1]

		delete(m.indexes, key)
	}
}

func indexMapLen[K comparable, V any](m *indexMap[K, V]) uint {
	return uint(len(m.keys))
}

func indexMapGetEntryAt[K comparable, V any](m *indexMap[K, V], idx uint) (key K, val V, ok bool) {
	ok = idx < uint(len(m.keys))
	if ok {
		key = m.keys[idx]
		val = m.values[idx]
	}
	return
}
