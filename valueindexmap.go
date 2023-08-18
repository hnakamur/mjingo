package mjingo

type valueIndexMap struct {
	indexes map[keyRef]uint
	keys    []keyRef
	values  []value
}

type keyRefValue struct {
	key keyRef
	val value
}

func valueIndexMapFromKeyRefValues(keyVals []keyRefValue) *valueIndexMap {
	m := newValueIndexMapWithCapacity(uint(len(keyVals)))
	for _, keyVal := range keyVals {
		m.Store(keyVal.key, keyVal.val)
	}
	return m
}

func newValueIndexMap() *valueIndexMap {
	return &valueIndexMap{
		indexes: make(map[keyRef]uint),
	}
}

func newValueIndexMapWithCapacity(capacity uint) *valueIndexMap {
	return &valueIndexMap{
		indexes: make(map[keyRef]uint, capacity),
		keys:    make([]keyRef, 0, capacity),
		values:  make([]value, 0, capacity),
	}
}

func (m *valueIndexMap) Load(key keyRef) (val value, ok bool) {
	var idx uint
	idx, ok = m.indexes[key]
	if ok {
		val = m.values[idx]
	}
	return
}

func (m *valueIndexMap) Store(key keyRef, val value) {
	idx, ok := m.indexes[key]
	if ok {
		m.values[idx] = val
	} else {
		m.indexes[key] = uint(len(m.keys))
		m.keys = append(m.keys, key)
		m.values = append(m.values, val)
	}
}

func (m *valueIndexMap) Delete(key keyRef) {
	idx, ok := m.indexes[key]
	if ok {
		i := int(idx)

		// https://github.com/golang/go/wiki/SliceTricks#delete

		copy(m.keys[i:], m.keys[i+1:])
		m.keys[len(m.keys)-1] = nil
		m.keys = m.keys[:len(m.keys)-1]

		copy(m.values[i:], m.values[i+1:])
		m.values[len(m.values)-1] = nil
		m.values = m.values[:len(m.values)-1]

		delete(m.indexes, key)
	}
}

func (m *valueIndexMap) Len() uint {
	return uint(len(m.keys))
}

func (m *valueIndexMap) GetEntryAt(idx uint) (key keyRef, val value, ok bool) {
	ok = idx < uint(len(m.keys))
	if ok {
		key = m.keys[idx]
		val = m.values[idx]
	}
	return
}

func (m *valueIndexMap) Clone() *valueIndexMap {
	l := m.Len()
	rv := newValueIndexMapWithCapacity(l)
	for i := uint(0); i < l; i++ {
		key, val, _ := m.GetEntryAt(i)
		rv.Store(key, val)
	}
	return rv
}
