package mjingo

import (
	"fmt"
)

type codeGenerator struct {
	instructions     instructions
	blocks           map[string]instructions
	pendingBlock     stack[pendingBlock]
	currentLine      uint32
	spanStack        stack[span]
	rawTemplateBytes uint
}

// Represents an open block of code that does not yet have updated
// jump targets.
type pendingBlock interface {
	typ() pendingBlockType
}

var _ = pendingBlock(branchPendingBlock{})
var _ = pendingBlock(loopPendingBlock{})
var _ = pendingBlock(scBoolPendingBlock{})

type branchPendingBlock struct{ jumpInst uint }
type loopPendingBlock struct{ iterInst uint }
type scBoolPendingBlock struct{ instructions []uint }

func (branchPendingBlock) typ() pendingBlockType { return pendingBlockTypeBranch }
func (loopPendingBlock) typ() pendingBlockType   { return pendingBlockTypeLoop }
func (scBoolPendingBlock) typ() pendingBlockType { return pendingBlockTypeScBool }

type pendingBlockType int

const (
	pendingBlockTypeBranch pendingBlockType = iota + 1
	pendingBlockTypeLoop
	pendingBlockTypeScBool
)

func newCodeGenerator(file, source string) *codeGenerator {
	return &codeGenerator{
		instructions: newInstructions(file, source),
		blocks:       make(map[string]instructions),
		pendingBlock: newStackWithCapacity[pendingBlock](32),
	}
}

func (g *codeGenerator) compileStmt(stmt statement) {
	switch st := stmt.(type) {
	case templateStmt:
		for _, node := range st.children {
			g.compileStmt(node)
		}
	case emitExprStmt:
		g.compileEmitExpr(emitExprStmt{expr: st.expr, span: st.span})
	case emitRawStmt:
		g.add(emitRawInstruction{val: st.raw})
		g.rawTemplateBytes += uint(len(st.raw))
	case ifCondStmt:
		g.compileIfStmt(st)
	}
}

func (g *codeGenerator) compileIfStmt(ifCond ifCondStmt) {
	g.setLineFromSpan(ifCond.span)
	g.compileExpr(ifCond.expr)
	g.startIf()
	for _, node := range ifCond.trueBody {
		g.compileStmt(node)
	}
	if len(ifCond.falseBody) > 0 {
		g.startElse()
		for _, node := range ifCond.falseBody {
			g.compileStmt(node)
		}
	}
	g.endIf()
}

func (g *codeGenerator) compileEmitExpr(exp emitExprStmt) {
	g.setLineFromSpan(exp.span)

	if _, ok := exp.expr.(callExpr); ok {
		panic("not implemented")
	}

	g.compileExpr(exp.expr)
	g.add(emitInstruction{})
}

func (g *codeGenerator) compileExpr(exp expression) {
	switch exp := exp.(type) {
	case varExpr:
		g.setLineFromSpan(exp.span)
		g.add(lookupInstruction{name: exp.id})
	case constExpr:
		g.setLineFromSpan(exp.span)
		g.add(loadConstInstruction{val: exp.value})
	case sliceExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		if exp.start.valid {
			g.compileExpr(exp.start.data)
		} else {
			g.add(loadConstInstruction{val: i64Value{n: int64(0)}})
		}
		if exp.stop.valid {
			g.compileExpr(exp.stop.data)
		} else {
			g.add(loadConstInstruction{val: valueNone})
		}
		if exp.step.valid {
			g.compileExpr(exp.step.data)
		} else {
			g.add(loadConstInstruction{val: i64Value{n: int64(1)}})
		}
		g.add(sliceInstruction{})
		g.popSpan()
	case unaryOpExpr:
		g.setLineFromSpan(exp.span)
		g.compileExpr(exp.expr)
		switch exp.op {
		case unaryOpTypeNot:
			g.add(notInstruction{})
		case unaryOpTypeNeg:
			g.addWithSpan(negInstruction{}, exp.span)
		}
	case binOpExpr:
		g.compileBinOp(exp)
	case ifExpr:
		g.setLineFromSpan(exp.span)
		g.compileExpr(exp.testExpr)
		g.startIf()
		g.compileExpr(exp.trueExpr)
		g.startElse()
		if exp.falseExpr.valid {
			g.compileExpr(exp.falseExpr.data)
		} else {
			g.add(loadConstInstruction{val: valueUndefined})
		}
		g.endIf()
	case getAttrExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.add(getAttrInstruction{name: exp.name})
		g.popSpan()
	case getItemExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.compileExpr(exp.subscriptExpr)
		g.add(getItemInstruction{})
		g.popSpan()
	case listExpr:
		if v := exp.asConst(); v.valid {
			g.add(loadConstInstruction{val: v.data})
		} else {
			g.setLineFromSpan(exp.span)
			for _, item := range exp.items {
				g.compileExpr(item)
			}
			g.add(buildListInstruction{count: uint(len(exp.items))})
		}
	case mapExpr:
		if v := exp.asConst(); v.valid {
			g.add(loadConstInstruction{val: v.data})
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
			g.add(buildMapInstruction{pairCount: uint(len(exp.keys))})
		}
	default:
		panic(fmt.Sprintf("not implemented for exprType: %s", exp.typ()))
	}
}

func (g *codeGenerator) startIf() {
	jumpInst := g.add(jumpIfFalseInstruction{jumpTarget: 0})
	g.pendingBlock.push(branchPendingBlock{jumpInst: jumpInst})
}

func (g *codeGenerator) startElse() {
	jumpInst := g.add(jumpInstruction{jumpTarget: 0})
	g.endCondition(jumpInst + 1)
	g.pendingBlock.push(branchPendingBlock{jumpInst: jumpInst})
}

func (g *codeGenerator) endIf() {
	g.endCondition(g.nextInstruction())
}

func (g *codeGenerator) endCondition(jumpInst uint) {
	b := g.pendingBlock.pop()
	if b == nil {
		panic("pendingBlock should not be empty in endCondition")
	}
	if b, ok := (*b).(branchPendingBlock); ok {
		switch g.instructions.instructions[b.jumpInst].(type) {
		case jumpIfFalseInstruction:
			g.instructions.instructions[b.jumpInst] = jumpIfFalseInstruction{jumpTarget: jumpInst}
		case jumpInstruction:
			g.instructions.instructions[b.jumpInst] = jumpInstruction{jumpTarget: jumpInst}
		}
	} else {
		panic("must be branchPendingBlock")
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

func (g *codeGenerator) nextInstruction() uint {
	return uint(len(g.instructions.instructions))
}

func (g *codeGenerator) compileBinOp(exp binOpExpr) {
	g.pushSpan(exp.span)
	var instr instruction
	switch exp.op {
	case binOpTypeEq:
		instr = eqInstruction{}
	case binOpTypeNe:
		instr = neInstruction{}
	case binOpTypeLt:
		instr = ltInstruction{}
	case binOpTypeLte:
		instr = lteInstruction{}
	case binOpTypeGt:
		instr = gtInstruction{}
	case binOpTypeGte:
		instr = gteInstruction{}
	case binOpTypeScAnd, binOpTypeScOr:
		panic("not implemented yet")
	case binOpTypeAdd:
		instr = addInstruction{}
	case binOpTypeSub:
		instr = subInstruction{}
	case binOpTypeMul:
		instr = mulInstruction{}
	case binOpTypeDiv:
		instr = divInstruction{}
	case binOpTypeFloorDiv:
		instr = intDivInstruction{}
	case binOpTypeRem:
		instr = remInstruction{}
	case binOpTypePow:
		instr = powInstruction{}
	case binOpTypeConcat:
		instr = stringConcatInstruction{}
	case binOpTypeIn:
		instr = inInstruction{}
	}
	g.compileExpr(exp.left)
	g.compileExpr(exp.right)
	g.add(instr)
	g.popSpan()
}
