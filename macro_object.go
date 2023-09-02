package mjingo

import (
	"fmt"
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
)

type macroData struct {
	name            string
	argSpec         []string
	macroRefID      uint
	closure         Value
	callerReference bool
}

type macro struct {
	data macroData
}

var _ = (object)((*macro)(nil))
var _ = (caller)((*macro)(nil))
var _ = (structObject)((*macro)(nil))

func (m *macro) String() string {
	return fmt.Sprintf("<macro %s>", m.data.name)
}

func (m *macro) Kind() objectKind { return objectKindStruct }

func (m *macro) Call(state *vmState, args []Value) (Value, error) {
	var kwargs *valueMap
	if len(args) > 0 {
		if mapVal, ok := args[len(args)-1].(mapValue); ok && mapVal.Type == mapTypeKwargs {
			kwargs = mapVal.Map
			args = args[:len(args)-1]
		}
	}

	if len(args) > len(m.data.argSpec) {
		return nil, newError(TooManyArguments, "")
	}

	kwargsUsed := hashset.NewStrHashSet()
	argValues := make([]Value, 0, len(m.data.argSpec))
	for i, name := range m.data.argSpec {
		var kwarg Value
		if kwargs != nil {
			kwarg, _ = kwargs.Get(keyRefFromValue(valueFromString(name)))
			// TODO: change to below as well as IndexMap
			// kwarg, _ = kwargs.Get(KeyRefFromString(name))
		}

		var arg Value
		switch {
		case i < len(args) && kwarg != nil:
			return nil, newError(TooManyArguments, fmt.Sprintf("duplicate argument `%s`", name))
		case i < len(args) && kwarg == nil:
			arg = args[i].clone()
		case i >= len(args) && kwarg != nil:
			kwargsUsed.Add(name)
			arg = kwarg.clone()
		default:
			arg = Undefined
		}
		argValues = append(argValues, arg)
	}

	caller := option.None[Value]()
	if m.data.callerReference {
		kwargsUsed.Add("caller")
		// option.AndThen(kwargs)
		caller = option.Some[Value](Undefined)
		if kwargs != nil {
			if v, ok := kwargs.Get(keyRefFromString("caller")); ok {
				caller = option.Some(v)
			}
		}
	}

	if kwargs != nil {
		for _, keyRef := range kwargs.Keys() {
			if optKey := keyRef.AsStr(); optKey.IsSome() {
				if !kwargsUsed.Contains(optKey.Unwrap()) {
					return nil, newError(TooManyArguments,
						fmt.Sprintf("unknown keyword argument `%s`", optKey.Unwrap()))
				}
			}
		}
	}

	instsAndOffset := state.macros[m.data.macroRefID]
	insts := instsAndOffset.a
	offset := instsAndOffset.b
	vm := newVirtualMachine(state.env)
	var b strings.Builder
	out := newOutput(&b)

	closure := m.data.closure.clone()

	if _, err := vm.evalMacro(insts, offset, closure, caller, out, state, argValues); err != nil {
		return nil, err
	}

	if _, ok := state.autoEscape.(AutoEscapeNone); !ok {
		return ValueFromSafeString(b.String()), nil
	}
	return valueFromString(b.String()), nil
}

func (m *macro) StaticFields() option.Option[[]string] {
	return option.Some([]string{"name", "arguments", "caller"})
}

func (m *macro) GetField(name string) option.Option[Value] {
	switch name {
	case "name":
		return option.Some(valueFromString(m.data.name))
	case "arguments":
		return option.Some(valueFromSlice(slicex.Map(m.data.argSpec, valueFromString)))
	case "caller":
		return option.Some(valueFromBool(m.data.callerReference))
	}
	return option.None[Value]()
}

func (m *macro) Fields() []string { return nil }
