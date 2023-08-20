package compiler

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hnakamur/mjingo/internal"
	"github.com/hnakamur/mjingo/internal/datast/option"
	"github.com/hnakamur/mjingo/value"
)

type tokenStream struct {
	iter     *tokenizeIterator
	curToken token
	curSpan  *internal.Span
	curErr   error
	lastSpan internal.Span
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

func (s *tokenStream) next() (token, *internal.Span, error) {
	tkn, spn, err := s.current()
	s.curToken, s.curSpan, s.curErr = s.iter.Next()
	if spn != nil {
		s.lastSpan = *spn
	}
	return tkn, spn, err
}

func (s *tokenStream) current() (token, *internal.Span, error) {
	return s.curToken, s.curSpan, s.curErr
}

func (s *tokenStream) expandSpan(span internal.Span) internal.Span {
	return internal.Span{
		StartLine:   span.StartLine,
		StartCol:    span.StartCol,
		StartOffset: span.StartOffset,
		EndLine:     s.lastSpan.EndLine,
		EndCol:      s.lastSpan.EndCol,
		EndOffset:   s.lastSpan.EndOffset,
	}
}

func (s *tokenStream) currentSpan() internal.Span {
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

func (p *parser) parseFilterExpr(expr expression) (expression, error) {
loop:
	for {
		tkn, _, err := p.stream.current()
		if err != nil {
			return nil, err
		}
		switch tkn := tkn.(type) {
		case pipeToken:
			if _, _, err := p.stream.next(); err != nil {
				return nil, err
			}
			var tkIdent token
			var spn internal.Span
			if tkIdent, spn, err = p.expectToken(isTokenOfType[identToken], "identifier"); err != nil {
				return nil, err
			}
			name := tkIdent.(identToken).ident
			var args []expression
			if matched, err := p.matchesToken(isTokenOfType[parenOpenToken]); err != nil {
				return nil, err
			} else if matched {
				args, err = p.parseArgs()
				if err != nil {
					return nil, err
				}
			}
			expr = filterExpr{
				name: name,
				expr: option.Some(expr),
				args: args,
				span: p.stream.expandSpan(spn),
			}
		case identToken:
			if tkn.ident == "is" {
				if _, _, err := p.stream.next(); err != nil {
					return nil, err
				}
				negated := false
				if matched, err := p.skipToken(isIdentTokenWithName("not")); err != nil {
					return nil, err
				} else if matched {
					negated = true
				}
				var name string
				var spn internal.Span
				if tk, sp, err := p.expectToken(isTokenOfType[identToken], "`(`"); err != nil {
					return nil, err
				} else {
					name = tk.(identToken).ident
					spn = sp
				}
				args := []expression{}
				if matched, err := p.matchesToken(isTokenOfType[parenOpenToken]); err != nil {
					return nil, err
				} else if matched {
					if a, err := p.parseArgs(); err != nil {
						return nil, err
					} else {
						args = a
					}
				}
				expr = testExpr{
					name: name,
					expr: expr,
					args: args,
					span: p.stream.expandSpan(spn),
				}
				if negated {
					expr = unaryOpExpr{
						op:   unaryOpTypeNot,
						expr: expr,
						span: p.stream.expandSpan(spn),
					}
				}
			} else {
				break loop
			}
		default:
			break loop
		}
	}
	return expr, nil
}

func (p *parser) parseArgs() ([]expression, error) {
	args := []expression{}
	firstSpan := option.None[internal.Span]()
	kwargs := []kwarg{}

	if _, _, err := p.expectToken(isTokenOfType[parenOpenToken], "`(`"); err != nil {
		return nil, err
	}
	for {
		if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
			return nil, err
		} else if matched {
			break
		}

		if len(args) != 0 || len(kwargs) != 0 {
			if _, _, err := p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
				return nil, err
			}
			if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
				return nil, err
			} else if matched {
				break
			}
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		// keyword argument
		getVarInVarAssign := func(expr expression) (option.Option[varExpr], error) {
			if varExp, ok := expr.(varExpr); ok {
				if matched, err := p.skipToken(isTokenOfType[assignToken]); err != nil {
					return option.None[varExpr](), err
				} else if matched {
					return option.Some(varExp), nil
				}
			}
			return option.None[varExpr](), nil
		}
		if optVarExp, err := getVarInVarAssign(expr); err != nil {
			return nil, err
		} else if option.IsSome(optVarExp) {
			varExp := option.Unwrap(optVarExp)
			if option.IsSome(firstSpan) {
				firstSpan = option.Some(varExp.span)
			}
			arg, err := p.parseExprNoIf()
			if err != nil {
				return nil, err
			}
			kwargs = append(kwargs, kwarg{key: varExp.id, arg: arg})
		} else if len(kwargs) != 0 {
			return nil, syntaxError("non-keyword arg after keyword arg")
		} else {
			args = append(args, expr)
		}
	}

	if len(kwargs) != 0 {
		args = append(args, kwargsExpr{
			pairs: kwargs,
			span:  p.stream.expandSpan(option.Unwrap(firstSpan)),
		})
	}

	return args, nil
}

func (p *parser) parsePostfix(exp expression, spn internal.Span) (expression, error) {
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

			start := option.None[expression]()
			stop := option.None[expression]()
			step := option.None[expression]()
			isSlice := false

			if matched, err := p.matchesToken(isTokenOfType[colonToken]); err != nil {
				return nil, err
			} else if !matched {
				if exp, err := p.parseExpr(); err != nil {
					return nil, err
				} else {
					start = option.Some(exp)
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
						stop = option.Some(exp)
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
							step = option.Some(exp)
						}
					}
				}
			}
			if _, _, err := p.expectToken(isTokenOfType[bracketCloseToken], "`]`"); err != nil {
				return nil, err
			}

			if !isSlice {
				if option.IsNone(start) {
					return nil, syntaxError("empty subscript")
				}
				exp = getItemExpr{
					expr:          exp,
					subscriptExpr: option.Unwrap(start),
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
			return makeConst(value.FromBool(true), *spn), nil
		case "false", "False":
			return makeConst(value.FromBool(false), *spn), nil
		case "none", "None":
			return makeConst(value.None, *spn), nil
		default:
			return varExpr{id: tkn.ident, span: *spn}, nil
		}
	case stringToken:
		return makeConst(value.FromString(tkn.s), *spn), nil
	case intToken:
		return makeConst(value.FromI64(tkn.n), *spn), nil
	case floatToken:
		return makeConst(value.FromF64(tkn.f), *spn), nil
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

func (p *parser) parseListExpr(spn internal.Span) (expression, error) {
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

func (p *parser) parseMapExpr(spn internal.Span) (expression, error) {
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

func (p *parser) parseTupleOrExpression(spn internal.Span) (expression, error) {
	// MiniJinja does not really have tuples, but it treats the tuple
	// syntax the same as lists.
	if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
		return nil, err
	} else if matched {
		return listExpr{
			items: []expression{},
			span:  p.stream.expandSpan(spn),
		}, nil
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if matched, err := p.matchesToken(isTokenOfType[commaToken]); err != nil {
		return nil, err
	} else if matched {
		items := []expression{expr}
		for {
			if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
				return nil, err
			} else if matched {
				break
			}
			if _, _, err := p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
				return nil, err
			}
			if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
				return nil, err
			} else if matched {
				break
			}
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			items = append(items, expr)
		}
		expr = listExpr{
			items: items,
			span:  p.stream.expandSpan(spn),
		}

	} else if _, _, err := p.expectToken(isTokenOfType[parenCloseToken], "`)`"); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *parser) parseUnaryOnly() (expression, error) {
	return p.unaryop(p.parseUnaryOnly, p.parsePrimary,
		func(tkn token) option.Option[unaryOpType] {
			if _, ok := tkn.(minusToken); ok {
				return option.Some(unaryOpTypeNeg)
			}
			return option.None[unaryOpType]()
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
	return p.binop(p.parseUnary, func(tkn token) option.Option[binOpType] {
		if _, ok := tkn.(powToken); ok {
			return option.Some(binOpTypePow)
		}
		return option.None[binOpType]()
	})
}

func (p *parser) parseMath2() (expression, error) {
	return p.binop(p.parsePow, func(tkn token) option.Option[binOpType] {
		switch tkn.(type) {
		case mulToken:
			return option.Some(binOpTypeMul)
		case divToken:
			return option.Some(binOpTypeDiv)
		case floorDivToken:
			return option.Some(binOpTypeFloorDiv)
		case modToken:
			return option.Some(binOpTypeRem)
		default:
			return option.None[binOpType]()
		}
	})
}

func (p *parser) parseConcat() (expression, error) {
	return p.binop(p.parseMath2, func(tkn token) option.Option[binOpType] {
		if _, ok := tkn.(tildeToken); ok {
			return option.Some(binOpTypeConcat)
		}
		return option.None[binOpType]()
	})
}

func (p *parser) parseMath1() (expression, error) {
	return p.binop(p.parseConcat, func(tkn token) option.Option[binOpType] {
		switch tkn.(type) {
		case plusToken:
			return option.Some(binOpTypeAdd)
		case minusToken:
			return option.Some(binOpTypeSub)
		default:
			return option.None[binOpType]()
		}
	})
}

func (p *parser) parseCompare() (expression, error) {
	spn := p.stream.lastSpan
	exp, err := p.parseMath1()
	if err != nil {
		return nil, err
	}
loop:
	for {
		negated := false
		tkn, _, err := p.stream.current()
		if err != nil {
			return nil, err
		}
		if tkn == nil {
			break
		}
		var op binOpType
		switch tk := tkn.(type) {
		case eqToken:
			op = binOpTypeEq
		case neToken:
			op = binOpTypeNe
		case ltToken:
			op = binOpTypeLt
		case lteToken:
			op = binOpTypeLte
		case gtToken:
			op = binOpTypeGt
		case gteToken:
			op = binOpTypeGte
		case identToken:
			switch tk.ident {
			case "in":
				op = binOpTypeIn
			case "not":
				tkn2, _, err := p.stream.next()
				if err != nil {
					return nil, err
				}
				if tkn2 == nil {
					break loop
				}
				if _, _, err := p.expectToken(isIdentTokenWithName("in"), "in"); err != nil {
					return nil, err
				}
				negated = true
				op = binOpTypeIn
			default:
				break loop
			}
		default:
			break loop
		}
		if !negated {
			tkn2, _, err := p.stream.next()
			if err != nil {
				return nil, err
			}
			if tkn2 == nil {
				break loop
			}
		}
		right, err := p.parseMath1()
		if err != nil {
			return nil, err
		}
		exp = binOpExpr{
			op:    op,
			left:  exp,
			right: right,
			span:  p.stream.expandSpan(spn),
		}
		if negated {
			exp = unaryOpExpr{
				op:   unaryOpTypeNot,
				expr: exp,
				span: p.stream.expandSpan(spn),
			}
		}
		spn = p.stream.lastSpan
	}
	return exp, nil
}

func (p *parser) parseNot() (expression, error) {
	return p.unaryop(p.parseNot, p.parseCompare,
		func(tkn token) option.Option[unaryOpType] {
			if isIdentTokenWithName("not")(tkn) {
				return option.Some(unaryOpTypeNot)
			}
			return option.None[unaryOpType]()
		})
}

func (p *parser) parseAnd() (expression, error) {
	return p.binop(p.parseNot, func(tkn token) option.Option[binOpType] {
		if isIdentTokenWithName("and")(tkn) {
			return option.Some(binOpTypeScAnd)
		}
		return option.None[binOpType]()
	})
}

func (p *parser) parseOr() (expression, error) {
	return p.binop(p.parseAnd, func(tkn token) option.Option[binOpType] {
		if isIdentTokenWithName("or")(tkn) {
			return option.Some(binOpTypeScOr)
		}
		return option.None[binOpType]()
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
			exp3 := option.None[expression]()
			if matched, err := p.skipToken(isIdentTokenWithName("else")); err != nil {
				return nil, err
			} else if matched {
				ex, err := p.parseIfExpr()
				if err != nil {
					return nil, err
				}
				exp3 = option.Some(ex)
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

func (p *parser) expectToken(f func(tkn token) bool, expected string) (token, internal.Span, error) {
	tkn, spn, err := p.stream.next()
	if err != nil {
		return nil, internal.Span{}, err
	}
	if tkn == nil {
		return nil, internal.Span{}, unexpectedEOF(expected)
	}
	if f(tkn) {
		return tkn, *spn, nil
	}
	return nil, internal.Span{}, unexpected(tkn, expected)
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
	return internal.NewError(internal.SyntaxError,
		fmt.Sprintf("unexpected %v, expected %s", unexpected, expected))
}

func unexpectedEOF(expected string) error {
	return unexpected("end of input", expected)
}

func makeConst(v value.Value, spn internal.Span) expression {
	return constExpr{value: v, span: spn}
}

func syntaxError(msg string) error {
	return internal.NewError(internal.SyntaxError, msg)
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
	case "set":
		res, err := p.parseSet()
		if err != nil {
			return nil, err
		}
		switch r := res.(type) {
		case setStmtSetParseResult:
			r.stmt.span = p.stream.expandSpan(spn)
			return r.stmt, nil
		case setBlockStmtSetParseResult:
			r.stmt.span = p.stream.expandSpan(spn)
			return r.stmt, nil
		default:
			panic("unreachable")
		}
	case "autoescape":
		st, err := p.parseAutoEscape()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "filter":
		st, err := p.parseFilterBlock()
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
	var spn internal.Span
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
	filterExpr := option.None[expression]()
	if matched, err := p.skipToken(isIdentTokenWithName("if")); err != nil {
		return forLoopStmt{}, err
	} else if matched {
		if exp, err := p.parseExpr(); err != nil {
			return forLoopStmt{}, err
		} else {
			filterExpr = option.Some(exp)
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

func (p *parser) parseSet() (setParseResult, error) {
	var target expression
	inParen := false
	if matched, err := p.skipToken(isTokenOfType[parenOpenToken]); err != nil {
		return nil, err
	} else if matched {
		target, err = p.parseAssignment()
		if _, _, err := p.expectToken(isTokenOfType[parenCloseToken], "`)`"); err != nil {
			return nil, err
		}
		inParen = true
	} else {
		target, err = p.parseAssignName()
		if err != nil {
			return nil, err
		}
	}

	isSetBlock := false
	if !inParen {
		if matched, err := p.matchesToken(func(tkn token) bool {
			return isTokenOfType[blockEndToken](tkn) || isTokenOfType[pipeToken](tkn)
		}); err != nil {
			return nil, err
		} else if matched {
			isSetBlock = true
		}
	}
	if isSetBlock {
		filter := option.None[expression]()
		if matched, err := p.skipToken(isTokenOfType[parenOpenToken]); err != nil {
			return nil, err
		} else if matched {
			if exp, err := p.parseFilterChain(); err != nil {
				return nil, err
			} else {
				filter = option.Some(exp)
			}
		}
		if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
			return nil, err
		}
		body, err := p.subparse(isIdentTokenWithName("endset"))
		if err != nil {
			return nil, err
		}
		if _, _, err := p.stream.next(); err != nil {
			return nil, err
		}
		return setBlockStmtSetParseResult{stmt: setBlockStmt{
			target: target,
			filter: filter,
			body:   body,
		}}, nil
	} else {
		if _, _, err := p.expectToken(isTokenOfType[assignToken], "assignment operator"); err != nil {
			return nil, err
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return setStmtSetParseResult{stmt: setStmt{
			target: target,
			expr:   expr,
		}}, nil
	}
}

func (p *parser) parseAutoEscape() (autoEscapeStmt, error) {
	enabled, err := p.parseExpr()
	if err != nil {
		return autoEscapeStmt{}, err
	}
	if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
		return autoEscapeStmt{}, err
	}
	body, err := p.subparse(isIdentTokenWithName("endautoescape"))
	if err != nil {
		return autoEscapeStmt{}, err
	}
	if _, _, err := p.stream.next(); err != nil {
		return autoEscapeStmt{}, err
	}
	return autoEscapeStmt{
		enabled: enabled,
		body:    body,
	}, nil
}

func (p *parser) parseFilterChain() (expression, error) {
	filter := option.None[expression]()

	for {
		if matched, err := p.matchesToken(isTokenOfType[blockEndToken]); err != nil {
			return nil, err
		} else if matched {
			break
		}
		if option.IsSome(filter) {
			if _, _, err := p.expectToken(isTokenOfType[pipeToken], "`|`"); err != nil {
				return nil, err
			}
		}
		var name string
		var spn internal.Span
		if tkn, s, err := p.expectToken(isTokenOfType[identToken], "identifier"); err != nil {
			return nil, err
		} else {
			name = tkn.(identToken).ident
			spn = s
		}
		args := []expression{}
		if matched, err := p.matchesToken(isTokenOfType[parenOpenToken]); err != nil {
			return nil, err
		} else if matched {
			if a, err := p.parseArgs(); err != nil {
				return nil, err
			} else {
				args = a
			}
		}
		filter = option.Some[expression](filterExpr{
			name: name,
			expr: filter,
			args: args,
			span: p.stream.expandSpan(spn),
		})
	}
	if option.IsSome(filter) {
		return option.Unwrap(filter), nil
	}
	return nil, syntaxError("expected a filter")
}

func (p *parser) parseFilterBlock() (filterBlockStmt, error) {
	filter, err := p.parseFilterChain()
	if err != nil {
		return filterBlockStmt{}, err
	}
	if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
		return filterBlockStmt{}, err
	}
	body, err := p.subparse(isIdentTokenWithName("endfilter"))
	if err != nil {
		return filterBlockStmt{}, err
	}
	if _, _, err := p.stream.next(); err != nil {
		return filterBlockStmt{}, err
	}
	return filterBlockStmt{
		filter: filter,
		body:   body,
	}, nil
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
	return ParseWithSyntax(source, filename, DefaultSyntaxConfig)
}

func ParseWithSyntax(source, filename string, syntax SyntaxConfig) (statement, error) {
	// we want to chop off a single newline at the end.  This means that a template
	// by default does not end in a newline which is a useful property to allow
	// inline templates to work.  If someone wants a trailing newline the expectation
	// is that the user adds it themselves for achieve consistency.
	source = strings.TrimSuffix(source, "\n")
	source = strings.TrimSuffix(source, "\r")

	parser := newParser(source, false, &syntax)
	return parser.parse()
}

func (p *parser) binop(next func() (expression, error), matchFn func(tkn token) option.Option[binOpType]) (expression, error) {
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
		if option.IsNone(opType) {
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
			op:    option.Unwrap(opType),
			left:  left,
			right: right,
			span:  p.stream.expandSpan(spn),
		}
	}
	return left, nil
}

func (p *parser) unaryop(opFn, next func() (expression, error), matchFn func(tkn token) option.Option[unaryOpType]) (expression, error) {
	spn := p.stream.currentSpan()
	tkn, _, err := p.stream.current()
	if err != nil {
		return nil, err
	}
	if tkn == nil {
		return next()
	}
	opType := matchFn(tkn)
	if option.IsNone(opType) {
		return next()
	}
	if _, _, err := p.stream.next(); err != nil {
		return nil, err
	}
	exp, err := opFn()
	if err != nil {
		return nil, err
	}
	return unaryOpExpr{op: option.Unwrap(opType), expr: exp, span: p.stream.expandSpan(spn)}, nil
}

type setParseResult interface {
	typ() setParseResultType
}

type setStmtSetParseResult struct{ stmt setStmt }
type setBlockStmtSetParseResult struct{ stmt setBlockStmt }

func (setStmtSetParseResult) typ() setParseResultType      { return setParseResultTypeSetStmt }
func (setBlockStmtSetParseResult) typ() setParseResultType { return setParseResultTypeSetBlockStmt }

var _ = setParseResult(setStmtSetParseResult{})
var _ = setParseResult(setBlockStmtSetParseResult{})

type setParseResultType uint

const (
	setParseResultTypeSetStmt setParseResultType = iota + 1
	setParseResultTypeSetBlockStmt
)
