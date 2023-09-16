package mjingo

import (
	"fmt"
)

// Caller is implemented by any value that has Call method.
type Caller interface {
	// Call is called when the object is invoked directly.
	//
	// To convert the arguments into arguments use the
	// ConvertArgToGoValue or ConvertVariadicArgsToGoValue function.
	Call(state *State, args []Value) (Value, error)
}

// CallMethoder is implemented by any value that has CallMethod method.
type CallMethoder interface {
	// CallMethod is called when the engine tries to call a method on the object.
	//
	// It's the responsibility of the implementer to ensure that an
	// error is generated if an invalid method is invoked.
	//
	// To convert the arguments into arguments use the
	// ConvertArgToGoValue or ConvertVariadicArgsToGoValue function.
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
