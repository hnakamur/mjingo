package vm

import (
	"fmt"
	"strings"

	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/internal/value"
)

type MacroData struct {
	name            string
	argSpec         []string
	macroRefID      uint
	closure         value.Value
	callerReference bool
}

type Macro struct {
	data MacroData
}

var _ = (value.Object)((*Macro)(nil))
var _ = (Caller)((*Macro)(nil))
var _ = (value.StructObject)((*Macro)(nil))

func (m *Macro) String() string {
	return fmt.Sprintf("<macro %s>", m.data.name)
}

func (m *Macro) Kind() value.ObjectKind { return value.ObjectKindStruct }

func (m *Macro) Call(state *State, args []value.Value) (value.Value, error) {
	var kwargs *value.ValueMap
	if len(args) > 0 {
		if mapVal, ok := args[len(args)-1].(value.MapValue); ok && mapVal.Type == value.MapTypeKwargs {
			kwargs = mapVal.Map
			args = args[:len(args)-1]
		}
	}

	if len(args) > len(m.data.argSpec) {
		return nil, common.NewError(common.TooManyArguments, "")
	}

	kwargsUsed := hashset.NewStrHashSet()
	argValues := make([]value.Value, 0, len(m.data.argSpec))
	for i, name := range m.data.argSpec {
		var kwarg value.Value
		if kwargs != nil {
			kwarg, _ = kwargs.Get(value.KeyRefFromValue(value.ValueFromString(name)))
			// TODO: change to below as well as IndexMap
			// kwarg, _ = kwargs.Get(KeyRefFromString(name))
		}

		var arg value.Value
		switch {
		case i < len(args) && kwarg != nil:
			return nil, common.NewError(common.TooManyArguments, fmt.Sprintf("duplicate argument `%s`", name))
		case i < len(args) && kwarg == nil:
			arg = args[i].Clone()
		case i >= len(args) && kwarg != nil:
			kwargsUsed.Add(name)
			arg = kwarg.Clone()
		default:
			arg = value.Undefined
		}
		argValues = append(argValues, arg)
	}

	caller := option.None[value.Value]()
	if m.data.callerReference {
		kwargsUsed.Add("caller")
		// option.AndThen(kwargs)
		caller = option.Some[value.Value](value.Undefined)
		if kwargs != nil {
			if v, ok := kwargs.Get(value.KeyRefFromString("caller")); ok {
				caller = option.Some(v)
			}
		}
	}

	if kwargs != nil {
		for _, keyRef := range kwargs.Keys() {
			if optKey := keyRef.AsStr(); optKey.IsSome() {
				if !kwargsUsed.Contains(optKey.Unwrap()) {
					return nil, common.NewError(common.TooManyArguments,
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

	closure := m.data.closure.Clone()

	if _, err := vm.evalMacro(insts, offset, closure, caller, out, state, argValues); err != nil {
		return nil, err
	}

	if _, ok := state.autoEscape.(AutoEscapeNone); !ok {
		return value.ValueFromSafeString(b.String()), nil
	}
	return value.ValueFromString(b.String()), nil
}

func (m *Macro) StaticFields() option.Option[[]string] {
	return option.Some([]string{"name", "arguments", "caller"})
}

func (m *Macro) GetField(name string) option.Option[value.Value] {
	switch name {
	case "name":
		return option.Some(value.ValueFromString(m.data.name))
	case "arguments":
		return option.Some(value.ValueFromSlice(slicex.Map(m.data.argSpec, value.ValueFromString)))
	case "caller":
		return option.Some(value.ValueFromBool(m.data.callerReference))
	}
	return option.None[value.Value]()
}

func (m *Macro) Fields() []string { return nil }
