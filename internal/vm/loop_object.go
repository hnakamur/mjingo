package vm

import (
	"fmt"
	"slices"

	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/value"
)

type LoopObject struct {
	len              uint
	idx              uint
	depth            uint
	valueTriple      [3]option.Option[value.Value]
	lastChangedValue []value.Value
}

var _ = (value.Object)((*LoopObject)(nil))
var _ = (CallMethoder)((*LoopObject)(nil))
var _ = (value.StructObject)((*LoopObject)(nil))

func (l *LoopObject) Kind() value.ObjectKind { return value.ObjectKindStruct }

func (l *LoopObject) CallMethod(state *State, name string, args []value.Value) (value.Value, error) {
	switch name {
	case "changed":
		if slices.Equal(l.lastChangedValue, args) {
			return value.ValueFromBool(false), nil
		}
		l.lastChangedValue = make([]value.Value, len(args))
		for i, arg := range args {
			l.lastChangedValue[i] = arg
		}
		return value.ValueFromBool(true), nil
	case "cycle":
		idx := l.idx % uint(len(args))
		if idx < uint(len(args)) {
			return args[idx].Clone(), nil
		}
		return value.Undefined, nil
	}
	return nil, common.NewError(common.UnknownMethod, fmt.Sprintf("loop object has no method named %s", name))
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

func (l *LoopObject) GetField(name string) option.Option[value.Value] {
	idx := l.idx
	// if we never iterated, then all attributes are undefined.
	// this can happen in some rare circumstances where the engine
	// did not manage to iterate
	if idx == ^uint(0) {
		return option.Some[value.Value](value.Undefined)
	}
	switch name {
	case "index0":
		return option.Some[value.Value](value.ValueFromI64(int64(idx)))
	case "index":
		return option.Some[value.Value](value.ValueFromI64(int64(idx + 1)))
	case "len":
		return option.Some[value.Value](value.ValueFromI64(int64(l.len)))
	case "revindex":
		// TODO: saturating_sub
		return option.Some[value.Value](value.ValueFromI64(int64(l.len - idx)))
	case "revindex0":
		// TODO: saturating_sub
		return option.Some[value.Value](value.ValueFromI64(int64(l.len - idx - 1)))
	case "first":
		return option.Some[value.Value](value.ValueFromBool(idx == 0))
	case "last":
		return option.Some[value.Value](value.ValueFromBool(l.len == 0 || idx == l.len-1))
	case "depth":
		return option.Some[value.Value](value.ValueFromI64(int64(l.depth + 1)))
	case "depth0":
		return option.Some[value.Value](value.ValueFromI64(int64(l.depth)))
	case "previtem":
		return option.Some[value.Value](l.valueTriple[0].UnwrapOr(value.Undefined).Clone())
	case "nextitem":
		return option.Some[value.Value](l.valueTriple[2].UnwrapOr(value.Undefined).Clone())
	}
	return option.None[value.Value]()
}
