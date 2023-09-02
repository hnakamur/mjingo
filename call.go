package mjingo

import (
	"fmt"
)

type Caller interface {
	Call(state *State, args []Value) (Value, error)
}

type CallMethoder interface {
	CallMethod(state *State, name string, args []Value) (Value, error)
}

func Call(receiver Value, state *State, args []Value) (Value, error) {
	if dyVal, ok := receiver.(DynamicValue); ok {
		if c, ok := dyVal.Dy.(Caller); ok {
			return c.Call(state, args)
		}
		return nil, NewError(InvalidOperation, "tried to call non callable object")
	}
	return notCallableValueType(receiver)
}

func notCallableValueType(v Value) (Value, error) {
	return nil, NewError(InvalidOperation,
		fmt.Sprintf("value of type %s is not callable", v.Kind()))
}

func CallMethod(receiver Value, state *State, name string, args []Value) (Value, error) {
	switch v := receiver.(type) {
	case MapValue:
		if val, ok := v.Map.Get(KeyRefFromString(name)); ok {
			if dyVal, ok := val.(DynamicValue); ok {
				if c, ok := dyVal.Dy.(Caller); ok {
					return c.Call(state, args)
				}
			}
			return notCallableValueType(val)
		}
	case DynamicValue:
		if c, ok := v.Dy.(CallMethoder); ok {
			return c.CallMethod(state, name, args)
		}
	}
	return nil, NewError(InvalidOperation,
		fmt.Sprintf("object has no method named %s", name))
}
