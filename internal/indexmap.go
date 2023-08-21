package internal

import "github.com/hnakamur/mjingo/internal/datast/indexmap"

type IndexMap indexmap.Map[KeyRef, Value]
type IndexMapEntry = indexmap.Entry[KeyRef, Value]

func NewIndexMap() *IndexMap {
	return (*IndexMap)(indexmap.New[KeyRef, Value]())
}

func NewIndexMapWithCapacity(capacity uint) *IndexMap {
	return (*IndexMap)(indexmap.WithCapacity[KeyRef, Value](capacity))
}

func NewIndexMapFromEntries(entries []IndexMapEntry) *IndexMap {
	return (*IndexMap)(indexmap.FromEntries[KeyRef, Value](entries))
}

func (m *IndexMap) Set(key KeyRef, value Value) {
	indexmap.Set((*indexmap.Map[KeyRef, Value])(m), key, value)
}

func (m *IndexMap) Get(key KeyRef) (Value, bool) {
	return indexmap.Get((*indexmap.Map[KeyRef, Value])(m), key)
}

func (m *IndexMap) Delete(key KeyRef) (Value, bool) {
	return indexmap.Delete((*indexmap.Map[KeyRef, Value])(m), key)
}

func (m *IndexMap) Len() uint {
	return indexmap.Len((*indexmap.Map[KeyRef, Value])(m))
}

func (m *IndexMap) EntryAt(i uint) (IndexMapEntry, bool) {
	return indexmap.EntryAt((*indexmap.Map[KeyRef, Value])(m), i)
}

func (m *IndexMap) Clone() *IndexMap {
	return (*IndexMap)(indexmap.Clone((*indexmap.Map[KeyRef, Value])(m)))
}

func (m *IndexMap) keys() []KeyRef {
	return indexmap.Keys((*indexmap.Map[KeyRef, Value])(m))
}
