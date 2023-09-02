package vm

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/value"
)

type Output struct {
	w            io.Writer
	captureStack []io.Writer
}

var _ = (io.Writer)((*Output)(nil))

func newOutput(w io.Writer) *Output {
	return &Output{w: w}
}

func newOutputNull() *Output {
	// The null writer also has a single entry on the discarding capture
	// stack.  In fact, `w` is more or less useless here as we always
	// shadow it.  This is done so that `is_discarding` returns true.
	return &Output{w: io.Discard, captureStack: []io.Writer{io.Discard}}
}

func (o *Output) target() io.Writer {
	if len(o.captureStack) > 0 {
		return o.captureStack[len(o.captureStack)-1]
	}
	return o.w
}

func (o *Output) isDiscarding() bool {
	return len(o.captureStack) > 0 && o.captureStack[len(o.captureStack)-1] == io.Discard
}

func (o *Output) Write(p []byte) (n int, err error) {
	return o.target().Write(p)
}

// Begins capturing into a string or discard.
func (o *Output) beginCapture(mode compiler.CaptureMode) {
	var w io.Writer
	switch mode {
	case compiler.CaptureModeCapture:
		w = new(strings.Builder)
	case compiler.CaptureModeDiscard:
		w = io.Discard
	default:
		panic("unreachable")
	}
	o.captureStack = append(o.captureStack, w)
}

// Ends capturing and returns the captured string as
func (o *Output) endCapture(escape AutoEscape) value.Value {
	if len(o.captureStack) == 0 {
		return value.Undefined
	}
	w := o.captureStack[len(o.captureStack)-1]
	o.captureStack = o.captureStack[:len(o.captureStack)-1]
	if builder, ok := w.(*strings.Builder); ok {
		str := builder.String()
		if _, ok := escape.(AutoEscapeNone); !ok {
			return value.ValueFromSafeString(str)
		} else {
			return value.ValueFromString(str)
		}
	}
	return value.Undefined
}

func writeString(o *Output, s string) error {
	_, err := io.WriteString(o, s)
	return err
}

func writeWithHTMLEscaping(o *Output, val value.Value) error {
	switch val.Kind() {
	case value.ValueKindUndefined, value.ValueKindNone, value.ValueKindBool, value.ValueKindNumber:
		return writeString(o, val.String())
	default:
		if optStr := val.AsStr(); optStr.IsSome() {
			// TODO: escape single quote with `&quot;` not `&#34;`
			return writeString(o, html.EscapeString(optStr.Unwrap()))
		}
		return writeString(o, html.EscapeString(val.String()))
	}
}

func writeEscaped(o *Output, autoEscape AutoEscape, val value.Value) error {
	// common case of safe strings or strings without auto escaping
	if val.IsSafe() || autoEscape.IsNone() {
		return writeString(o, val.String())
	}

	switch esc := autoEscape.(type) {
	case AutoEscapeNone:
		return writeString(o, val.String())
	case AutoEscapeHTML:
		return writeWithHTMLEscaping(o, val)
	case AutoEscapeJSON:
		panic("not implemented")
	case AutoEscapeCustom:
		panic(fmt.Sprintf("not implemented for custom auto escape name=%s", esc.Name))
	}
	return nil
}
