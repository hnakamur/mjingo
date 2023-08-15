package mjingo

type statement interface {
	typ() stmtType
}

type templateStmt struct {
	children []statement
	span     span
}
type emitExprStmt struct {
	expr expression
	span span
}
type emitRawStmt struct {
	raw  string
	span span
}
type forLoopStmt struct {
	target     expression
	iter       expression
	filterExpr option[expression]
	recursive  bool
	body       []statement
	elseBody   []statement
	span       span
}
type ifCondStmt struct {
	expr      expression
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
	target expression
	expr   expression
	span   span
}
type setBlockStmt struct {
	target expression
	filter option[expression]
	body   []statement
	span   span
}
type autoEscapeStmt struct {
	enabled expression
	body    []statement
	span    span
}
type filterBlockStmt struct {
	filter expression
	body   []statement
	span   span
}
type blockStmt struct {
	name string
	body []statement
	span span
}
type importStmt struct {
	expr expression
	name expression
	span span
}
type fromImportStmt struct {
	expr  expression
	names []importName
	span  span
}
type extendsStmt struct {
	name expression
	span span
}
type includeStmt struct {
	name          expression
	ignoreMissing bool
	span          span
}
type macroStmt struct {
	name     string
	args     []expression
	defaults []expression
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
	lhs expression
	rhs expression
}

type importName struct {
	name expression
	as   option[expression]
}

type call struct {
	expr expression
	args []expression
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
	span span
}

type constExpr struct {
	value value
	span  span
}

type sliceExpr struct {
	expr  expression
	start option[expression]
	stop  option[expression]
	step  option[expression]
	span  span
}

type unaryOpExpr struct {
	op   unaryOpType
	expr expression
	span span
}

type binOpExpr struct {
	op    binOpKind
	left  expression
	right expression
	span  span
}

type ifExpr struct {
	testExpr  expression
	trueExpr  expression
	falseExpr option[expression]
	span      span
}

type filterExpr struct {
	name string
	expr option[expression]
	args []expression
	span span
}

type testExpr struct {
	name string
	expr expression
	args []expression
	span span
}

type getAttrExpr struct {
	expr expression
	name string
	span span
}

type getItemExpr struct {
	expr          expression
	subscriptExpr expression
	span          span
}

type callExpr struct {
	expr expression
	args []expression
	span span
}

type listExpr struct {
	items []expression
	span  span
}

type mapExpr struct {
	keys   []expression
	values []expression
	span   span
}

type kwargsExpr struct {
	pairs []kwarg
	span  span
}

type kwarg struct {
	key string
	arg expression
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

type unaryOpType int

const (
	unaryOpTypeNot unaryOpType = iota + 1
	unaryOpTypeNeg
)

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

func (l listExpr) asConst() option[value] {
	for _, item := range l.items {
		if _, ok := item.(constExpr); !ok {
			return option[value]{}
		}
	}

	seq := make([]value, 0, len(l.items))
	for _, item := range l.items {
		if item, ok := item.(constExpr); ok {
			seq = append(seq, item.value)
		}
	}
	return option[value]{valid: true, data: seqValue{items: seq}}
}

func (m mapExpr) asConst() option[value] {
	for _, key := range m.keys {
		if _, ok := key.(constExpr); !ok {
			return option[value]{}
		}
	}
	for _, val := range m.values {
		if _, ok := val.(constExpr); !ok {
			return option[value]{}
		}
	}

	rv := make(map[string]value, len(m.keys))
	for i, key := range m.keys {
		val := m.values[i]
		if key.typ() == exprTypeConst && val.typ() == exprTypeConst {
			// implmentation here is different from minijinja
			if keyStr := key.(constExpr).value.asStr(); keyStr.valid {
				rv[keyStr.data] = val.(constExpr).value
			}
		}
	}
	return option[value]{valid: true, data: mapValue{m: rv}}
}
