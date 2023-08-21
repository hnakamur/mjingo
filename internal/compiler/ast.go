package compiler

import (
	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/value"
)

type statement interface {
	typ() stmtType
}

type templateStmt struct {
	children []statement
	span     internal.Span
}
type emitExprStmt struct {
	expr expression
	span internal.Span
}
type emitRawStmt struct {
	raw  string
	span internal.Span
}
type forLoopStmt struct {
	target     expression
	iter       expression
	filterExpr option.Option[expression]
	recursive  bool
	body       []statement
	elseBody   []statement
	span       internal.Span
}
type ifCondStmt struct {
	expr      expression
	trueBody  []statement
	falseBody []statement
	span      internal.Span
}
type withBlockStmt struct {
	assignments []assignment
	body        []statement
	span        internal.Span
}
type setStmt struct {
	target expression
	expr   expression
	span   internal.Span
}
type setBlockStmt struct {
	target expression
	filter option.Option[expression]
	body   []statement
	span   internal.Span
}
type autoEscapeStmt struct {
	enabled expression
	body    []statement
	span    internal.Span
}
type filterBlockStmt struct {
	filter expression
	body   []statement
	span   internal.Span
}
type blockStmt struct {
	name string
	body []statement
	span internal.Span
}
type importStmt struct {
	expr expression
	name expression
	span internal.Span
}
type fromImportStmt struct {
	expr  expression
	names []importName
	span  internal.Span
}
type extendsStmt struct {
	name expression
	span internal.Span
}
type includeStmt struct {
	name          expression
	ignoreMissing bool
	span          internal.Span
}
type macroStmt struct {
	name     string
	args     []expression
	defaults []expression
	body     []statement
	span     internal.Span
}
type callBlockStmt struct {
	call      call
	macroDecl macroStmt
	span      internal.Span
}
type doStmt struct {
	call call
	span internal.Span
}

type assignment struct {
	lhs expression
	rhs expression
}

type importName struct {
	name expression
	as   option.Option[expression]
}

type call struct {
	expr expression
	args []expression
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

type expression interface {
	typ() exprType
}

type varExpr struct {
	id   string
	span internal.Span
}

type constExpr struct {
	value value.Value
	span  internal.Span
}

type sliceExpr struct {
	expr  expression
	start option.Option[expression]
	stop  option.Option[expression]
	step  option.Option[expression]
	span  internal.Span
}

type unaryOpExpr struct {
	op   unaryOpType
	expr expression
	span internal.Span
}

type binOpExpr struct {
	op    binOpType
	left  expression
	right expression
	span  internal.Span
}

type ifExpr struct {
	testExpr  expression
	trueExpr  expression
	falseExpr option.Option[expression]
	span      internal.Span
}

type filterExpr struct {
	name string
	expr option.Option[expression]
	args []expression
	span internal.Span
}

type testExpr struct {
	name string
	expr expression
	args []expression
	span internal.Span
}

type getAttrExpr struct {
	expr expression
	name string
	span internal.Span
}

type getItemExpr struct {
	expr          expression
	subscriptExpr expression
	span          internal.Span
}

type callExpr struct {
	call call
	span internal.Span
}

type listExpr struct {
	items []expression
	span  internal.Span
}

type mapExpr struct {
	keys   []expression
	values []expression
	span   internal.Span
}

type kwargsExpr struct {
	pairs []kwargExpr
	span  internal.Span
}

type kwargExpr struct {
	key string
	arg expression
}

func (e kwargsExpr) asConst() option.Option[value.Value] {
	if !slicex.All(e.pairs, func(x kwargExpr) bool {
		_, ok := x.arg.(constExpr)
		return ok
	}) {
		return option.None[value.Value]()
	}

	rv := value.NewIndexMapWithCapacity(uint(len(e.pairs)))
	for _, pair := range e.pairs {
		if v, ok := pair.arg.(constExpr); ok {
			rv.Set(value.KeyRefFromValue(value.FromString(pair.key)), v.value.Clone())
		}
	}
	return option.Some(value.FromKwargs(value.Kwargs{Values: *rv}))
}

var _ = expression(varExpr{})
var _ = expression(constExpr{})
var _ = expression(sliceExpr{})
var _ = expression(unaryOpExpr{})
var _ = expression(binOpExpr{})
var _ = expression(ifExpr{})
var _ = expression(filterExpr{})
var _ = expression(testExpr{})
var _ = expression(getAttrExpr{})
var _ = expression(getItemExpr{})
var _ = expression(callExpr{})
var _ = expression(listExpr{})
var _ = expression(mapExpr{})
var _ = expression(kwargsExpr{})

func (varExpr) typ() exprType     { return exprTypeVar }
func (constExpr) typ() exprType   { return exprTypeConst }
func (sliceExpr) typ() exprType   { return exprTypeSlice }
func (unaryOpExpr) typ() exprType { return exprTypeUnaryOp }
func (binOpExpr) typ() exprType   { return exprTypeBinOp }
func (ifExpr) typ() exprType      { return exprTypeIfExpr }
func (filterExpr) typ() exprType  { return exprTypeFilter }
func (testExpr) typ() exprType    { return exprTypeTest }
func (getAttrExpr) typ() exprType { return exprTypeGetAttr }
func (getItemExpr) typ() exprType { return exprTypeGetItem }
func (callExpr) typ() exprType    { return exprTypeCall }
func (listExpr) typ() exprType    { return exprTypeList }
func (mapExpr) typ() exprType     { return exprTypeMap }
func (kwargsExpr) typ() exprType  { return exprTypeKwargs }

type exprType int

const (
	exprTypeVar exprType = iota + 1
	exprTypeConst
	exprTypeSlice
	exprTypeUnaryOp
	exprTypeBinOp
	exprTypeIfExpr
	exprTypeFilter
	exprTypeTest
	exprTypeGetAttr
	exprTypeGetItem
	exprTypeCall
	exprTypeList
	exprTypeMap
	exprTypeKwargs
)

func (k exprType) String() string {
	switch k {
	case exprTypeVar:
		return "var"
	case exprTypeConst:
		return "const"
	case exprTypeSlice:
		return "slice"
	case exprTypeUnaryOp:
		return "unaryOp"
	case exprTypeBinOp:
		return "binOp"
	case exprTypeIfExpr:
		return "ifExpr"
	case exprTypeFilter:
		return "filter"
	case exprTypeTest:
		return "test"
	case exprTypeGetAttr:
		return "getAttr"
	case exprTypeGetItem:
		return "getItem"
	case exprTypeCall:
		return "call"
	case exprTypeList:
		return "list"
	case exprTypeMap:
		return "map"
	case exprTypeKwargs:
		return "kwargs"
	default:
		panic("invalid exprType")
	}
}

func (k exprType) Description() string {
	switch k {
	case exprTypeVar:
		return "variable"
	case exprTypeConst:
		return "constant"
	case exprTypeSlice, exprTypeUnaryOp, exprTypeBinOp, exprTypeIfExpr, exprTypeGetAttr, exprTypeGetItem:
		return "expression"
	case exprTypeTest:
		return "test expression"
	case exprTypeFilter:
		return "filter expression"
	case exprTypeCall:
		return "call"
	case exprTypeList:
		return "list literal"
	case exprTypeMap:
		return "map literal"
	case exprTypeKwargs:
		return "keyword arguments"
	default:
		panic("invalid exprType")
	}
}

type unaryOpType int

const (
	unaryOpTypeNot unaryOpType = iota + 1
	unaryOpTypeNeg
)

type binOpType int

const (
	binOpTypeEq binOpType = iota + 1
	binOpTypeNe
	binOpTypeLt
	binOpTypeLte
	binOpTypeGt
	binOpTypeGte
	binOpTypeScAnd
	binOpTypeScOr
	binOpTypeAdd
	binOpTypeSub
	binOpTypeMul
	binOpTypeDiv
	binOpTypeFloorDiv
	binOpTypeRem
	binOpTypePow
	binOpTypeConcat
	binOpTypeIn
)

func (l listExpr) asConst() option.Option[value.Value] {
	for _, item := range l.items {
		if _, ok := item.(constExpr); !ok {
			return option.None[value.Value]()
		}
	}

	seq := make([]value.Value, 0, len(l.items))
	for _, item := range l.items {
		if item, ok := item.(constExpr); ok {
			seq = append(seq, item.value)
		}
	}
	return option.Some(value.FromSlice(seq))
}

func (m mapExpr) asConst() option.Option[value.Value] {
	for _, key := range m.keys {
		if _, ok := key.(constExpr); !ok {
			return option.None[value.Value]()
		}
	}
	for _, val := range m.values {
		if _, ok := val.(constExpr); !ok {
			return option.None[value.Value]()
		}
	}

	rv := value.NewIndexMapWithCapacity(uint(len(m.keys)))
	for i, key := range m.keys {
		val := m.values[i]
		if key.typ() == exprTypeConst && val.typ() == exprTypeConst {
			keyRf := value.KeyRefFromValue(key.(constExpr).value)
			rv.Set(keyRf, val.(constExpr).value)
		}
	}
	return option.Some(value.FromIndexMap(rv))
}

type callType interface {
	kind() callTypeKind
}

type callTypeFunction struct{ name string }
type callTypeMethod struct {
	expr expression
	name string
}
type callTypeBlock struct{ name string }
type callTypeObject struct{ expr expression }

func (callTypeFunction) kind() callTypeKind { return callTypeKindFunction }
func (callTypeMethod) kind() callTypeKind   { return callTypeKindMethod }
func (callTypeBlock) kind() callTypeKind    { return callTypeKindBlock }
func (callTypeObject) kind() callTypeKind   { return callTypeKindObject }

var _ = (callType)(callTypeFunction{})
var _ = (callType)(callTypeMethod{})
var _ = (callType)(callTypeBlock{})
var _ = (callType)(callTypeObject{})

type callTypeKind uint

const (
	callTypeKindFunction callTypeKind = iota + 1
	callTypeKindMethod
	callTypeKindBlock
	callTypeKindObject
)

func (c *call) identityCall() callType {
	switch exp := c.expr.(type) {
	case varExpr:
		return callTypeFunction{name: exp.id}
	case getAttrExpr:
		if varExp, ok := exp.expr.(varExpr); ok {
			if varExp.id == "self" {
				return callTypeBlock{name: exp.name}
			}
		}
		return callTypeMethod{expr: exp.expr, name: exp.name}
	default:
		return callTypeObject{expr: c.expr}
	}
}
