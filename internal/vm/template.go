package vm

import (
	"strings"

	"github.com/hnakamur/mjingo/internal/compiler"
	"github.com/hnakamur/mjingo/value"
)

type Template struct {
	env               *Environment
	compiled          *compiledTemplate
	initialAutoEscape compiler.AutoEscape
}

func (t *Template) render(context any) (string, error) {
	var b strings.Builder
	root := context.(value.Value)
	out := newOutput(&b)
	if err := t._eval(root, out); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (t *Template) _eval(root value.Value, out *Output) error {
	vm := newVirtualMachine(t.env)
	if _, err := vm.eval(t.compiled.instructions, root, t.compiled.blocks,
		out, t.initialAutoEscape); err != nil {
		return err
	}
	return nil
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
