package internal


type closure struct {
	values map[string]Value
}

func newClosure() closure {
	return closure{values: make(map[string]Value)}
}

func (c closure) clone() closure {
	values := make(map[string]Value, len(c.values))
	for key, val := range c.values {
		values[key] = val
	}
	return closure{values: values}
}
