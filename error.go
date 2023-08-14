package mjingo

import (
	"fmt"
	"strings"
)

type Error struct {
	kind   ErrorKind
	detail option[string]
	name   option[string]
	lineno uint
	span   option[span]
}

type ErrorKind int

const (
	// A non primitive value was encountered where one was expected.
	NonPrimitive ErrorKind = 1
	// A value is not valid for a key in a map.
	NonKey ErrorKind = 2
	// An invalid operation was attempted.
	InvalidOperation ErrorKind = 3
	// The template has a syntax error
	SyntaxError ErrorKind = 4
	// A template was not found.
	TemplateNotFound ErrorKind = 5
	// Too many arguments were passed to a function.
	TooManyArguments ErrorKind = 6
	// A expected argument was missing
	MissingArgument ErrorKind = 7
	// A filter is unknown
	UnknownFilter ErrorKind = 8
	// A test is unknown
	UnknownTest ErrorKind = 9
	// A function is unknown
	UnknownFunction ErrorKind = 10
	// Un unknown method was called
	UnknownMethod ErrorKind = 11
	// A bad escape sequence in a string was encountered.
	BadEscape ErrorKind = 12
	// An operation on an undefined value was attempted.
	UndefinedError ErrorKind = 13
	// Not able to serialize this value.
	BadSerialization ErrorKind = 14
	// Not able to deserialize this value.
	CannotDeserialize ErrorKind = 15
	// An error happened in an include.
	BadInclude ErrorKind = 16
	// An error happened in a super block.
	EvalBlock ErrorKind = 17
	// Unable to unpack a value.
	CannotUnpack ErrorKind = 18
	// Failed writing output.
	WriteFailure ErrorKind = 19
	// Engine ran out of fuel
	OutOfFuel ErrorKind = 20
	// Error creating aho-corasick delimiters
	InvalidDelimiter ErrorKind = 21
	// An unknown block was called
	UnknownBlock ErrorKind = 22
)

func (k ErrorKind) String() string {
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
	b.WriteString(e.kind.String())
	if e.detail.valid {
		b.WriteString(": ")
		b.WriteString(e.detail.data)
	}
	if e.name.valid {
		fmt.Fprintf(&b, " (in %s:%d)", e.name.data, e.lineno)
	}
	return b.String()
}

func (e *Error) line() option[uint] {
	if e.lineno > 0 {
		return option[uint]{valid: true, data: e.lineno}
	} else {
		return option[uint]{}
	}
}

func (e *Error) setFilenameAndLine(filename string, lineno uint) {
	e.name = option[string]{valid: true, data: filename}
	e.lineno = lineno
}

func (e *Error) setFilenameAndSpan(filename string, spn span) {
	e.name = option[string]{valid: true, data: filename}
	e.span = option[span]{valid: true, data: spn}
	e.lineno = uint(spn.startLine)
}
