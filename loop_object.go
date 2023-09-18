package mjingo

import (
	"fmt"
	"slices"

	"github.com/hnakamur/mjingo/option"
)

type loopObject struct {
	len              uint
	idx              uint
	depth            uint
	valueTriple      [3]option.Option[Value]
	lastChangedValue []Value
}

var _ = (Object)((*loopObject)(nil))
var _ = (CallMethoder)((*loopObject)(nil))
var _ = (StructObject)((*loopObject)(nil))

func (l *loopObject) Kind() ObjectKind { return ObjectKindStruct }

func (l *loopObject) CallMethod(state *State, name string, args []Value) (Value, error) {
	switch name {
	case "changed":
		if slices.Equal(l.lastChangedValue, args) {
			return valueFromBool(false), nil
		}
		l.lastChangedValue = make([]Value, len(args))
		for i, arg := range args {
			l.lastChangedValue[i] = arg
		}
		return valueFromBool(true), nil
	case "cycle":
		idx := l.idx % uint(len(args))
		if idx < uint(len(args)) {
			return args[idx].clone(), nil
		}
		return Undefined, nil
	}
	return Value{}, NewError(UnknownMethod, fmt.Sprintf("loop object has no method named %s", name))
}

func (*loopObject) StaticFields() option.Option[[]string] {
	return option.Some([]string{
		"index0",
		"index",
		"length",
		"revindex",
		"revindex0",
		"first",
		"last",
		"depth",
		"depth0",
		"previtem",
		"nextitem",
	})
}

func (*loopObject) Fields() []string { return nil }

func (l *loopObject) GetField(name string) option.Option[Value] {
	idx := l.idx
	// if we never iterated, then all attributes are undefined.
	// this can happen in some rare circumstances where the engine
	// did not manage to iterate
	if idx == ^uint(0) {
		return option.Some[Value](Undefined)
	}
	switch name {
	case "index0":
		return option.Some[Value](valueFromI64(int64(idx)))
	case "index":
		return option.Some[Value](valueFromI64(int64(idx + 1)))
	case "length":
		return option.Some[Value](valueFromI64(int64(l.len)))
	case "revindex":
		return option.Some[Value](valueFromI64(int64(uintSaturatingSub(l.len, idx))))
	case "revindex0":
		return option.Some[Value](valueFromI64(int64(uintSaturatingSub(uintSaturatingSub(l.len, idx), 1))))
	case "first":
		return option.Some[Value](valueFromBool(idx == 0))
	case "last":
		return option.Some[Value](valueFromBool(l.len == 0 || idx == l.len-1))
	case "depth":
		return option.Some[Value](valueFromI64(int64(l.depth + 1)))
	case "depth0":
		return option.Some[Value](valueFromI64(int64(l.depth)))
	case "previtem":
		return option.Some[Value](l.valueTriple[0].UnwrapOr(Undefined).clone())
	case "nextitem":
		return option.Some[Value](l.valueTriple[2].UnwrapOr(Undefined).clone())
	}
	return option.None[Value]()
}

func uintSaturatingSub(x, y uint) uint {
	if y > x {
		return 0
	}
	return x - y
}
