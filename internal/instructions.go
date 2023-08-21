package internal

import (
	"cmp"
	"slices"

	"github.com/hnakamur/mjingo/internal/datast/option"
)

type CaptureMode uint8

const (
	CaptureModeCapture CaptureMode = iota + 1
	CaptureModeDiscard
)

const (
	// This loop has the loop var.
	LoopFlagWithLoopVar = 1

	// This loop is recursive.
	LoopFlagRecursive = 2
)

// This macro uses the caller var.
const macroCaller = 2

// Go type to represent locals.
type LocalID = uint8

// The maximum number of filters/tests that can be cached.
const MaxLocals = 50

type Instructions struct {
	instructions []Instruction
	lineInfos    []lineInfo
	spanInfos    []spanInfo
	name         string
	source       string
}

func (i *Instructions) Instructions() []Instruction { return i.instructions }

func (i *Instructions) Name() string { return i.name }

var emptyInstructions = Instructions{
	name: "<unknown>",
}

type Instruction interface {
	Typ() instType
}

type EmitRawInstruction struct{ Val string }
type StoreLocalInstruction struct{ Name string }
type LookupInstruction struct{ Name string }
type GetAttrInstruction struct{ Name string }
type GetItemInstruction struct{}
type SliceInstruction struct{}
type LoadConstInstruction struct{ Val Value }
type BuildMapInstruction struct{ PairCount uint }
type BuildKwargsInstruction struct{ PairCount uint }
type BuildListInstruction struct{ Count uint }
type UnpackListInstruction struct{ Count uint }
type ListAppendInstruction struct{}
type AddInstruction struct{}
type SubInstruction struct{}
type MulInstruction struct{}
type DivInstruction struct{}
type IntDivInstruction struct{}
type RemInstruction struct{}
type PowInstruction struct{}
type NegInstruction struct{}
type EqInstruction struct{}
type NeInstruction struct{}
type GtInstruction struct{}
type GteInstruction struct{}
type LtInstruction struct{}
type LteInstruction struct{}
type NotInstruction struct{}
type StringConcatInstruction struct{}
type InInstruction struct{}
type ApplyFilterInstruction struct {
	Name     string
	ArgCount uint
	LocalID  LocalID
}
type PerformTestInstruction struct {
	Name     string
	ArgCount uint
	LocalID  LocalID
}
type EmitInstruction struct{}
type PushLoopInstruction struct{ Flags uint8 }
type PushWithInstruction struct{}
type IterateInstruction struct{ JumpTarget uint }
type PushDidNotIterateInstruction struct{}
type PopFrameInstruction struct{}
type JumpInstruction struct{ JumpTarget uint }
type JumpIfFalseInstruction struct{ JumpTarget uint }
type JumpIfFalseOrPopInstruction struct{ JumpTarget uint }
type JumpIfTrueOrPopInstruction struct{ JumpTarget uint }
type PushAutoEscapeInstruction struct{}
type PopAutoEscapeInstruction struct{}
type BeginCaptureInstruction struct{ Mode CaptureMode }
type EndCaptureInstruction struct{}
type CallFunctionInstruction struct {
	Name     string
	ArgCount uint
}
type CallMethodInstruction struct {
	Name     string
	ArgCount uint
}
type CallObjectInstruction struct{ ArgCount uint }
type DupTopInstruction struct{}
type DiscardTopInstruction struct{}
type FastSuperInstruction struct{}
type FastRecurseInstruction struct{}
type CallBlockInstruction struct{ Name string }
type LoadBlocksInstruction struct{}
type IncludeInstruction struct{ IgnoreMissing bool }
type ExportLocalsInstruction struct{}
type BuildMacroInstruction struct {
	Name string
	Size uint
	Kind uint8
}
type ReturnInstruction struct{}
type IsUndefinedInstruction struct{}
type EncloseInstruction struct{ Name string }
type GetClosureInstruction struct{}

type lineInfo struct {
	firstInstruction uint32
	line             uint32
}

type spanInfo struct {
	firstInstruction uint32
	span             option.Option[Span]
}

var _ = Instruction(EmitRawInstruction{})
var _ = Instruction(StoreLocalInstruction{})
var _ = Instruction(LookupInstruction{})
var _ = Instruction(GetAttrInstruction{})
var _ = Instruction(GetItemInstruction{})
var _ = Instruction(SliceInstruction{})
var _ = Instruction(LoadConstInstruction{})
var _ = Instruction(BuildMapInstruction{})
var _ = Instruction(BuildKwargsInstruction{})
var _ = Instruction(BuildListInstruction{})
var _ = Instruction(UnpackListInstruction{})
var _ = Instruction(ListAppendInstruction{})
var _ = Instruction(AddInstruction{})
var _ = Instruction(SubInstruction{})
var _ = Instruction(MulInstruction{})
var _ = Instruction(DivInstruction{})
var _ = Instruction(IntDivInstruction{})
var _ = Instruction(RemInstruction{})
var _ = Instruction(PowInstruction{})
var _ = Instruction(NegInstruction{})
var _ = Instruction(EqInstruction{})
var _ = Instruction(NeInstruction{})
var _ = Instruction(GtInstruction{})
var _ = Instruction(GteInstruction{})
var _ = Instruction(LtInstruction{})
var _ = Instruction(LteInstruction{})
var _ = Instruction(NotInstruction{})
var _ = Instruction(StringConcatInstruction{})
var _ = Instruction(InInstruction{})
var _ = Instruction(ApplyFilterInstruction{})
var _ = Instruction(PerformTestInstruction{})
var _ = Instruction(EmitInstruction{})
var _ = Instruction(PushLoopInstruction{})
var _ = Instruction(PushWithInstruction{})
var _ = Instruction(IterateInstruction{})
var _ = Instruction(PushDidNotIterateInstruction{})
var _ = Instruction(PopFrameInstruction{})
var _ = Instruction(JumpInstruction{})
var _ = Instruction(JumpIfFalseInstruction{})
var _ = Instruction(JumpIfFalseOrPopInstruction{})
var _ = Instruction(JumpIfTrueOrPopInstruction{})
var _ = Instruction(PushAutoEscapeInstruction{})
var _ = Instruction(PopAutoEscapeInstruction{})
var _ = Instruction(BeginCaptureInstruction{})
var _ = Instruction(EndCaptureInstruction{})
var _ = Instruction(CallFunctionInstruction{})
var _ = Instruction(CallMethodInstruction{})
var _ = Instruction(CallObjectInstruction{})
var _ = Instruction(DupTopInstruction{})
var _ = Instruction(DiscardTopInstruction{})
var _ = Instruction(FastSuperInstruction{})
var _ = Instruction(FastRecurseInstruction{})
var _ = Instruction(CallBlockInstruction{})
var _ = Instruction(LoadBlocksInstruction{})
var _ = Instruction(IncludeInstruction{})
var _ = Instruction(ExportLocalsInstruction{})
var _ = Instruction(BuildMacroInstruction{})
var _ = Instruction(ReturnInstruction{})
var _ = Instruction(IsUndefinedInstruction{})
var _ = Instruction(EncloseInstruction{})
var _ = Instruction(GetClosureInstruction{})

func (EmitRawInstruction) Typ() instType           { return instTypeEmitRaw }
func (StoreLocalInstruction) Typ() instType        { return instTypeStoreLocal }
func (LookupInstruction) Typ() instType            { return instTypeLookup }
func (GetAttrInstruction) Typ() instType           { return instTypeGetAttr }
func (GetItemInstruction) Typ() instType           { return instTypeGetItem }
func (SliceInstruction) Typ() instType             { return instTypeSlice }
func (LoadConstInstruction) Typ() instType         { return instTypeLoadConst }
func (BuildMapInstruction) Typ() instType          { return instTypeBuildMap }
func (BuildKwargsInstruction) Typ() instType       { return instTypeBuildKwargs }
func (BuildListInstruction) Typ() instType         { return instTypeBuildList }
func (UnpackListInstruction) Typ() instType        { return instTypeUnpackList }
func (ListAppendInstruction) Typ() instType        { return instTypeListAppend }
func (AddInstruction) Typ() instType               { return instTypeAdd }
func (SubInstruction) Typ() instType               { return instTypeSub }
func (MulInstruction) Typ() instType               { return instTypeMul }
func (DivInstruction) Typ() instType               { return instTypeDiv }
func (IntDivInstruction) Typ() instType            { return instTypeIntDiv }
func (RemInstruction) Typ() instType               { return instTypeRem }
func (PowInstruction) Typ() instType               { return instTypePow }
func (NegInstruction) Typ() instType               { return instTypeNeg }
func (EqInstruction) Typ() instType                { return instTypeEq }
func (NeInstruction) Typ() instType                { return instTypeNe }
func (GtInstruction) Typ() instType                { return instTypeGt }
func (GteInstruction) Typ() instType               { return instTypeGte }
func (LtInstruction) Typ() instType                { return instTypeLt }
func (LteInstruction) Typ() instType               { return instTypeLte }
func (NotInstruction) Typ() instType               { return instTypeNot }
func (StringConcatInstruction) Typ() instType      { return instTypeStringConcat }
func (InInstruction) Typ() instType                { return instTypeIn }
func (ApplyFilterInstruction) Typ() instType       { return instTypeApplyFilter }
func (PerformTestInstruction) Typ() instType       { return instTypePerformTest }
func (EmitInstruction) Typ() instType              { return instTypeEmit }
func (PushLoopInstruction) Typ() instType          { return instTypePushLoop }
func (PushWithInstruction) Typ() instType          { return instTypePushWith }
func (IterateInstruction) Typ() instType           { return instTypeIterate }
func (PushDidNotIterateInstruction) Typ() instType { return instTypePushDidNotIterate }
func (PopFrameInstruction) Typ() instType          { return instTypePopFrame }
func (JumpInstruction) Typ() instType              { return instTypeJump }
func (JumpIfFalseInstruction) Typ() instType       { return instTypeJumpIfFalse }
func (JumpIfFalseOrPopInstruction) Typ() instType  { return instTypeJumpIfFalseOrPop }
func (JumpIfTrueOrPopInstruction) Typ() instType   { return instTypeJumpIfTrueOrPop }
func (PushAutoEscapeInstruction) Typ() instType    { return instTypePushAutoEscape }
func (PopAutoEscapeInstruction) Typ() instType     { return instTypePopAutoEscape }
func (BeginCaptureInstruction) Typ() instType      { return instTypeBeginCapture }
func (EndCaptureInstruction) Typ() instType        { return instTypeEndCapture }
func (CallFunctionInstruction) Typ() instType      { return instTypeCallFunction }
func (CallMethodInstruction) Typ() instType        { return instTypeCallMethod }
func (CallObjectInstruction) Typ() instType        { return instTypeCallObject }
func (DupTopInstruction) Typ() instType            { return instTypeDupTop }
func (DiscardTopInstruction) Typ() instType        { return instTypeDiscardTop }
func (FastSuperInstruction) Typ() instType         { return instTypeFastSuper }
func (FastRecurseInstruction) Typ() instType       { return instTypeFastRecurse }
func (CallBlockInstruction) Typ() instType         { return instTypeCallBlock }
func (LoadBlocksInstruction) Typ() instType        { return instTypeLoadBlocks }
func (IncludeInstruction) Typ() instType           { return instTypeInclude }
func (ExportLocalsInstruction) Typ() instType      { return instTypeExportLocals }
func (BuildMacroInstruction) Typ() instType        { return instTypeBuildMacro }
func (ReturnInstruction) Typ() instType            { return instTypeReturn }
func (IsUndefinedInstruction) Typ() instType       { return instTypeIsUndefined }
func (EncloseInstruction) Typ() instType           { return instTypeEnclose }
func (GetClosureInstruction) Typ() instType        { return instTypeGetClosure }

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

	// Loads a constant valu.Value.
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

	// Negates the valu.Value.
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

	// Jump if the stack top evaluates to false or pops the valu.Value
	instTypeJumpIfFalseOrPop

	// Jump if the stack top evaluates to true or pops the valu.Value
	instTypeJumpIfTrueOrPop

	// Sets the auto escape flag to the current valu.Value.
	instTypePushAutoEscape

	// Resets the auto escape flag to the previous valu.Value.
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

	// True if the valu.Value is undefined
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

func newInstructions(name, source string) Instructions {
	return Instructions{
		instructions: make([]Instruction, 0, 128),
		lineInfos:    make([]lineInfo, 0, 128),
		spanInfos:    make([]spanInfo, 0, 128),
		name:         name,
		source:       source,
	}
}

func (i *Instructions) add(instr Instruction) uint {
	rv := len(i.instructions)
	i.instructions = append(i.instructions, instr)
	return uint(rv)
}

func (i *Instructions) addLineRecord(instr uint, line uint32) {
	sameLoc := false
	if len(i.lineInfos) > 0 {
		lastLoc := i.lineInfos[len(i.lineInfos)-1]
		sameLoc = lastLoc.line == line
	}
	if !sameLoc {
		i.lineInfos = append(i.lineInfos, lineInfo{firstInstruction: uint32(instr), line: line})
	}
}

func (i *Instructions) addWithLine(instr Instruction, line uint32) uint {
	rv := i.add(instr)
	i.addLineRecord(rv, line)
	return rv
}

func (i *Instructions) addWithSpan(instr Instruction, spn Span) uint {
	rv := i.add(instr)

	sameLoc := false
	if len(i.spanInfos) > 0 {
		lastLoc := i.spanInfos[len(i.spanInfos)-1]
		sameLoc = option.IsSome(lastLoc.span) && option.Unwrap(lastLoc.span) == spn
	}
	if !sameLoc {
		i.spanInfos = append(i.spanInfos,
			spanInfo{firstInstruction: uint32(rv), span: option.Some(spn)})
	}

	i.addLineRecord(rv, spn.StartLine)
	return rv
}

func (i *Instructions) GetLine(idx uint) option.Option[uint] {
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

func (i *Instructions) GetSpan(idx uint) option.Option[Span] {
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
	return option.None[Span]()
}
