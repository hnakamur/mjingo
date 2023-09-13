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

func valueCall(receiver Value, state *State, args []Value) (Value, error) {
	if dyVal, ok := receiver.data.(dynamicValue); ok {
		if c, ok := dyVal.Dy.(Caller); ok {
			return c.Call(state, args)
		}
		return Value{}, NewError(InvalidOperation, "tried to call non callable object")
	}
	return notCallableValueType(receiver)
}

func notCallableValueType(v Value) (Value, error) {
	return Value{}, NewError(InvalidOperation,
		fmt.Sprintf("value of type %s is not callable", v.kind()))
}

func callMethod(receiver Value, state *State, name string, args []Value) (Value, error) {
	switch v := receiver.data.(type) {
	case mapValue:
		if val, ok := v.Map.Get(keyRefFromString(name)); ok {
			if dyVal, ok := val.data.(dynamicValue); ok {
				if c, ok := dyVal.Dy.(Caller); ok {
					return c.Call(state, args)
				}
			}
			return notCallableValueType(val)
		}
	case dynamicValue:
		if c, ok := v.Dy.(CallMethoder); ok {
			return c.CallMethod(state, name, args)
		}
	}
	return Value{}, NewError(InvalidOperation,
		fmt.Sprintf("object has no method named %s", name))
}
