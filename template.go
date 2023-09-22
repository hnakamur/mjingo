package mjingo

import (
	"strings"
)

// Template represents a template.
//
// Templates are stored in the [Environment] as bytecode instructions.  With the
// [Environment.GetTemplate] method that is looked up and returned in form of
// this handle.  Such a template can be cheaply copied as it only holds references.
//
// To render the [Template.Render] method can be used.
type Template struct {
	env               *Environment
	compiled          *compiledTemplate
	initialAutoEscape AutoEscape
}

// Render renders the template into a string.
//
// The provided value is used as the initial context for the template.  It
// can be any Value.
//
// For very large contexts and to avoid the overhead of serialization of
// potentially unused values, you might consider using a dynamic
// [StructObject] as value.
func (t *Template) Render(context Value) (string, error) {
	var b strings.Builder
	out := newOutput(&b)
	if err := t._eval(context, out); err != nil {
		return "", err
	}
	return b.String(), nil
}

// EvalToState evaluates the template into a [`State`].
//
// This evaluates the template, discards the output and returns the final
// `State` for introspection.  From there global variables or blocks
// can be accessed.  What this does is quite similar to how the engine
// interally works with tempaltes that are extended or imported from.
func (t *Template) EvalToState(context Value) (*State, error) {
	out := newOutputNull()
	vm := newVirtualMachine(t.env)
	_, state, err := vm.eval(t.compiled.instructions, context, t.compiled.blocks, out, t.initialAutoEscape)
	return state, err
}

func (t *Template) name() string {
	return t.compiled.instructions.Name()
}

func (t *Template) _eval(root Value, out *output) error {
	vm := newVirtualMachine(t.env)
	_, _, err := vm.eval(t.compiled.instructions, root, t.compiled.blocks,
		out, t.initialAutoEscape)
	return err
}

func (t *Template) instructionsAndBlocks() (insts instructions, blocks map[string]instructions, err error) {
	return t.compiled.instructions, t.compiled.blocks, nil
}

type compiledTemplate struct {
	instructions   instructions
	blocks         map[string]instructions
	bufferSizeHint uint
	syntax         *syntaxConfig
}

func newCompiledTemplate(name, source string, syntax syntaxConfig, keepTrailingNewline bool) (*compiledTemplate, error) {
	return attachBasicDebugInfo[*compiledTemplate](source)(newCompiledTemplateImpl(name, source, syntax, keepTrailingNewline))
}

func newCompiledTemplateImpl(name, source string, syntax syntaxConfig, keepTrailingNewline bool) (*compiledTemplate, error) {
	st, err := parseWithSyntax(source, name, syntax, keepTrailingNewline)
	if err != nil {
		return nil, err
	}
	gen := newCodeGenerator(name, source)
	gen.CompileStmt(st)
	instructions, blocks := gen.Finish()
	return &compiledTemplate{
		instructions: instructions,
		blocks:       blocks,
		syntax:       &syntax,
	}, nil
}
