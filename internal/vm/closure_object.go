package vm

import "github.com/hnakamur/mjingo/value"

type closure struct {
	values map[string]value.Value
}

func newClosure() closure {
	return closure{values: make(map[string]value.Value)}
}

func (c closure) clone() closure {
	values := make(map[string]value.Value, len(c.values))
	for key, val := range c.values {
		values[key] = val
	}
	return closure{values: values}
}
