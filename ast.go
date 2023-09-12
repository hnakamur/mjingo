package mjingo

import (
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/option"
)

type statement interface {
	typ() stmtType
}

type templateStmt struct {
	children []statement
	span     span
}
type emitExprStmt struct {
	expr astExpr
	span span
}
type emitRawStmt struct {
	raw  string
	span span
}
type forLoopStmt struct {
	target     astExpr
	iter       astExpr
	filterExpr option.Option[astExpr]
	recursive  bool
	body       []statement
	elseBody   []statement
	span       span
}
type ifCondStmt struct {
	expr      astExpr
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
	target astExpr
	expr   astExpr
	span   span
}
type setBlockStmt struct {
	target astExpr
	filter option.Option[astExpr]
	body   []statement
	span   span
}
type autoEscapeStmt struct {
	enabled astExpr
	body    []statement
	span    span
}
type filterBlockStmt struct {
	filter astExpr
	body   []statement
	span   span
}
type blockStmt struct {
	name string
	body []statement
	span span
}
type importStmt struct {
	expr astExpr
	name astExpr
	span span
}
type fromImportStmt struct {
	expr  astExpr
	names []importName
	span  span
}
type extendsStmt struct {
	name astExpr
	span span
}
type includeStmt struct {
	name          astExpr
	ignoreMissing bool
	span          span
}
type macroStmt struct {
	name     string
	args     []astExpr
	defaults []astExpr
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
	lhs astExpr
	rhs astExpr
}

type importName struct {
	name astExpr
	as   option.Option[astExpr]
}

type call struct {
	expr astExpr
	args []astExpr
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

type astExpr interface {
	typ() exprType
}

type varExpr struct {
	id   string
	span span
}

type constExpr struct {
	val  Value
	span span
}

type sliceExpr struct {
	expr  astExpr
	start option.Option[astExpr]
	stop  option.Option[astExpr]
	step  option.Option[astExpr]
	span  span
}

type unaryOpExpr struct {
	op   unaryOpType
	expr astExpr
	span span
}

type binOpExpr struct {
	op    binOpType
	left  astExpr
	right astExpr
	span  span
}

type ifExpr struct {
	testExpr  astExpr
	trueExpr  astExpr
	falseExpr option.Option[astExpr]
	span      span
}

type filterExpr struct {
	name string
	expr option.Option[astExpr]
	args []astExpr
	span span
}

type testExpr struct {
	name string
	expr astExpr
	args []astExpr
	span span
}

type getAttrExpr struct {
	expr astExpr
	name string
	span span
}

type getItemExpr struct {
	expr          astExpr
	subscriptExpr astExpr
	span          span
}

type callExpr struct {
	call call
	span span
}

type listExpr struct {
	items []astExpr
	span  span
}

type mapExpr struct {
	keys   []astExpr
	values []astExpr
	span   span
}

type kwargsExpr struct {
	pairs []kwargExpr
	span  span
}

type kwargExpr struct {
	key string
	arg astExpr
}

func (e kwargsExpr) asConst() option.Option[Value] {
	if !slicex.All(e.pairs, func(x kwargExpr) bool {
		_, ok := x.arg.(constExpr)
		return ok
	}) {
		return option.None[Value]()
	}

	rv := valueMapWithCapacity(uint(len(e.pairs)))
	for _, pair := range e.pairs {
		if v, ok := pair.arg.(constExpr); ok {
			rv.Set(keyRefFromValue(valueFromString(pair.key)), v.val.clone())
		}
	}
	return option.Some(valueFromKwargs(newKwargs(*rv)))
}

var _ = astExpr(varExpr{})
var _ = astExpr(constExpr{})
var _ = astExpr(sliceExpr{})
var _ = astExpr(unaryOpExpr{})
var _ = astExpr(binOpExpr{})
var _ = astExpr(ifExpr{})
var _ = astExpr(filterExpr{})
var _ = astExpr(testExpr{})
var _ = astExpr(getAttrExpr{})
var _ = astExpr(getItemExpr{})
var _ = astExpr(callExpr{})
var _ = astExpr(listExpr{})
var _ = astExpr(mapExpr{})
var _ = astExpr(kwargsExpr{})

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

func (l listExpr) asConst() option.Option[Value] {
	for _, item := range l.items {
		if _, ok := item.(constExpr); !ok {
			return option.None[Value]()
		}
	}

	seq := make([]Value, 0, len(l.items))
	for _, item := range l.items {
		if item, ok := item.(constExpr); ok {
			seq = append(seq, item.val)
		}
	}
	return option.Some(valueFromSlice(seq))
}

func (m mapExpr) asConst() option.Option[Value] {
	for _, key := range m.keys {
		if _, ok := key.(constExpr); !ok {
			return option.None[Value]()
		}
	}
	for _, val := range m.values {
		if _, ok := val.(constExpr); !ok {
			return option.None[Value]()
		}
	}

	rv := valueMapWithCapacity(uint(len(m.keys)))
	for i, key := range m.keys {
		val := m.values[i]
		if key.typ() == exprTypeConst && val.typ() == exprTypeConst {
			keyRf := keyRefFromValue(key.(constExpr).val)
			rv.Set(keyRf, val.(constExpr).val)
		}
	}
	return option.Some(valueFromIndexMap(rv))
}

type callType interface {
	kind() callTypeKind
}

type callTypeFunction struct{ name string }
type callTypeMethod struct {
	expr astExpr
	name string
}
type callTypeBlock struct{ name string }
type callTypeObject struct{ expr astExpr }

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
