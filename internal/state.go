package internal

import (
	"github.com/hnakamur/mjingo/internal/datast/option"
)

type State struct {
	env          *Environment
	ctx          context
	currentBlock option.Option[string]
	autoEscape   AutoEscape
	instructions Instructions
	blocks       map[string]blockStack
}

type locals = map[string]Value

type blockStack struct {
	instrs []Instructions
	depth  uint
}

func (s *State) name() string {
	return s.instructions.Name()
}

func (s *State) undefinedBehavior() UndefinedBehavior {
	return s.env.undefinedBehavior
}

func (s *State) lookup(name string) option.Option[Value] {
	return s.ctx.load(s.env, name)
}

func newBlockStack(instrs Instructions) blockStack {
	return blockStack{instrs: []Instructions{instrs}, depth: 0}
}

func (b *blockStack) instructions() Instructions {
	return b.instrs[b.depth]
}

func (b *blockStack) push() bool {
	if b.depth+1 < uint(len(b.instrs)) {
		b.depth++
		return true
	}
	return false
}

func (b *blockStack) pop() {
	if b.depth > 0 {
		b.depth--
	}
}

func (b *blockStack) appendInstructions(instrs Instructions) {
	b.instrs = append(b.instrs, instrs)
}
