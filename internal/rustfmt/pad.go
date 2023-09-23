package rustfmt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type PadAdapter struct {
	inner padAdapterInner
}

var _ io.Writer = (*PadAdapter)(nil)

type padAdapterInner struct {
	writer io.Writer
	state  padAdapterState
}

type padAdapterState struct {
	onNewline bool
}

func NewPadAdapter(w io.Writer, onNewline bool) *PadAdapter {
	return &PadAdapter{
		inner: padAdapterInner{
			writer: w,
			state: padAdapterState{
				onNewline: onNewline,
			},
		},
	}
}

func (a *PadAdapter) Write(p []byte) (n int, err error) { return a.inner.Write(p) }

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

type PadFormatAdapter struct {
	inner padAdapterInner
}

var _ fmt.State = (*PadFormatAdapter)(nil)

func NewPadFormatAdapter(f fmt.State, onNewline bool) *PadFormatAdapter {
	return &PadFormatAdapter{
		inner: padAdapterInner{
			writer: f,
			state: padAdapterState{
				onNewline: onNewline,
			},
		},
	}
}

func (a *PadFormatAdapter) Write(p []byte) (n int, err error) { return a.inner.Write(p) }

func (a *PadFormatAdapter) fmt() fmt.State { return a.inner.writer.(fmt.State) }

// Flag implements fmt.State.
func (a *PadFormatAdapter) Flag(c int) bool { return a.fmt().Flag(c) }

// Precision implements fmt.State.
func (a *PadFormatAdapter) Precision() (prec int, ok bool) { return a.fmt().Precision() }

// Width implements fmt.State.
func (a *PadFormatAdapter) Width() (wid int, ok bool) { return a.fmt().Width() }
