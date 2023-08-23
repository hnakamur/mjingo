package internal

import (
	"fmt"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type codeGenerator struct {
	instructions     Instructions
	blocks           map[string]Instructions
	pendingBlock     stack[pendingBlock]
	currentLine      uint32
	spanStack        stack[Span]
	filterLocalIds   map[string]LocalID
	testLocalIds     map[string]LocalID
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

func NewCodeGenerator(file, source string) *codeGenerator {
	return &codeGenerator{
		instructions:   newInstructions(file, source),
		blocks:         make(map[string]Instructions),
		pendingBlock:   newStackWithCapacity[pendingBlock](32),
		filterLocalIds: make(map[string]LocalID),
		testLocalIds:   make(map[string]LocalID),
	}
}

func (g *codeGenerator) CompileStmt(stmt statement) {
	switch st := stmt.(type) {
	case templateStmt:
		for _, node := range st.children {
			g.CompileStmt(node)
		}
	case emitExprStmt:
		g.compileEmitExpr(emitExprStmt{expr: st.expr, span: st.span})
	case emitRawStmt:
		g.add(EmitRawInstruction{Val: st.raw})
		g.rawTemplateBytes += uint(len(st.raw))
	case forLoopStmt:
		g.compileForLoop(st)
	case ifCondStmt:
		g.compileIfStmt(st)
	case withBlockStmt:
		g.setLineFromSpan(st.span)
		for _, assign := range st.assignments {
			g.compileExpr(assign.rhs)
			g.compileAssignment(assign.lhs)
		}
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(PopFrameInstruction{})
	case setStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.expr)
		g.compileAssignment(st.target)
	case setBlockStmt:
		g.setLineFromSpan(st.span)
		g.add(BeginCaptureInstruction{Mode: CaptureModeCapture})
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(EndCaptureInstruction{})
		if option.IsSome(st.filter) {
			g.compileExpr(option.Unwrap(st.filter))
		}
		g.compileAssignment(st.target)
	case autoEscapeStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.enabled)
		g.add(PushAutoEscapeInstruction{})
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(PopAutoEscapeInstruction{})
	case filterBlockStmt:
		g.setLineFromSpan(st.span)
		g.add(BeginCaptureInstruction{Mode: CaptureModeCapture})
		for _, node := range st.body {
			g.CompileStmt(node)
		}
		g.add(EndCaptureInstruction{})
		g.compileExpr(st.filter)
		g.add(EmitInstruction{})
	case blockStmt:
		g.compileBlock(st)
	case extendsStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.name)
		g.addWithSpan(LoadBlocksInstruction{}, st.span)
	case includeStmt:
		g.setLineFromSpan(st.span)
		g.compileExpr(st.name)
		g.addWithSpan(IncludeInstruction{IgnoreMissing: st.ignoreMissing}, st.span)
	case macroStmt:
		g.compileMacro(st)
	case callBlockStmt:
		g.compileCallBlock(st)
	case doStmt:
		g.compileDo(st)
	default:
		panic(fmt.Sprintf("not implemented, st=%+v (%T)", st, st))
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
	g.add(CallBlockInstruction{Name: block.name})
}

func (g *codeGenerator) compileEmitExpr(exp emitExprStmt) {
	g.setLineFromSpan(exp.span)

	if callExpr, ok := exp.expr.(callExpr); ok {
		switch ct := callExpr.call.identityCall().(type) {
		case callTypeFunction:
			if ct.name == "super" && len(callExpr.call.args) == 0 {
				g.addWithSpan(FastSuperInstruction{}, callExpr.span)
				return
			} else if ct.name == "loop" && len(callExpr.call.args) == 1 {
				g.compileExpr(callExpr.call.args[0])
				g.add(FastRecurseInstruction{})
				return
			}
		case callTypeBlock:
			g.add(CallBlockInstruction{Name: ct.name})
			return
		}
	}
	g.compileExpr(exp.expr)
	g.add(EmitInstruction{})
}

func (g *codeGenerator) compileForLoop(forLoop forLoopStmt) {
	g.setLineFromSpan(forLoop.span)
	if option.IsSome(forLoop.filterExpr) {
		// filter expressions work like a nested for loop without
		// the special loop variable that append into a new list
		// just outside of the loop.
		g.add(BuildListInstruction{Count: 0})
		g.compileExpr(forLoop.iter)
		g.startForLoop(false, false)
		g.add(DupTopInstruction{})
		g.compileAssignment(forLoop.target)
		g.compileExpr(option.Unwrap(forLoop.filterExpr))
		g.startIf()
		g.add(ListAppendInstruction{})
		g.startElse()
		g.add(DiscardTopInstruction{})
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
func (g *codeGenerator) compileAssignment(expr expression) {
	switch exp := expr.(type) {
	case varExpr:
		g.add(StoreLocalInstruction{Name: exp.id})
	case listExpr:
		g.pushSpan(exp.span)
		g.add(UnpackListInstruction{Count: uint(len(exp.items))})
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
	inst := g.add(JumpInstruction{JumpTarget: ^uint(0)})
	j := len(macroDecl.defaults) - 1
	for i := len(macroDecl.args) - 1; i >= 0; i-- {
		if j >= 0 {
			g.add(DupTopInstruction{})
			g.add(IsUndefinedInstruction{})
			g.startIf()
			g.add(DiscardTopInstruction{})
			g.compileExpr(macroDecl.defaults[j])
			g.endIf()
			j--
		}
		g.compileAssignment(macroDecl.args[i])
	}
	for _, node := range macroDecl.body {
		g.CompileStmt(node)
	}
	g.add(ReturnInstruction{})
	undeclared := findMacroClosure(macroDecl)
	callerReference := undeclared.Contains("caller")
	undeclared.Delete("caller")
	macroInst := g.nextInstruction()
	for _, name := range undeclared.Keys() {
		g.add(EncloseInstruction{Name: name})
	}
	g.add(GetClosureInstruction{})
	ids := make([]Value, 0, len(macroDecl.args))
	for _, arg := range macroDecl.args {
		if varExp, ok := arg.(varExpr); ok {
			ids = append(ids, ValueFromString(varExp.id))
		} else {
			panic("unreachable")
		}
	}
	g.add(LoadConstInstruction{Val: ValueFromSlice(ids)})
	flags := uint8(0)
	if callerReference {
		flags |= macroCaller
	}
	g.add(BuildMacroInstruction{Name: macroDecl.name, Offset: inst + 1, Flags: flags})
	if g.instructions.instructions[inst].Typ() == instTypeJump {
		g.instructions.instructions[inst] = JumpInstruction{JumpTarget: macroInst}
	} else {
		panic("unreachable")
	}
}

func (g *codeGenerator) compileMacro(macroDecl macroStmt) {
	g.compileMacroExpression(macroDecl)
	g.add(StoreLocalInstruction{Name: macroDecl.name})
}

func (g *codeGenerator) compileCallBlock(callBlock callBlockStmt) {
	g.compileCall(callBlock.call, callBlock.span, option.Some(callBlock.macroDecl))
	g.add(EmitInstruction{})
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

func (g *codeGenerator) compileExpr(exp expression) {
	switch exp := exp.(type) {
	case varExpr:
		g.setLineFromSpan(exp.span)
		g.add(LookupInstruction{Name: exp.id})
	case constExpr:
		g.setLineFromSpan(exp.span)
		g.add(LoadConstInstruction{Val: exp.val})
	case sliceExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		if option.IsSome(exp.start) {
			g.compileExpr(option.Unwrap(exp.start))
		} else {
			g.add(LoadConstInstruction{Val: ValueFromI64(int64(0))})
		}
		if option.IsSome(exp.stop) {
			g.compileExpr(option.Unwrap(exp.stop))
		} else {
			g.add(LoadConstInstruction{Val: None})
		}
		if option.IsSome(exp.step) {
			g.compileExpr(option.Unwrap(exp.step))
		} else {
			g.add(LoadConstInstruction{Val: ValueFromI64(int64(1))})
		}
		g.add(SliceInstruction{})
		g.popSpan()
	case unaryOpExpr:
		g.setLineFromSpan(exp.span)
		g.compileExpr(exp.expr)
		switch exp.op {
		case unaryOpTypeNot:
			g.add(NotInstruction{})
		case unaryOpTypeNeg:
			g.addWithSpan(NegInstruction{}, exp.span)
		}
	case binOpExpr:
		g.compileBinOp(exp)
	case ifExpr:
		g.setLineFromSpan(exp.span)
		g.compileExpr(exp.testExpr)
		g.startIf()
		g.compileExpr(exp.trueExpr)
		g.startElse()
		if option.IsSome(exp.falseExpr) {
			g.compileExpr(option.Unwrap(exp.falseExpr))
		} else {
			g.add(LoadConstInstruction{Val: Undefined})
		}
		g.endIf()
	case filterExpr:
		g.pushSpan(exp.span)
		if option.IsSome(exp.expr) {
			g.compileExpr(option.Unwrap(exp.expr))
		}
		for _, arg := range exp.args {
			g.compileExpr(arg)
		}
		localID := getLocalID(g.testLocalIds, exp.name)
		g.add(ApplyFilterInstruction{Name: exp.name, ArgCount: uint(len(exp.args)) + 1, LocalID: localID})
		g.popSpan()
	case testExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		for _, arg := range exp.args {
			g.compileExpr(arg)
		}
		localID := getLocalID(g.testLocalIds, exp.name)
		g.add(PerformTestInstruction{Name: exp.name, ArgCount: uint(len(exp.args)) + 1, LocalID: localID})
		g.popSpan()
	case getAttrExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.add(GetAttrInstruction{Name: exp.name})
		g.popSpan()
	case getItemExpr:
		g.pushSpan(exp.span)
		g.compileExpr(exp.expr)
		g.compileExpr(exp.subscriptExpr)
		g.add(GetItemInstruction{})
		g.popSpan()
	case callExpr:
		g.compileCall(exp.call, exp.span, option.None[macroStmt]())
	case listExpr:
		if v := exp.asConst(); option.IsSome(v) {
			g.add(LoadConstInstruction{Val: option.Unwrap(v)})
		} else {
			g.setLineFromSpan(exp.span)
			for _, item := range exp.items {
				g.compileExpr(item)
			}
			g.add(BuildListInstruction{Count: uint(len(exp.items))})
		}
	case mapExpr:
		if v := exp.asConst(); option.IsSome(v) {
			g.add(LoadConstInstruction{Val: option.Unwrap(v)})
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
			g.add(BuildMapInstruction{PairCount: uint(len(exp.keys))})
		}
	case kwargsExpr:
		optVal := exp.asConst()
		if option.IsSome(optVal) {
			g.add(LoadConstInstruction{Val: option.Unwrap(optVal)})
		} else {
			g.setLineFromSpan(exp.span)
			for _, pair := range exp.pairs {
				g.add(LoadConstInstruction{Val: ValueFromString(pair.key)})
				g.compileExpr(pair.arg)
			}
			g.add(BuildKwargsInstruction{PairCount: uint(len(exp.pairs))})
		}
	default:
		panic(fmt.Sprintf("not implemented for exprType: %s", exp.typ()))
	}
}

func (g *codeGenerator) compileCall(c call, spn Span, caller option.Option[macroStmt]) {
	g.pushSpan(spn)
	switch ct := c.identityCall().(type) {
	case callTypeFunction:
		argCount := g.compileCallArgs(c.args, caller)
		g.add(CallFunctionInstruction{Name: ct.name, ArgCount: argCount})
	case callTypeBlock:
		g.add(BeginCaptureInstruction{Mode: CaptureModeCapture})
		g.add(CallBlockInstruction{Name: ct.name})
		g.add(EndCaptureInstruction{})
	case callTypeMethod:
		g.compileExpr(ct.expr)
		argCount := g.compileCallArgs(c.args, caller)
		g.add(CallMethodInstruction{Name: ct.name, ArgCount: argCount + 1})
	case callTypeObject:
		g.compileExpr(ct.expr)
		argCount := g.compileCallArgs(c.args, caller)
		g.add(CallObjectInstruction{ArgCount: argCount + 1})
	}
	g.popSpan()
}

func (g *codeGenerator) compileCallArgs(args []expression, caller option.Option[macroStmt]) uint {
	if option.IsSome(caller) {
		return g.compileCallArgsWithCaller(args, option.Unwrap(caller))
	}
	for _, arg := range args {
		g.compileExpr(arg)
	}
	return uint(len(args))
}

func (g *codeGenerator) compileCallArgsWithCaller(args []expression, caller macroStmt) uint {
	injectedCaller := false

	// try to add the caller to already existing keyword arguments.
	for _, arg := range args {
		if m, ok := arg.(kwargsExpr); ok {
			g.setLineFromSpan(m.span)
			for _, pair := range m.pairs {
				g.add(LoadConstInstruction{Val: ValueFromString(pair.key)})
				g.compileExpr(pair.arg)
			}
			g.add(LoadConstInstruction{Val: ValueFromString("caller")})
			g.compileMacroExpression(caller)
			g.add(BuildKwargsInstruction{PairCount: uint(len(m.pairs)) + 1})
			injectedCaller = true
		} else {
			g.compileExpr(arg)
		}
	}

	// if there are no keyword args so far, create a new kwargs object
	// and add caller to that.
	if !injectedCaller {
		g.add(LoadConstInstruction{Val: ValueFromString("caller")})
		g.compileMacroExpression(caller)
		g.add(BuildKwargsInstruction{PairCount: 1})
		return uint(len(args)) + 1
	}
	return uint(len(args))
}

func (g *codeGenerator) startForLoop(withLoopVar, recursive bool) {
	flags := uint8(0)
	if withLoopVar {
		flags |= LoopFlagWithLoopVar
	}
	if recursive {
		flags |= LoopFlagRecursive
	}
	g.add(PushLoopInstruction{Flags: flags})
	iterInst := g.add(IterateInstruction{JumpTarget: 0})
	g.pendingBlock.push(loopPendingBlock{iterInst: iterInst})
}

func (g *codeGenerator) endForLoop(pushDidNotIterate bool) {
	b := g.pendingBlock.pop()
	if b == nil {
		panic("pendingBlock should not be empty in endForLoop")
	}
	if b, ok := (*b).(loopPendingBlock); ok {
		g.add(JumpInstruction{JumpTarget: b.iterInst})
		loopEnd := g.nextInstruction()
		if pushDidNotIterate {
			g.add(PushDidNotIterateInstruction{})
		}
		g.add(PopFrameInstruction{})
		if _, ok := g.instructions.instructions[b.iterInst].(IterateInstruction); ok {
			g.instructions.instructions[b.iterInst] = IterateInstruction{JumpTarget: loopEnd}
		} else {
			panic("must be iterateInstruction")
		}
	} else {
		panic("must be loopPendingBlock")
	}
}

func (g *codeGenerator) startIf() {
	jumpInst := g.add(JumpIfFalseInstruction{JumpTarget: 0})
	g.pendingBlock.push(branchPendingBlock{jumpInst: jumpInst})
}

func (g *codeGenerator) startElse() {
	jumpInst := g.add(JumpInstruction{JumpTarget: 0})
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
		case JumpIfFalseInstruction:
			g.instructions.instructions[b.jumpInst] = JumpIfFalseInstruction{JumpTarget: jumpInst}
		case JumpInstruction:
			g.instructions.instructions[b.jumpInst] = JumpInstruction{JumpTarget: jumpInst}
		}
	} else {
		panic("must be branchPendingBlock")
	}
}

func (g *codeGenerator) Finish() (Instructions, map[string]Instructions) {
	return g.instructions, g.blocks
}

func (g *codeGenerator) setLine(lineno uint32) {
	g.currentLine = lineno
}

func (g *codeGenerator) setLineFromSpan(spn Span) {
	g.setLine(spn.StartLine)
}

func (g *codeGenerator) pushSpan(spn Span) {
	g.spanStack.push(spn)
	g.setLineFromSpan(spn)
}

func (g *codeGenerator) popSpan() {
	g.spanStack.pop()
}

func (g *codeGenerator) add(instr Instruction) uint {
	if spn := g.spanStack.peek(); spn != nil {
		if spn.StartLine == g.currentLine {
			return g.instructions.addWithSpan(instr, *spn)
		}
	}
	return g.instructions.addWithLine(instr, g.currentLine)
}

func (g *codeGenerator) addWithSpan(instr Instruction, spn Span) uint {
	return g.instructions.addWithSpan(instr, spn)
}

func (g *codeGenerator) nextInstruction() uint {
	return uint(len(g.instructions.instructions))
}

func (g *codeGenerator) newSubgenerator() *codeGenerator {
	sub := NewCodeGenerator(g.instructions.name, g.instructions.source)
	sub.currentLine = g.currentLine
	if !g.spanStack.empty() {
		sub.spanStack.push(*g.spanStack.peek())
	}
	return sub
}

func (g *codeGenerator) finishSubgenerator(sub *codeGenerator) Instructions {
	g.currentLine = sub.currentLine
	insts, blocks := sub.finish()
	for name, block := range blocks {
		g.blocks[name] = block
	}
	return insts
}

func (g *codeGenerator) compileBinOp(exp binOpExpr) {
	g.pushSpan(exp.span)
	var instr Instruction
	switch exp.op {
	case binOpTypeEq:
		instr = EqInstruction{}
	case binOpTypeNe:
		instr = NeInstruction{}
	case binOpTypeLt:
		instr = LtInstruction{}
	case binOpTypeLte:
		instr = LteInstruction{}
	case binOpTypeGt:
		instr = GtInstruction{}
	case binOpTypeGte:
		instr = GteInstruction{}
	case binOpTypeScAnd, binOpTypeScOr:
		panic("not implemented yet")
	case binOpTypeAdd:
		instr = AddInstruction{}
	case binOpTypeSub:
		instr = SubInstruction{}
	case binOpTypeMul:
		instr = MulInstruction{}
	case binOpTypeDiv:
		instr = DivInstruction{}
	case binOpTypeFloorDiv:
		instr = IntDivInstruction{}
	case binOpTypeRem:
		instr = RemInstruction{}
	case binOpTypePow:
		instr = PowInstruction{}
	case binOpTypeConcat:
		instr = StringConcatInstruction{}
	case binOpTypeIn:
		instr = InInstruction{}
	}
	g.compileExpr(exp.left)
	g.compileExpr(exp.right)
	g.add(instr)
	g.popSpan()
}

func (g *codeGenerator) finish() (Instructions, map[string]Instructions) {
	if !g.pendingBlock.empty() {
		panic("unreachable")
	}
	return g.instructions, g.blocks
}

func getLocalID(ids map[string]LocalID, name string) LocalID {
	if id, ok := ids[name]; ok {
		return id
	} else if len(ids) >= MaxLocals {
		return ^LocalID(0)
	} else {
		nextID := LocalID(len(ids))
		ids[name] = nextID
		return nextID
	}
}
