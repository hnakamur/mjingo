package mjingo

import (
	"io"
	"strings"
)

type captureMode uint8

const (
	captureModeCapture captureMode = iota + 1
	captureModeDiscard
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
func (o *Output) beginCapture(mode captureMode) {
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

// Ends capturing and returns the captured string as value.
func (o *Output) endCapture(escape autoEscape) value {
	if len(o.captureStack) == 0 {
		return valueUndefined
	}
	w := o.captureStack[len(o.captureStack)-1]
	o.captureStack = o.captureStack[:len(o.captureStack)-1]
	if builder, ok := w.(*strings.Builder); ok {
		str := builder.String()
		if _, ok := escape.(autoEscapeNone); !ok {
			return newValueFromSafeString(str)
		} else {
			return newValueFromString(str)
		}
	}
	return valueUndefined
}
