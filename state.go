package mjingo

import (
	"github.com/hnakamur/mjingo/internal/datast/hashset"
	stackpkg "github.com/hnakamur/mjingo/internal/datast/stack"
	"github.com/hnakamur/mjingo/option"
)

// State provides access to the current execution state of the engine.
//
// A read only reference is passed to filter functions and similar objects to
// allow limited interfacing with the engine.  The state is useful to look up
// information about the engine in filter, test or global functions.  It not
// only provides access to the template environment but also the context
// variables of the engine, the current auto escaping behavior as well as the
// auto escape flag.
//
// In some testing scenarios or more advanced use cases you might need to get
// a [State].  The state is managed as part of the template execution but the
// initial state can be retrieved via [Template.NewState].
// The most common way to get hold of the state however is via functions of filters.
type State struct {
	env             *Environment
	ctx             context
	currentBlock    option.Option[string]
	autoEscape      AutoEscape
	instructions    instructions
	blocks          map[string]*blockStack
	loadedTemplates hashset.StrHashSet
	macros          stackpkg.Stack[macroStackElem]
}

type locals = map[string]Value

type blockStack struct {
	instrs []instructions
	depth  uint
}

type macroStackElem struct {
	insts  instructions
	offset uint
}

// Env returns a reference to the current environment.
func (s *State) Env() *Environment { return s.env }

// AutoEscape returns the current value of the auto escape flag.
func (s *State) AutoEscape() AutoEscape { return s.autoEscape }

// Name returns the name of the current template.
func (s *State) Name() string {
	return s.instructions.Name()
}

// UndefinedBehavior returns the current undefined behavior.
func (s *State) UndefinedBehavior() UndefinedBehavior {
	return s.env.undefinedBehavior
}

func (s *State) lookup(name string) option.Option[Value] {
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
