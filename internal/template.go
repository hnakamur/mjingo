package internal

import (
	"strings"
)

type Template struct {
	env               *Environment
	compiled          *compiledTemplate
	initialAutoEscape AutoEscape
}

func (t *Template) Render(context any) (string, error) {
	var b strings.Builder
	root := context.(Value)
	out := newOutput(&b)
	if err := t._eval(root, out); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (t *Template) name() string {
	return t.compiled.instructions.Name()
}

func (t *Template) _eval(root Value, out *Output) error {
	vm := newVirtualMachine(t.env)
	if _, err := vm.eval(t.compiled.instructions, root, t.compiled.blocks,
		out, t.initialAutoEscape); err != nil {
		return err
	}
	return nil
}

func (t *Template) instructionsAndBlocks() (instructions Instructions, blocks map[string]Instructions, err error) {
	return t.compiled.instructions, t.compiled.blocks, nil
}

type compiledTemplate struct {
	instructions   Instructions
	blocks         map[string]Instructions
	bufferSizeHint uint
	syntax         *SyntaxConfig
}

func newCompiledTemplate(name, source string, syntax SyntaxConfig) (*compiledTemplate, error) {
	st, err := ParseWithSyntax(source, name, syntax)
	if err != nil {
		return nil, err
	}
	gen := NewCodeGenerator(name, source)
	gen.CompileStmt(st)
	instructions, blocks := gen.Finish()
	return &compiledTemplate{
		instructions: instructions,
		blocks:       blocks,
		syntax:       &syntax,
	}, nil
}
