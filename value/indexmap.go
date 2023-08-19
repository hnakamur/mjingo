package value

import "github.com/hnakamur/mjingo/internal/datast/indexmap"

type IndexMap = indexmap.Map[KeyRef, Value]
type IndexMapEntry = indexmap.Entry[KeyRef, Value]

func NewIndexMap() *IndexMap {
	return indexmap.New[KeyRef, Value]()
}

func NewIndexMapWithCapacity(capacity uint) *IndexMap {
	return indexmap.WithCapacity[KeyRef, Value](capacity)
}

func NewIndexMapFromEntries(entries []IndexMapEntry) *IndexMap {
	return indexmap.FromEntries[KeyRef, Value](entries)
}

func IndexMapSet(m *IndexMap, key KeyRef, value Value) {
	indexmap.Set(m, key, value)
}

func IndexMapGet(m *IndexMap, key KeyRef) (Value, bool) {
	return indexmap.Get(m, key)
}

func IndexMapDelete(m *IndexMap, key KeyRef) (Value, bool) {
	return indexmap.Delete(m, key)
}

func IndexMapLen(m *IndexMap) uint {
	return indexmap.Len(m)
}

func IndexMapEntryAt(m *IndexMap, i uint) (IndexMapEntry, bool) {
	return indexmap.EntryAt(m, i)
}
