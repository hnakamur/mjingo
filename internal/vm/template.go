package vm

import (
	"strings"

	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/internal/value"
)

type Template struct {
	env               *Environment
	compiled          *compiledTemplate
	initialAutoEscape AutoEscape
}

func (t *Template) Render(context value.Value) (string, error) {
	var b strings.Builder
	out := newOutput(&b)
	if err := t._eval(context, out); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (t *Template) name() string {
	return t.compiled.instructions.Name()
}

func (t *Template) _eval(root value.Value, out *Output) error {
	vm := newVirtualMachine(t.env)
	if _, err := vm.eval(t.compiled.instructions, root, t.compiled.blocks,
		out, t.initialAutoEscape); err != nil {
		return err
	}
	return nil
}

func (t *Template) instructionsAndBlocks() (instructions compiler.Instructions, blocks map[string]compiler.Instructions, err error) {
	return t.compiled.instructions, t.compiled.blocks, nil
}

type compiledTemplate struct {
	instructions   compiler.Instructions
	blocks         map[string]compiler.Instructions
	bufferSizeHint uint
	syntax         *compiler.SyntaxConfig
}

func newCompiledTemplate(name, source string, syntax compiler.SyntaxConfig) (*compiledTemplate, error) {
	st, err := compiler.ParseWithSyntax(source, name, syntax)
	if err != nil {
		return nil, err
	}
	gen := compiler.NewCodeGenerator(name, source)
	gen.CompileStmt(st)
	instructions, blocks := gen.Finish()
	return &compiledTemplate{
		instructions: instructions,
		blocks:       blocks,
		syntax:       &syntax,
	}, nil
}
