package mjingo

import (
	"io"
	"strings"
)

type Template struct {
	env      *Environment
	compiled *compiledTemplate
}

func (t *Template) render(context any) (string, error) {
	var b strings.Builder
	root := context.(value)
	if err := t._eval(root, &b); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (t *Template) _eval(root value, out io.Writer) error {
	vm := newVirtualMachine(t.env)
	if _, err := vm.eval(t.compiled.instructions, root, t.compiled.blocks, out); err != nil {
		return err
	}
	return nil
}

type compiledTemplate struct {
	instructions   instructions
	blocks         map[string]instructions
	bufferSizeHint uint
	syntax         *SyntaxConfig
}

func newCompiledTemplate(name, source string, syntax SyntaxConfig) (*compiledTemplate, error) {
	st, err := parseWithSyntax(source, name, syntax)
	if err != nil {
		return nil, err
	}
	gen := newCodeGenerator(name, source)
	gen.compileStmt(st)
	instructions, blocks := gen.finish()
	return &compiledTemplate{
		instructions: instructions,
		blocks:       blocks,
		syntax:       &syntax,
	}, nil
}
