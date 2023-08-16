package mjingo

import (
	"cmp"
	"slices"
)

const (
	// / This loop has the loop var.
	loopFlagWithLoopVar = 1

	// / This loop is recursive.
	loopFlagRecursive = 2
)

// Go type to represent locals.
type localId = uint8

// The maximum number of filters/tests that can be cached.
const maxLocals = 50

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

type instruction interface {
	typ() instType
}

type emitRawInstruction struct{ val string }
type storeLocalInstruction struct{ name string }
type lookupInstruction struct{ name string }
type getAttrInstruction struct{ name string }
type getItemInstruction struct{}
type sliceInstruction struct{}
type loadConstInstruction struct{ val value }
type buildMapInstruction struct{ pairCount uint }
type buildKwargsInstruction struct{ pairCount uint }
type buildListInstruction struct{ count uint }
type unpackListInstruction struct{ count uint }
type listAppendInstruction struct{}
type addInstruction struct{}
type subInstruction struct{}
type mulInstruction struct{}
type divInstruction struct{}
type intDivInstruction struct{}
type remInstruction struct{}
type powInstruction struct{}
type negInstruction struct{}
type eqInstruction struct{}
type neInstruction struct{}
type gtInstruction struct{}
type gteInstruction struct{}
type ltInstruction struct{}
type lteInstruction struct{}
type notInstruction struct{}
type stringConcatInstruction struct{}
type inInstruction struct{}
type applyFilterInstruction struct {
	str     string
	size    uint
	localId localId
}
type performTestInstruction struct {
	name     string
	argCount uint
	localId  localId
}
type emitInstruction struct{}
type pushLoopInstruction struct{ flags uint8 }
type pushWithInstruction struct{}
type iterateInstruction struct{ jumpTarget uint }
type pushDidNotIterateInstruction struct{}
type popFrameInstruction struct{}
type jumpInstruction struct{ jumpTarget uint }
type jumpIfFalseInstruction struct{ jumpTarget uint }
type jumpIfFalseOrPopInstruction struct{ jumpTarget uint }
type jumpIfTrueOrPopInstruction struct{ jumpTarget uint }
type pushAutoEscapeInstruction struct{}
type popAutoEscapeInstruction struct{}
type beginCaptureInstruction struct{ mode captureMode }
type endCaptureInstruction struct{}
type callFunctionInstruction struct {
	name string
	size uint
}
type callMethodInstruction struct {
	name string
	size uint
}
type callObjectInstruction struct{ argCount uint }
type dupTopInstruction struct{}
type discardTopInstruction struct{}
type fastSuperInstruction struct{}
type fastRecurseInstruction struct{}
type callBlockInstruction struct{ name string }
type loadBlocksInstruction struct{}
type includeInstruction struct{ ignoreMissing bool }
type exportLocalsInstruction struct{}
type buildMacroInstruction struct {
	name string
	size uint
	kind uint8
}
type returnInstruction struct{}
type isUndefinedInstruction struct{}
type encloseInstruction struct{ name string }
type getClosureInstruction struct{}

type lineInfo struct {
	firstInstruction uint32
	line             uint32
}

type spanInfo struct {
	firstInstruction uint32
	span             option[span]
}

var _ = instruction(emitRawInstruction{})
var _ = instruction(storeLocalInstruction{})
var _ = instruction(lookupInstruction{})
var _ = instruction(getAttrInstruction{})
var _ = instruction(getItemInstruction{})
var _ = instruction(sliceInstruction{})
var _ = instruction(loadConstInstruction{})
var _ = instruction(buildMapInstruction{})
var _ = instruction(buildKwargsInstruction{})
var _ = instruction(buildListInstruction{})
var _ = instruction(unpackListInstruction{})
var _ = instruction(listAppendInstruction{})
var _ = instruction(addInstruction{})
var _ = instruction(subInstruction{})
var _ = instruction(mulInstruction{})
var _ = instruction(divInstruction{})
var _ = instruction(intDivInstruction{})
var _ = instruction(remInstruction{})
var _ = instruction(powInstruction{})
var _ = instruction(negInstruction{})
var _ = instruction(eqInstruction{})
var _ = instruction(neInstruction{})
var _ = instruction(gtInstruction{})
var _ = instruction(gteInstruction{})
var _ = instruction(ltInstruction{})
var _ = instruction(lteInstruction{})
var _ = instruction(notInstruction{})
var _ = instruction(stringConcatInstruction{})
var _ = instruction(inInstruction{})
var _ = instruction(applyFilterInstruction{})
var _ = instruction(performTestInstruction{})
var _ = instruction(emitInstruction{})
var _ = instruction(pushLoopInstruction{})
var _ = instruction(pushWithInstruction{})
var _ = instruction(iterateInstruction{})
var _ = instruction(pushDidNotIterateInstruction{})
var _ = instruction(popFrameInstruction{})
var _ = instruction(jumpInstruction{})
var _ = instruction(jumpIfFalseInstruction{})
var _ = instruction(jumpIfFalseOrPopInstruction{})
var _ = instruction(jumpIfTrueOrPopInstruction{})
var _ = instruction(pushAutoEscapeInstruction{})
var _ = instruction(popAutoEscapeInstruction{})
var _ = instruction(beginCaptureInstruction{})
var _ = instruction(endCaptureInstruction{})
var _ = instruction(callFunctionInstruction{})
var _ = instruction(callMethodInstruction{})
var _ = instruction(callObjectInstruction{})
var _ = instruction(dupTopInstruction{})
var _ = instruction(discardTopInstruction{})
var _ = instruction(fastSuperInstruction{})
var _ = instruction(fastRecurseInstruction{})
var _ = instruction(callBlockInstruction{})
var _ = instruction(loadBlocksInstruction{})
var _ = instruction(includeInstruction{})
var _ = instruction(exportLocalsInstruction{})
var _ = instruction(buildMacroInstruction{})
var _ = instruction(returnInstruction{})
var _ = instruction(isUndefinedInstruction{})
var _ = instruction(encloseInstruction{})
var _ = instruction(getClosureInstruction{})

func (emitRawInstruction) typ() instType           { return instTypeEmitRaw }
func (storeLocalInstruction) typ() instType        { return instTypeStoreLocal }
func (lookupInstruction) typ() instType            { return instTypeLookup }
func (getAttrInstruction) typ() instType           { return instTypeGetAttr }
func (getItemInstruction) typ() instType           { return instTypeGetItem }
func (sliceInstruction) typ() instType             { return instTypeSlice }
func (loadConstInstruction) typ() instType         { return instTypeLoadConst }
func (buildMapInstruction) typ() instType          { return instTypeBuildMap }
func (buildKwargsInstruction) typ() instType       { return instTypeBuildKwargs }
func (buildListInstruction) typ() instType         { return instTypeBuildList }
func (unpackListInstruction) typ() instType        { return instTypeUnpackList }
func (listAppendInstruction) typ() instType        { return instTypeListAppend }
func (addInstruction) typ() instType               { return instTypeAdd }
func (subInstruction) typ() instType               { return instTypeSub }
func (mulInstruction) typ() instType               { return instTypeMul }
func (divInstruction) typ() instType               { return instTypeDiv }
func (intDivInstruction) typ() instType            { return instTypeIntDiv }
func (remInstruction) typ() instType               { return instTypeRem }
func (powInstruction) typ() instType               { return instTypePow }
func (negInstruction) typ() instType               { return instTypeNeg }
func (eqInstruction) typ() instType                { return instTypeEq }
func (neInstruction) typ() instType                { return instTypeNe }
func (gtInstruction) typ() instType                { return instTypeGt }
func (gteInstruction) typ() instType               { return instTypeGte }
func (ltInstruction) typ() instType                { return instTypeLt }
func (lteInstruction) typ() instType               { return instTypeLte }
func (notInstruction) typ() instType               { return instTypeNot }
func (stringConcatInstruction) typ() instType      { return instTypeStringConcat }
func (inInstruction) typ() instType                { return instTypeIn }
func (applyFilterInstruction) typ() instType       { return instTypeApplyFilter }
func (performTestInstruction) typ() instType       { return instTypePerformTest }
func (emitInstruction) typ() instType              { return instTypeEmit }
func (pushLoopInstruction) typ() instType          { return instTypePushLoop }
func (pushWithInstruction) typ() instType          { return instTypePushWith }
func (iterateInstruction) typ() instType           { return instTypeIterate }
func (pushDidNotIterateInstruction) typ() instType { return instTypePushDidNotIterate }
func (popFrameInstruction) typ() instType          { return instTypePopFrame }
func (jumpInstruction) typ() instType              { return instTypeJump }
func (jumpIfFalseInstruction) typ() instType       { return instTypeJumpIfFalse }
func (jumpIfFalseOrPopInstruction) typ() instType  { return instTypeJumpIfFalseOrPop }
func (jumpIfTrueOrPopInstruction) typ() instType   { return instTypeJumpIfTrueOrPop }
func (pushAutoEscapeInstruction) typ() instType    { return instTypePushAutoEscape }
func (popAutoEscapeInstruction) typ() instType     { return instTypePopAutoEscape }
func (beginCaptureInstruction) typ() instType      { return instTypeBeginCapture }
func (endCaptureInstruction) typ() instType        { return instTypeEndCapture }
func (callFunctionInstruction) typ() instType      { return instTypeCallFunction }
func (callMethodInstruction) typ() instType        { return instTypeCallMethod }
func (callObjectInstruction) typ() instType        { return instTypeCallObject }
func (dupTopInstruction) typ() instType            { return instTypeDupTop }
func (discardTopInstruction) typ() instType        { return instTypeDiscardTop }
func (fastSuperInstruction) typ() instType         { return instTypeFastSuper }
func (fastRecurseInstruction) typ() instType       { return instTypeFastRecurse }
func (callBlockInstruction) typ() instType         { return instTypeCallBlock }
func (loadBlocksInstruction) typ() instType        { return instTypeLoadBlocks }
func (includeInstruction) typ() instType           { return instTypeInclude }
func (exportLocalsInstruction) typ() instType      { return instTypeExportLocals }
func (buildMacroInstruction) typ() instType        { return instTypeBuildMacro }
func (returnInstruction) typ() instType            { return instTypeReturn }
func (isUndefinedInstruction) typ() instType       { return instTypeIsUndefined }
func (encloseInstruction) typ() instType           { return instTypeEnclose }
func (getClosureInstruction) typ() instType        { return instTypeGetClosure }

/*
instTypeEmitRaw
instTypeStoreLocal
instTypeLookup
instTypeGetAttr
instTypeGetItem
instTypeSlice
instTypeLoadConst
instTypeBuildMap
instTypeBuildKwargs
instTypeBuildList
instTypeUnpackList
instTypeListAppend
instTypeAdd
instTypeSub
instTypeMul
instTypeDiv
instTypeIntDiv
instTypeRem
instTypePow
instTypeNeg
instTypeEq
instTypeNe
instTypeGt
instTypeGte
instTypeLt
instTypeLte
instTypeNot
instTypeStringConcat
instTypeIn
instTypeApplyFilter
instTypePerformTest
instTypeEmit
instTypePushLoop
instTypePushWith
instTypeIterate
instTypePushDidNotIterate
instTypePopFrame
instTypeJump
instTypeJumpIfFalse
instTypeJumpIfFalseOrPop
instTypeJumpIfTrueOrPop
instTypePushAutoEscape
instTypePopAutoEscape
instTypeBeginCapture
instTypeEndCapture
instTypeCallFunction
instTypeCallMethod
instTypeCallObject
instTypeDupTop
instTypeDiscardTop
instTypeFastSuper
instTypeFastRecurse
instTypeCallBlock
instTypeLoadBlocks
instTypeInclude
instTypeExportLocals
instTypeBuildMacro
instTypeReturn
instTypeIsUndefined
instTypeEnclose
instTypeGetClosure
*/

type instType uint

const (
	// Emits raw source
	instTypeEmitRaw instType = iota + 1

	// Stores a variable (only possible in for loops)
	instTypeStoreLocal

	// Load a variable
	instTypeLookup

	// Looks up an attribute.
	instTypeGetAttr

	// Looks up an item.
	instTypeGetItem

	// Performs a slice operation.
	instTypeSlice

	// Loads a constant value.
	instTypeLoadConst

	// Builds a map of the last n pairs on the stack.
	instTypeBuildMap

	// Builds a kwargs map of the last n pairs on the stack.
	instTypeBuildKwargs

	// Builds a list of the last n pairs on the stack.
	instTypeBuildList

	// Unpacks a list into N stack items.
	instTypeUnpackList

	// Appends to the list.
	instTypeListAppend

	// Add the top two values
	instTypeAdd

	// Subtract the top two values
	instTypeSub

	// Multiply the top two values
	instTypeMul

	// Divide the top two values
	instTypeDiv

	// Integer divide the top two values as "integer".
	//
	// Note that in MiniJinja this currently uses an euclidean
	// division to match the rem implementation.  In Python this
	// instead uses a flooring division and a flooring remainder.
	instTypeIntDiv

	// Calculate the remainder the top two values
	instTypeRem

	// x to the power of y.
	instTypePow

	// Negates the value.
	instTypeNeg

	// `=` operator
	instTypeEq

	// `!=` operator
	instTypeNe

	// `>` operator
	instTypeGt

	// `>=` operator
	instTypeGte

	// `<` operator
	instTypeLt

	// `<=` operator
	instTypeLte

	// Unary not
	instTypeNot

	// String concatenation operator
	instTypeStringConcat

	// Performs a containment check
	instTypeIn

	// Apply a filter.
	instTypeApplyFilter

	// Perform a filter.
	instTypePerformTest

	// Emit the stack top as output
	instTypeEmit

	// Starts a loop
	//
	// The argument are loop flags.
	instTypePushLoop

	// Starts a with block.
	instTypePushWith

	// Does a single loop iteration
	//
	// The argument is the jump target for when the loop
	// ends and must point to a `PopFrame` instruction.
	instTypeIterate

	// Push a bool that indicates that the loop iterated.
	instTypePushDidNotIterate

	// Pops the topmost frame
	instTypePopFrame

	// Jump to a specific instruction
	instTypeJump

	// Jump if the stack top evaluates to false
	instTypeJumpIfFalse

	// Jump if the stack top evaluates to false or pops the value
	instTypeJumpIfFalseOrPop

	// Jump if the stack top evaluates to true or pops the value
	instTypeJumpIfTrueOrPop

	// Sets the auto escape flag to the current value.
	instTypePushAutoEscape

	// Resets the auto escape flag to the previous value.
	instTypePopAutoEscape

	// Begins capturing of output (false) or discard (true).
	instTypeBeginCapture

	// Ends capturing of output.
	instTypeEndCapture

	// Calls a global function
	instTypeCallFunction

	// Calls a method
	instTypeCallMethod

	// Calls an object
	instTypeCallObject

	// Duplicates the top item
	instTypeDupTop

	// Discards the top item
	instTypeDiscardTop

	// A fast super instruction without intermediate capturing.
	instTypeFastSuper

	// A fast loop recurse instruction without intermediate capturing.
	instTypeFastRecurse

	// Call into a block.
	instTypeCallBlock

	// Loads block from a template with name on stack ("extends")
	instTypeLoadBlocks

	// Includes another template.
	instTypeInclude

	// Builds a module
	instTypeExportLocals

	// Builds a macro on the stack.
	instTypeBuildMacro

	// Breaks from the interpreter loop (exists a function)
	instTypeReturn

	// True if the value is undefined
	instTypeIsUndefined

	// Encloses a variable.
	instTypeEnclose

	// Returns the closure of this context level.
	instTypeGetClosure
)

func (k instType) String() string {
	switch k {
	case instTypeEmitRaw:
		return "emitRaw"
	case instTypeStoreLocal:
		return "storeLocal"
	case instTypeLookup:
		return "lookup"
	case instTypeGetAttr:
		return "getAttr"
	case instTypeGetItem:
		return "getItem"
	case instTypeSlice:
		return "slice"
	case instTypeLoadConst:
		return "loadConst"
	case instTypeBuildMap:
		return "buildMap"
	case instTypeBuildKwargs:
		return "buildKwargs"
	case instTypeBuildList:
		return "buildList"
	case instTypeUnpackList:
		return "unpackList"
	case instTypeListAppend:
		return "listAppend"
	case instTypeAdd:
		return "add"
	case instTypeSub:
		return "sub"
	case instTypeMul:
		return "mul"
	case instTypeDiv:
		return "div"
	case instTypeIntDiv:
		return "intDiv"
	case instTypeRem:
		return "rem"
	case instTypePow:
		return "pow"
	case instTypeNeg:
		return "neg"
	case instTypeEq:
		return "eq"
	case instTypeNe:
		return "ne"
	case instTypeGt:
		return "gt"
	case instTypeGte:
		return "gte"
	case instTypeLt:
		return "lt"
	case instTypeLte:
		return "lte"
	case instTypeNot:
		return "not"
	case instTypeStringConcat:
		return "stringConcat"
	case instTypeIn:
		return "in"
	case instTypeApplyFilter:
		return "applyFilter"
	case instTypePerformTest:
		return "performTest"
	case instTypeEmit:
		return "emit"
	case instTypePushLoop:
		return "pushLoop"
	case instTypePushWith:
		return "pushWith"
	case instTypeIterate:
		return "iterate"
	case instTypePushDidNotIterate:
		return "pushDidNotIterate"
	case instTypePopFrame:
		return "popFrame"
	case instTypeJump:
		return "jump"
	case instTypeJumpIfFalse:
		return "jumpIfFalse"
	case instTypeJumpIfFalseOrPop:
		return "jumpIfFalseOrPop"
	case instTypeJumpIfTrueOrPop:
		return "jumpIfTrueOrPop"
	case instTypePushAutoEscape:
		return "pushAutoEscape"
	case instTypePopAutoEscape:
		return "popAutoEscape"
	case instTypeBeginCapture:
		return "beginCapture"
	case instTypeEndCapture:
		return "endCapture"
	case instTypeCallFunction:
		return "callFunction"
	case instTypeCallMethod:
		return "callMethod"
	case instTypeCallObject:
		return "callObject"
	case instTypeDupTop:
		return "dupTop"
	case instTypeDiscardTop:
		return "discardTop"
	case instTypeFastSuper:
		return "fastSuper"
	case instTypeFastRecurse:
		return "fastRecurse"
	case instTypeCallBlock:
		return "callBlock"
	case instTypeLoadBlocks:
		return "loadBlocks"
	case instTypeInclude:
		return "include"
	case instTypeExportLocals:
		return "exportLocals"
	case instTypeBuildMacro:
		return "buildMacro"
	case instTypeReturn:
		return "return"
	case instTypeIsUndefined:
		return "isUndefined"
	case instTypeEnclose:
		return "enclose"
	case instTypeGetClosure:
		return "getClosure"
	default:
		panic("invalid instType")
	}
}

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
