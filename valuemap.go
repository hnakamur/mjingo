package mjingo

import (
	"github.com/hnakamur/mjingo/internal/datast/indexmap"
)

type ValueMap = indexmap.Map[KeyRef, Value]
type ValueMapEntry = indexmap.Entry[KeyRef, Value]

func NewValueMap() *ValueMap {
	return indexmap.New[KeyRef, Value]()
}

func ValueMapWithCapacity(capacity uint) *ValueMap {
	return indexmap.WithCapacity[KeyRef, Value](capacity)
}

func ValueMapFromEntries(entries []ValueMapEntry) *ValueMap {
	return indexmap.FromEntries[KeyRef, Value](entries)
}
