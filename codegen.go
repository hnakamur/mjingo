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
		g.add(instruction{kind: instructionKindEmitRaw, data: s.raw})
		g.rawTemplateBytes += uint(len(s.raw))
	}
}

func (g *codeGenerator) compileEmitExpr(exp emitExprStmt) {
	g.setLineFromSpan(exp.span)

	if _, ok := exp.expr.(callExpr); ok {
		panic("not implemented")
	}

	g.compileExpr(exp.expr)
	g.add(instruction{kind: instructionKindEmit})
}

func (g *codeGenerator) compileExpr(exp expression) {
	switch exp := exp.(type) {
	case varExpr:
		g.setLineFromSpan(exp.span)
		g.add(instruction{kind: instructionKindLookup, data: exp.id})
	case constExpr:
		g.setLineFromSpan(exp.span)
		g.add(instruction{kind: instructionKindLoadConst, data: exp.value})
	case sliceExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		if exp.start.valid {
			g.compileExpr(exp.start.data)
		} else {
			g.add(instruction{kind: instructionKindLoadConst, data: i64Value{n: int64(0)}})
		}
		if exp.stop.valid {
			g.compileExpr(exp.stop.data)
		} else {
			g.add(instruction{kind: instructionKindLoadConst, data: valueNone})
		}
		if exp.step.valid {
			g.compileExpr(exp.step.data)
		} else {
			g.add(instruction{kind: instructionKindLoadConst, data: i64Value{n: int64(1)}})
		}
		g.add(instruction{kind: instructionKindSlice})
		g.popSpan()
	case unaryOpExpr:
		g.setLineFromSpan(exp.span)
		g.compileExpr(exp.expr)
		switch exp.op {
		case unaryOpTypeNot:
			g.add(instruction{kind: instructionKindNot})
		case unaryOpTypeNeg:
			g.addWithSpan(instruction{kind: instructionKindNeg}, exp.span)
		}
	case binOpExpr:
		g.compileBinOp(exp)
	case getAttrExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.add(instruction{kind: instructionKindGetAttr, data: exp.name})
		g.popSpan()
	case getItemExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.compileExpr(exp.subscriptExpr)
		g.add(instruction{kind: instructionKindGetItem})
		g.popSpan()
	case listExpr:
		if v := exp.asConst(); v.valid {
			g.add(instruction{kind: instructionKindLoadConst, data: v.data})
		} else {
			g.setLineFromSpan(exp.span)
			for _, item := range exp.items {
				g.compileExpr(item)
			}
			g.add(instruction{
				kind: instructionKindBuildList,
				data: buildListInstructionData(len(exp.items)),
			})
		}
	case mapExpr:
		if v := exp.asConst(); v.valid {
			g.add(instruction{kind: instructionKindLoadConst, data: v.data})
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
			g.add(instruction{
				kind: instructionKindBuildMap,
				data: buildMapInstructionData(len(exp.keys)),
			})
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
		instr = instruction{kind: instructionKindEq}
	case binOpTypeNe:
		instr = instruction{kind: instructionKindNe}
	case binOpTypeLt:
		instr = instruction{kind: instructionKindLt}
	case binOpTypeLte:
		instr = instruction{kind: instructionKindLte}
	case binOpTypeGt:
		instr = instruction{kind: instructionKindGt}
	case binOpTypeGte:
		instr = instruction{kind: instructionKindGte}
	case binOpTypeScAnd, binOpTypeScOr:
		panic("not implemented yet")
	case binOpTypeAdd:
		instr = instruction{kind: instructionKindAdd}
	case binOpTypeSub:
		instr = instruction{kind: instructionKindSub}
	case binOpTypeMul:
		instr = instruction{kind: instructionKindMul}
	case binOpTypeDiv:
		instr = instruction{kind: instructionKindDiv}
	case binOpTypeFloorDiv:
		instr = instruction{kind: instructionKindIntDiv}
	case binOpTypeRem:
		instr = instruction{kind: instructionKindRem}
	case binOpTypePow:
		instr = instruction{kind: instructionKindPow}
	case binOpTypeConcat:
		instr = instruction{kind: instructionKindStringConcat}
	case binOpTypeIn:
		instr = instruction{kind: instructionKindIn}
	}
	g.compileExpr(exp.left)
	g.compileExpr(exp.right)
	g.add(instr)
	g.popSpan()
}
