package internal

import (
	"fmt"
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type Error struct {
	typ    ErrorType
	detail string
	name   option.Option[string]
	lineno uint
	span   option.Option[Span]
	source error
}

func NewError(typ ErrorType, detail string) *Error {
	return &Error{typ: typ, detail: detail}
}

type ErrorType int

const (
	// A non primitive value was encountered where one was expected.
	NonPrimitive ErrorType = 1
	// A value is not valid for a key in a map.
	NonKey ErrorType = 2
	// An invalid operation was attempted.
	InvalidOperation ErrorType = 3
	// The template has a syntax error
	SyntaxError ErrorType = 4
	// A template was not found.
	TemplateNotFound ErrorType = 5
	// Too many arguments were passed to a function.
	TooManyArguments ErrorType = 6
	// A expected argument was missing
	MissingArgument ErrorType = 7
	// A filter is unknown
	UnknownFilter ErrorType = 8
	// A test is unknown
	UnknownTest ErrorType = 9
	// A function is unknown
	UnknownFunction ErrorType = 10
	// Un unknown method was called
	UnknownMethod ErrorType = 11
	// A bad escape sequence in a string was encountered.
	BadEscape ErrorType = 12
	// An operation on an undefined value was attempted.
	UndefinedError ErrorType = 13
	// Not able to serialize this
	BadSerialization ErrorType = 14
	// Not able to deserialize this
	CannotDeserialize ErrorType = 15
	// An error happened in an include.
	BadInclude ErrorType = 16
	// An error happened in a super block.
	EvalBlock ErrorType = 17
	// Unable to unpack a
	CannotUnpack ErrorType = 18
	// Failed writing output.
	WriteFailure ErrorType = 19
	// Engine ran out of fuel
	OutOfFuel ErrorType = 20
	// Error creating aho-corasick delimiters
	InvalidDelimiter ErrorType = 21
	// An unknown block was called
	UnknownBlock ErrorType = 22
)

func (k ErrorType) String() string {
	switch k {
	case NonPrimitive:
		return "not a primitive"
	case NonKey:
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
	case WriteFailure:
		return "failed to write output"
	case CannotDeserialize:
		return "cannot deserialize"
	case OutOfFuel:
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

func (e *Error) Line() option.Option[uint] {
	if e.lineno > 0 {
		return option.Some(e.lineno)
	}
	return option.None[uint]()
}

func (e *Error) SetFilenameAndLine(filename string, lineno uint) {
	e.name = option.Some(filename)
	e.lineno = lineno
}

func (e *Error) SetFilenameAndSpan(filename string, spn Span) {
	e.name = option.Some(filename)
	e.span = option.Some(spn)
	e.lineno = uint(spn.StartLine)
}

func (e *Error) WithSource(err error) *Error {
	e.source = err
	return e
}
