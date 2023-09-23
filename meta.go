package mjingo

import (
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/internal/datast/slicex"
	"github.com/hnakamur/mjingo/internal/datast/stack"
	"github.com/hnakamur/mjingo/option"
)

type assignmentTracker struct {
	out       *hashset.StrHashSet
	nestedOut option.Option[*hashset.StrHashSet]
	assigned  stack.Stack[*hashset.StrHashSet]
}

func (t *assignmentTracker) isAssigned(name string) bool {
	return slicex.Any(t.assigned, func(s *hashset.StrHashSet) bool {
		return s.Contains(name)
	})
}

func (t *assignmentTracker) assign(name string) {
	t.assigned[len(t.assigned)-1].Add(name)
}

func (t *assignmentTracker) assignNested(name string) {
	if t.nestedOut.IsSome() {
		s := t.nestedOut.Unwrap()
		s.Add(name)
	}
}

func (t *assignmentTracker) push() {
	t.assigned.Push(hashset.NewStrHashSet())
}

func (t *assignmentTracker) pop() {
	t.assigned.Pop()
}

func findMacroClosure(m macroStmt) *hashset.StrHashSet {
	state := assignmentTracker{
		out:       hashset.NewStrHashSet(),
		nestedOut: option.None[*hashset.StrHashSet](),
		assigned:  []*hashset.StrHashSet{hashset.NewStrHashSet()},
	}
	for _, arg := range m.args {
		trackAssign(arg, &state)
	}
	for _, node := range m.body {
		trackWalk(node, &state)
	}
	return state.out
}

func trackAssign(expr astExpr, state *assignmentTracker) {
	switch exp := expr.(type) {
	case varExpr:
		state.assign(exp.id)
	case listExpr:
		for _, x := range exp.items {
			trackAssign(x, state)
		}
	}
}

func trackVisitExprOpt(expr option.Option[astExpr], state *assignmentTracker) {
	if expr.IsSome() {
		trackVisitExpr(expr.Unwrap(), state)
	}
}

func trackVisitExpr(expr astExpr, state *assignmentTracker) {
	switch exp := expr.(type) {
	case varExpr:
		if !state.isAssigned(exp.id) {
			state.out.Add(exp.id)
			// if we are not tracking nested assignments, we can consider a variable
			// to be assigned the first time we perform a lookup.
			if state.nestedOut.IsNone() {
				state.assign(exp.id)
			} else {
				state.assignNested(exp.id)
			}
		}
	case constExpr: // do nothing
	case unaryOpExpr:
		trackVisitExpr(exp.expr, state)
	case binOpExpr:
		trackVisitExpr(exp.left, state)
		trackVisitExpr(exp.right, state)
	case ifExpr:
		trackVisitExpr(exp.testExpr, state)
		trackVisitExpr(exp.trueExpr, state)
		trackVisitExprOpt(exp.falseExpr, state)
	case filterExpr:
		trackVisitExprOpt(exp.expr, state)
		trackVisitExpressions(exp.args, state)
	case testExpr:
		trackVisitExpr(exp.expr, state)
		trackVisitExpressions(exp.args, state)
	case getAttrExpr:
		// if we are tracking nested, we check if we have a chain of attribute
		// lookups that terminate in a variable lookup.  In that case we can
		// assign the nested lookup.
		if state.nestedOut.IsSome() {
			attrs := []string{exp.name}
			ptr := &exp.expr
		loop:
			for {
				switch exp2 := (*ptr).(type) {
				case varExpr:
					if !state.isAssigned(exp2.id) {
						var b strings.Builder
						b.WriteString(exp2.id)
						for i := len(attrs) - 1; i >= 0; i-- {
							b.WriteRune('.')
							b.WriteString(attrs[i])
						}
						state.assignNested(b.String())
						return
					}
				case getAttrExpr:
					attrs = append(attrs, exp2.name)
					ptr = &exp2.expr
					continue
				default:
					break loop
				}
			}
		}
		trackVisitExpr(exp.expr, state)
	case getItemExpr:
		trackVisitExpr(exp.expr, state)
		trackVisitExpr(exp.subscriptExpr, state)
	case sliceExpr:
		trackVisitExprOpt(exp.start, state)
		trackVisitExprOpt(exp.stop, state)
		trackVisitExprOpt(exp.step, state)
	case callExpr:
		trackVisitExpr(exp.call.data.expr, state)
		trackVisitExpressions(exp.call.data.args, state)
	case listExpr:
		trackVisitExpressions(exp.items, state)
	case mapExpr:
		for i, key := range exp.keys {
			val := exp.values[i]
			trackVisitExpr(key, state)
			trackVisitExpr(val, state)
		}
	case kwargsExpr:
		for _, pair := range exp.pairs {
			trackVisitExpr(pair.arg, state)
		}
	}
}

func trackVisitExpressions(expressions []astExpr, state *assignmentTracker) {
	for _, expr := range expressions {
		trackVisitExpr(expr, state)
	}
}

func trackWalkStatements(statements []statement, state *assignmentTracker) {
	for _, node := range statements {
		trackWalk(node, state)
	}
}

func trackWalk(node statement, state *assignmentTracker) {
	switch st := node.(type) {
	case templateStmt:
		state.assign("self")
		trackWalkStatements(st.children, state)
	case emitExprStmt:
		trackVisitExpr(st.expr, state)
	case emitRawStmt: // do nothing
	case forLoopStmt:
		state.push()
		state.assign("loop")
		trackVisitExpr(st.iter, state)
		trackAssign(st.target, state)
		trackVisitExprOpt(st.filterExpr, state)
		trackWalkStatements(st.body, state)
		state.pop()
		state.push()
		trackWalkStatements(st.elseBody, state)
		state.pop()
	case ifCondStmt:
		trackVisitExpr(st.expr, state)
		state.push()
		trackWalkStatements(st.trueBody, state)
		state.pop()
		state.push()
		trackWalkStatements(st.falseBody, state)
		state.pop()
	case withBlockStmt:
		state.push()
		for _, assign := range st.assignments {
			trackAssign(assign.lhs, state)
			trackVisitExpr(assign.rhs, state)
		}
		trackWalkStatements(st.body, state)
		state.pop()
	case setStmt:
		trackAssign(st.target, state)
		trackVisitExpr(st.expr, state)
	case autoEscapeStmt:
		state.push()
		trackWalkStatements(st.body, state)
		state.pop()
	case filterBlockStmt:
		state.push()
		trackWalkStatements(st.body, state)
		state.pop()
	case setBlockStmt:
		trackAssign(st.target, state)
		state.push()
		trackWalkStatements(st.body, state)
		state.pop()
	case blockStmt:
		state.push()
		state.assign("super")
		trackWalkStatements(st.body, state)
		state.pop()
	case extendsStmt, includeStmt: // do nothing
	case importStmt:
		trackAssign(st.name, state)
	case fromImportStmt:
		for _, name := range st.names {
			trackAssign(name.as.UnwrapOr(name.name), state)
		}
	case macroStmt:
		state.assign(st.name)
	case callBlockStmt: // do nothing
	case doStmt:
		trackVisitExpr(st.call.data.expr, state)
		trackVisitExpressions(st.call.data.args, state)
	}
}
