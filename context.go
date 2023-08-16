package mjingo

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
	recurseJumpTarget option[uint]
	// if we're popping the frame, do we want to jump somewhere?  The
	// first item is the target jump instruction, the second argument
	// tells us if we need to end capturing.
	currentRecursionJump option[recursionJump]
	iterator             valueIterator
	object               loop
}

type loop struct {
	len              uint
	idx              uint // atomic.Uint64
	depth            uint
	valueTriple      optValueTriple
	lastChangedValue option[[]value]
}

type optValueTriple [3]option[value]

type recursionJump struct {
	target     uint
	endCapture bool
}

type frame struct {
	locals      locals
	ctx         value
	currentLoop option[loopState]

	// normally a frame does not carry a closure, but it can when a macro is
	// declared.  Once that happens, all writes to the frames locals are also
	// duplicated into the closure.  Macros declared on that level, then share
	// the closure object to enclose the parent values.  This emulates the
	// behavior of closures in Jinja2.
	closure any
}

func newContext(f frame) *context {
	stack := make([]frame, 0, 12)
	stack = append(stack, f)
	return &context{stack: stack, outerStackDepth: 0}
}

func (c *context) store(key string, val value) {
	top := &c.stack[len(c.stack)-1]
	// TODO: implement for top.closure
	top.locals[key] = val
}

func (c *context) load(env *Environment, key string) option[value] {
	for i := len(c.stack) - 1; i >= 0; i-- {
		frame := c.stack[i]

		// look at locals first
		if v, ok := frame.locals[key]; ok {
			return option[value]{valid: true, data: v}
		}

		// if we are a loop, check if we are looking up the special loop var.
		if frame.currentLoop.valid {
			l := frame.currentLoop.data
			if l.withLoopVar && key == "loop" {
				panic("not implemented")
			}
		}

		// perform a fast lookup.  This one will not produce errors if the
		// context is undefined or of the wrong type.
		if rv := frame.ctx.getAttrFast(key); rv.valid {
			return rv
		}
	}
	panic("not implemented")
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
func (c *context) currentLoop() option[*loopState] {
	for i := len(c.stack) - 1; i >= 0; i-- {
		frame := &c.stack[i]
		if frame.currentLoop.valid {
			return option[*loopState]{valid: true, data: &frame.currentLoop.data}
		}
	}
	return option[*loopState]{}
}

func (c *context) depth() uint {
	return c.outerStackDepth + uint(len(c.stack))
}

func (c *context) checkDepth() error {
	if c.depth() > contextStackMaxRecursion {
		return newError(InvalidOperation, "recursion limit exceeded")
	}
	return nil
}

func newFrame(ctx value) *frame {
	return &frame{
		locals: make(map[string]value),
		ctx:    ctx,
	}
}

func newFrameDefault() *frame {
	return newFrame(valueUndefined)
}

type vmStack struct {
	values []value
}

func newVMStack() vmStack {
	return vmStack{values: make([]value, 0, 16)}
}

func (s *vmStack) push(arg value) {
	s.values = append(s.values, arg)
}

func (s *vmStack) pop() value {
	v := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return v
}

func (s *vmStack) sliceTop(n uint) []value {
	return s.values[uint(len(s.values))-n:]
}

func (s *vmStack) dropTop(n uint) {
	s.values = s.values[:uint(len(s.values))-n]
}

func (s *vmStack) tryPop() option[value] {
	if len(s.values) == 0 {
		return option[value]{}
	}
	return option[value]{valid: true, data: s.pop()}
}

func (s *vmStack) peek() *value {
	return &s.values[len(s.values)-1]
}
