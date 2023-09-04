package mjingo

import (
	"strings"
)

type Template struct {
	env               *Environment
	compiled          *compiledTemplate
	initialAutoEscape AutoEscape
}

func (t *Template) Render(context Value) (string, error) {
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

func (t *Template) _eval(root Value, out *output) error {
	vm := newVirtualMachine(t.env)
	if _, err := vm.eval(t.compiled.instructions, root, t.compiled.blocks,
		out, t.initialAutoEscape); err != nil {
		return err
	}
	return nil
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