package mjingo

import "fmt"

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
		g.setLineFromSpan(exp.span)
		g.add(instruction{kind: instructionKindLookup, data: data.id})
	case exprKindConst:
		data := exp.data.(constExprData)
		g.setLineFromSpan(exp.span)
		g.add(instruction{kind: instructionKindLoadConst, data: data.value})
	case exprKindGetAttr:
		data := exp.data.(getAttrExprData)
		g.pushSpan(exp.span)
		g.compileExpr(data.expr)
		g.add(instruction{kind: instructionKindGetAttr, data: data.name})
		g.popSpan()
	default:
		panic(fmt.Sprintf("not implemented for exprKind: %s", exp.kind))
	}
}

func (g *codeGenerator) finish() (instructions, map[string]instructions) {
	return g.instructions, g.blocks
}

func (g *codeGenerator) setLine(lineno uint32) {
	g.currentLine = lineno
}

func (g *codeGenerator) setLineFromSpan(spn span) {
	g.setLine(spn.startLine)
}

func (g *codeGenerator) pushSpan(spn span) {
	g.spanStack.push(spn)
	g.setLineFromSpan(spn)
}

func (g *codeGenerator) popSpan() {
	g.spanStack.pop()
}

func (g *codeGenerator) add(instr instruction) {
	if spn := g.spanStack.peek(); spn != nil {
		if spn.startLine == g.currentLine {
			panic("not implemented")
		}
	}
	g.instructions.addWithLine(instr, g.currentLine)
}
