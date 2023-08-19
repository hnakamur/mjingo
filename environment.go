package mjingo

import (
	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/vm"
)

type Environment = vm.Environment
type Template = vm.Template

type AutoEscape = compiler.AutoEscape
type AutoEscapeNone = compiler.AutoEscapeNone
type AutoEscapeHTML = compiler.AutoEscapeHTML
type AutoEscapeJSON = compiler.AutoEscapeJSON
type AutoEscapeCustom = compiler.AutoEscapeCustom

type UndefinedBehavior = compiler.UndefinedBehavior

const (
	// The default, somewhat lenient undefined behavior.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorLenient = compiler.UndefinedBehaviorLenient

	// Like `Lenient`, but also allows chaining of undefined lookups.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** allowed (returns [`undefined`](Value::UNDEFINED))
	UndefinedBehaviorChainable = compiler.UndefinedBehaviorChainable

	// Complains very quickly about undefined values.
	//
	// * **printing:** fails
	// * **iteration:** fails
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorStrict = compiler.UndefinedBehaviorStrict

	UndefinedBehaviorDefault = UndefinedBehaviorLenient
)

type Error = internal.Error

func NewEnvironment() *Environment {
	return vm.NewEnvironment()
}
