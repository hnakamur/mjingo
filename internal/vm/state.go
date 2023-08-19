package vm

import (
	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/value"
)

type State struct {
	env          *Environment
	ctx          context
	currentBlock option.Option[string]
	autoEscape   compiler.AutoEscape
	instructions compiler.Instructions
	blocks       map[string]blockStack
}

type locals = map[string]value.Value

type blockStack struct {
	instrs []compiler.Instructions
	depth  uint
}

func (s *State) name() string {
	return s.instructions.Name()
}

func (s *State) undefinedBehavior() compiler.UndefinedBehavior {
	return s.env.undefinedBehavior
}

func (s *State) lookup(name string) option.Option[value.Value] {
	return s.ctx.load(s.env, name)
}

func newBlockStack(instrs compiler.Instructions) blockStack {
	return blockStack{instrs: []compiler.Instructions{instrs}, depth: 0}
}

func (b *blockStack) instructions() compiler.Instructions {
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

func (b *blockStack) appendInstructions(instrs compiler.Instructions) {
	b.instrs = append(b.instrs, instrs)
}
