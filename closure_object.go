package mjingo

import (
	"github.com/hnakamur/mjingo/internal/datast/option"
)

type Closure struct {
	values map[string]Value
}

var _ = (Object)((*Closure)(nil))
var _ = (StructObject)((*Closure)(nil))

func newClosure() Closure {
	return Closure{values: make(map[string]Value)}
}

func (c Closure) clone() Closure {
	values := make(map[string]Value, len(c.values))
	for key, val := range c.values {
		values[key] = val
	}
	return Closure{values: values}
}

// Stores a value by key in the closure.
func (c *Closure) store(key string, val Value) {
	c.values[key] = val
}

// Upset a value into the closure.
func (c *Closure) storeIfMissing(key string, f func() Value) {
	if _, ok := c.values[key]; !ok {
		c.values[key] = f()
	}
}

func (c *Closure) Kind() ObjectKind { return ObjectKindStruct }

func (c *Closure) StaticFields() option.Option[[]string] { return option.None[[]string]() }

func (c *Closure) Fields() []string {
	keys := make([]string, 0, len(c.values))
	for key := range c.values {
		keys = append(keys, key)
	}
	return keys
}

func (c *Closure) GetField(name string) option.Option[Value] {
	val, ok := c.values[name]
	if ok {
		return option.Some(val)
	}
	return option.None[Value]()
}
