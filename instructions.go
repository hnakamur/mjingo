package mjingo

import (
	"cmp"
	"slices"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type captureMode uint8

const (
	captureModeCapture captureMode = iota + 1
	captureModeDiscard
)

const (
	// This loop has the loop var.
	loopFlagWithLoopVar = 1

	// This loop is recursive.
	loopFlagRecursive = 2
)

// This macro uses the caller var.
const macroCaller = 2

// Go type to represent locals.
type localID = uint8

// The maximum number of filters/tests that can be cached.
const maxLocals = 50

type instructions struct {
	instructions []instruction
	lineInfos    []lineInfo
	spanInfos    []spanInfo
	name         string
	source       string
}

func (i *instructions) Instructions() []instruction { return i.instructions }

func (i *instructions) Name() string { return i.name }

var emptyInstructions = instructions{
	name: "<unknown>",
}

type instruction interface {
	Typ() instType
}

type emitRawInstruction struct{ Val string }
type storeLocalInstruction struct{ Name string }
type lookupInstruction struct{ Name string }
type getAttrInstruction struct{ Name string }
type getItemInstruction struct{}
type sliceInstruction struct{}
type loadConstInstruction struct{ Val Value }
type buildMapInstruction struct{ PairCount uint }
type buildKwargsInstruction struct{ PairCount uint }
type buildListInstruction struct{ Count uint }
type unpackListInstruction struct{ Count uint }
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
	Name     string
	ArgCount uint
	LocalID  localID
}
type performTestInstruction struct {
	Name     string
	ArgCount uint
	LocalID  localID
}
type emitInstruction struct{}
type pushLoopInstruction struct{ Flags uint8 }
type pushWithInstruction struct{}
type iterateInstruction struct{ JumpTarget uint }
type pushDidNotIterateInstruction struct{}
type popFrameInstruction struct{}
type jumpInstruction struct{ JumpTarget uint }
type jumpIfFalseInstruction struct{ JumpTarget uint }
type jumpIfFalseOrPopInstruction struct{ JumpTarget uint }
type jumpIfTrueOrPopInstruction struct{ JumpTarget uint }
type pushAutoEscapeInstruction struct{}
type popAutoEscapeInstruction struct{}
type beginCaptureInstruction struct{ Mode captureMode }
type endCaptureInstruction struct{}
type callFunctionInstruction struct {
	Name     string
	ArgCount uint
}
type callMethodInstruction struct {
	Name     string
	ArgCount uint
}
type callObjectInstruction struct{ ArgCount uint }
type dupTopInstruction struct{}
type discardTopInstruction struct{}
type fastSuperInstruction struct{}
type fastRecurseInstruction struct{}
type callBlockInstruction struct{ Name string }
type loadBlocksInstruction struct{}
type includeInstruction struct{ IgnoreMissing bool }
type exportLocalsInstruction struct{}
type buildMacroInstruction struct {
	Name   string
	Offset uint
	Flags  uint8
}
type returnInstruction struct{}
type isUndefinedInstruction struct{}
type encloseInstruction struct{ Name string }
type getClosureInstruction struct{}

type lineInfo struct {
	firstInstruction uint32
	line             uint32
}

type spanInfo struct {
	firstInstruction uint32
	span             option.Option[span]
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

func (emitRawInstruction) Typ() instType           { return instTypeEmitRaw }
func (storeLocalInstruction) Typ() instType        { return instTypeStoreLocal }
func (lookupInstruction) Typ() instType            { return instTypeLookup }
func (getAttrInstruction) Typ() instType           { return instTypeGetAttr }
func (getItemInstruction) Typ() instType           { return instTypeGetItem }
func (sliceInstruction) Typ() instType             { return instTypeSlice }
func (loadConstInstruction) Typ() instType         { return instTypeLoadConst }
func (buildMapInstruction) Typ() instType          { return instTypeBuildMap }
func (buildKwargsInstruction) Typ() instType       { return instTypeBuildKwargs }
func (buildListInstruction) Typ() instType         { return instTypeBuildList }
func (unpackListInstruction) Typ() instType        { return instTypeUnpackList }
func (listAppendInstruction) Typ() instType        { return instTypeListAppend }
func (addInstruction) Typ() instType               { return instTypeAdd }
func (subInstruction) Typ() instType               { return instTypeSub }
func (mulInstruction) Typ() instType               { return instTypeMul }
func (divInstruction) Typ() instType               { return instTypeDiv }
func (intDivInstruction) Typ() instType            { return instTypeIntDiv }
func (remInstruction) Typ() instType               { return instTypeRem }
func (powInstruction) Typ() instType               { return instTypePow }
func (negInstruction) Typ() instType               { return instTypeNeg }
func (eqInstruction) Typ() instType                { return instTypeEq }
func (neInstruction) Typ() instType                { return instTypeNe }
func (gtInstruction) Typ() instType                { return instTypeGt }
func (gteInstruction) Typ() instType               { return instTypeGte }
func (ltInstruction) Typ() instType                { return instTypeLt }
func (lteInstruction) Typ() instType               { return instTypeLte }
func (notInstruction) Typ() instType               { return instTypeNot }
func (stringConcatInstruction) Typ() instType      { return instTypeStringConcat }
func (inInstruction) Typ() instType                { return instTypeIn }
func (applyFilterInstruction) Typ() instType       { return instTypeApplyFilter }
func (performTestInstruction) Typ() instType       { return instTypePerformTest }
func (emitInstruction) Typ() instType              { return instTypeEmit }
func (pushLoopInstruction) Typ() instType          { return instTypePushLoop }
func (pushWithInstruction) Typ() instType          { return instTypePushWith }
func (iterateInstruction) Typ() instType           { return instTypeIterate }
func (pushDidNotIterateInstruction) Typ() instType { return instTypePushDidNotIterate }
func (popFrameInstruction) Typ() instType          { return instTypePopFrame }
func (jumpInstruction) Typ() instType              { return instTypeJump }
func (jumpIfFalseInstruction) Typ() instType       { return instTypeJumpIfFalse }
func (jumpIfFalseOrPopInstruction) Typ() instType  { return instTypeJumpIfFalseOrPop }
func (jumpIfTrueOrPopInstruction) Typ() instType   { return instTypeJumpIfTrueOrPop }
func (pushAutoEscapeInstruction) Typ() instType    { return instTypePushAutoEscape }
func (popAutoEscapeInstruction) Typ() instType     { return instTypePopAutoEscape }
func (beginCaptureInstruction) Typ() instType      { return instTypeBeginCapture }
func (endCaptureInstruction) Typ() instType        { return instTypeEndCapture }
func (callFunctionInstruction) Typ() instType      { return instTypeCallFunction }
func (callMethodInstruction) Typ() instType        { return instTypeCallMethod }
func (callObjectInstruction) Typ() instType        { return instTypeCallObject }
func (dupTopInstruction) Typ() instType            { return instTypeDupTop }
func (discardTopInstruction) Typ() instType        { return instTypeDiscardTop }
func (fastSuperInstruction) Typ() instType         { return instTypeFastSuper }
func (fastRecurseInstruction) Typ() instType       { return instTypeFastRecurse }
func (callBlockInstruction) Typ() instType         { return instTypeCallBlock }
func (loadBlocksInstruction) Typ() instType        { return instTypeLoadBlocks }
func (includeInstruction) Typ() instType           { return instTypeInclude }
func (exportLocalsInstruction) Typ() instType      { return instTypeExportLocals }
func (buildMacroInstruction) Typ() instType        { return instTypeBuildMacro }
func (returnInstruction) Typ() instType            { return instTypeReturn }
func (isUndefinedInstruction) Typ() instType       { return instTypeIsUndefined }
func (encloseInstruction) Typ() instType           { return instTypeEnclose }
func (getClosureInstruction) Typ() instType        { return instTypeGetClosure }

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

	// Loads a constant Value.
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

	// Negates the Value.
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

	// Jump if the stack top evaluates to false or pops the Value
	instTypeJumpIfFalseOrPop

	// Jump if the stack top evaluates to true or pops the Value
	instTypeJumpIfTrueOrPop

	// Sets the auto escape flag to the current Value.
	instTypePushAutoEscape

	// Resets the auto escape flag to the previous Value.
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

	// True if the Value is undefined
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
		sameLoc = lastLoc.span.IsSome() && lastLoc.span.Unwrap() == spn
	}
	if !sameLoc {
		i.spanInfos = append(i.spanInfos,
			spanInfo{firstInstruction: uint32(rv), span: option.Some(spn)})
	}

	i.addLineRecord(rv, spn.StartLine)
	return rv
}

func (i *instructions) GetLine(idx uint) option.Option[uint] {
	n, found := slices.BinarySearchFunc(i.lineInfos,
		lineInfo{firstInstruction: uint32(idx)},
		func(a, b lineInfo) int {
			return cmp.Compare(a.firstInstruction, b.firstInstruction)
		})
	if found {
		return option.Some(uint(i.lineInfos[n].line))
	}
	if n != 0 {
		return option.Some(uint(i.lineInfos[n-1].line))
	}
	return option.None[uint]()
}

func (i *instructions) GetSpan(idx uint) option.Option[span] {
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
	return option.None[span]()
}
