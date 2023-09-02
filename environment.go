package mjingo

import (
	"github.com/hnakamur/mjingo/internal/common"
	"github.com/hnakamur/mjingo/internal/vm"
)

type Environment = vm.Environment
type Template = vm.Template

type AutoEscape = vm.AutoEscape
type AutoEscapeNone = vm.AutoEscapeNone
type AutoEscapeHTML = vm.AutoEscapeHTML
type AutoEscapeJSON = vm.AutoEscapeJSON
type AutoEscapeCustom = vm.AutoEscapeCustom

type UndefinedBehavior = vm.UndefinedBehavior

const (
	// The default, somewhat lenient undefined behavior.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorLenient = vm.UndefinedBehaviorLenient

	// Like `Lenient`, but also allows chaining of undefined lookups.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** allowed (returns [`undefined`](Value::UNDEFINED))
	UndefinedBehaviorChainable = vm.UndefinedBehaviorChainable

	// Complains very quickly about undefined values.
	//
	// * **printing:** fails
	// * **iteration:** fails
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorStrict = vm.UndefinedBehaviorStrict

	UndefinedBehaviorDefault = UndefinedBehaviorLenient
)

type Error = common.Error

func NewEnvironment() *Environment {
	return vm.NewEnvironment()
}
