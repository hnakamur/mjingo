package mjingo

import (
	"fmt"
	"strings"

	"github.com/hnakamur/mjingo/option"
)

// Error represents template errors.
//
// If debug mode is enabled a template error contains additional debug
// information that can be displayed by formatting an error with the
// alternative formatting (DebugString()).  That information
// is also shown for the [DebugString] display where the extended information
// is hidden when the alternative formatting is used.
//
// Since mjingo takes advantage of chained errors it's recommended
// to render the entire chain to better understand the causes.
type Error struct {
	typ    ErrorType
	detail string
	name   option.Option[string]
	lineno uint
	span   option.Option[span]
	source error
}

// NewError creates a new error with kind and detail.
func NewError(typ ErrorType, detail string) *Error {
	return &Error{typ: typ, detail: detail}
}

// ErrorType describes the error kind.
type ErrorType int

const (
	// NonPrimitive represents a non primitive value was encountered where one was expected.
	NonPrimitive ErrorType = 1
	// A value is not valid for a key in a map.
	nonKey ErrorType = 2
	// InvalidOperation is an invalid operation was attempted.
	InvalidOperation ErrorType = 3
	// SyntaxError represents the template has a syntax error
	SyntaxError ErrorType = 4
	// TemplateNotFound represents a template was not found.
	TemplateNotFound ErrorType = 5
	// TooManyArguments represents too many arguments were passed to a function.
	TooManyArguments ErrorType = 6
	// MissingArgument represents a expected argument was missing
	MissingArgument ErrorType = 7
	// UnknownFilter represents a filter is unknown
	UnknownFilter ErrorType = 8
	// UnknownTest represents A test is unknown
	UnknownTest ErrorType = 9
	// UnknownFunction represents a function is unknown
	UnknownFunction ErrorType = 10
	// UnknownMethod represents an unknown method was called
	UnknownMethod ErrorType = 11
	// BadEscape represents a bad escape sequence in a string was encountered.
	BadEscape ErrorType = 12
	// UndefinedError represents an operation on an undefined value was attempted.
	UndefinedError ErrorType = 13
	// BadSerialization represents not able to serialize this
	BadSerialization ErrorType = 14
	// Not able to deserialize this
	cannotDeserialize ErrorType = 15
	// BadInclude represents an error happened in an include.
	BadInclude ErrorType = 16
	// EvalBlock represents an error happened in a super block.
	EvalBlock ErrorType = 17
	// CannotUnpack represents unable to unpack a value.
	CannotUnpack ErrorType = 18
	// Failed writing output.
	writeFailure ErrorType = 19
	// Engine ran out of fuel
	outOfFuel ErrorType = 20
	// InvalidDelimiter represents error creating aho-corasick delimiters
	InvalidDelimiter ErrorType = 21
	// UnknownBlock represents an unknown block was called
	UnknownBlock ErrorType = 22
)

func (k ErrorType) String() string {
	switch k {
	case NonPrimitive:
		return "not a primitive"
	case nonKey:
		return "not a key type"
	case InvalidOperation:
		return "invalid operation"
	case SyntaxError:
		return "syntax error"
	case TemplateNotFound:
		return "template not found"
	case TooManyArguments:
		return "too many arguments"
	case MissingArgument:
		return "missing argument"
	case UnknownFilter:
		return "unknown filter"
	case UnknownFunction:
		return "unknown function"
	case UnknownTest:
		return "unknown test"
	case UnknownMethod:
		return "unknown method"
	case BadEscape:
		return "bad string escape"
	case UndefinedError:
		return "undefined value"
	case BadSerialization:
		return "could not serialize to value"
	case BadInclude:
		return "could not render include"
	case EvalBlock:
		return "could not render block"
	case CannotUnpack:
		return "cannot unpack"
	case writeFailure:
		return "failed to write output"
	case cannotDeserialize:
		return "cannot deserialize"
	case outOfFuel:
		return "engine ran out of fuel"
	case InvalidDelimiter:
		return "invalid custom delimiters"
	case UnknownBlock:
		return "unknown block"
	default:
		panic("unknown error kind")
	}
}

func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(e.typ.String())
	if e.detail != "" {
		b.WriteString(": ")
		b.WriteString(e.detail)
	}
	if e.name.IsSome() {
		fmt.Fprintf(&b, " (in %s:%d)", e.name.Unwrap(), e.lineno)
	}
	return b.String()
}

// Type returns the error kind
func (e *Error) Type() ErrorType { return e.typ }

func (e *Error) line() option.Option[uint] {
	if e.lineno > 0 {
		return option.Some(e.lineno)
	}
	return option.None[uint]()
}

func (e *Error) setFilenameAndLine(filename string, lineno uint) {
	e.name = option.Some(filename)
	e.lineno = lineno
}

func (e *Error) setFilenameAndSpan(filename string, spn span) {
	e.name = option.Some(filename)
	e.span = option.Some(spn)
	e.lineno = uint(spn.StartLine)
}

func (e *Error) withSource(err error) *Error {
	e.source = err
	return e
}
