package mjingo

import (
	"fmt"
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/option"
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

var _ = (Object)((*macro)(nil))
var _ = (Caller)((*macro)(nil))
var _ = (StructObject)((*macro)(nil))

func (m *macro) String() string {
	return fmt.Sprintf("<macro %s>", m.data.name)
}

func (m *macro) Kind() ObjectKind { return ObjectKindStruct }

func (m *macro) Call(state *State, args []Value) (Value, error) {
	var kwargs *valueMap
	if len(args) > 0 {
		if mapVal, ok := args[len(args)-1].data.(mapValue); ok && mapVal.Type == mapTypeKwargs {
			kwargs = mapVal.Map
			args = args[:len(args)-1]
		}
	}

	if len(args) > len(m.data.argSpec) {
		return Value{}, NewError(TooManyArguments, "")
	}

	kwargsUsed := hashset.NewStrHashSet()
	argValues := make([]Value, 0, len(m.data.argSpec))
	for i, name := range m.data.argSpec {
		var kwarg Value
		if kwargs != nil {
			kwarg, _ = kwargs.Get(keyRefFromValue(valueFromString(name)))
		}

		var arg Value
		switch {
		case i < len(args) && kwarg.data != nil:
			return Value{}, NewError(TooManyArguments, fmt.Sprintf("duplicate argument `%s`", name))
		case i < len(args) && kwarg.data == nil:
			arg = args[i].clone()
		case i >= len(args) && kwarg.data != nil:
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
					return Value{}, NewError(TooManyArguments,
						fmt.Sprintf("unknown keyword argument `%s`", optKey.Unwrap()))
				}
			}
		}
	}

	instsAndOffset := state.macros[m.data.macroRefID]
	insts := instsAndOffset.insts
	offset := instsAndOffset.offset
	vm := newVirtualMachine(state.env)
	var b strings.Builder
	out := newOutput(&b)

	closure := m.data.closure.clone()

	if _, err := vm.evalMacro(insts, offset, closure, caller, out, state, argValues); err != nil {
		return Value{}, err
	}

	if _, ok := state.autoEscape.(autoEscapeNone); !ok {
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
