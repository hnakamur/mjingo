package rustfmt

import (
	"fmt"
	"io"
)

const (
	DisplayVerb = '\U0001f5a5'
	DebugVerb   = '\U0001f41e'

	DisplayString = "%\U0001f5a5"
	DebugString   = "%\U0001f41e"

	PrettyFlag = '#'

	DisplayPrettyString = "%#\U0001f5a5"
	DebugPrettyString   = "%#\U0001f41e"
)

// RustFormatter is a marker interface to indicate a type supports
// FmtDisplayVerb and FmtDebugVerb in fmt.Formatter.
type Formatter interface {
	fmt.Formatter
	SupportRustFormat()
}

type DebugStruct struct {
	name        string
	indentLevel uint
	fields      []debugStructField
}

type debugStructField struct {
	name  string
	value any
}

func NewDebugStruct(name string) *DebugStruct {
	return &DebugStruct{name: name}
}

func (s *DebugStruct) Field(name string, value any) *DebugStruct {
	s.fields = append(s.fields, debugStructField{name: name, value: value})
	return s
}

func (s DebugStruct) Format(f fmt.State, verb rune) {
	switch verb {
	case DisplayVerb:
		io.WriteString(f, s.name)
		io.WriteString(f, " { ")
		for i, field := range s.fields {
			if i > 0 {
				io.WriteString(f, ", ")
			}
			io.WriteString(f, field.name)
			io.WriteString(f, ": ")
			FormatAnyValue(f, verb, field.value)
		}
		io.WriteString(f, " }")
	case DebugVerb:
		io.WriteString(f, s.name)
		io.WriteString(f, " {\n")
		w := NewPadFormatAdapter(f, true)
		for _, field := range s.fields {
			io.WriteString(w, field.name)
			io.WriteString(w, ": ")
			FormatAnyValue(w, verb, field.value)
			io.WriteString(w, ",\n")
		}
		io.WriteString(f, "}")
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods DebugStruct
		type debugStruct hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), debugStruct(s))
	}
}

func FormatAnyValue(f fmt.State, verb rune, val any) {
	if rf, ok := val.(Formatter); ok {
		rf.Format(f, verb)
	} else {
		fmt.Fprintf(f, "%v", val)
	}
}

type DebugList struct {
	entries []any
	inner   debugElem
}

func NewDebugList(entries []any) *DebugList {
	return &DebugList{entries: entries}
}

func (d *DebugList) Format(f fmt.State, verb rune) {
	switch verb {
	case DisplayVerb, DebugVerb:
		io.WriteString(f, "[")
		for _, entry := range d.entries {
			d.inner.SetData(entry).Format(f, verb)
		}
		io.WriteString(f, "]")
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods DebugList
		type debugList hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), debugList(*d))
	}
}

type debugElem struct {
	hasFields bool
	data      any
}

func (d *debugElem) SetData(data any) *debugElem {
	d.data = data
	return d
}

func (d *debugElem) Format(f fmt.State, verb rune) {
	switch verb {
	case DebugVerb:
		if !d.hasFields {
			io.WriteString(f, "\n")
		}
		w := NewPadFormatAdapter(f, true)
		FormatAnyValue(w, verb, d.data)
		io.WriteString(w, ",\n")
		d.hasFields = true
	case DisplayVerb:
		if d.hasFields {
			io.WriteString(f, ", ")
			d.hasFields = true
		}
		FormatAnyValue(f, verb, d.data)
		d.hasFields = true
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods debugElem
		type debugInner hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), debugInner(*d))
	}
}
