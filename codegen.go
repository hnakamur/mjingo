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

func (g *codeGenerator) compileStmt(s statement) {
	switch s := s.(type) {
	case templateStmt:
		for _, node := range s.children {
			g.compileStmt(node)
		}
	case emitExprStmt:
		g.compileEmitExpr(emitExprStmt{expr: s.expr, span: s.span})
	case emitRawStmt:
		g.add(emitRawInst{val: s.raw})
		g.rawTemplateBytes += uint(len(s.raw))
	}
}

func (g *codeGenerator) compileEmitExpr(exp emitExprStmt) {
	g.setLineFromSpan(exp.span)

	if _, ok := exp.expr.(callExpr); ok {
		panic("not implemented")
	}

	g.compileExpr(exp.expr)
	g.add(emitInst{})
}

func (g *codeGenerator) compileExpr(exp expression) {
	switch exp := exp.(type) {
	case varExpr:
		g.setLineFromSpan(exp.span)
		g.add(lookupInst{name: exp.id})
	case constExpr:
		g.setLineFromSpan(exp.span)
		g.add(loadConstInst{val: exp.value})
	case sliceExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		if exp.start.valid {
			g.compileExpr(exp.start.data)
		} else {
			g.add(loadConstInst{val: i64Value{n: int64(0)}})
		}
		if exp.stop.valid {
			g.compileExpr(exp.stop.data)
		} else {
			g.add(loadConstInst{val: valueNone})
		}
		if exp.step.valid {
			g.compileExpr(exp.step.data)
		} else {
			g.add(loadConstInst{val: i64Value{n: int64(1)}})
		}
		g.add(sliceInst{})
		g.popSpan()
	case unaryOpExpr:
		g.setLineFromSpan(exp.span)
		g.compileExpr(exp.expr)
		switch exp.op {
		case unaryOpTypeNot:
			g.add(notInst{})
		case unaryOpTypeNeg:
			g.addWithSpan(negInst{}, exp.span)
		}
	case binOpExpr:
		g.compileBinOp(exp)
	case getAttrExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.add(getAttrInst{name: exp.name})
		g.popSpan()
	case getItemExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.compileExpr(exp.subscriptExpr)
		g.add(getItemInst{})
		g.popSpan()
	case listExpr:
		if v := exp.asConst(); v.valid {
			g.add(loadConstInst{val: v.data})
		} else {
			g.setLineFromSpan(exp.span)
			for _, item := range exp.items {
				g.compileExpr(item)
			}
			g.add(buildListInst{count: uint(len(exp.items))})
		}
	case mapExpr:
		if v := exp.asConst(); v.valid {
			g.add(loadConstInst{val: v.data})
		} else {
			g.setLineFromSpan(exp.span)
			if len(exp.keys) != len(exp.values) {
				panic("mismatch length of keys and values for a map")
			}
			for i, key := range exp.keys {
				v := exp.values[i]
				g.compileExpr(key)
				g.compileExpr(v)
			}
			g.add(buildMapInst{pairCount: uint(len(exp.keys))})
		}
	default:
		panic(fmt.Sprintf("not implemented for exprType: %s", exp.typ()))
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

func (g *codeGenerator) addWithSpan(instr instruction, spn span) uint {
	return g.instructions.addWithSpan(instr, spn)
}

func (g *codeGenerator) compileBinOp(exp binOpExpr) {
	g.pushSpan(exp.span)
	var instr instruction
	switch exp.op {
	case binOpTypeEq:
		instr = eqInst{}
	case binOpTypeNe:
		instr = neInst{}
	case binOpTypeLt:
		instr = ltInst{}
	case binOpTypeLte:
		instr = lteInst{}
	case binOpTypeGt:
		instr = gtInst{}
	case binOpTypeGte:
		instr = gteInst{}
	case binOpTypeScAnd, binOpTypeScOr:
		panic("not implemented yet")
	case binOpTypeAdd:
		instr = addInst{}
	case binOpTypeSub:
		instr = subInst{}
	case binOpTypeMul:
		instr = mulInst{}
	case binOpTypeDiv:
		instr = divInst{}
	case binOpTypeFloorDiv:
		instr = intDivInst{}
	case binOpTypeRem:
		instr = remInst{}
	case binOpTypePow:
		instr = powInst{}
	case binOpTypeConcat:
		instr = stringConcatInst{}
	case binOpTypeIn:
		instr = inInst{}
	}
	g.compileExpr(exp.left)
	g.compileExpr(exp.right)
	g.add(instr)
	g.popSpan()
}
