package mjingo

import (
	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/option"
	stackpkg "github.com/hnakamur/mjingo/internal/datast/stack"
)

type State interface {
	Env() *Environment
	Name() string
	AutoEscape() AutoEscape
	UndefinedBehavior() UndefinedBehavior
}

type vmState struct {
	env             *Environment
	ctx             context
	currentBlock    option.Option[string]
	autoEscape      AutoEscape
	instructions    instructions
	blocks          map[string]*blockStack
	loadedTemplates hashset.StrHashSet
	macros          stackpkg.Stack[macroStackElem]
}

var _ State = ((*vmState)(nil))

type locals = map[string]Value

type blockStack struct {
	instrs []instructions
	depth  uint
}

type macroStackElem struct {
	insts  instructions
	offset uint
}

func (s *vmState) Env() *Environment { return s.env }

func (s *vmState) AutoEscape() AutoEscape { return s.autoEscape }

func (s *vmState) Name() string {
	return s.instructions.Name()
}

func (s *vmState) UndefinedBehavior() UndefinedBehavior {
	return s.env.undefinedBehavior
}

func (s *vmState) lookup(name string) option.Option[Value] {
	return s.ctx.load(s.env, name)
}

func newBlockStack(instrs instructions) *blockStack {
	return &blockStack{instrs: []instructions{instrs}, depth: 0}
}

func (b *blockStack) instructions() instructions {
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

func (b *blockStack) appendInstructions(instrs instructions) {
	b.instrs = append(b.instrs, instrs)
}
