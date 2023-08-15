package mjingo

import (
	"fmt"
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
			ident := tkn.s
			if ident == "is" {
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
			if tkn, _, err := p.expectToken(isTokenOfType(tokenTypeIdent), "identifier"); err != nil {
				return nil, err
			} else {
				name := tkn.(identToken).s
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

			if matched, err := p.matchesToken(isTokenOfType(tokenTypeColon)); err != nil {
				return nil, err
			} else if !matched {
				if exp, err := p.parseExpr(); err != nil {
					return nil, err
				} else {
					start = option[expression]{valid: true, data: exp}
				}
			}
			if matched, err := p.skipToken(tokenTypeColon); err != nil {
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
				if matched, err := p.skipToken(tokenTypeColon); err != nil {
					return nil, err
				} else if matched {
					if matched, err := p.matchesToken(isTokenOfType(tokenTypeBracketClose)); err != nil {
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
			if _, _, err := p.expectToken(isTokenOfType(tokenTypeBracketClose), "`]`"); err != nil {
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
	return withRecursionGuard(p, p.parsePrimaryImpl)
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
		ident := tkn.s
		switch ident {
		case "true", "True":
			return makeConst(boolValue{b: true}, *spn), nil
		case "false", "False":
			return makeConst(boolValue{b: false}, *spn), nil
		case "none", "None":
			return makeConst(valueNone, *spn), nil
		default:
			return varExpr{id: ident, span: *spn}, nil
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
		if matched, err := p.skipToken(tokenTypeBracketClose); err != nil {
			return nil, err
		} else if matched {
			break
		}
		if len(items) > 0 {
			if _, _, err := p.expectToken(isTokenOfType(tokenTypeComma), "`,`"); err != nil {
				return nil, err
			}
			if matched, err := p.skipToken(tokenTypeBracketClose); err != nil {
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
		if matched, err := p.skipToken(tokenTypeBraceClose); err != nil {
			return nil, err
		} else if matched {
			break
		}
		if len(keys) > 0 {
			if _, _, err := p.expectToken(isTokenOfType(tokenTypeComma), "`,`"); err != nil {
				return nil, err
			}
			if matched, err := p.skipToken(tokenTypeBraceClose); err != nil {
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
		if _, _, err := p.expectToken(isTokenOfType(tokenTypeColon), "`:`"); err != nil {
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
	return p.parseUnary()
}

func (p *parser) parseMath2() (expression, error) {
	return p.parsePow()
}

func (p *parser) parseConcat() (expression, error) {
	return p.parseMath2()
}

func (p *parser) parseMath1() (expression, error) {
	return p.binop(p.parseConcat, func(tkn token) option[binOpKind] {
		switch tkn.(type) {
		case plusToken:
			return option[binOpKind]{valid: true, data: binOpKindAdd}
		case minusToken:
			return option[binOpKind]{valid: true, data: binOpKindSub}
		default:
			return option[binOpKind]{}
		}
	})
}

func (p *parser) parseCompare() (expression, error) {
	return p.parseMath1()
}

func (p *parser) parseNot() (expression, error) {
	return p.parseCompare()
}

func (p *parser) parseAnd() (expression, error) {
	return p.parseNot()
}

func (p *parser) parseOr() (expression, error) {
	return p.parseAnd()
}

func (p *parser) parseIfExpr() (expression, error) {
	exp, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	// not implemented
	return exp, nil
}

func (p *parser) parseExpr() (expression, error) {
	return withRecursionGuard(p, p.parseIfExpr)
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

func (p *parser) skipToken(k tokenType) (matched bool, err error) {
	if err = p.stream.curErr; err != nil {
		return false, err
	}
	if p.stream.curToken.typ() == k {
		p.stream.next()
		return true, nil
	}
	return false, nil
}

const parseMaxRecursion = 150

func withRecursionGuard(p *parser, f func() (expression, error)) (expression, error) {
	p.depth++
	if p.depth > parseMaxRecursion {
		return nil, syntaxError("template exceeds maximum recursion limits")
	}
	defer func() { p.depth-- }()
	return f()
}

func unexpected(unexpected any, expected string) error {
	return &Error{
		kind: SyntaxError,
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
		kind:   SyntaxError,
		detail: option[string]{valid: true, data: msg},
	}
}

func (p *parser) subparse(endCheck func(*token) bool) ([]statement, error) {
	var rv []statement
	for {
		tkn, spn, err := p.stream.next()
		if err != nil {
			return nil, err
		}
		if tkn == nil {
			break
		}
		switch tkn := tkn.(type) {
		case templateDataToken:
			raw := tkn.s
			rv = append(rv, emitRawStmt{raw: raw, span: *spn})
		case variableStartToken:
			exp, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			rv = append(rv, emitExprStmt{expr: exp, span: p.stream.expandSpan(*spn)})
			if _, _, err := p.expectToken(isTokenOfType(tokenTypeVariableEnd), "end of variable block"); err != nil {
				return nil, err
			}
		case blockStartToken:
			panic("not implemented")
		default:
			panic("lexer produced garbage")
		}
	}
	return rv, nil
}

func (p *parser) parse() (statement, error) {
	spn := p.stream.lastSpan
	ss, err := p.subparse(func(*token) bool { return false })
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

func isTokenOfType(k tokenType) func(tkn token) bool {
	return func(tkn token) bool {
		return tkn.typ() == k
	}
}

func (p *parser) binop(next func() (expression, error), matchFn func(tkn token) option[binOpKind]) (expression, error) {
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
		opKind := matchFn(tkn)
		if !opKind.valid {
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
			op:    opKind.data,
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
	opKind := matchFn(tkn)
	if !opKind.valid {
		return next()
	}
	if _, _, err := p.stream.next(); err != nil {
		return nil, err
	}
	exp, err := opFn()
	if err != nil {
		return nil, err
	}
	return unaryOpExpr{op: opKind.data, expr: exp, span: p.stream.expandSpan(spn)}, nil
}
