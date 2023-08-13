package mjingo

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
	gen.compileStmt(*st)
	instructions, blocks := gen.finish()
	return &compiledTemplate{
		instructions: instructions,
		blocks:       blocks,
		syntax:       &syntax,
	}, nil
}
