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
	kind   ErrorKind
	detail string
	name   option.Option[string]
	lineno uint
	span   option.Option[span]
	source error
}

// NewError creates a new error with kind and detail.
func NewError(kind ErrorKind, detail string) *Error {
	return &Error{kind: kind, detail: detail}
}

// ErrorKind describes the error kind.
type ErrorKind int

const (
	// NonPrimitive represents a non primitive value was encountered where one was expected.
	NonPrimitive ErrorKind = 1
	// A value is not valid for a key in a map.
	nonKey ErrorKind = 2
	// InvalidOperation is an invalid operation was attempted.
	InvalidOperation ErrorKind = 3
	// SyntaxError represents the template has a syntax error
	SyntaxError ErrorKind = 4
	// TemplateNotFound represents a template was not found.
	TemplateNotFound ErrorKind = 5
	// TooManyArguments represents too many arguments were passed to a function.
	TooManyArguments ErrorKind = 6
	// MissingArgument represents a expected argument was missing
	MissingArgument ErrorKind = 7
	// UnknownFilter represents a filter is unknown
	UnknownFilter ErrorKind = 8
	// UnknownTest represents A test is unknown
	UnknownTest ErrorKind = 9
	// UnknownFunction represents a function is unknown
	UnknownFunction ErrorKind = 10
	// UnknownMethod represents an unknown method was called
	UnknownMethod ErrorKind = 11
	// BadEscape represents a bad escape sequence in a string was encountered.
	BadEscape ErrorKind = 12
	// UndefinedError represents an operation on an undefined value was attempted.
	UndefinedError ErrorKind = 13
	// BadSerialization represents not able to serialize this
	BadSerialization ErrorKind = 14
	// Not able to deserialize this
	cannotDeserialize ErrorKind = 15
	// BadInclude represents an error happened in an include.
	BadInclude ErrorKind = 16
	// EvalBlock represents an error happened in a super block.
	EvalBlock ErrorKind = 17
	// CannotUnpack represents unable to unpack a value.
	CannotUnpack ErrorKind = 18
	// Failed writing output.
	writeFailure ErrorKind = 19
	// Engine ran out of fuel
	outOfFuel ErrorKind = 20
	// InvalidDelimiter represents error creating aho-corasick delimiters
	InvalidDelimiter ErrorKind = 21
	// UnknownBlock represents an unknown block was called
	UnknownBlock ErrorKind = 22
)

func (k ErrorKind) String() string {
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
	b.WriteString(e.kind.String())
	if e.detail != "" {
		b.WriteString(": ")
		b.WriteString(e.detail)
	}
	if e.name.IsSome() {
		fmt.Fprintf(&b, " (in %s:%d)", e.name.Unwrap(), e.lineno)
	}
	return b.String()
}

// Kind returns the error kind
func (e *Error) Kind() ErrorKind { return e.kind }

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
