package mjingo

import (
	"github.com/hnakamur/mjingo/option"
)

type closureObject struct {
	values map[string]Value
}

var _ = (Object)((*closureObject)(nil))
var _ = (StructObject)((*closureObject)(nil))

func newClosure() closureObject {
	return closureObject{values: make(map[string]Value)}
}

func (c closureObject) clone() closureObject {
	values := make(map[string]Value, len(c.values))
	for key, val := range c.values {
		values[key] = val
	}
	return closureObject{values: values}
}

// Stores a value by key in the closure.
func (c *closureObject) store(key string, val Value) {
	c.values[key] = val
}

// Upset a value into the closure.
func (c *closureObject) storeIfMissing(key string, f func() Value) {
	if _, ok := c.values[key]; !ok {
		c.values[key] = f()
	}
}

func (c *closureObject) Kind() ObjectKind { return ObjectKindStruct }

func (c *closureObject) StaticFields() option.Option[[]string] { return option.None[[]string]() }

func (c *closureObject) Fields() []string {
	keys := make([]string, 0, len(c.values))
	for key := range c.values {
		keys = append(keys, key)
	}
	return keys
}

func (c *closureObject) GetField(name string) option.Option[Value] {
	val, ok := c.values[name]
	if ok {
		return option.Some(val)
	}
	return option.None[Value]()
}
