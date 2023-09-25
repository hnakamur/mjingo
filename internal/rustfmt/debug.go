package rustfmt

import (
	"fmt"
	"io"

	"github.com/hnakamur/mjingo/internal/datast/indexmap"
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
	SupportsCustomVerb(verb rune) bool
}

type DebugStruct struct {
	name        string
	indentLevel uint
	fields      []debugStructField
}

var _ Formatter = DebugStruct{}

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

// SupportRustFormat implements Formatter.
func (DebugStruct) SupportsCustomVerb(verb rune) bool { return verb == DebugVerb }

func (s DebugStruct) Format(f fmt.State, verb rune) {
	switch verb {
	case DebugVerb:
		if f.Flag(PrettyFlag) {
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
		} else {
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
		}
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

var _ Formatter = (*DebugList)(nil)

func NewDebugList(entries []any) *DebugList {
	return &DebugList{entries: entries}
}

// SupportRustFormat implements Formatter.
func (DebugList) SupportsCustomVerb(verb rune) bool { return verb == DebugVerb }

func (d *DebugList) Format(f fmt.State, verb rune) {
	switch verb {
	case DebugVerb:
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

var _ Formatter = (*debugElem)(nil)

func (d *debugElem) SetData(data any) *debugElem {
	d.data = data
	return d
}

// SupportRustFormat implements Formatter.
func (*debugElem) SupportsCustomVerb(verb rune) bool { return verb == DebugVerb }

func (d *debugElem) Format(f fmt.State, verb rune) {
	switch verb {
	case DebugVerb:
		if f.Flag(PrettyFlag) {
			if !d.hasFields {
				io.WriteString(f, "\n")
			}
			w := NewPadFormatAdapter(f, true)
			FormatAnyValue(w, verb, d.data)
			io.WriteString(w, ",\n")
			d.hasFields = true
		} else {
			if d.hasFields {
				io.WriteString(f, ", ")
				d.hasFields = true
			}
			FormatAnyValue(f, verb, d.data)
			d.hasFields = true
		}
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods debugElem
		type debugElem hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), debugElem(*d))
	}
}

type DebugMap[K indexmap.HashEqualer, V any] struct {
	m indexmap.Map[K, V]
}

func NewDebugMap[K indexmap.HashEqualer, V any](m indexmap.Map[K, V]) *DebugMap[K, V] {
	return &DebugMap[K, V]{m: m}
}

// SupportRustFormat implements Formatter.
func (*DebugMap[K, V]) SupportsCustomVerb(verb rune) bool { return verb == DebugVerb }

func (m *DebugMap[K, V]) Format(f fmt.State, verb rune) {
	switch verb {
	case DebugVerb:
		if f.Flag(PrettyFlag) {
			io.WriteString(f, "{")
			w := NewPadFormatAdapter(f, true)
			for i := uint(0); i < m.m.Len(); i++ {
				if i == 0 {
					io.WriteString(f, "\n")
				}
				e, _ := m.m.EntryAt(i)
				FormatValue(w, verb, e.Key, "%q")
				io.WriteString(w, ": ")
				FormatAnyValue(w, verb, e.Value)
				io.WriteString(w, ",\n")
			}
			io.WriteString(f, "}")
		} else {
			io.WriteString(f, "{")
			for i := uint(0); i < m.m.Len(); i++ {
				if i > 0 {
					io.WriteString(f, ", ")
				}
				e, _ := m.m.EntryAt(i)
				FormatValue(f, verb, e.Key, "%q")
				io.WriteString(f, ": ")
				FormatAnyValue(f, verb, e.Value)
			}
			io.WriteString(f, "}")
		}
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods DebugMap[K, V]
		type DebugMap[K indexmap.HashEqualer, V any] hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), DebugMap[K, V](*m))
	}
}

func FormatValue(f fmt.State, verb rune, val any, fallbackFormat string) {
	if rf, ok := val.(Formatter); ok {
		rf.Format(f, verb)
	} else {
		fmt.Fprintf(f, fallbackFormat, val)
	}
}
