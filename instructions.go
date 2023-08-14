package mjingo

import (
	"cmp"
	"slices"
)

type instructions struct {
	instructions []instruction
	lineInfos    []lineInfo
	spanInfos    []spanInfo
	name         string
	source       string
}

var emptyInstructions = instructions{
	name: "<unknown>",
}

type instructionKind uint

const (
	// Emits raw source
	instructionKindEmitRaw instructionKind = iota + 1

	// Stores a variable (only possible in for loops)
	instructionKindStoreLocal

	// Load a variable
	instructionKindLookup

	// Looks up an attribute.
	instructionKindGetAttr

	// Looks up an item.
	instructionKindGetItem

	// Performs a slice operation.
	instructionKindSlice

	// Loads a constant value.
	instructionKindLoadConst

	// Builds a map of the last n pairs on the stack.
	instructionKindBuildMap

	// Builds a kwargs map of the last n pairs on the stack.
	instructionKindBuildKwargs

	// Builds a list of the last n pairs on the stack.
	instructionKindBuildList

	// Unpacks a list into N stack items.
	instructionKindUnpackList

	// Appends to the list.
	instructionKindListAppend

	// Add the top two values
	instructionKindAdd

	// Subtract the top two values
	instructionKindSub

	// Multiply the top two values
	instructionKindMul

	// Divide the top two values
	instructionKindDiv

	// Integer divide the top two values as "integer".
	//
	// Note that in MiniJinja this currently uses an euclidean
	// division to match the rem implementation.  In Python this
	// instead uses a flooring division and a flooring remainder.
	instructionKindIntDiv

	// Calculate the remainder the top two values
	instructionKindRem

	// x to the power of y.
	instructionKindPow

	// Negates the value.
	instructionKindNeg

	// `=` operator
	instructionKindEq

	// `!=` operator
	instructionKindNe

	// `>` operator
	instructionKindGt

	// `>=` operator
	instructionKindGte

	// `<` operator
	instructionKindLt

	// `<=` operator
	instructionKindLte

	// Unary not
	instructionKindNot

	// String concatenation operator
	instructionKindStringConcat

	// Performs a containment check
	instructionKindIn

	// Apply a filter.
	instructionKindApplyFilter

	// Perform a filter.
	instructionKindPerformTest

	// Emit the stack top as output
	instructionKindEmit

	// Starts a loop
	//
	// The argument are loop flags.
	instructionKindPushLoop

	// Starts a with block.
	instructionKindPushWith

	// Does a single loop iteration
	//
	// The argument is the jump target for when the loop
	// ends and must point to a `PopFrame` instruction.
	instructionKindIterate

	// Push a bool that indicates that the loop iterated.
	instructionKindPushDidNotIterate

	// Pops the topmost frame
	instructionKindPopFrame

	// Jump to a specific instruction
	instructionKindJump

	// Jump if the stack top evaluates to false
	instructionKindJumpIfFalse

	// Jump if the stack top evaluates to false or pops the value
	instructionKindJumpIfFalseOrPop

	// Jump if the stack top evaluates to true or pops the value
	instructionKindJumpIfTrueOrPop

	// Sets the auto escape flag to the current value.
	instructionKindPushAutoEscape

	// Resets the auto escape flag to the previous value.
	instructionKindPopAutoEscape

	// Begins capturing of output (false) or discard (true).
	instructionKindBeginCapture

	// Ends capturing of output.
	instructionKindEndCapture

	// Calls a global function
	instructionKindCallFunction

	// Calls a method
	instructionKindCallMethod

	// Calls an object
	instructionKindCallObject

	// Duplicates the top item
	instructionKindDupTop

	// Discards the top item
	instructionKindDiscardTop

	// A fast super instruction without intermediate capturing.
	instructionKindFastSuper

	// A fast loop recurse instruction without intermediate capturing.
	instructionKindFastRecurse

	// Call into a block.
	instructionKindCallBlock

	// Loads block from a template with name on stack ("extends")
	instructionKindLoadBlocks

	// Includes another template.
	instructionKindInclude

	// Builds a module
	instructionKindExportLocals

	// Builds a macro on the stack.
	instructionKindBuildMacro

	// Breaks from the interpreter loop (exists a function)
	instructionKindReturn

	// True if the value is undefined
	instructionKindIsUndefined

	// Encloses a variable.
	instructionKindEnclose

	// Returns the closure of this context level.
	instructionKindGetClosure
)

func (k instructionKind) String() string {
	switch k {
	case instructionKindEmitRaw:
		return "emitRaw"
	case instructionKindStoreLocal:
		return "storeLocal"
	case instructionKindLookup:
		return "lookup"
	case instructionKindGetAttr:
		return "getAttr"
	case instructionKindGetItem:
		return "getItem"
	case instructionKindSlice:
		return "slice"
	case instructionKindLoadConst:
		return "loadConst"
	case instructionKindBuildMap:
		return "buildMap"
	case instructionKindBuildKwargs:
		return "buildKwargs"
	case instructionKindBuildList:
		return "buildList"
	case instructionKindUnpackList:
		return "unpackList"
	case instructionKindListAppend:
		return "listAppend"
	case instructionKindAdd:
		return "add"
	case instructionKindSub:
		return "sub"
	case instructionKindMul:
		return "mul"
	case instructionKindDiv:
		return "div"
	case instructionKindIntDiv:
		return "intDiv"
	case instructionKindRem:
		return "rem"
	case instructionKindPow:
		return "pow"
	case instructionKindNeg:
		return "neg"
	case instructionKindEq:
		return "eq"
	case instructionKindNe:
		return "ne"
	case instructionKindGt:
		return "gt"
	case instructionKindGte:
		return "gte"
	case instructionKindLt:
		return "lt"
	case instructionKindLte:
		return "lte"
	case instructionKindNot:
		return "not"
	case instructionKindStringConcat:
		return "stringConcat"
	case instructionKindIn:
		return "in"
	case instructionKindApplyFilter:
		return "applyFilter"
	case instructionKindPerformTest:
		return "performTest"
	case instructionKindEmit:
		return "emit"
	case instructionKindPushLoop:
		return "pushLoop"
	case instructionKindPushWith:
		return "pushWith"
	case instructionKindIterate:
		return "iterate"
	case instructionKindPushDidNotIterate:
		return "pushDidNotIterate"
	case instructionKindPopFrame:
		return "popFrame"
	case instructionKindJump:
		return "jump"
	case instructionKindJumpIfFalse:
		return "jumpIfFalse"
	case instructionKindJumpIfFalseOrPop:
		return "jumpIfFalseOrPop"
	case instructionKindJumpIfTrueOrPop:
		return "jumpIfTrueOrPop"
	case instructionKindPushAutoEscape:
		return "pushAutoEscape"
	case instructionKindPopAutoEscape:
		return "popAutoEscape"
	case instructionKindBeginCapture:
		return "beginCapture"
	case instructionKindEndCapture:
		return "endCapture"
	case instructionKindCallFunction:
		return "callFunction"
	case instructionKindCallMethod:
		return "callMethod"
	case instructionKindCallObject:
		return "callObject"
	case instructionKindDupTop:
		return "dupTop"
	case instructionKindDiscardTop:
		return "discardTop"
	case instructionKindFastSuper:
		return "fastSuper"
	case instructionKindFastRecurse:
		return "fastRecurse"
	case instructionKindCallBlock:
		return "callBlock"
	case instructionKindLoadBlocks:
		return "loadBlocks"
	case instructionKindInclude:
		return "include"
	case instructionKindExportLocals:
		return "exportLocals"
	case instructionKindBuildMacro:
		return "buildMacro"
	case instructionKindReturn:
		return "return"
	case instructionKindIsUndefined:
		return "isUndefined"
	case instructionKindEnclose:
		return "enclose"
	case instructionKindGetClosure:
		return "getClosure"
	default:
		panic("invalid instructionKind")
	}
}

type emitRawInstructionData = string
type storeLocalInstructionData = string
type lookupInstructionData = string
type getAttrInstructionData = string
type loadConstInstructionData = value
type buildMapInstructionData = uint
type buildKwargsInstructionData = uint
type buildListInstructionData = uint
type unpackListInstructionData = uint

type applyFilterInstructionData struct {
	str     string
	size    uint
	localId localId
}

type performTestInstructionData struct {
	str     string
	size    uint
	localId localId
}

type iterateInstructionData uint
type jumpInstructionData uint
type jumpIfFalseInstructionData uint
type jumpIfFalseOrPopInstructionData uint
type jumpIfTrueOrPopInstructionData uint
type beginCaptureInstructionData captureMode

type callFunctionInstructionData struct {
	name string
	size uint
}

type callMethodInstructionData struct {
	name string
	size uint
}

type callObjectInstructionData uint
type callBlockInstructionData string
type includeInstructionData bool

type buildMacroInstructionData struct {
	name string
	size uint
	kind uint8
}

type encloseInstructionData string

type instruction struct {
	kind instructionKind
	data any
}

type lineInfo struct {
	firstInstruction uint32
	line             uint32
}

type spanInfo struct {
	firstInstruction uint32
	span             option[span]
}

type localId uint8

func newInstructions(name, source string) instructions {
	return instructions{
		instructions: make([]instruction, 0, 128),
		lineInfos:    make([]lineInfo, 0, 128),
		spanInfos:    make([]spanInfo, 0, 128),
		name:         name,
		source:       source,
	}
}

func (i *instructions) add(instr instruction) uint {
	rv := len(i.instructions)
	i.instructions = append(i.instructions, instr)
	return uint(rv)
}

func (i *instructions) addLineRecord(instr uint, line uint32) {
	sameLoc := false
	if len(i.lineInfos) > 0 {
		lastLoc := i.lineInfos[len(i.lineInfos)-1]
		sameLoc = lastLoc.line == line
	}
	if !sameLoc {
		i.lineInfos = append(i.lineInfos, lineInfo{firstInstruction: uint32(instr), line: line})
	}
}

func (i *instructions) addWithLine(instr instruction, line uint32) uint {
	rv := i.add(instr)
	i.addLineRecord(rv, line)
	return rv
}

func (i *instructions) addWithSpan(instr instruction, spn span) uint {
	rv := i.add(instr)

	sameLoc := false
	if len(i.spanInfos) > 0 {
		lastLoc := i.spanInfos[len(i.spanInfos)-1]
		sameLoc = lastLoc.span.valid && lastLoc.span.data == spn
	}
	if !sameLoc {
		i.spanInfos = append(i.spanInfos,
			spanInfo{firstInstruction: uint32(rv), span: option[span]{valid: true, data: spn}})
	}

	i.addLineRecord(rv, spn.startLine)
	return rv
}

func (i *instructions) getLine(idx uint) option[uint] {
	n, found := slices.BinarySearchFunc(i.lineInfos,
		lineInfo{firstInstruction: uint32(idx)},
		func(a, b lineInfo) int {
			return cmp.Compare(a.firstInstruction, b.firstInstruction)
		})
	if found {
		return option[uint]{valid: true, data: uint(i.lineInfos[n].line)}
	}
	if n != 0 {
		return option[uint]{valid: true, data: uint(i.lineInfos[n-1].line)}
	}
	return option[uint]{}
}

func (i *instructions) getSpan(idx uint) option[span] {
	n, found := slices.BinarySearchFunc(i.spanInfos,
		spanInfo{firstInstruction: uint32(idx)},
		func(a, b spanInfo) int {
			return cmp.Compare(a.firstInstruction, b.firstInstruction)
		})
	if found {
		return i.spanInfos[n].span
	}
	if n != 0 {
		return i.spanInfos[n-1].span
	}
	return option[span]{}
}
