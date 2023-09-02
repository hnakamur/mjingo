package mjingo

import (
	"github.com/hnakamur/mjingo/internal/datast/indexmap"
)

type valueMap = indexmap.Map[keyRef, Value]
type valueMapEntry = indexmap.Entry[keyRef, Value]

func newValueMap() *valueMap {
	return indexmap.New[keyRef, Value]()
}

func valueMapWithCapacity(capacity uint) *valueMap {
	return indexmap.WithCapacity[keyRef, Value](capacity)
}

func valueMapFromEntries(entries []valueMapEntry) *valueMap {
	return indexmap.FromEntries[keyRef, Value](entries)
}
