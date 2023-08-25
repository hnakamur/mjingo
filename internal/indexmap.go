package internal

import (
	"github.com/hnakamur/mjingo/internal/datast/indexmap"
)

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
