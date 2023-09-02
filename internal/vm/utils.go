package vm

import (
	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/value"
)

// Controls the autoescaping behavior.
type AutoEscape interface {
	typ() AutoEscapeType
	IsNone() bool
}

type AutoEscapeNone struct{}
type AutoEscapeHTML struct{}
type AutoEscapeJSON struct{}
type AutoEscapeCustom struct{ Name string }

func (AutoEscapeNone) typ() AutoEscapeType   { return autoEscapeTypeNone }
func (AutoEscapeHTML) typ() AutoEscapeType   { return autoEscapeTypeHTML }
func (AutoEscapeJSON) typ() AutoEscapeType   { return autoEscapeTypeJSON }
func (AutoEscapeCustom) typ() AutoEscapeType { return autoEscapeTypeCustom }

func (AutoEscapeNone) IsNone() bool   { return true }
func (AutoEscapeHTML) IsNone() bool   { return false }
func (AutoEscapeJSON) IsNone() bool   { return false }
func (AutoEscapeCustom) IsNone() bool { return false }

var _ = (AutoEscape)(AutoEscapeNone{})
var _ = (AutoEscape)(AutoEscapeHTML{})
var _ = (AutoEscape)(AutoEscapeJSON{})
var _ = (AutoEscape)(AutoEscapeCustom{})

type AutoEscapeType uint

const (
	autoEscapeTypeNone AutoEscapeType = iota
	autoEscapeTypeHTML
	autoEscapeTypeJSON
	autoEscapeTypeCustom
)

type UndefinedBehavior uint

const (
	// The default, somewhat lenient undefined behavior.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorLenient UndefinedBehavior = iota

	// Like `Lenient`, but also allows chaining of undefined lookups.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** allowed (returns [`undefined`](Value::UNDEFINED))
	UndefinedBehaviorChainable

	// Complains very quickly about undefined values.
	//
	// * **printing:** fails
	// * **iteration:** fails
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorStrict

	UndefinedBehaviorDefault = UndefinedBehaviorLenient
)

func (b UndefinedBehavior) HandleUndefined(parentWasUndefined bool) (value.Value, error) {
	switch {
	case (b == UndefinedBehaviorLenient && !parentWasUndefined) || b == UndefinedBehaviorChainable:
		return value.Undefined, nil
	case (b == UndefinedBehaviorLenient && parentWasUndefined) || b == UndefinedBehaviorStrict:
		return nil, common.NewError(common.UndefinedError, "")
	default:
		panic("unreachable")
	}
}

// Tries to iterate over a valu.Value while handling the undefined valu.Value.
//
// If the valu.Value is undefined, then iteration fails if the behavior is set to strict,
// otherwise it succeeds with an empty iteration.  This is also internally used in the
// engine to convert values to lists.
func (b UndefinedBehavior) TryIter(val value.Value) (value.Iterator, error) {
	if err := b.AssertIterable(val); err != nil {
		return value.Iterator{}, err
	}
	iter, err := val.TryIter()
	if err != nil {
		return value.Iterator{}, err
	}
	return iter, nil
}

// Are we strict on iteration?
func (b UndefinedBehavior) AssertIterable(val value.Value) error {
	if b == UndefinedBehaviorStrict && val.IsUndefined() {
		return common.NewError(common.UndefinedError, "")
	}
	return nil
}
