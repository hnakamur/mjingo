package vm

import (
	"io"
	"strings"

	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/valu"
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

// Ends capturing and returns the captured string as value.
func (o *Output) endCapture(escape compiler.AutoEscape) valu.Value {
	if len(o.captureStack) == 0 {
		return valu.Undefined
	}
	w := o.captureStack[len(o.captureStack)-1]
	o.captureStack = o.captureStack[:len(o.captureStack)-1]
	if builder, ok := w.(*strings.Builder); ok {
		str := builder.String()
		if _, ok := escape.(compiler.AutoEscapeNone); !ok {
			return valu.FromSafeString(str)
		} else {
			return valu.FromString(str)
		}
	}
	return valu.Undefined
}
