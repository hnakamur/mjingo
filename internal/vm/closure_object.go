package vm

import (
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/value"
)

type Closure struct {
	values map[string]value.Value
}

var _ = (value.Object)((*Closure)(nil))
var _ = (value.StructObject)((*Closure)(nil))

func newClosure() Closure {
	return Closure{values: make(map[string]value.Value)}
}

func (c Closure) clone() Closure {
	values := make(map[string]value.Value, len(c.values))
	for key, val := range c.values {
		values[key] = val
	}
	return Closure{values: values}
}

// Stores a value by key in the closure.
func (c *Closure) store(key string, val value.Value) {
	c.values[key] = val
}

// Upset a value into the closure.
func (c *Closure) storeIfMissing(key string, f func() value.Value) {
	if _, ok := c.values[key]; !ok {
		c.values[key] = f()
	}
}

func (c *Closure) Kind() value.ObjectKind { return value.ObjectKindStruct }

func (c *Closure) StaticFields() option.Option[[]string] { return option.None[[]string]() }

func (c *Closure) Fields() []string {
	keys := make([]string, 0, len(c.values))
	for key := range c.values {
		keys = append(keys, key)
	}
	return keys
}

func (c *Closure) GetField(name string) option.Option[value.Value] {
	val, ok := c.values[name]
	if ok {
		return option.Some(val)
	}
	return option.None[value.Value]()
}
