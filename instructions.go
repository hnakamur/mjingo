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

type instruction interface {
	typ() instType
}

type emitRawInst struct{ val string }
type storeLocalInst struct{ name string }
type lookupInst struct{ name string }
type getAttrInst struct{ name string }
type getItemInst struct{}
type sliceInst struct{}
type loadConstInst struct{ val value }
type buildMapInst struct{ pairCount uint }
type buildKwargsInst struct{ pairCount uint }
type buildListInst struct{ count uint }
type unpackListInst struct{ count uint }
type listAppendInst struct{}
type addInst struct{}
type subInst struct{}
type mulInst struct{}
type divInst struct{}
type intDivInst struct{}
type remInst struct{}
type powInst struct{}
type negInst struct{}
type eqInst struct{}
type neInst struct{}
type gtInst struct{}
type gteInst struct{}
type ltInst struct{}
type lteInst struct{}
type notInst struct{}
type stringConcatInst struct{}
type inInst struct{}
type applyFilterInst struct {
	str     string
	size    uint
	localId localId
}
type performTestInst struct {
	str     string
	size    uint
	localId localId
}
type emitInst struct{}
type pushLoopInst struct{ flags uint8 }
type pushWithInst struct{}
type iterateInst struct{ jumpTarget uint }
type pushDidNotIterateInst struct{}
type popFrameInst struct{}
type jumpInst struct{ jumpTarget uint }
type jumpIfFalseInst struct{ jumpTarget uint }
type jumpIfFalseOrPopInst struct{ jumpTarget uint }
type jumpIfTrueOrPopInst struct{ jumpTarget uint }
type pushAutoEscapeInst struct{}
type popAutoEscapeInst struct{}
type beginCaptureInst struct{ mode captureMode }
type endCaptureInst struct{}
type callFunctionInst struct {
	name string
	size uint
}
type callMethodInst struct {
	name string
	size uint
}
type callObjectInst struct{ argCount uint }
type dupTopInst struct{}
type discardTopInst struct{}
type fastSuperInst struct{}
type fastRecurseInst struct{}
type callBlockInst struct{ name string }
type loadBlocksInst struct{}
type includeInst struct{ ignoreMissing bool }
type exportLocalsInst struct{}
type buildMacroInst struct {
	name string
	size uint
	kind uint8
}
type returnInst struct{}
type isUndefinedInst struct{}
type encloseInst struct{ name string }
type getClosureInst struct{}

type lineInfo struct {
	firstInstruction uint32
	line             uint32
}

type spanInfo struct {
	firstInstruction uint32
	span             option[span]
}

var _ = instruction(emitRawInst{})
var _ = instruction(storeLocalInst{})
var _ = instruction(lookupInst{})
var _ = instruction(getAttrInst{})
var _ = instruction(getItemInst{})
var _ = instruction(sliceInst{})
var _ = instruction(loadConstInst{})
var _ = instruction(buildMapInst{})
var _ = instruction(buildKwargsInst{})
var _ = instruction(buildListInst{})
var _ = instruction(unpackListInst{})
var _ = instruction(listAppendInst{})
var _ = instruction(addInst{})
var _ = instruction(subInst{})
var _ = instruction(mulInst{})
var _ = instruction(divInst{})
var _ = instruction(intDivInst{})
var _ = instruction(remInst{})
var _ = instruction(powInst{})
var _ = instruction(negInst{})
var _ = instruction(eqInst{})
var _ = instruction(neInst{})
var _ = instruction(gtInst{})
var _ = instruction(gteInst{})
var _ = instruction(ltInst{})
var _ = instruction(lteInst{})
var _ = instruction(notInst{})
var _ = instruction(stringConcatInst{})
var _ = instruction(inInst{})
var _ = instruction(applyFilterInst{})
var _ = instruction(performTestInst{})
var _ = instruction(emitInst{})
var _ = instruction(pushLoopInst{})
var _ = instruction(pushWithInst{})
var _ = instruction(iterateInst{})
var _ = instruction(pushDidNotIterateInst{})
var _ = instruction(popFrameInst{})
var _ = instruction(jumpInst{})
var _ = instruction(jumpIfFalseInst{})
var _ = instruction(jumpIfFalseOrPopInst{})
var _ = instruction(jumpIfTrueOrPopInst{})
var _ = instruction(pushAutoEscapeInst{})
var _ = instruction(popAutoEscapeInst{})
var _ = instruction(beginCaptureInst{})
var _ = instruction(endCaptureInst{})
var _ = instruction(callFunctionInst{})
var _ = instruction(callMethodInst{})
var _ = instruction(callObjectInst{})
var _ = instruction(dupTopInst{})
var _ = instruction(discardTopInst{})
var _ = instruction(fastSuperInst{})
var _ = instruction(fastRecurseInst{})
var _ = instruction(callBlockInst{})
var _ = instruction(loadBlocksInst{})
var _ = instruction(includeInst{})
var _ = instruction(exportLocalsInst{})
var _ = instruction(buildMacroInst{})
var _ = instruction(returnInst{})
var _ = instruction(isUndefinedInst{})
var _ = instruction(encloseInst{})
var _ = instruction(getClosureInst{})

func (emitRawInst) typ() instType           { return instTypeEmitRaw }
func (storeLocalInst) typ() instType        { return instTypeStoreLocal }
func (lookupInst) typ() instType            { return instTypeLookup }
func (getAttrInst) typ() instType           { return instTypeGetAttr }
func (getItemInst) typ() instType           { return instTypeGetItem }
func (sliceInst) typ() instType             { return instTypeSlice }
func (loadConstInst) typ() instType         { return instTypeLoadConst }
func (buildMapInst) typ() instType          { return instTypeBuildMap }
func (buildKwargsInst) typ() instType       { return instTypeBuildKwargs }
func (buildListInst) typ() instType         { return instTypeBuildList }
func (unpackListInst) typ() instType        { return instTypeUnpackList }
func (listAppendInst) typ() instType        { return instTypeListAppend }
func (addInst) typ() instType               { return instTypeAdd }
func (subInst) typ() instType               { return instTypeSub }
func (mulInst) typ() instType               { return instTypeMul }
func (divInst) typ() instType               { return instTypeDiv }
func (intDivInst) typ() instType            { return instTypeIntDiv }
func (remInst) typ() instType               { return instTypeRem }
func (powInst) typ() instType               { return instTypePow }
func (negInst) typ() instType               { return instTypeNeg }
func (eqInst) typ() instType                { return instTypeEq }
func (neInst) typ() instType                { return instTypeNe }
func (gtInst) typ() instType                { return instTypeGt }
func (gteInst) typ() instType               { return instTypeGte }
func (ltInst) typ() instType                { return instTypeLt }
func (lteInst) typ() instType               { return instTypeLte }
func (notInst) typ() instType               { return instTypeNot }
func (stringConcatInst) typ() instType      { return instTypeStringConcat }
func (inInst) typ() instType                { return instTypeIn }
func (applyFilterInst) typ() instType       { return instTypeApplyFilter }
func (performTestInst) typ() instType       { return instTypePerformTest }
func (emitInst) typ() instType              { return instTypeEmit }
func (pushLoopInst) typ() instType          { return instTypePushLoop }
func (pushWithInst) typ() instType          { return instTypePushWith }
func (iterateInst) typ() instType           { return instTypeIterate }
func (pushDidNotIterateInst) typ() instType { return instTypePushDidNotIterate }
func (popFrameInst) typ() instType          { return instTypePopFrame }
func (jumpInst) typ() instType              { return instTypeJump }
func (jumpIfFalseInst) typ() instType       { return instTypeJumpIfFalse }
func (jumpIfFalseOrPopInst) typ() instType  { return instTypeJumpIfFalseOrPop }
func (jumpIfTrueOrPopInst) typ() instType   { return instTypeJumpIfTrueOrPop }
func (pushAutoEscapeInst) typ() instType    { return instTypePushAutoEscape }
func (popAutoEscapeInst) typ() instType     { return instTypePopAutoEscape }
func (beginCaptureInst) typ() instType      { return instTypeBeginCapture }
func (endCaptureInst) typ() instType        { return instTypeEndCapture }
func (callFunctionInst) typ() instType      { return instTypeCallFunction }
func (callMethodInst) typ() instType        { return instTypeCallMethod }
func (callObjectInst) typ() instType        { return instTypeCallObject }
func (dupTopInst) typ() instType            { return instTypeDupTop }
func (discardTopInst) typ() instType        { return instTypeDiscardTop }
func (fastSuperInst) typ() instType         { return instTypeFastSuper }
func (fastRecurseInst) typ() instType       { return instTypeFastRecurse }
func (callBlockInst) typ() instType         { return instTypeCallBlock }
func (loadBlocksInst) typ() instType        { return instTypeLoadBlocks }
func (includeInst) typ() instType           { return instTypeInclude }
func (exportLocalsInst) typ() instType      { return instTypeExportLocals }
func (buildMacroInst) typ() instType        { return instTypeBuildMacro }
func (returnInst) typ() instType            { return instTypeReturn }
func (isUndefinedInst) typ() instType       { return instTypeIsUndefined }
func (encloseInst) typ() instType           { return instTypeEnclose }
func (getClosureInst) typ() instType        { return instTypeGetClosure }

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
