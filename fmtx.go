package mjingo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type debugStruct struct {
	name        string
	indentLevel uint
	fields      []debugStructField
}

type debugStructField struct {
	name  string
	value any
}

func newDebugStruct(name string) *debugStruct {
	return &debugStruct{name: name}
}

func (s *debugStruct) field(name string, value any) *debugStruct {
	s.fields = append(s.fields, debugStructField{name: name, value: value})
	return s
}

func (s debugStruct) Format(f fmt.State, verb rune) {
	switch verb {
	case 's':
		io.WriteString(f, s.name)
		io.WriteString(f, " { ")
		for i, field := range s.fields {
			if i > 0 {
				io.WriteString(f, ", ")
			}
			io.WriteString(f, field.name)
			io.WriteString(f, ": ")
			fmt.Fprintf(f, "%v", field.value)
		}
		io.WriteString(f, " }")
	case 'q':
		io.WriteString(f, s.name)
		io.WriteString(f, " {\n")
		w := newPadAdapter(f, true)
		for _, field := range s.fields {
			io.WriteString(w, field.name)
			io.WriteString(w, ": ")
			fmt.Fprintf(w, "%v", field.value)
			io.WriteString(w, ",\n")
		}
		io.WriteString(f, "}")
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods debugStruct
		type debugStruct hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), debugStruct(s))
	}
}

type debugList struct {
	entries []any
	inner   debugElem
}

func newDebugList(entries []any) *debugList {
	return &debugList{entries: entries}
}

func (d *debugList) Format(f fmt.State, verb rune) {
	switch verb {
	case 'q', 's':
		io.WriteString(f, "[")
		for _, entry := range d.entries {
			d.inner.SetData(entry).Format(f, verb)
		}
		io.WriteString(f, "]")
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods debugList
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
	case 'q':
		if !d.hasFields {
			io.WriteString(f, "\n")
		}
		w := newPadAdapter(f, true)
		fmt.Fprintf(w, "%v", d.data)
		io.WriteString(w, ",\n")
		d.hasFields = true
	case 's':
		if d.hasFields {
			io.WriteString(f, ", ")
			d.hasFields = true
		}
		fmt.Fprintf(f, "%v", d.data)
		d.hasFields = true
	default:
		// https://github.com/golang/go/issues/51195#issuecomment-1563538796
		type hideMethods debugElem
		type debugInner hideMethods
		fmt.Fprintf(f, fmt.FormatString(f, verb), debugInner(*d))
	}
}

type padAdapter struct {
	inner padAdapterInner
}

var _ io.Writer = (*padAdapter)(nil)

type padAdapterInner struct {
	writer io.Writer
	state  padAdapterState
}

type padAdapterState struct {
	onNewline bool
}

func newPadAdapter(w io.Writer, onNewline bool) *padAdapter {
	return &padAdapter{
		inner: padAdapterInner{
			writer: w,
			state: padAdapterState{
				onNewline: onNewline,
			},
		},
	}
}

func (a *padAdapter) Write(p []byte) (n int, err error) { return a.inner.Write(p) }

func (a *padAdapterInner) Write(p []byte) (n int, err error) {
	br := bufio.NewReader(bytes.NewReader(p))
	for {
		var line []byte
		line, err = br.ReadBytes('\n')
		if err == io.EOF && len(line) == 0 {
			return n, nil
		}
		if a.state.onNewline {
			n2, err2 := io.WriteString(a.writer, "    ")
			n += n2
			if err2 != nil {
				return n, err2
			}
		}
		a.state.onNewline = bytes.HasSuffix(line, []byte{'\n'})
		if len(line) > 0 {
			n2, err2 := a.writer.Write(line)
			n += n2
			if err2 != nil {
				return n, err2
			}
		}
		if err != nil {
			if err == io.EOF {
				return n, nil
			}
			return n, err
		}
	}
}

type padFormatAdapter struct {
	inner padAdapterInner
}

var _ fmt.State = (*padFormatAdapter)(nil)

func newPadFormatAdapter(f fmt.State, onNewline bool) *padFormatAdapter {
	return &padFormatAdapter{
		inner: padAdapterInner{
			writer: f,
			state: padAdapterState{
				onNewline: onNewline,
			},
		},
	}
}

func (a *padFormatAdapter) Write(p []byte) (n int, err error) { return a.inner.Write(p) }

func (a *padFormatAdapter) fmt() fmt.State { return a.inner.writer.(fmt.State) }

// Flag implements fmt.State.
func (a *padFormatAdapter) Flag(c int) bool { return a.fmt().Flag(c) }

// Precision implements fmt.State.
func (a *padFormatAdapter) Precision() (prec int, ok bool) { return a.fmt().Precision() }

// Width implements fmt.State.
func (a *padFormatAdapter) Width() (wid int, ok bool) { return a.fmt().Width() }
