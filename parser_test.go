package mjingo

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	debugStringSpan := func(spn span) string {
		return fmt.Sprintf("%d:%d-%d:%d", spn.StartLine, spn.StartCol, spn.EndLine, spn.EndCol)
	}

	const indent1 = "    "

	var debugStringExprHelper func(exp astExpr, w io.Writer, prefix, indent string)
	debugStringExprHelper = func(exp astExpr, w io.Writer, prefix, indent string) {
		switch e := exp.(type) {
		case constExpr:
			fmt.Fprintf(w, "%sConst {\n", prefix)
			fmt.Fprintf(w, "%s%svalue: %s,\n", indent, indent1, e.val.DebugString())
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case callExpr:
			fmt.Fprintf(w, "%sCall {\n", prefix)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(e.call.expr, w, "", indent+indent1)
			if len(e.call.args) == 0 {
				fmt.Fprintf(w, "%s%sargs: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sargs: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, arg := range e.call.args {
					debugStringExprHelper(arg, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case varExpr:
			fmt.Fprintf(w, "%sVar {\n", prefix)
			fmt.Fprintf(w, "%s%sid: %q,\n", indent, indent1, e.id)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case getAttrExpr:
			fmt.Fprintf(w, "%sGetAttr {\n", prefix)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(e.expr, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%sname: %q,\n", indent, indent1, e.name)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case kwargsExpr:
			fmt.Fprintf(w, "%sKwargs {\n", prefix)
			fmt.Fprintf(w, "%s%spairs: [\n", indent, indent1)
			nextIndet := indent + indent1 + indent1 + indent1
			for _, pair := range e.pairs {
				fmt.Fprintf(w, "%s%s%s(\n", indent, indent1, indent1)
				fmt.Fprintf(w, "%s%s%s%s%q,\n", indent, indent1, indent1, indent1, pair.key)
				debugStringExprHelper(pair.arg, w, nextIndet, nextIndet)
				fmt.Fprintf(w, "%s%s%s),\n", indent, indent1, indent1)
			}
			fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case binOpExpr:
			fmt.Fprintf(w, "%sBinOp {\n", prefix)
			fmt.Fprintf(w, "%s%sop: %s,\n", indent, indent1, e.op)
			fmt.Fprintf(w, "%s%sleft: ", indent, indent1)
			debugStringExprHelper(e.left, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%sright: ", indent, indent1)
			debugStringExprHelper(e.right, w, "", indent+indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case unaryOpExpr:
			fmt.Fprintf(w, "%sUnaryOp {\n", prefix)
			fmt.Fprintf(w, "%s%sop: %s,\n", indent, indent1, e.op)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(e.expr, w, "", indent+indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case filterExpr:
			fmt.Fprintf(w, "%sFilter {\n", prefix)
			fmt.Fprintf(w, "%s%sname: %q,\n", indent, indent1, e.name)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			if e.expr.IsSome() {
				fmt.Fprint(w, "Some(\n")
				nextIndent := indent + indent1 + indent1
				debugStringExprHelper(e.expr.Unwrap(), w, nextIndent, nextIndent)
				fmt.Fprintf(w, "%s%s),\n", indent, indent1)
			} else {
				fmt.Fprint(w, "None,\n")
			}
			if len(e.args) == 0 {
				fmt.Fprintf(w, "%s%sargs: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sargs: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, arg := range e.args {
					debugStringExprHelper(arg, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case listExpr:
			fmt.Fprintf(w, "%sList {\n", prefix)
			if len(e.items) == 0 {
				fmt.Fprintf(w, "%s%sitems: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sitems: [\n", indent, indent1)
				nextIndet := indent + indent1 + indent1
				for _, item := range e.items {
					debugStringExprHelper(item, w, nextIndet, nextIndet)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case getItemExpr:
			fmt.Fprintf(w, "%sGetItem {\n", prefix)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			nextIndet := indent + indent1
			debugStringExprHelper(e.expr, w, "", nextIndet)
			fmt.Fprintf(w, "%s%ssubscript_expr: ", indent, indent1)
			debugStringExprHelper(e.subscriptExpr, w, "", nextIndet)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case ifExpr:
			fmt.Fprintf(w, "%sIfExpr {\n", prefix)
			fmt.Fprintf(w, "%s%stest_expr: ", indent, indent1)
			nextIndet := indent + indent1
			debugStringExprHelper(e.testExpr, w, "", nextIndet)
			fmt.Fprintf(w, "%s%strue_expr: ", indent, indent1)
			debugStringExprHelper(e.trueExpr, w, "", nextIndet)
			fmt.Fprintf(w, "%s%sfalse_expr: ", indent, indent1)
			if e.falseExpr.IsSome() {
				fmt.Fprint(w, "Some(\n")
				nextIndent := indent + indent1 + indent1
				debugStringExprHelper(e.falseExpr.Unwrap(), w, nextIndent, nextIndent)
				fmt.Fprintf(w, "%s%s),\n", indent, indent1)
			} else {
				fmt.Fprint(w, "None,\n")
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case mapExpr:
			fmt.Fprintf(w, "%sMap {\n", prefix)
			if len(e.keys) == 0 {
				fmt.Fprintf(w, "%s%skeys: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%skeys: [\n", indent, indent1)
				nextIndet := indent + indent1 + indent1
				for _, key := range e.keys {
					debugStringExprHelper(key, w, nextIndet, nextIndet)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			if len(e.values) == 0 {
				fmt.Fprintf(w, "%s%svalues: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%svalues: [\n", indent, indent1)
				nextIndet := indent + indent1 + indent1
				for _, val := range e.values {
					debugStringExprHelper(val, w, nextIndet, nextIndet)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		case testExpr:
			fmt.Fprintf(w, "%sTest {\n", prefix)
			fmt.Fprintf(w, "%s%sname: %q,\n", indent, indent1, e.name)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(e.expr, w, "", indent+indent1)
			if len(e.args) == 0 {
				fmt.Fprintf(w, "%s%sargs: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sargs: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, arg := range e.args {
					debugStringExprHelper(arg, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(e.span))
		default:
			panic(fmt.Sprintf("not implemented yet for expression type: %T", exp))
		}
	}

	var debugStringStmtHelper func(stmt statement, w io.Writer, prefix, indent string)
	debugStringStmtHelper = func(stmt statement, w io.Writer, prefix, indent string) {
		switch st := stmt.(type) {
		case templateStmt:
			fmt.Fprintf(w, "%sTemplate {\n", prefix)
			fmt.Fprintf(w, "%s%schildren: [\n", indent, indent1)
			nextIndent := indent + indent1 + indent1
			for _, child := range st.children {
				debugStringStmtHelper(child, w, nextIndent, nextIndent)
			}
			fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case autoEscapeStmt:
			fmt.Fprintf(w, "%sAutoEscape {\n", prefix)
			fmt.Fprintf(w, "%s%senabled: ", indent, indent1)
			debugStringExprHelper(st.enabled, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%sbody: [\n", indent, indent1)
			nextIndent := indent + indent1 + indent1
			for _, child := range st.body {
				debugStringStmtHelper(child, w, nextIndent, nextIndent)
			}
			fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case emitRawStmt:
			fmt.Fprintf(w, "%sEmitRaw {\n", prefix)
			fmt.Fprintf(w, "%s%sraw: %q,\n", indent, indent1, st.raw)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case emitExprStmt:
			fmt.Fprintf(w, "%sEmitExpr {\n", prefix)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(st.expr, w, "", indent+indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case blockStmt:
			fmt.Fprintf(w, "%sBlock {\n", prefix)
			fmt.Fprintf(w, "%s%sname: %q,\n", indent, indent1, st.name)
			if len(st.body) == 0 {
				fmt.Fprintf(w, "%s%sbody: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sbody: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.body {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case extendsStmt:
			fmt.Fprintf(w, "%sExtends {\n", prefix)
			fmt.Fprintf(w, "%s%sname: ", indent, indent1)
			debugStringExprHelper(st.name, w, "", indent+indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case filterBlockStmt:
			fmt.Fprintf(w, "%sFilterBlock {\n", prefix)
			fmt.Fprintf(w, "%s%sfilter: ", indent, indent1)
			debugStringExprHelper(st.filter, w, "", indent+indent1)
			if len(st.body) == 0 {
				fmt.Fprintf(w, "%s%sbody: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sbody: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.body {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case forLoopStmt:
			fmt.Fprintf(w, "%sForLoop {\n", prefix)
			fmt.Fprintf(w, "%s%starget: ", indent, indent1)
			debugStringExprHelper(st.target, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%siter: ", indent, indent1)
			debugStringExprHelper(st.iter, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%sfilter_expr: ", indent, indent1)
			if st.filterExpr.IsSome() {
				fmt.Fprint(w, "Some(\n")
				nextIndent := indent + indent1 + indent1
				debugStringExprHelper(st.filterExpr.Unwrap(), w, nextIndent, nextIndent)
				fmt.Fprintf(w, "%s%s),\n", indent, indent1)
			} else {
				fmt.Fprint(w, "None,\n")
			}
			fmt.Fprintf(w, "%s%srecursive: %v,\n", indent, indent1, st.recursive)
			if len(st.body) == 0 {
				fmt.Fprintf(w, "%s%sbody: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sbody: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.body {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			if len(st.elseBody) == 0 {
				fmt.Fprintf(w, "%s%selse_body: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%selse_body: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.elseBody {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case ifCondStmt:
			fmt.Fprintf(w, "%sIfCond {\n", prefix)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(st.expr, w, "", indent+indent1)
			if len(st.trueBody) == 0 {
				fmt.Fprintf(w, "%s%strue_body: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%strue_body: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.trueBody {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			if len(st.falseBody) == 0 {
				fmt.Fprintf(w, "%s%sfalse_body: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sfalse_body: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.falseBody {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case fromImportStmt:
			fmt.Fprintf(w, "%sFromImport {\n", prefix)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(st.expr, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%snames: [\n", indent, indent1)
			nextIndent := indent + indent1 + indent1 + indent1
			for _, name := range st.names {
				fmt.Fprintf(w, "%s%s%s(\n", indent, indent1, indent1)
				debugStringExprHelper(name.name, w, nextIndent, nextIndent)
				if name.as.IsSome() {
					fmt.Fprintf(w, "%s%s%s%sSome(\n", indent, indent1, indent1, indent1)
					debugStringExprHelper(name.as.Unwrap(), w, nextIndent+indent1, nextIndent+indent1)
					fmt.Fprintf(w, "%s%s%s%s),\n", indent, indent1, indent1, indent1)
				} else {
					fmt.Fprintf(w, "%s%s%s%sNone,\n", indent, indent1, indent1, indent1)
				}
				fmt.Fprintf(w, "%s%s%s),\n", indent, indent1, indent1)
			}
			fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case importStmt:
			fmt.Fprintf(w, "%sImport {\n", prefix)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(st.expr, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%sname: ", indent, indent1)
			debugStringExprHelper(st.name, w, "", indent+indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case includeStmt:
			fmt.Fprintf(w, "%sInclude {\n", prefix)
			fmt.Fprintf(w, "%s%sname: ", indent, indent1)
			debugStringExprHelper(st.name, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%signore_missing: %v,\n", indent, indent1, st.ignoreMissing)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case macroStmt:
			fmt.Fprintf(w, "%sMacro {\n", prefix)
			fmt.Fprintf(w, "%s%sname: %q,\n", indent, indent1, st.name)
			if len(st.args) == 0 {
				fmt.Fprintf(w, "%s%sargs: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sargs: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, arg := range st.args {
					debugStringExprHelper(arg, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			if len(st.defaults) == 0 {
				fmt.Fprintf(w, "%s%sdefaults: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sdefaults: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.defaults {
					debugStringExprHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			if len(st.body) == 0 {
				fmt.Fprintf(w, "%s%sbody: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sbody: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.body {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case setStmt:
			fmt.Fprintf(w, "%sSet {\n", prefix)
			fmt.Fprintf(w, "%s%starget: ", indent, indent1)
			debugStringExprHelper(st.target, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%sexpr: ", indent, indent1)
			debugStringExprHelper(st.expr, w, "", indent+indent1)
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case setBlockStmt:
			fmt.Fprintf(w, "%sSetBlock {\n", prefix)
			fmt.Fprintf(w, "%s%starget: ", indent, indent1)
			debugStringExprHelper(st.target, w, "", indent+indent1)
			fmt.Fprintf(w, "%s%sfilter: ", indent, indent1)
			if st.filter.IsSome() {
				fmt.Fprint(w, "Some(\n")
				nextIndent := indent + indent1 + indent1
				debugStringExprHelper(st.filter.Unwrap(), w, nextIndent, nextIndent)
				fmt.Fprintf(w, "%s%s),\n", indent, indent1)
			} else {
				fmt.Fprint(w, "None,\n")
			}
			if len(st.body) == 0 {
				fmt.Fprintf(w, "%s%sbody: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sbody: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.body {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		case withBlockStmt:
			fmt.Fprintf(w, "%sWithBlock {\n", prefix)
			if len(st.assignments) == 0 {
				fmt.Fprintf(w, "%s%sassignments: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sassignments: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1 + indent1
				for _, assign := range st.assignments {
					fmt.Fprintf(w, "%s%s%s(\n", indent, indent1, indent1)
					debugStringExprHelper(assign.lhs, w, nextIndent, nextIndent)
					debugStringExprHelper(assign.rhs, w, nextIndent, nextIndent)
					fmt.Fprintf(w, "%s%s%s),\n", indent, indent1, indent1)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			if len(st.body) == 0 {
				fmt.Fprintf(w, "%s%sbody: [],\n", indent, indent1)
			} else {
				fmt.Fprintf(w, "%s%sbody: [\n", indent, indent1)
				nextIndent := indent + indent1 + indent1
				for _, child := range st.body {
					debugStringStmtHelper(child, w, nextIndent, nextIndent)
				}
				fmt.Fprintf(w, "%s%s],\n", indent, indent1)
			}
			fmt.Fprintf(w, "%s} @ %s,\n", indent, debugStringSpan(st.span))
		default:
			panic(fmt.Sprintf("not implemented yet for statement type: %T", stmt))
		}
	}

	debugStringStmt := func(stmt statement) string {
		var b strings.Builder
		b.WriteString("Ok(\n")
		debugStringStmtHelper(stmt, &b, indent1, indent1)
		b.WriteString(")\n")
		return b.String()
	}

	debugStringErr := func(err error) string {
		var merr *Error
		if errors.As(err, &merr) {
			return fmt.Sprintf("Err(\n"+
				"    Error {\n"+
				"        kind: %s,\n"+
				"        detail: %q,\n"+
				"        name: %q,\n"+
				"        line: %d,\n"+
				"    },\n"+
				")\n",
				merr.Kind().debugString(), merr.detail, merr.name.Unwrap(),
				merr.lineno.UnwrapOr(0))
		}
		return fmt.Sprintf("Err(\n"+
			"    %s\n"+
			")\n", err)
	}

	inputFilenames := mustGlob(t, []string{"tests", "parser-inputs"}, []string{"*.txt"})
	for _, inputFilename := range inputFilenames {
		inputFileBasename := filepath.Base(inputFilename)
		t.Run(inputFileBasename, func(t *testing.T) {
			inputContent := mustReadFile(t, inputFilename)
			ast, err := parse(inputContent, inputFileBasename)
			var got string
			if err != nil {
				got = debugStringErr(err)
			} else {
				got = debugStringStmt(ast)
			}
			checkResultWithSnapshotFile(t, got, inputFilename)
		})
	}
}
