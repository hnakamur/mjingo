package mjingo

import (
	"fmt"
	"slices"
	"strings"
)

type tokenStream struct {
	iter     *tokenizeIterator
	curToken token
	curSpan  *span
	curErr   error
	lastSpan span
}

func newTokenStream(source string, inExpr bool, syntax *SyntaxConfig) *tokenStream {
	iter := newTokenizeIterator(source, inExpr, syntax)
	tkn, spn, err := iter.Next()

	return &tokenStream{
		iter:     iter,
		curToken: tkn,
		curSpan:  spn,
		curErr:   err,
	}
}

func (s *tokenStream) next() (token, *span, error) {
	tkn, spn, err := s.current()
	s.curToken, s.curSpan, s.curErr = s.iter.Next()
	if spn != nil {
		s.lastSpan = *spn
	}
	return tkn, spn, err
}

func (s *tokenStream) current() (token, *span, error) {
	return s.curToken, s.curSpan, s.curErr
}

func (s *tokenStream) expandSpan(span span) span {
	span.endLine = s.lastSpan.endLine
	span.endCol = s.lastSpan.endCol
	span.endOffset = s.lastSpan.endOffset
	return span
}

func (s *tokenStream) currentSpan() span {
	if s.curSpan != nil {
		return *s.curSpan
	}
	return s.lastSpan
}

type parser struct {
	stream  *tokenStream
	inMacro bool
	blocks  map[string]struct{}
	depth   uint
}

func newParser(source string, inExpr bool, syntax *SyntaxConfig) *parser {
	return &parser{
		stream: newTokenStream(source, inExpr, syntax),
		blocks: make(map[string]struct{}),
	}
}

func (p *parser) parseFilterExpr(exp expression) (expression, error) {
loop:
	for {
		tkn, _, err := p.stream.current()
		if err != nil {
			return nil, err
		}
		switch tkn := tkn.(type) {
		case pipeToken:
			panic("not implemented")
		case identToken:
			if tkn.ident == "is" {
				panic("not implemented")
			} else {
				break loop
			}
		default:
			break loop
		}
	}
	return exp, nil
}

func (p *parser) parsePostfix(exp expression, spn span) (expression, error) {
loop:
	for {
		nextSpan := p.stream.currentSpan()
		tkn, _, err := p.stream.current()
		if err != nil {
			return nil, err
		}
		switch tkn.(type) {
		case dotToken:
			if _, _, err := p.stream.next(); err != nil {
				return nil, err
			}
			if tkn, _, err := p.expectToken(isTokenOfType[identToken], "identifier"); err != nil {
				return nil, err
			} else {
				name := tkn.(identToken).ident
				exp = getAttrExpr{
					expr: exp,
					name: name,
					span: p.stream.expandSpan(spn),
				}
			}
		case bracketOpenToken:
			if _, _, err := p.stream.next(); err != nil {
				return nil, err
			}

			start := option[expression]{}
			stop := option[expression]{}
			step := option[expression]{}
			isSlice := false

			if matched, err := p.matchesToken(isTokenOfType[colonToken]); err != nil {
				return nil, err
			} else if !matched {
				if exp, err := p.parseExpr(); err != nil {
					return nil, err
				} else {
					start = option[expression]{valid: true, data: exp}
				}
			}
			if matched, err := p.skipToken(isTokenOfType[colonToken]); err != nil {
				return nil, err
			} else if matched {
				isSlice = true
				if matched, err := p.matchesToken(func(tkn token) bool {
					return tkn.typ() == tokenTypeBracketClose || tkn.typ() == tokenTypeColon
				}); err != nil {
					return nil, err
				} else if !matched {
					if exp, err := p.parseExpr(); err != nil {
						return nil, err
					} else {
						stop = option[expression]{valid: true, data: exp}
					}
				}
				if matched, err := p.skipToken(isTokenOfType[colonToken]); err != nil {
					return nil, err
				} else if matched {
					if matched, err := p.matchesToken(isTokenOfType[bracketCloseToken]); err != nil {
						return nil, err
					} else if !matched {
						if exp, err := p.parseExpr(); err != nil {
							return nil, err
						} else {
							step = option[expression]{valid: true, data: exp}
						}
					}
				}
			}
			if _, _, err := p.expectToken(isTokenOfType[bracketCloseToken], "`]`"); err != nil {
				return nil, err
			}

			if !isSlice {
				if !start.valid {
					return nil, syntaxError("empty subscript")
				}
				exp = getItemExpr{
					expr:          exp,
					subscriptExpr: start.data,
					span:          p.stream.expandSpan(spn),
				}
			} else {
				exp = sliceExpr{
					expr:  exp,
					start: start,
					stop:  stop,
					step:  step,
					span:  p.stream.expandSpan(spn),
				}
			}

		case parenOpenToken:
			panic("not implemented")
		default:
			break loop
		}
		spn = nextSpan
	}
	return exp, nil
}

func (p *parser) parsePrimary() (expression, error) {
	return p.withRecursionGuardExpr(p.parsePrimaryImpl)
}

func (p *parser) parsePrimaryImpl() (expression, error) {
	tkn, spn, err := p.stream.next()
	if err != nil {
		return nil, err
	}
	if tkn == nil {
		return nil, unexpectedEOF("expression")
	}
	switch tkn := tkn.(type) {
	case identToken:
		switch tkn.ident {
		case "true", "True":
			return makeConst(boolValue{b: true}, *spn), nil
		case "false", "False":
			return makeConst(boolValue{b: false}, *spn), nil
		case "none", "None":
			return makeConst(valueNone, *spn), nil
		default:
			return varExpr{id: tkn.ident, span: *spn}, nil
		}
	case stringToken:
		return makeConst(stringValue{s: tkn.s}, *spn), nil
	case intToken:
		return makeConst(i64Value{n: tkn.n}, *spn), nil
	case floatToken:
		return makeConst(f64Value{f: tkn.f}, *spn), nil
	case parenOpenToken:
		return p.parseTupleOrExpression(*spn)
	case bracketOpenToken:
		return p.parseListExpr(*spn)
	case braceOpenToken:
		return p.parseMapExpr(*spn)
	default:
		return nil, syntaxError(fmt.Sprintf("expected %v", tkn))
	}
}

func (p *parser) parseListExpr(spn span) (expression, error) {
	var items []expression
	for {
		if matched, err := p.skipToken(isTokenOfType[bracketCloseToken]); err != nil {
			return nil, err
		} else if matched {
			break
		}
		if len(items) > 0 {
			if _, _, err := p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
				return nil, err
			}
			if matched, err := p.skipToken(isTokenOfType[bracketCloseToken]); err != nil {
				return nil, err
			} else if matched {
				break
			}
		}
		if item, err := p.parseExpr(); err != nil {
			return nil, err
		} else {
			items = append(items, item)
		}
	}
	return listExpr{items: items, span: p.stream.expandSpan(spn)}, nil
}

func (p *parser) parseMapExpr(spn span) (expression, error) {
	var keys, values []expression
	for {
		if matched, err := p.skipToken(isTokenOfType[braceCloseToken]); err != nil {
			return nil, err
		} else if matched {
			break
		}
		if len(keys) > 0 {
			if _, _, err := p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
				return nil, err
			}
			if matched, err := p.skipToken(isTokenOfType[braceCloseToken]); err != nil {
				return nil, err
			} else if matched {
				break
			}
		}
		if key, err := p.parseExpr(); err != nil {
			return nil, err
		} else {
			keys = append(keys, key)
		}
		if _, _, err := p.expectToken(isTokenOfType[colonToken], "`:`"); err != nil {
			return nil, err
		}
		if value, err := p.parseExpr(); err != nil {
			return nil, err
		} else {
			values = append(values, value)
		}
	}
	return mapExpr{keys: keys, values: values, span: p.stream.expandSpan(spn)}, nil
}

func (p *parser) parseTupleOrExpression(spn span) (expression, error) {
	// MiniJinja does not really have tuples, but it treats the tuple
	// syntax the same as lists.
	panic("not implemented")
}

func (p *parser) parseUnaryOnly() (expression, error) {
	return p.unaryop(p.parseUnaryOnly, p.parsePrimary,
		func(tkn token) option[unaryOpType] {
			if _, ok := tkn.(minusToken); ok {
				return option[unaryOpType]{valid: true, data: unaryOpTypeNeg}
			}
			return option[unaryOpType]{}
		})
}

func (p *parser) parseUnary() (expression, error) {
	spn := p.stream.currentSpan()
	exp, err := p.parseUnaryOnly()
	if err != nil {
		return nil, err
	}
	exp, err = p.parsePostfix(exp, spn)
	if err != nil {
		return nil, err
	}
	return p.parseFilterExpr(exp)
}

func (p *parser) parsePow() (expression, error) {
	return p.binop(p.parseUnary, func(tkn token) option[binOpType] {
		if _, ok := tkn.(powToken); ok {
			return option[binOpType]{valid: true, data: binOpTypePow}
		}
		return option[binOpType]{}
	})
}

func (p *parser) parseMath2() (expression, error) {
	return p.binop(p.parsePow, func(tkn token) option[binOpType] {
		switch tkn.(type) {
		case mulToken:
			return option[binOpType]{valid: true, data: binOpTypeMul}
		case divToken:
			return option[binOpType]{valid: true, data: binOpTypeDiv}
		case floorDivToken:
			return option[binOpType]{valid: true, data: binOpTypeFloorDiv}
		case modToken:
			return option[binOpType]{valid: true, data: binOpTypeRem}
		default:
			return option[binOpType]{}
		}
	})
}

func (p *parser) parseConcat() (expression, error) {
	return p.binop(p.parseMath2, func(tkn token) option[binOpType] {
		if _, ok := tkn.(tildeToken); ok {
			return option[binOpType]{valid: true, data: binOpTypeConcat}
		}
		return option[binOpType]{}
	})
}

func (p *parser) parseMath1() (expression, error) {
	return p.binop(p.parseConcat, func(tkn token) option[binOpType] {
		switch tkn.(type) {
		case plusToken:
			return option[binOpType]{valid: true, data: binOpTypeAdd}
		case minusToken:
			return option[binOpType]{valid: true, data: binOpTypeSub}
		default:
			return option[binOpType]{}
		}
	})
}

func (p *parser) parseCompare() (expression, error) {
	return p.parseMath1()
}

func (p *parser) parseNot() (expression, error) {
	return p.unaryop(p.parseNot, p.parseCompare,
		func(tkn token) option[unaryOpType] {
			if isIdentTokenWithName("not")(tkn) {
				return option[unaryOpType]{valid: true, data: unaryOpTypeNot}
			}
			return option[unaryOpType]{}
		})
}

func (p *parser) parseAnd() (expression, error) {
	return p.binop(p.parseNot, func(tkn token) option[binOpType] {
		if isIdentTokenWithName("and")(tkn) {
			return option[binOpType]{valid: true, data: binOpTypeScAnd}
		}
		return option[binOpType]{}
	})
}

func (p *parser) parseOr() (expression, error) {
	return p.binop(p.parseAnd, func(tkn token) option[binOpType] {
		if isIdentTokenWithName("or")(tkn) {
			return option[binOpType]{valid: true, data: binOpTypeScOr}
		}
		return option[binOpType]{}
	})
}

func (p *parser) parseIfExpr() (expression, error) {
	spn := p.stream.lastSpan
	exp, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	for {
		if matched, err := p.skipToken(isIdentTokenWithName("if")); err != nil {
			return nil, err
		} else if matched {
			exp2, err := p.parseOr()
			if err != nil {
				return nil, err
			}
			exp3 := option[expression]{}
			if matched, err := p.skipToken(isIdentTokenWithName("else")); err != nil {
				return nil, err
			} else if matched {
				ex, err := p.parseIfExpr()
				if err != nil {
					return nil, err
				}
				exp3 = option[expression]{valid: true, data: ex}
			}
			exp = ifExpr{
				testExpr:  exp2,
				trueExpr:  exp,
				falseExpr: exp3,
				span:      p.stream.expandSpan(spn),
			}
			spn = p.stream.lastSpan
		} else {
			break
		}
	}
	return exp, nil
}

func (p *parser) parseExpr() (expression, error) {
	return p.withRecursionGuardExpr(p.parseIfExpr)
}

func (p *parser) parseExprNoIf() (expression, error) {
	return p.parseOr()
}

func (p *parser) expectToken(f func(tkn token) bool, expected string) (token, span, error) {
	tkn, spn, err := p.stream.next()
	if err != nil {
		return nil, span{}, err
	}
	if tkn == nil {
		return nil, span{}, unexpectedEOF(expected)
	}
	if f(tkn) {
		return tkn, *spn, nil
	}
	return nil, span{}, unexpected(tkn, expected)
}

func (p *parser) matchesToken(f func(tkn token) bool) (bool, error) {
	tkn, _, err := p.stream.current()
	if err != nil {
		return false, err
	}
	if tkn == nil {
		return false, nil
	}
	return f(tkn), nil
}

func (p *parser) skipToken(f func(token) bool) (matched bool, err error) {
	if err = p.stream.curErr; err != nil {
		return false, err
	}
	if f(p.stream.curToken) {
		p.stream.next()
		return true, nil
	}
	return false, nil
}

const parseMaxRecursion = 150

func (p *parser) withRecursionGuardExpr(f func() (expression, error)) (expression, error) {
	p.depth++
	if p.depth > parseMaxRecursion {
		return nil, syntaxError("template exceeds maximum recursion limits")
	}
	defer func() { p.depth-- }()
	return f()
}

func (p *parser) withRecursionGuardStmt(f func() (statement, error)) (statement, error) {
	p.depth++
	if p.depth > parseMaxRecursion {
		return nil, syntaxError("template exceeds maximum recursion limits")
	}
	defer func() { p.depth-- }()
	return f()
}

func unexpected(unexpected any, expected string) error {
	return &Error{
		typ: SyntaxError,
		detail: option[string]{
			valid: true,
			data:  fmt.Sprintf("unexpected %v, expected %s", unexpected, expected),
		},
	}
}

func unexpectedEOF(expected string) error {
	return unexpected("end of input", expected)
}

func makeConst(v value, spn span) expression {
	return constExpr{value: v, span: spn}
}

func syntaxError(msg string) error {
	return &Error{
		typ:    SyntaxError,
		detail: option[string]{valid: true, data: msg},
	}
}

func (p *parser) subparse(endCheck func(token) bool) ([]statement, error) {
	var rv []statement
	for {
		tkn, spn, err := p.stream.next()
		if err != nil {
			return nil, err
		}
		if tkn == nil {
			break
		}
		switch tk := tkn.(type) {
		case templateDataToken:
			raw := tk.s
			rv = append(rv, emitRawStmt{raw: raw, span: *spn})
		case variableStartToken:
			exp, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			rv = append(rv, emitExprStmt{expr: exp, span: p.stream.expandSpan(*spn)})
			if _, _, err := p.expectToken(isTokenOfType[variableEndToken], "end of variable block"); err != nil {
				return nil, err
			}
		case blockStartToken:
			tkn, _, err := p.stream.current()
			if err != nil {
				return nil, err
			}
			if tkn == nil {
				return nil, syntaxError("unexpected end of input, expected keyword")
			}
			if endCheck(tkn) {
				return rv, nil
			}
			if st, err := p.parseStmt(); err != nil {
				return nil, err
			} else {
				rv = append(rv, st)
			}
			if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
				return nil, err
			}
		default:
			panic("lexer produced garbage")
		}
	}
	return rv, nil
}

func (p *parser) parseStmt() (statement, error) {
	return p.withRecursionGuardStmt(p.parseStmtUnprotected)
}

func (p *parser) parseStmtUnprotected() (statement, error) {
	tkn, spn, err := p.expectToken(func(token) bool { return true }, "block keyword")
	if err != nil {
		return nil, err
	}

	var ident string
	if identTkn, ok := tkn.(identToken); ok {
		ident = identTkn.ident
	} else {
		return nil, syntaxError(fmt.Sprintf("unknown %s, expected statement", tkn))
	}

	switch ident {
	case "for":
		st, err := p.parseForStmt()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "if":
		st, err := p.parseIfCond()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "with":
		st, err := p.parseWithBlock()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	default:
		return nil, syntaxError(fmt.Sprintf("unknown statement %s", ident))
	}
}

var reservedNames = []string{"true", "True", "false", "False", "none", "None", "loop", "self"}

func (p *parser) parseAssignName() (expression, error) {
	var id string
	var spn span
	if tkn, sp, err := p.expectToken(isTokenOfType[identToken], "identifier"); err != nil {
		return nil, err
	} else {
		id = tkn.(identToken).ident
		spn = sp
	}

	if slices.Contains(reservedNames, id) {
		return nil, syntaxError(fmt.Sprintf("cannot assign to reserved variable name %s", id))
	}
	return varExpr{id: id, span: spn}, nil
}

func (p *parser) parseAssignment() (expression, error) {
	spn := p.stream.currentSpan()
	var items []expression
	isTuple := false

	for {
		if len(items) > 0 {
			if _, _, err := p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
				return nil, err
			}
		}
		if matched, err := p.matchesToken(func(tkn token) bool {
			return isTokenOfType[parenCloseToken](tkn) ||
				isTokenOfType[variableEndToken](tkn) ||
				isTokenOfType[blockEndToken](tkn) ||
				isIdentTokenWithName("in")(tkn)
		}); err != nil {
			return nil, err
		} else if matched {
			break
		}

		var item expression
		if matched, err := p.skipToken(isTokenOfType[parenOpenToken]); err != nil {
			return nil, err
		} else if matched {
			if rv, err := p.parseAssignment(); err != nil {
				return nil, err
			} else {
				if _, _, err := p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
					return nil, err
				}
				item = rv
			}
		} else {
			if exp, err := p.parseAssignName(); err != nil {
				return nil, err
			} else {
				item = exp
			}
		}
		items = append(items, item)

		if matched, err := p.matchesToken(isTokenOfType[commaToken]); err != nil {
			return nil, err
		} else if matched {
			isTuple = true
		} else {
			break
		}
	}

	if !isTuple && len(items) == 1 {
		return items[0], nil
	}
	return listExpr{items: items, span: p.stream.expandSpan(spn)}, nil
}

func (p *parser) parseForStmt() (forLoopStmt, error) {
	target, err := p.parseAssignment()
	if err != nil {
		return forLoopStmt{}, err
	}
	if _, _, err := p.expectToken(isIdentTokenWithName("in"), "in"); err != nil {
		return forLoopStmt{}, err
	}
	iter, err := p.parseExprNoIf()
	if err != nil {
		return forLoopStmt{}, err
	}
	filterExpr := option[expression]{}
	if matched, err := p.skipToken(isIdentTokenWithName("if")); err != nil {
		return forLoopStmt{}, err
	} else if matched {
		if exp, err := p.parseExpr(); err != nil {
			return forLoopStmt{}, err
		} else {
			filterExpr = option[expression]{valid: true, data: exp}
		}
	}
	recursive := false
	if matched, err := p.skipToken(isIdentTokenWithName("recursive")); err != nil {
		return forLoopStmt{}, err
	} else if matched {
		recursive = true
	}
	if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
		return forLoopStmt{}, err
	}
	body, err := p.subparse(func(tkn token) bool {
		return isIdentTokenWithName("endfor")(tkn) || isIdentTokenWithName("else")(tkn)
	})
	if err != nil {
		return forLoopStmt{}, err
	}
	elseBody := []statement{}
	if matched, err := p.skipToken(isIdentTokenWithName("else")); err != nil {
		return forLoopStmt{}, err
	} else if matched {
		if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
			return forLoopStmt{}, err
		}
		elseBody, err = p.subparse(isIdentTokenWithName("endfor"))
		if err != nil {
			return forLoopStmt{}, err
		}
	}
	if _, _, err := p.stream.next(); err != nil {
		return forLoopStmt{}, err
	}
	return forLoopStmt{
		target:     target,
		iter:       iter,
		filterExpr: filterExpr,
		recursive:  recursive,
		body:       body,
		elseBody:   elseBody,
	}, nil
}

func (p *parser) parseIfCond() (ifCondStmt, error) {
	exp, err := p.parseExprNoIf()
	if err != nil {
		return ifCondStmt{}, err
	}
	if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
		return ifCondStmt{}, err
	}

	trueBody, err := p.subparse(isIdentTokenWithName("endif", "else", "elif"))
	if err != nil {
		return ifCondStmt{}, err
	}

	var falseBody []statement
	tkn, spn, err := p.stream.next()
	if err != nil {
		return ifCondStmt{}, err
	}
	switch {
	case isIdentTokenWithName("else")(tkn):
		if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
			return ifCondStmt{}, err
		}
		falseBody, err = p.subparse(isIdentTokenWithName("endif"))
		if _, _, err := p.stream.next(); err != nil {
			return ifCondStmt{}, err
		}
	case isIdentTokenWithName("elif")(tkn):
		st, err := p.parseIfCond()
		if err != nil {
			return ifCondStmt{}, err
		}
		st.span = p.stream.expandSpan(*spn)
		falseBody = []statement{st}
	}

	return ifCondStmt{expr: exp, trueBody: trueBody, falseBody: falseBody}, nil
}

func (p *parser) parseWithBlock() (withBlockStmt, error) {
	var assignments []assignment

	for {
		if matched, err := p.matchesToken(isTokenOfType[blockEndToken]); err != nil {
			return withBlockStmt{}, err
		} else if matched {
			break
		}
		var target expression
		if matched, err := p.skipToken(isTokenOfType[parenOpenToken]); err != nil {
			return withBlockStmt{}, err
		} else if matched {
			target, err = p.parseAssignment()
			if err != nil {
				return withBlockStmt{}, err
			}
			if _, _, err := p.expectToken(isTokenOfType[parenCloseToken], "`)`"); err != nil {
				return withBlockStmt{}, err
			}
		} else {
			target, err = p.parseAssignName()
			if err != nil {
				return withBlockStmt{}, err
			}
		}
		if _, _, err := p.expectToken(isTokenOfType[assignToken], "assignment operator"); err != nil {
			return withBlockStmt{}, err
		}
		if exp, err := p.parseExpr(); err != nil {
			return withBlockStmt{}, err
		} else {
			assignments = append(assignments, assignment{lhs: target, rhs: exp})
		}
	}

	if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
		return withBlockStmt{}, err
	}
	body, err := p.subparse(isIdentTokenWithName("endwith"))
	if err != nil {
		return withBlockStmt{}, err
	}
	if _, _, err := p.stream.next(); err != nil {
		return withBlockStmt{}, err
	}
	return withBlockStmt{assignments: assignments, body: body}, nil
}

func (p *parser) parse() (statement, error) {
	spn := p.stream.lastSpan
	ss, err := p.subparse(func(token) bool { return false })
	if err != nil {
		return nil, err
	}
	return templateStmt{children: ss, span: p.stream.expandSpan(spn)}, nil
}

func parse(source, filename string) (statement, error) {
	return parseWithSyntax(source, filename, DefaultSyntaxConfig)
}

func parseWithSyntax(source, filename string, syntax SyntaxConfig) (statement, error) {
	// we want to chop off a single newline at the end.  This means that a template
	// by default does not end in a newline which is a useful property to allow
	// inline templates to work.  If someone wants a trailing newline the expectation
	// is that the user adds it themselves for achieve consistency.
	source = strings.TrimSuffix(source, "\n")
	source = strings.TrimSuffix(source, "\r")

	parser := newParser(source, false, &syntax)
	return parser.parse()
}

func (p *parser) binop(next func() (expression, error), matchFn func(tkn token) option[binOpType]) (expression, error) {
	spn := p.stream.currentSpan()
	left, err := next()
	if err != nil {
		return nil, err
	}
	for {
		tkn, _, err := p.stream.current()
		if err != nil {
			return nil, err
		}
		if tkn == nil {
			break
		}
		opType := matchFn(tkn)
		if !opType.valid {
			break
		}
		if _, _, err := p.stream.next(); err != nil {
			return nil, err
		}
		right, err := next()
		if err != nil {
			return nil, err
		}
		left = binOpExpr{
			op:    opType.data,
			left:  left,
			right: right,
			span:  p.stream.expandSpan(spn),
		}
	}
	return left, nil
}

func (p *parser) unaryop(opFn, next func() (expression, error), matchFn func(tkn token) option[unaryOpType]) (expression, error) {
	spn := p.stream.currentSpan()
	tkn, _, err := p.stream.current()
	if err != nil {
		return nil, err
	}
	if tkn == nil {
		return next()
	}
	opType := matchFn(tkn)
	if !opType.valid {
		return next()
	}
	if _, _, err := p.stream.next(); err != nil {
		return nil, err
	}
	exp, err := opFn()
	if err != nil {
		return nil, err
	}
	return unaryOpExpr{op: opType.data, expr: exp, span: p.stream.expandSpan(spn)}, nil
}
