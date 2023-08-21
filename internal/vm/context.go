package vm

import (
	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/value"
)

// The maximum recursion in the VM.  Normally each stack frame
// adds one to this counter (eg: every time a frame is added).
// However in some situations more depth is pushed if the cost
// of the stack frame is higher.
const contextStackMaxRecursion = 500

type context struct {
	stack           []frame
	outerStackDepth uint
}

type loopState struct {
	withLoopVar       bool
	recurseJumpTarget option.Option[uint]
	// if we're popping the frame, do we want to jump somewhere?  The
	// first item is the target jump instruction, the second argument
	// tells us if we need to end capturing.
	currentRecursionJump option.Option[recursionJump]
	iterator             value.Iterator
	object               loop
}

type loop struct {
	len              uint
	idx              uint // atomic.Uint64
	depth            uint
	valueTriple      optValueTriple
	lastChangedValue option.Option[[]value.Value]
}

type optValueTriple [3]option.Option[value.Value]

type recursionJump struct {
	target     uint
	endCapture bool
}

type frame struct {
	locals      locals
	ctx         value.Value
	currentLoop option.Option[loopState]

	// normally a frame does not carry a closure, but it can when a macro is
	// declared.  Once that happens, all writes to the frames locals are also
	// duplicated into the closure.  Macros declared on that level, then share
	// the closure object to enclose the parent values.  This emulates the
	// behavior of closures in Jinja2.
	closure option.Option[closure]
}

func newContext(f frame) *context {
	stack := make([]frame, 0, 12)
	stack = append(stack, f)
	return &context{stack: stack, outerStackDepth: 0}
}

func (c *context) store(key string, val value.Value) {
	top := &c.stack[len(c.stack)-1]
	// TODO: implement for top.closure
	top.locals[key] = val
}

func (c *context) closure() closure {
	top := &c.stack[len(c.stack)-1]
	if option.IsNone(top.closure) {
		top.closure = option.Some(newClosure())
	}
	return option.Unwrap(top.closure).clone()
}

func (c *context) load(env *Environment, key string) option.Option[value.Value] {
	for i := len(c.stack) - 1; i >= 0; i-- {
		frame := c.stack[i]

		// look at locals first
		if v, ok := frame.locals[key]; ok {
			return option.Some(v)
		}

		// if we are a loop, check if we are looking up the special loop var.
		if option.IsSome(frame.currentLoop) {
			l := option.Unwrap(frame.currentLoop)
			if l.withLoopVar && key == "loop" {
				panic("not implemented")
			}
		}

		// perform a fast lookup.  This one will not produce errors if the
		// context is undefined or of the wrong type.
		if rv := frame.ctx.GetAttrFast(key); option.IsSome(rv) {
			return rv
		}
	}
	return env.getGlobal(key)
}

func (c *context) pushFrame(f frame) error {
	if err := c.checkDepth(); err != nil {
		return err
	}
	c.stack = append(c.stack, f)
	return nil
}

func (c *context) popFrame() frame {
	f := c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
	return f
}

// Returns the current innermost loop.
func (c *context) currentLoop() option.Option[*loopState] {
	for i := len(c.stack) - 1; i >= 0; i-- {
		frame := &c.stack[i]
		if option.IsSome(frame.currentLoop) {
			return option.Some(option.AsPtr(&frame.currentLoop))
		}
	}
	return option.None[*loopState]()
}

func (c *context) depth() uint {
	return c.outerStackDepth + uint(len(c.stack))
}

func (c *context) checkDepth() error {
	if c.depth() > contextStackMaxRecursion {
		return internal.NewError(internal.InvalidOperation, "recursion limit exceeded")
	}
	return nil
}

func newFrame(ctx value.Value) *frame {
	return &frame{
		locals:      make(map[string]value.Value),
		ctx:         ctx,
		currentLoop: option.None[loopState](),
		closure:     option.None[closure](),
	}
}

func newFrameDefault() *frame {
	return newFrame(value.Undefined)
}
