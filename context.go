package mjingo

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
	iterator             any
	object               loop
}

type loop struct {
	len   uint
	idx   uint // atomic.Uint64
	depth uint
}

type recursionJump struct {
	target     uint
	endCapture bool
}

type frame struct {
	locals      locals
	ctx         value
	currentLoop option[loopState]

	closure any
}

func newContext(f frame) *context {
	stack := make([]frame, 0, 12)
	stack = append(stack, f)
	return &context{stack: stack, outerStackDepth: 0}
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

func newFrame(ctx value) *frame {
	return &frame{
		locals: make(map[string]value),
		ctx:    ctx,
	}
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
