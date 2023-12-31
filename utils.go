package mjingo

// AutoEscape controls the autoescaping behavior.
type AutoEscape interface {
	typ() autoEscapeType
	isNone() bool
}

// AutoEscapeNone do not apply auto escaping.
var AutoEscapeNone AutoEscape

// AutoEscapeHTML use HTML auto escaping rules.
//
// Any value will be converted into a string and the following characters
// will be escaped in ways compatible to XML and HTML: `<`, `>`, `&`, `"`,
// `'`, and `/`.
var AutoEscapeHTML AutoEscape

// AutoEscapeJSON use escaping rules suitable for JSON/JavaScript or YAML.
//
// Any value effectively ends up being serialized to JSON upon printing.  The
// serialized values will be compatible with JavaScript and YAML as well.
var AutoEscapeJSON AutoEscape

func init() {
	AutoEscapeNone = autoEscapeNone{}
	AutoEscapeHTML = autoEscapeHTML{}
	AutoEscapeJSON = autoEscapeJSON{}
}

type autoEscapeNone struct{}
type autoEscapeHTML struct{}
type autoEscapeJSON struct{}
type autoEscapeCustom struct{ name string }

func (autoEscapeNone) typ() autoEscapeType   { return autoEscapeTypeNone }
func (autoEscapeHTML) typ() autoEscapeType   { return autoEscapeTypeHTML }
func (autoEscapeJSON) typ() autoEscapeType   { return autoEscapeTypeJSON }
func (autoEscapeCustom) typ() autoEscapeType { return autoEscapeTypeCustom }

func (autoEscapeNone) isNone() bool   { return true }
func (autoEscapeHTML) isNone() bool   { return false }
func (autoEscapeJSON) isNone() bool   { return false }
func (autoEscapeCustom) isNone() bool { return false }

var _ = (AutoEscape)(autoEscapeNone{})
var _ = (AutoEscape)(autoEscapeHTML{})
var _ = (AutoEscape)(autoEscapeJSON{})
var _ = (AutoEscape)(autoEscapeCustom{})

type autoEscapeType uint

const (
	autoEscapeTypeNone autoEscapeType = iota
	autoEscapeTypeHTML
	autoEscapeTypeJSON
	autoEscapeTypeCustom
)

// UndefinedBehavior defines the behavior of undefined values in the engine.
//
// At present there are three types of behaviors available which mirror the behaviors
// that Jinja2 provides out of the box.
type UndefinedBehavior uint

const (
	// UndefinedBehaviorLenient is the default, somewhat lenient undefined behavior.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorLenient UndefinedBehavior = iota

	// UndefinedBehaviorChainable is like `Lenient`, but also allows chaining of undefined lookups.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** allowed (returns [`undefined`](Value::UNDEFINED))
	UndefinedBehaviorChainable

	// UndefinedBehaviorStrict complains very quickly about undefined values.
	//
	// * **printing:** fails
	// * **iteration:** fails
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorStrict

	UndefinedBehaviorDefault = UndefinedBehaviorLenient
)

func (b UndefinedBehavior) handleUndefined(parentWasUndefined bool) (Value, error) {
	switch {
	case (b == UndefinedBehaviorLenient && !parentWasUndefined) || b == UndefinedBehaviorChainable:
		return Undefined, nil
	case (b == UndefinedBehaviorLenient && parentWasUndefined) || b == UndefinedBehaviorStrict:
		return Value{}, NewError(UndefinedError, "")
	default:
		panic("unreachable")
	}
}

// Tries to iterate over a Value while handling the undefined Value.
//
// If the Value is undefined, then iteration fails if the behavior is set to strict,
// otherwise it succeeds with an empty iteration.  This is also internally used in the
// engine to convert values to lists.
func (b UndefinedBehavior) tryIter(val Value) (iterator, error) {
	if err := b.assertIterable(val); err != nil {
		return iterator{}, err
	}
	iter, err := val.tryIter()
	if err != nil {
		return iterator{}, err
	}
	return iter, nil
}

// Are we strict on iteration?
func (b UndefinedBehavior) assertIterable(val Value) error {
	if b == UndefinedBehaviorStrict && val.isUndefined() {
		return NewError(UndefinedError, "")
	}
	return nil
}
