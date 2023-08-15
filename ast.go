package mjingo

import (
	"fmt"
	"strings"
)

type spanned[T any] struct {
	data T
	span span
}

type stmtKind int

const (
	stmtKindTemplate stmtKind = iota + 1
	stmtKindEmitExpr
	stmtKindEmitRaw
	stmtKindForLoop
	stmtKindIfCond
	stmtKindWithBlock
	stmtKindSet
	stmtKindSetBlock
	stmtKindAutoEscape
	stmtKindFilterBlock
	stmtKindBlock
	stmtKindImport
	stmtKindFromImport
	stmtKindExtends
	stmtKindInclude
	stmtKindMacro
	stmtKindCallBlock
	stmtKindDo
)

func (k stmtKind) String() string {
	switch k {
	case stmtKindTemplate:
		return "template"
	case stmtKindEmitExpr:
		return "emitExpr"
	case stmtKindEmitRaw:
		return "emitRaw"
	case stmtKindForLoop:
		return "forLoop"
	case stmtKindIfCond:
		return "ifCond"
	case stmtKindWithBlock:
		return "withBlock"
	case stmtKindSet:
		return "set"
	case stmtKindSetBlock:
		return "setBlock"
	case stmtKindAutoEscape:
		return "autoEscape"
	case stmtKindFilterBlock:
		return "filterBlock"
	case stmtKindBlock:
		return "block"
	case stmtKindImport:
		return "import"
	case stmtKindFromImport:
		return "fromImport"
	case stmtKindExtends:
		return "extends"
	case stmtKindInclude:
		return "include"
	case stmtKindMacro:
		return "macro"
	case stmtKindCallBlock:
		return "callBlock"
	case stmtKindDo:
		return "do"
	default:
		panic("invalid stmtKind")
	}
}

type stmt struct {
	kind stmtKind
	data any
	span span
}

func (s stmt) String() string {
	switch s.kind {
	case stmtKindTemplate:
		return fmt.Sprintf("(template %s)", s.data)
	case stmtKindEmitExpr:
		return fmt.Sprintf("(emitExpr %s)", s.data)
	case stmtKindEmitRaw:
		return fmt.Sprintf("(emitRaw %s)", s.data)
	case stmtKindForLoop:
		return "forLoop"
	case stmtKindIfCond:
		return "ifCond"
	case stmtKindWithBlock:
		return "withBlock"
	case stmtKindSet:
		return "set"
	case stmtKindSetBlock:
		return "setBlock"
	case stmtKindAutoEscape:
		return "autoEscape"
	case stmtKindFilterBlock:
		return "filterBlock"
	case stmtKindBlock:
		return "block"
	case stmtKindImport:
		return "import"
	case stmtKindFromImport:
		return "fromImport"
	case stmtKindExtends:
		return "extends"
	case stmtKindInclude:
		return "include"
	case stmtKindMacro:
		return "macro"
	case stmtKindCallBlock:
		return "callBlock"
	case stmtKindDo:
		return "do"
	default:
		panic("invalid stmtKind")
	}
}

type templateStmtData struct {
	children []stmt
}

func (s templateStmtData) String() string {
	var b strings.Builder
	b.WriteRune('[')
	for i, c := range s.children {
		if i > 0 {
			b.WriteRune(' ')
		}
		b.WriteString(c.String())
	}
	b.WriteRune(']')
	return b.String()
}

type emitExprStmtData struct {
	expr expr
}

func (s emitExprStmtData) String() string {
	return s.expr.String()
}

type emitRawStmtData struct {
	raw string
}

func (s emitRawStmtData) String() string {
	return fmt.Sprintf("%q", s.raw)
}

type forLoopStmtData struct {
	target     expr
	iter       expr
	filterExpr option[expr]
	recursive  bool
	body       []stmt
	elseBody   []stmt
}

type ifCondStmtData struct {
	expr      expr
	trueBody  []stmt
	falseBody []stmt
}

type withBlockStmtData struct {
	assignments []assignment
	body        []stmt
}

type assignment struct {
	lhs expr
	rhs expr
}

type setBlockStmtData struct {
	target expr
	expr   expr
}

type setStmtData struct {
	target expr
	filter option[expr]
	body   []stmt
}

type autoEscapeStmtData struct {
	enabled expr
	body    []stmt
}

type filterBlockStmtData struct {
	filter expr
	body   []stmt
}

type blockStmtData struct {
	target expr
	filter option[expr]
	body   []stmt
}

type importStmtData struct {
	expr expr
	name expr
}

type fromImportStmtData struct {
	expr  expr
	names []importName
}

type importName struct {
	name expr
	as   option[expr]
}

type extendsStmtData struct {
	name expr
}

type includeStmtData struct {
	name          expr
	ignoreMissing bool
}

type macroStmtData struct {
	name     string
	args     []expr
	defaults []expr
	body     []stmt
}

type callBlockStmtData struct {
	call      spanned[call]
	macroDecl spanned[macroStmtData]
}

type doStmtData struct {
	call spanned[call]
}

type call struct {
	expr expr
	args []expr
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
