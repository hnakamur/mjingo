package mjingo

import (
	"github.com/hnakamur/mjingo/internal/datast/stack"
	"github.com/hnakamur/mjingo/option"
)

type codeGenerator struct {
	instructions     instructions
	blocks           map[string]instructions
	pendingBlock     stack.Stack[pendingBlock]
	currentLine      uint32
	spanStack        stack.Stack[span]
	filterLocalIds   map[string]localID
	testLocalIds     map[string]localID
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
		instructions:   newInstructions(file, source),
		blocks:         make(map[string]instructions),
		pendingBlock:   stack.NewStackWithCapacity[pendingBlock](32),
		filterLocalIds: make(map[string]localID),
		testLocalIds:   make(map[string]localID),
	}
}

func (g *codeGenerator) CompileStmt(stmt statement) {
	switch st := stmt.(type) {
	case templateStmt:
		g.setLineFromSpan(st.span)
		for _, node := range st.children {
			g.CompileStmt(node)
		}
	case emitExprStmt:
		g.compileEmitExpr(emitExprStmt{expr: st.expr, span: st.span})
	case emitRawStmt:
		g.setLineFromSpan(st.span)
		g.add(emitRawInstruction{Val: st.raw})
		g.rawTemplateBytes += uint(len(st.raw))
	case forLoopStmt:
		g.compileForLoop(st)
	case ifCondStmt:
		g.compileIfStmt(st)
	case withBlockStmt:
		g.setLineFromSpan(st.span)
		g.add(pushWithInstruction{})
		for _, assign := range st.assignments {
			g.compileExpr(assign.rhs)
			g.compileAssignment(assign.lhs)
		}
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(popFrameInstruction{})
	case setStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.expr)
		g.compileAssignment(st.target)
	case setBlockStmt:
		g.setLineFromSpan(st.span)
		g.add(beginCaptureInstruction{Mode: captureModeCapture})
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(endCaptureInstruction{})
		if st.filter.IsSome() {
			g.compileExpr(st.filter.Unwrap())
		}
		g.compileAssignment(st.target)
	case autoEscapeStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.enabled)
		g.add(pushAutoEscapeInstruction{})
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(popAutoEscapeInstruction{})
	case filterBlockStmt:
		g.setLineFromSpan(st.span)
		g.add(beginCaptureInstruction{Mode: captureModeCapture})
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(endCaptureInstruction{})
		g.compileExpr(st.filter)
		g.add(emitInstruction{})
	case blockStmt:
		g.compileBlock(st)
	case importStmt:
		g.add(beginCaptureInstruction{Mode: captureModeDiscard})
		g.add(pushWithInstruction{})
		g.compileExpr(st.expr)
		g.addWithSpan(includeInstruction{IgnoreMissing: false}, st.span)
		g.add(exportLocalsInstruction{})
		g.add(popFrameInstruction{})
		g.compileAssignment(st.name)
		g.add(endCaptureInstruction{})
	case fromImportStmt:
		g.add(beginCaptureInstruction{Mode: captureModeDiscard})
		g.add(pushWithInstruction{})
		g.compileExpr(st.expr)
		g.addWithSpan(includeInstruction{IgnoreMissing: false}, st.span)
		for _, importName := range st.names {
			g.compileExpr(importName.name)
		}
		g.add(popFrameInstruction{})
		for _, importName := range st.names {
			g.compileAssignment(importName.as.UnwrapOr(importName.name))
		}
		g.add(endCaptureInstruction{})
	case extendsStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.name)
		g.addWithSpan(loadBlocksInstruction{}, st.span)
	case includeStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.name)
		g.addWithSpan(includeInstruction{IgnoreMissing: st.ignoreMissing}, st.span)
	case macroStmt:
		g.compileMacro(st)
	case callBlockStmt:
		g.compileCallBlock(st)
	case doStmt:
		g.compileDo(st)
	default:
		panic("unreachable")
	}
}

func (g *codeGenerator) compileBlock(block blockStmt) {
	g.setLineFromSpan(block.span)
	sub := g.newSubgenerator()
	for _, node := range block.body {
		sub.CompileStmt(node)
	}
	insts := g.finishSubgenerator(sub)
	g.blocks[block.name] = insts
	g.add(callBlockInstruction{Name: block.name})
}

func (g *codeGenerator) compileEmitExpr(exp emitExprStmt) {
	g.setLineFromSpan(exp.span)

	if callExpr, ok := exp.expr.(callExpr); ok {
		switch ct := callExpr.call.identityCall().(type) {
		case callTypeFunction:
			if ct.name == "super" && len(callExpr.call.args) == 0 {
				g.addWithSpan(fastSuperInstruction{}, callExpr.span)
				return
			} else if ct.name == "loop" && len(callExpr.call.args) == 1 {
				g.compileExpr(callExpr.call.args[0])
				g.add(fastRecurseInstruction{})
				return
			}
		case callTypeBlock:
			g.add(callBlockInstruction{Name: ct.name})
			return
		}
	}
	g.compileExpr(exp.expr)
	g.add(emitInstruction{})
}

func (g *codeGenerator) compileForLoop(forLoop forLoopStmt) {
	g.setLineFromSpan(forLoop.span)
	if forLoop.filterExpr.IsSome() {
		// filter expressions work like a nested for loop without
		// the special loop variable that append into a new list
		// just outside of the loop.
		g.add(buildListInstruction{Count: 0})
		g.compileExpr(forLoop.iter)
		g.startForLoop(false, false)
		g.add(dupTopInstruction{})
		g.compileAssignment(forLoop.target)
		g.compileExpr(forLoop.filterExpr.Unwrap())
		g.startIf()
		g.add(listAppendInstruction{})
		g.startElse()
		g.add(discardTopInstruction{})
		g.endIf()
		g.endForLoop(false)
	} else {
		g.compileExpr(forLoop.iter)
	}
	g.startForLoop(true, forLoop.recursive)
	g.compileAssignment(forLoop.target)
	for _, node := range forLoop.body {
		g.CompileStmt(node)
	}
	g.endForLoop(len(forLoop.elseBody) != 0)
	if len(forLoop.elseBody) != 0 {
		g.startIf()
		for _, node := range forLoop.elseBody {
			g.CompileStmt(node)
		}
		g.endIf()
	}
}
func (g *codeGenerator) compileAssignment(expr astExpr) {
	switch exp := expr.(type) {
	case varExpr:
		g.add(storeLocalInstruction{Name: exp.id})
	case listExpr:
		g.pushSpan(exp.span)
		g.add(unpackListInstruction{Count: uint(len(exp.items))})
		for _, expr := range exp.items {
			g.compileAssignment(expr)
		}
		g.popSpan()
	default:
		panic("unreachable")
	}
}

func (g *codeGenerator) compileMacroExpression(macroDecl macroStmt) {
	g.setLineFromSpan(macroDecl.span)
	inst := g.add(jumpInstruction{JumpTarget: ^uint(0)})
	j := len(macroDecl.defaults) - 1
	for i := len(macroDecl.args) - 1; i >= 0; i-- {
		if j >= 0 {
			g.add(dupTopInstruction{})
			g.add(isUndefinedInstruction{})
			g.startIf()
			g.add(discardTopInstruction{})
			g.compileExpr(macroDecl.defaults[j])
			g.endIf()
			j--
		}
		g.compileAssignment(macroDecl.args[i])
	}
	for _, node := range macroDecl.body {
		g.CompileStmt(node)
	}
	g.add(returnInstruction{})
	undeclared := findMacroClosure(macroDecl)
	callerReference := undeclared.Contains("caller")
	undeclared.Delete("caller")
	macroInst := g.nextInstruction()
	for _, name := range undeclared.Keys() {
		g.add(encloseInstruction{Name: name})
	}
	g.add(getClosureInstruction{})
	ids := make([]Value, 0, len(macroDecl.args))
	for _, arg := range macroDecl.args {
		if varExp, ok := arg.(varExpr); ok {
			ids = append(ids, valueFromString(varExp.id))
		} else {
			panic("unreachable")
		}
	}
	g.add(loadConstInstruction{Val: valueFromSlice(ids)})
	flags := uint8(0)
	if callerReference {
		flags |= macroCaller
	}
	g.add(buildMacroInstruction{Name: macroDecl.name, Offset: inst + 1, Flags: flags})
	if g.instructions.instructions[inst].Typ() == instTypeJump {
		g.instructions.instructions[inst] = jumpInstruction{JumpTarget: macroInst}
	} else {
		panic("unreachable")
	}
}

func (g *codeGenerator) compileMacro(macroDecl macroStmt) {
	g.compileMacroExpression(macroDecl)
	g.add(storeLocalInstruction{Name: macroDecl.name})
}

func (g *codeGenerator) compileCallBlock(callBlock callBlockStmt) {
	g.compileCall(callBlock.call, callBlock.span, option.Some(callBlock.macroDecl))
	g.add(emitInstruction{})
}

func (g *codeGenerator) compileDo(doTag doStmt) {
	g.compileCall(doTag.call, doTag.span, option.None[macroStmt]())
}

func (g *codeGenerator) compileIfStmt(ifCond ifCondStmt) {
	g.setLineFromSpan(ifCond.span)
	g.compileExpr(ifCond.expr)
	g.startIf()
	for _, node := range ifCond.trueBody {
		g.CompileStmt(node)
	}
	if len(ifCond.falseBody) > 0 {
		g.startElse()
		for _, node := range ifCond.falseBody {
			g.CompileStmt(node)
		}
	}
	g.endIf()
}

func (g *codeGenerator) compileExpr(exp astExpr) {
	switch exp := exp.(type) {
	case varExpr:
		g.setLineFromSpan(exp.span)
		g.add(lookupInstruction{Name: exp.id})
	case constExpr:
		g.setLineFromSpan(exp.span)
		g.add(loadConstInstruction{Val: exp.val})
	case sliceExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		if exp.start.IsSome() {
			g.compileExpr(exp.start.Unwrap())
		} else {
			g.add(loadConstInstruction{Val: valueFromI64(int64(0))})
		}
		if exp.stop.IsSome() {
			g.compileExpr(exp.stop.Unwrap())
		} else {
			g.add(loadConstInstruction{Val: none})
		}
		if exp.step.IsSome() {
			g.compileExpr(exp.step.Unwrap())
		} else {
			g.add(loadConstInstruction{Val: valueFromI64(int64(1))})
		}
		g.add(sliceInstruction{})
		g.popSpan()
	case unaryOpExpr:
		g.setLineFromSpan(exp.span)
		switch exp.op {
		case unaryOpTypeNot:
			g.compileExpr(exp.expr)
			g.add(notInstruction{})
		case unaryOpTypeNeg:
			// common case: negative numbers.  In that case we
			// directly negate them if this is possible without
			// an error.
			if c, ok := exp.expr.(constExpr); ok {
				negated, err := opNeg(c.val)
				if err == nil {
					g.add(loadConstInstruction{Val: negated})
					return
				}
			}
			g.compileExpr(exp.expr)
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
		if exp.falseExpr.IsSome() {
			g.compileExpr(exp.falseExpr.Unwrap())
		} else {
			g.add(loadConstInstruction{Val: Undefined})
		}
		g.endIf()
	case filterExpr:
		g.pushSpan(exp.span)
		if exp.expr.IsSome() {
			g.compileExpr(exp.expr.Unwrap())
		}
		for _, arg := range exp.args {
			g.compileExpr(arg)
		}
		localID := getLocalID(g.testLocalIds, exp.name)
		g.add(applyFilterInstruction{Name: exp.name, ArgCount: uint(len(exp.args)) + 1, LocalID: localID})
		g.popSpan()
	case testExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		for _, arg := range exp.args {
			g.compileExpr(arg)
		}
		localID := getLocalID(g.testLocalIds, exp.name)
		g.add(performTestInstruction{Name: exp.name, ArgCount: uint(len(exp.args)) + 1, LocalID: localID})
		g.popSpan()
	case getAttrExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.add(getAttrInstruction{Name: exp.name})
		g.popSpan()
	case getItemExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.compileExpr(exp.subscriptExpr)
		g.add(getItemInstruction{})
		g.popSpan()
	case callExpr:
		g.compileCall(exp.call, exp.span, option.None[macroStmt]())
	case listExpr:
		if v := exp.asConst(); v.IsSome() {
			g.add(loadConstInstruction{Val: v.Unwrap()})
		} else {
			g.setLineFromSpan(exp.span)
			for _, item := range exp.items {
				g.compileExpr(item)
			}
			g.add(buildListInstruction{Count: uint(len(exp.items))})
		}
	case mapExpr:
		if v := exp.asConst(); v.IsSome() {
			g.add(loadConstInstruction{Val: v.Unwrap()})
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
			g.add(buildMapInstruction{PairCount: uint(len(exp.keys))})
		}
	case kwargsExpr:
		optVal := exp.asConst()
		if optVal.IsSome() {
			g.add(loadConstInstruction{Val: optVal.Unwrap()})
		} else {
			g.setLineFromSpan(exp.span)
			for _, pair := range exp.pairs {
				g.add(loadConstInstruction{Val: valueFromString(pair.key)})
				g.compileExpr(pair.arg)
			}
			g.add(buildKwargsInstruction{PairCount: uint(len(exp.pairs))})
		}
	default:
		panic("unreachable")
	}
}

func (g *codeGenerator) compileCall(c call, spn span, caller option.Option[macroStmt]) {
	g.pushSpan(spn)
	switch ct := c.identityCall().(type) {
	case callTypeFunction:
		argCount := g.compileCallArgs(c.args, caller)
		g.add(callFunctionInstruction{Name: ct.name, ArgCount: argCount})
	case callTypeBlock:
		g.add(beginCaptureInstruction{Mode: captureModeCapture})
		g.add(callBlockInstruction{Name: ct.name})
		g.add(endCaptureInstruction{})
	case callTypeMethod:
		g.compileExpr(ct.expr)
		argCount := g.compileCallArgs(c.args, caller)
		g.add(callMethodInstruction{Name: ct.name, ArgCount: argCount + 1})
	case callTypeObject:
		g.compileExpr(ct.expr)
		argCount := g.compileCallArgs(c.args, caller)
		g.add(callObjectInstruction{ArgCount: argCount + 1})
	}
	g.popSpan()
}

func (g *codeGenerator) compileCallArgs(args []astExpr, caller option.Option[macroStmt]) uint {
	if caller.IsSome() {
		return g.compileCallArgsWithCaller(args, caller.Unwrap())
	}
	for _, arg := range args {
		g.compileExpr(arg)
	}
	return uint(len(args))
}

func (g *codeGenerator) compileCallArgsWithCaller(args []astExpr, caller macroStmt) uint {
	injectedCaller := false

	// try to add the caller to already existing keyword arguments.
	for _, arg := range args {
		if m, ok := arg.(kwargsExpr); ok {
			g.setLineFromSpan(m.span)
			for _, pair := range m.pairs {
				g.add(loadConstInstruction{Val: valueFromString(pair.key)})
				g.compileExpr(pair.arg)
			}
			g.add(loadConstInstruction{Val: valueFromString("caller")})
			g.compileMacroExpression(caller)
			g.add(buildKwargsInstruction{PairCount: uint(len(m.pairs)) + 1})
			injectedCaller = true
		} else {
			g.compileExpr(arg)
		}
	}

	// if there are no keyword args so far, create a new kwargs object
	// and add caller to that.
	if !injectedCaller {
		g.add(loadConstInstruction{Val: valueFromString("caller")})
		g.compileMacroExpression(caller)
		g.add(buildKwargsInstruction{PairCount: 1})
		return uint(len(args)) + 1
	}
	return uint(len(args))
}

func (g *codeGenerator) startForLoop(withLoopVar, recursive bool) {
	flags := uint8(0)
	if withLoopVar {
		flags |= loopFlagWithLoopVar
	}
	if recursive {
		flags |= loopFlagRecursive
	}
	g.add(pushLoopInstruction{Flags: flags})
	iterInst := g.add(iterateInstruction{JumpTarget: 0})
	g.pendingBlock.Push(loopPendingBlock{iterInst: iterInst})
}

func (g *codeGenerator) endForLoop(pushDidNotIterate bool) {
	b := g.pendingBlock.Pop()
	if b == nil {
		panic("pendingBlock should not be empty in endForLoop")
	}
	if b, ok := b.(loopPendingBlock); ok {
		g.add(jumpInstruction{JumpTarget: b.iterInst})
		loopEnd := g.nextInstruction()
		if pushDidNotIterate {
			g.add(pushDidNotIterateInstruction{})
		}
		g.add(popFrameInstruction{})
		if _, ok := g.instructions.instructions[b.iterInst].(iterateInstruction); ok {
			g.instructions.instructions[b.iterInst] = iterateInstruction{JumpTarget: loopEnd}
		} else {
			panic("must be iterateInstruction")
		}
	} else {
		panic("must be loopPendingBlock")
	}
}

func (g *codeGenerator) startIf() {
	jumpInst := g.add(jumpIfFalseInstruction{JumpTarget: 0})
	g.pendingBlock.Push(branchPendingBlock{jumpInst: jumpInst})
}

func (g *codeGenerator) startElse() {
	jumpInst := g.add(jumpInstruction{JumpTarget: 0})
	g.endCondition(jumpInst + 1)
	g.pendingBlock.Push(branchPendingBlock{jumpInst: jumpInst})
}

func (g *codeGenerator) endIf() {
	g.endCondition(g.nextInstruction())
}

// Starts a short cirquited bool block.
func (g *codeGenerator) startScBool() {
	g.pendingBlock.Push(scBoolPendingBlock{})
}

// Emits a short circuited bool operator.
func (g *codeGenerator) scBool(and bool) {
	if blk, ok := g.pendingBlock.Peek(); ok {
		var inst instruction
		if and {
			inst = jumpIfFalseOrPopInstruction{JumpTarget: ^uint(0)}
		} else {
			inst = jumpIfTrueOrPopInstruction{JumpTarget: ^uint(0)}
		}
		instIdx := g.instructions.add(inst)
		scBoolBlk := blk.(scBoolPendingBlock)
		scBoolBlk.instructions = append(scBoolBlk.instructions, instIdx)
		g.pendingBlock[len(g.pendingBlock)-1] = scBoolBlk
	}
}

// Ends a short circuited bool block.
func (g *codeGenerator) endScBool() {
	end := g.nextInstruction()
	if !g.pendingBlock.Empty() {
		blk := g.pendingBlock.Pop()
		if scBoolBlk, ok := blk.(scBoolPendingBlock); ok {
			for _, instIdx := range scBoolBlk.instructions {
				switch g.instructions.instructions[instIdx].(type) {
				case jumpIfFalseOrPopInstruction:
					g.instructions.instructions[instIdx] = jumpIfFalseOrPopInstruction{JumpTarget: end}
				case jumpIfTrueOrPopInstruction:
					g.instructions.instructions[instIdx] = jumpIfTrueOrPopInstruction{JumpTarget: end}
				default:
					panic("unreachable")
				}
			}
		}
	}
}

func (g *codeGenerator) endCondition(jumpInst uint) {
	if g.pendingBlock.Empty() {
		panic("pendingBlock should not be empty in endCondition")
	}
	b := g.pendingBlock.Pop()
	if b, ok := b.(branchPendingBlock); ok {
		switch g.instructions.instructions[b.jumpInst].(type) {
		case jumpIfFalseInstruction:
			g.instructions.instructions[b.jumpInst] = jumpIfFalseInstruction{JumpTarget: jumpInst}
		case jumpInstruction:
			g.instructions.instructions[b.jumpInst] = jumpInstruction{JumpTarget: jumpInst}
		}
	} else {
		panic("must be branchPendingBlock")
	}
}

func (g *codeGenerator) Finish() (instructions, map[string]instructions) {
	return g.instructions, g.blocks
}

func (g *codeGenerator) setLine(lineno uint32) {
	g.currentLine = lineno
}

func (g *codeGenerator) setLineFromSpan(spn span) {
	g.setLine(spn.StartLine)
}

func (g *codeGenerator) pushSpan(spn span) {
	g.spanStack.Push(spn)
	g.setLineFromSpan(spn)
}

func (g *codeGenerator) popSpan() {
	g.spanStack.Pop()
}

func (g *codeGenerator) add(instr instruction) uint {
	if spn, ok := g.spanStack.Peek(); ok {
		if spn.StartLine == g.currentLine {
			return g.instructions.addWithSpan(instr, spn)
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

func (g *codeGenerator) newSubgenerator() *codeGenerator {
	sub := newCodeGenerator(g.instructions.name, g.instructions.source)
	sub.currentLine = g.currentLine
	if !g.spanStack.Empty() {
		v, _ := g.spanStack.Peek()
		sub.spanStack.Push(v)
	}
	return sub
}

func (g *codeGenerator) finishSubgenerator(sub *codeGenerator) instructions {
	g.currentLine = sub.currentLine
	insts, blocks := sub.finish()
	for name, block := range blocks {
		g.blocks[name] = block
	}
	return insts
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
		g.startScBool()
		g.compileExpr(exp.left)
		g.scBool(exp.op == binOpTypeScAnd)
		g.compileExpr(exp.right)
		g.endScBool()
		g.popSpan()
		return
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

func (g *codeGenerator) finish() (instructions, map[string]instructions) {
	if !g.pendingBlock.Empty() {
		panic("unreachable")
	}
	return g.instructions, g.blocks
}

func getLocalID(ids map[string]localID, name string) localID {
	if id, ok := ids[name]; ok {
		return id
	} else if len(ids) >= maxLocals {
		return ^localID(0)
	} else {
		nextID := localID(len(ids))
		ids[name] = nextID
		return nextID
	}
}
