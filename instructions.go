package mjingo

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

type emitRawInstructionData string
type storeLocalInstructionData string
type lookupInstructionData string
type getAttrInstructionData string
type loatConstInstructionData value
type buildMapInstructionData uint
type buildKwargsInstructionData uint
type buildListInstructionData uint
type unpackListInstructionData uint

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
