package mjingo

import (
	"fmt"
)

type statement interface {
	typ() stmtType
}

type templateStmt struct {
	children []statement
	span     span
}
type emitExprStmt struct {
	expr expr
	span span
}
type emitRawStmt struct {
	raw  string
	span span
}
type forLoopStmt struct {
	target     expr
	iter       expr
	filterExpr option[expr]
	recursive  bool
	body       []statement
	elseBody   []statement
	span       span
}
type ifCondStmt struct {
	expr      expr
	trueBody  []statement
	falseBody []statement
	span      span
}
type withBlockStmt struct {
	assignments []assignment
	body        []statement
	span        span
}
type setStmt struct {
	target expr
	expr   expr
	span   span
}
type setBlockStmt struct {
	target expr
	filter option[expr]
	body   []statement
	span   span
}
type autoEscapeStmt struct {
	enabled expr
	body    []statement
	span    span
}
type filterBlockStmt struct {
	filter expr
	body   []statement
	span   span
}
type blockStmt struct {
	name string
	body []statement
	span span
}
type importStmt struct {
	expr expr
	name expr
	span span
}
type fromImportStmt struct {
	expr  expr
	names []importName
	span  span
}
type extendsStmt struct {
	name expr
	span span
}
type includeStmt struct {
	name          expr
	ignoreMissing bool
	span          span
}
type macroStmt struct {
	name     string
	args     []expr
	defaults []expr
	body     []statement
	span     span
}
type callBlockStmt struct {
	call      call
	macroDecl macroStmt
	span      span
}
type doStmt struct {
	call call
	span span
}

type assignment struct {
	lhs expr
	rhs expr
}

type importName struct {
	name expr
	as   option[expr]
}

type call struct {
	expr expr
	args []expr
	span span
}

var _ = statement(templateStmt{})
var _ = statement(emitExprStmt{})
var _ = statement(emitRawStmt{})
var _ = statement(forLoopStmt{})
var _ = statement(ifCondStmt{})
var _ = statement(withBlockStmt{})
var _ = statement(setStmt{})
var _ = statement(setBlockStmt{})
var _ = statement(autoEscapeStmt{})
var _ = statement(filterBlockStmt{})
var _ = statement(blockStmt{})
var _ = statement(importStmt{})
var _ = statement(fromImportStmt{})
var _ = statement(extendsStmt{})
var _ = statement(includeStmt{})
var _ = statement(macroStmt{})
var _ = statement(callBlockStmt{})
var _ = statement(doStmt{})

func (templateStmt) typ() stmtType    { return stmtTypeTemplate }
func (emitExprStmt) typ() stmtType    { return stmtTypeEmitExpr }
func (emitRawStmt) typ() stmtType     { return stmtTypeEmitRaw }
func (forLoopStmt) typ() stmtType     { return stmtTypeForLoop }
func (ifCondStmt) typ() stmtType      { return stmtTypeIfCond }
func (withBlockStmt) typ() stmtType   { return stmtTypeWithBlock }
func (setStmt) typ() stmtType         { return stmtTypeSet }
func (setBlockStmt) typ() stmtType    { return stmtTypeSetBlock }
func (autoEscapeStmt) typ() stmtType  { return stmtTypeAutoEscape }
func (filterBlockStmt) typ() stmtType { return stmtTypeFilterBlock }
func (blockStmt) typ() stmtType       { return stmtTypeBlock }
func (importStmt) typ() stmtType      { return stmtTypeImport }
func (fromImportStmt) typ() stmtType  { return stmtTypeFromImport }
func (extendsStmt) typ() stmtType     { return stmtTypeExtends }
func (includeStmt) typ() stmtType     { return stmtTypeInclude }
func (macroStmt) typ() stmtType       { return stmtTypeMacro }
func (callBlockStmt) typ() stmtType   { return stmtTypeCallBlock }
func (doStmt) typ() stmtType          { return stmtTypeDo }

type spanned[T any] struct {
	data T
	span span
}

type stmtType int

const (
	stmtTypeTemplate stmtType = iota + 1
	stmtTypeEmitExpr
	stmtTypeEmitRaw
	stmtTypeForLoop
	stmtTypeIfCond
	stmtTypeWithBlock
	stmtTypeSet
	stmtTypeSetBlock
	stmtTypeAutoEscape
	stmtTypeFilterBlock
	stmtTypeBlock
	stmtTypeImport
	stmtTypeFromImport
	stmtTypeExtends
	stmtTypeInclude
	stmtTypeMacro
	stmtTypeCallBlock
	stmtTypeDo
)

func (k stmtType) String() string {
	switch k {
	case stmtTypeTemplate:
		return "template"
	case stmtTypeEmitExpr:
		return "emitExpr"
	case stmtTypeEmitRaw:
		return "emitRaw"
	case stmtTypeForLoop:
		return "forLoop"
	case stmtTypeIfCond:
		return "ifCond"
	case stmtTypeWithBlock:
		return "withBlock"
	case stmtTypeSet:
		return "set"
	case stmtTypeSetBlock:
		return "setBlock"
	case stmtTypeAutoEscape:
		return "autoEscape"
	case stmtTypeFilterBlock:
		return "filterBlock"
	case stmtTypeBlock:
		return "block"
	case stmtTypeImport:
		return "import"
	case stmtTypeFromImport:
		return "fromImport"
	case stmtTypeExtends:
		return "extends"
	case stmtTypeInclude:
		return "include"
	case stmtTypeMacro:
		return "macro"
	case stmtTypeCallBlock:
		return "callBlock"
	case stmtTypeDo:
		return "do"
	default:
		panic("invalid stmtType")
	}
}

type exprKind int

const (
	exprKindVar exprKind = iota + 1
	exprKindConst
	exprKindSlice
	exprKindUnaryOp
	exprKindBinOp
	exprKindIfExpr
	exprKindFilter
	exprKindTest
	exprKindGetAttr
	exprKindGetItem
	exprKindCall
	exprKindList
	exprKindMap
	exprKindKwargs
)

func (k exprKind) String() string {
	switch k {
	case exprKindVar:
		return "var"
	case exprKindConst:
		return "const"
	case exprKindSlice:
		return "slice"
	case exprKindUnaryOp:
		return "unaryOp"
	case exprKindBinOp:
		return "binOp"
	case exprKindIfExpr:
		return "ifExpr"
	case exprKindFilter:
		return "filter"
	case exprKindTest:
		return "test"
	case exprKindGetAttr:
		return "getAttr"
	case exprKindGetItem:
		return "getItem"
	case exprKindCall:
		return "call"
	case exprKindList:
		return "list"
	case exprKindMap:
		return "map"
	case exprKindKwargs:
		return "kwargs"
	default:
		panic("invalid exprKind")
	}
}

type expr struct {
	kind exprKind
	data any
	span span
}

func (e expr) String() string {
	switch e.kind {
	case exprKindVar:
		return fmt.Sprintf("(var %s)", e.data)
	case exprKindConst:
		return "const"
	case exprKindSlice:
		return "slice"
	case exprKindUnaryOp:
		return "unaryOp"
	case exprKindBinOp:
		return "binOp"
	case exprKindIfExpr:
		return "ifExpr"
	case exprKindFilter:
		return "filter"
	case exprKindTest:
		return "test"
	case exprKindGetAttr:
		return "getAttr"
	case exprKindGetItem:
		return "getItem"
	case exprKindCall:
		return "call"
	case exprKindList:
		return "list"
	case exprKindMap:
		return "map"
	case exprKindKwargs:
		return "kwargs"
	default:
		panic("invalid exprKind")
	}
}

type varExprData struct {
	id string
}

type constExprData struct {
	value value
}

type sliceExprData struct {
	expr  expr
	start option[expr]
	stop  option[expr]
	step  option[expr]
}

type unaryOpKind int

const (
	unaryOpKindNot unaryOpKind = iota + 1
	unaryOpKindNeg
)

type unaryOpExprData struct {
	op   unaryOpKind
	expr expr
}

type binOpKind int

const (
	binOpKindEq binOpKind = iota + 1
	binOpKindNe
	binOpKindLt
	binOpKindLte
	binOpKindGt
	binOpKindGte
	binOpKindScAnd
	binOpKindScOr
	binOpKindAdd
	binOpKindSub
	binOpKindMul
	binOpKindDiv
	binOpKindFloorDiv
	binOpKindRem
	binOpKindPow
	binOpKindConcat
	binOpKindIn
)

type binOpExprData struct {
	op    binOpKind
	left  expr
	right expr
}

type ifExprExprData struct {
	testExpr  expr
	trueExpr  expr
	falseExpr option[expr]
}

type filterExprData struct {
	name string
	expr option[expr]
	args []expr
}

type testExprData struct {
	name string
	expr expr
	args []expr
}

type getAttrExprData struct {
	expr expr
	name string
}

type getItemExprData struct {
	expr          expr
	subscriptExpr expr
}

type callExprData struct {
	expr expr
	args []expr
}

type listExprData struct {
	items []expr
}

func (l *listExprData) asConst() option[value] {
	for _, item := range l.items {
		if item.kind != exprKindConst {
			return option[value]{}
		}
	}

	seq := make([]value, 0, len(l.items))
	for _, item := range l.items {
		if item.kind == exprKindConst {
			data := item.data.(constExprData)
			seq = append(seq, data.value)
		}
	}
	return option[value]{valid: true, data: seqValue{items: seq}}
}

type kwargsExprData struct {
	pairs []kwarg
}

type kwarg struct {
	key string
	arg expr
}

type mapExprData struct {
	keys   []expr
	values []expr
}

func (m *mapExprData) asConst() option[value] {
	for _, key := range m.keys {
		if key.kind != exprKindConst {
			return option[value]{}
		}
	}
	for _, v := range m.values {
		if v.kind != exprKindConst {
			return option[value]{}
		}
	}

	rv := make(map[string]value, len(m.keys))
	for i, key := range m.keys {
		v := m.values[i]
		if key.kind == exprKindConst && v.kind == exprKindConst {
			keyData := key.data.(constExprData)
			// implmentation here is different from minijinja
			if keyStr := keyData.value.asStr(); keyStr.valid {
				valData := v.data.(constExprData)
				rv[keyStr.data] = valData.value
			}
		}
	}
	return option[value]{valid: true, data: mapValue{m: rv}}
}
