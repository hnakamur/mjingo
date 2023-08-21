package mjingo

import (
	"github.com/hnakamur/mjingo/internal"
)

type Environment = internal.Environment
type Template = internal.Template

type AutoEscape = internal.AutoEscape
type AutoEscapeNone = internal.AutoEscapeNone
type AutoEscapeHTML = internal.AutoEscapeHTML
type AutoEscapeJSON = internal.AutoEscapeJSON
type AutoEscapeCustom = internal.AutoEscapeCustom

type UndefinedBehavior = internal.UndefinedBehavior

const (
	// The default, somewhat lenient undefined behavior.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorLenient = internal.UndefinedBehaviorLenient

	// Like `Lenient`, but also allows chaining of undefined lookups.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** allowed (returns [`undefined`](Value::UNDEFINED))
	UndefinedBehaviorChainable = internal.UndefinedBehaviorChainable

	// Complains very quickly about undefined values.
	//
	// * **printing:** fails
	// * **iteration:** fails
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorStrict = internal.UndefinedBehaviorStrict

	UndefinedBehaviorDefault = UndefinedBehaviorLenient
)

type Error = internal.Error

func NewEnvironment() *Environment {
	return internal.NewEnvironment()
}
