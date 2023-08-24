package internal

import (
	"fmt"
	"slices"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type LoopObject struct {
	len              uint
	idx              uint
	depth            uint
	valueTriple      [3]option.Option[Value]
	lastChangedValue []Value
}

var _ = (Object)((*LoopObject)(nil))
var _ = (CallMethoder)((*LoopObject)(nil))
var _ = (StructObject)((*LoopObject)(nil))

func (l *LoopObject) Kind() ObjectKind { return ObjectKindStruct }

func (l *LoopObject) Clone() *LoopObject { return &*l }

func (l *LoopObject) CallMethod(state *State, name string, args []Value) (Value, error) {
	switch name {
	case "changed":
		if slices.Equal(l.lastChangedValue, args) {
			return ValueFromBool(false), nil
		}
		l.lastChangedValue = args
		return ValueFromBool(true), nil
	case "cycle":
		idx := l.idx % uint(len(args))
		if idx < uint(len(args)) {
			return args[idx].Clone(), nil
		}
		return Undefined, nil
	}
	return nil, NewError(UnknownMethod, fmt.Sprintf("loop object has no method named %s", name))
}

func (*LoopObject) StaticFields() option.Option[[]string] {
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

func (*LoopObject) Fields() []string { return nil }

func (l *LoopObject) GetField(name string) option.Option[Value] {
	idx := l.idx
	// if we never iterated, then all attributes are undefined.
	// this can happen in some rare circumstances where the engine
	// did not manage to iterate
	if idx == ^uint(0) {
		return option.Some[Value](Undefined)
	}
	switch name {
	case "index0":
		return option.Some[Value](ValueFromI64(int64(idx)))
	case "index":
		return option.Some[Value](ValueFromI64(int64(idx + 1)))
	case "len":
		return option.Some[Value](ValueFromI64(int64(l.len)))
	case "revindex":
		// TODO: saturating_sub
		return option.Some[Value](ValueFromI64(int64(l.len - idx)))
	case "revindex0":
		// TODO: saturating_sub
		return option.Some[Value](ValueFromI64(int64(l.len - idx - 1)))
	case "first":
		return option.Some[Value](ValueFromBool(idx == 0))
	case "last":
		return option.Some[Value](ValueFromBool(l.len == 0 || idx == l.len-1))
	case "depth":
		return option.Some[Value](ValueFromI64(int64(l.depth + 1)))
	case "depth0":
		return option.Some[Value](ValueFromI64(int64(l.depth)))
	case "previtem":
		return option.Some[Value](l.valueTriple[0].UnwrapOr(Undefined).Clone())
	case "nextitem":
		return option.Some[Value](l.valueTriple[2].UnwrapOr(Undefined).Clone())
	}
	return option.None[Value]()
}
