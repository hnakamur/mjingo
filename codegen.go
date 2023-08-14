package mjingo

import (
	"fmt"
)

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
		g.compileEmitExpr(spanned[emitExprStmtData]{data: expr, span: s.span})
	case stmtKindEmitRaw:
		raw := s.data.(emitRawStmtData)
		g.add(instruction{kind: instructionKindEmitRaw, data: raw.raw})
		g.rawTemplateBytes += uint(len(raw.raw))
	}
}

func (g *codeGenerator) compileEmitExpr(exp spanned[emitExprStmtData]) {
	g.setLineFromSpan(exp.span)

	if exp.data.expr.kind == exprKindCall {
		panic("not implemented")
	}

	g.compileExpr(exp.data.expr)
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
	case exprKindSlice:
		data := exp.data.(sliceExprData)
		g.pushSpan(exp.span)
		g.compileExpr(data.expr)
		if data.start.valid {
			g.compileExpr(data.start.data)
		} else {
			g.add(instruction{kind: instructionKindLoadConst, data: value{kind: valueKindI64, data: int64(0)}})
		}
		if data.stop.valid {
			g.compileExpr(data.stop.data)
		} else {
			g.add(instruction{kind: instructionKindLoadConst, data: valueNone})
		}
		if data.step.valid {
			g.compileExpr(data.step.data)
		} else {
			g.add(instruction{kind: instructionKindLoadConst, data: value{kind: valueKindI64, data: int64(1)}})
		}
		g.add(instruction{kind: instructionKindSlice})
		g.popSpan()
	case exprKindGetAttr:
		data := exp.data.(getAttrExprData)
		g.pushSpan(exp.span)
		g.compileExpr(data.expr)
		g.add(instruction{kind: instructionKindGetAttr, data: data.name})
		g.popSpan()
	case exprKindGetItem:
		data := exp.data.(getItemExprData)
		g.pushSpan(exp.span)
		g.compileExpr(data.expr)
		g.compileExpr(data.subscriptExpr)
		g.add(instruction{kind: instructionKindGetItem})
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

func (g *codeGenerator) add(instr instruction) uint {
	if spn := g.spanStack.peek(); spn != nil {
		if spn.startLine == g.currentLine {
			return g.instructions.addWithSpan(instr, *spn)
		}
	}
	return g.instructions.addWithLine(instr, g.currentLine)
}
