package mjingo

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
	stmtKindDo
)

type stmt struct {
	kind stmtKind
	data any
	span span
}

type templateStmtData struct {
	children []stmt
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

type expr struct {
	kind exprKind
	data any
	span span
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

type unaryOpData struct {
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

type binOpData struct {
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
