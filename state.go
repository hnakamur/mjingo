package mjingo

import (
	"strings"

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

// RenderBlock renders a block with the given name into a string.
//
// This method works like [Template.Render] but
// it only renders a specific block in the template.  The first argument is
// the name of the block.
//
// Note that rendering a block is a stateful operation.  If an error
// is returned the module has to be re-created as the internal state
// can end up corrupted.  This also means you can only render blocks
// if you have a mutable reference to the state which is not possible
// from within filters or similar.
func (s *State) RenderBlock(block string) (string, error) {
	var b strings.Builder
	out := newOutput(&b)
	_, err := newVirtualMachine(s.env).callBlock(block, s, out)
	if err != nil {
		return "", err
	}
	return b.String(), nil
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
