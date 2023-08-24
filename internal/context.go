package internal

import (
	"github.com/hnakamur/mjingo/internal/datast/option"
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
	iterator             Iterator
	object               loop
}

type loop struct {
	len              uint
	idx              uint // atomic.Uint64
	depth            uint
	valueTriple      optValueTriple
	lastChangedValue option.Option[[]Value]
}

type optValueTriple [3]option.Option[Value]

type recursionJump struct {
	target     uint
	endCapture bool
}

type frame struct {
	locals      locals
	ctx         Value
	currentLoop option.Option[loopState]

	// normally a frame does not carry a closure, but it can when a macro is
	// declared.  Once that happens, all writes to the frames locals are also
	// duplicated into the closure.  Macros declared on that level, then share
	// the closure object to enclose the parent values.  This emulates the
	// behavior of closures in Jinja2.
	closure option.Option[Closure]
}

func newContext(f frame) *context {
	stack := make([]frame, 0, 12)
	stack = append(stack, f)
	return &context{stack: stack, outerStackDepth: 0}
}

func (c *context) store(key string, val Value) {
	top := &c.stack[len(c.stack)-1]
	if top.closure.IsSome() {
		(&top.closure).AsPtr().store(key, val.Clone())
	}
	top.locals[key] = val
}

func (c *context) closure() Closure {
	top := &c.stack[len(c.stack)-1]
	if top.closure.IsNone() {
		top.closure = option.Some(newClosure())
	}
	return top.closure.Unwrap().clone()
}

func (c *context) takeClosure() option.Option[Closure] {
	top := &c.stack[len(c.stack)-1]
	rv := top.closure
	top.closure = option.None[Closure]()
	return rv
}

func (c *context) resetClosure(closure option.Option[Closure]) {
	top := &c.stack[len(c.stack)-1]
	top.closure = closure
}

func (c *context) load(env *Environment, key string) option.Option[Value] {
	for i := len(c.stack) - 1; i >= 0; i-- {
		frame := c.stack[i]

		// look at locals first
		if v, ok := frame.locals[key]; ok {
			return option.Some(v)
		}

		// if we are a loop, check if we are looking up the special loop var.
		if frame.currentLoop.IsSome() {
			l := frame.currentLoop.Unwrap()
			if l.withLoopVar && key == "loop" {
				panic("not implemented")
			}
		}

		// perform a fast lookup.  This one will not produce errors if the
		// context is undefined or of the wrong type.
		if rv := frame.ctx.GetAttrFast(key); rv.IsSome() {
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

// Returns the current locals mutably.
func (c *context) currentLocals() *locals {
	return &c.stack[len(c.stack)-1].locals
}

// Returns the current innermost loop.
func (c *context) currentLoop() option.Option[*loopState] {
	for i := len(c.stack) - 1; i >= 0; i-- {
		frame := &c.stack[i]
		if frame.currentLoop.IsSome() {
			return option.Some((&frame.currentLoop).AsPtr())
		}
	}
	return option.None[*loopState]()
}

func (c *context) depth() uint {
	return c.outerStackDepth + uint(len(c.stack))
}

func (c *context) incrDepth(delta uint) error {
	c.outerStackDepth += delta
	return c.checkDepth()
}

func (c *context) decrDepth(delta uint) {
	c.outerStackDepth -= delta
}

func (c *context) checkDepth() error {
	if c.depth() > contextStackMaxRecursion {
		return NewError(InvalidOperation, "recursion limit exceeded")
	}
	return nil
}

func newFrame(ctx Value) *frame {
	return &frame{
		locals:      make(map[string]Value),
		ctx:         ctx,
		currentLoop: option.None[loopState](),
		closure:     option.None[Closure](),
	}
}

func newFrameDefault() *frame {
	return newFrame(Undefined)
}
