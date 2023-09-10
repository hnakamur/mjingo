package mjingo

import (
	"fmt"
	"io"
	"strings"
)

type output struct {
	w            io.Writer
	captureStack []io.Writer
}

var _ = (io.Writer)((*output)(nil))

func newOutput(w io.Writer) *output {
	return &output{w: w}
}

func newOutputNull() *output {
	// The null writer also has a single entry on the discarding capture
	// stack.  In fact, `w` is more or less useless here as we always
	// shadow it.  This is done so that `is_discarding` returns true.
	return &output{w: io.Discard, captureStack: []io.Writer{io.Discard}}
}

func (o *output) target() io.Writer {
	if len(o.captureStack) > 0 {
		return o.captureStack[len(o.captureStack)-1]
	}
	return o.w
}

func (o *output) isDiscarding() bool {
	return len(o.captureStack) > 0 && o.captureStack[len(o.captureStack)-1] == io.Discard
}

func (o *output) Write(p []byte) (n int, err error) {
	return o.target().Write(p)
}

// Begins capturing into a string or discard.
func (o *output) beginCapture(mode captureMode) {
	var w io.Writer
	switch mode {
	case captureModeCapture:
		w = new(strings.Builder)
	case captureModeDiscard:
		w = io.Discard
	default:
		panic("unreachable")
	}
	o.captureStack = append(o.captureStack, w)
}

// Ends capturing and returns the captured string as
func (o *output) endCapture(escape AutoEscape) Value {
	if len(o.captureStack) == 0 {
		return Undefined
	}
	w := o.captureStack[len(o.captureStack)-1]
	o.captureStack = o.captureStack[:len(o.captureStack)-1]
	if builder, ok := w.(*strings.Builder); ok {
		str := builder.String()
		if _, ok := escape.(autoEscapeNone); !ok {
			return ValueFromSafeString(str)
		} else {
			return valueFromString(str)
		}
	}
	return Undefined
}

func writeString(o *output, s string) error {
	_, err := io.WriteString(o, s)
	return err
}

func writeWithHTMLEscaping(o *output, val Value) error {
	switch val.kind() {
	case valueKindUndefined, valueKindNone, valueKindBool, valueKindNumber:
		return writeString(o, val.String())
	default:
		str, err := valueTryToGoString(val)
		if err != nil {
			str = val.String()
		}
		return writeString(o, htmlEscapeString(str))
	}
}

func writeEscaped(o *output, autoEscape AutoEscape, val Value) error {
	// common case of safe strings or strings without auto escaping
	if val.isSafe() || autoEscape.isNone() {
		return writeString(o, val.String())
	}

	switch esc := autoEscape.(type) {
	case autoEscapeNone:
		return writeString(o, val.String())
	case autoEscapeHTML:
		return writeWithHTMLEscaping(o, val)
	case autoEscapeJSON:
		panic("not implemented")
	case autoEscapeCustom:
		panic(fmt.Sprintf("not implemented for custom auto escape name=%s", esc.name))
	}
	return nil
}

var htmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`'`, "&#x27;",
	`/`, "&#x2f;",
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&quot;",
)

func htmlEscapeString(s string) string {
	return htmlEscaper.Replace(s)
}
