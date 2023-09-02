package vm

import (
	"fmt"

	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/value"
)

type Caller interface {
	Call(state *State, args []value.Value) (value.Value, error)
}

type CallMethoder interface {
	CallMethod(state *State, name string, args []value.Value) (value.Value, error)
}

func Call(receiver value.Value, state *State, args []value.Value) (value.Value, error) {
	if dyVal, ok := receiver.(value.DynamicValue); ok {
		if c, ok := dyVal.Dy.(Caller); ok {
			return c.Call(state, args)
		}
		return nil, common.NewError(common.InvalidOperation, "tried to call non callable object")
	}
	return notCallableValueType(receiver)
}

func notCallableValueType(v value.Value) (value.Value, error) {
	return nil, common.NewError(common.InvalidOperation,
		fmt.Sprintf("value of type %s is not callable", v.Kind()))
}

func CallMethod(receiver value.Value, state *State, name string, args []value.Value) (value.Value, error) {
	switch v := receiver.(type) {
	case value.MapValue:
		if val, ok := v.Map.Get(value.KeyRefFromString(name)); ok {
			if dyVal, ok := val.(value.DynamicValue); ok {
				if c, ok := dyVal.Dy.(Caller); ok {
					return c.Call(state, args)
				}
			}
			return notCallableValueType(val)
		}
	case value.DynamicValue:
		if c, ok := v.Dy.(CallMethoder); ok {
			return c.CallMethod(state, name, args)
		}
	}
	return nil, common.NewError(common.InvalidOperation,
		fmt.Sprintf("object has no method named %s", name))
}
