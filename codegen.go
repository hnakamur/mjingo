package mjingo

type codeGenerator struct {
	instructions     instructions
	blocks           map[string]instructions
	currentLine      uint32
	spanStack        stack[span]
	rawTemplateBytes uint
}

func newCodeGenerator(file, source string) *codeGenerator {
	return &codeGenerator{
		instructions: newInstructions(file, source),
		blocks:       make(map[string]instructions),
	}
}

func (g *codeGenerator) compileStmt(s stmt) {
	switch s.kind {
	case stmtKindTemplate:
		data := s.data.(templateStmtData)
		for _, node := range data.children {
			g.compileStmt(node)
		}
	case stmtKindEmitExpr:
		expr := s.data.(emitExprStmtData)
		g.compileEmitExpr(expr)
	case stmtKindEmitRaw:
		raw := s.data.(emitRawStmtData)
		g.add(instruction{kind: instructionKindEmitRaw, data: raw.raw})
		g.rawTemplateBytes += uint(len(raw.raw))
	}
}

func (g *codeGenerator) compileEmitExpr(data emitExprStmtData) {
	g.compileExpr(data.expr)
	g.add(instruction{kind: instructionKindEmit})
}

func (g *codeGenerator) compileExpr(exp expr) {
	switch exp.kind {
	case exprKindVar:
		data := exp.data.(varExprData)
		g.add(instruction{kind: instructionKindLookup, data: data.id})
	default:
		panic("not implemented")
	}
}

func (g *codeGenerator) add(instr instruction) {
	if spn := g.spanStack.peek(); spn != nil {
		if spn.startLine == g.currentLine {
			panic("not implemented")
		}
	}
	g.instructions.addWithLine(instr, g.currentLine)
}

func (g *codeGenerator) finish() (instructions, map[string]instructions) {
	return g.instructions, g.blocks
}
