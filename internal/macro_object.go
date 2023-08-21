package internal

import (
	"fmt"
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
)

type MacroData struct {
	name            string
	argSpec         []string
	macroRefID      uint
	closure         Value
	callerReference bool
}

type Macro struct {
	data MacroData
}

var _ = (Object)((*Macro)(nil))
var _ = (Caller)((*Macro)(nil))
var _ = (StructObject)((*Macro)(nil))

func (m *Macro) String() string {
	return fmt.Sprintf("<macro %s>", m.data.name)
}

func (m *Macro) Kind() ObjectKind { return ObjectKindStruct }

func (m *Macro) Call(state *State, args []Value) (Value, error) {
	var kwargs *IndexMap
	if len(args) > 0 {
		if mapVal, ok := args[len(args)-1].(mapValue); ok {
			if mapVal.mapTyp == mapTypeKwargs {
				kwargs = mapVal.m
				args = args[:len(args)-1]
			}
		}
	}

	if len(args) > len(m.data.argSpec) {
		return nil, NewError(TooManyArguments, "")
	}

	kwargsUsed := hashset.NewStrHashSet()
	argValues := make([]Value, 0, len(m.data.argSpec))
	for i, name := range m.data.argSpec {
		var kwarg Value
		if kwargs != nil {
			kwarg, _ = kwargs.Get(KeyRefFromString(name))
		}

		var arg Value
		switch {
		case i < len(args) && kwarg != nil:
			return nil, NewError(TooManyArguments, fmt.Sprintf("duplicate argument `%s`", name))
		case i < len(args) && kwarg == nil:
			arg = args[i].Clone()
		case i >= len(args) && kwarg != nil:
			kwargsUsed.Add(name)
			arg = kwarg.Clone()
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
			if v, ok := kwargs.Get(KeyRefFromString("caller")); ok {
				caller = option.Some(v)
			}
		}
	}

	if kwargs != nil {
		for _, keyRef := range kwargs.keys() {
			if optKey := keyRef.AsStr(); option.IsSome(optKey) {
				if !kwargsUsed.Contains(option.Unwrap(optKey)) {
					return nil, NewError(TooManyArguments,
						fmt.Sprintf("unknown keyword argument `%s`", option.Unwrap(optKey)))
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

	closure := m.data.closure.Clone()

	if _, err := vm.evalMacro(insts, offset, closure, caller, out, state, argValues); err != nil {
		return nil, err
	}

	if _, ok := state.autoEscape.(AutoEscapeNone); !ok {
		return ValueFromSafeString(b.String()), nil
	}
	return ValueFromString(b.String()), nil
}

func (m *Macro) StaticFields() option.Option[[]string] {
	return option.Some([]string{"name", "arguments", "caller"})
}

func (m *Macro) GetField(name string) option.Option[Value] {
	switch name {
	case "name":
		return option.Some(ValueFromString(m.data.name))
	case "arguments":
		return option.Some(ValueFromSlice(slicex.Map(m.data.argSpec, ValueFromString)))
	case "caller":
		return option.Some(ValueFromBool(m.data.callerReference))
	}
	return option.None[Value]()
}

func (m *Macro) Fields() []string { return nil }
