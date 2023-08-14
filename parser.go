package mjingo

import (
	"fmt"
	"strings"
)

type tokenStream struct {
	iter     *tokenizeIterator
	curToken *token
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

func (s *tokenStream) next() (*token, *span, error) {
	tkn, spn, err := s.current()
	s.curToken, s.curSpan, s.curErr = s.iter.Next()
	if spn != nil {
		s.lastSpan = *spn
	}
	return tkn, spn, err
}

func (s *tokenStream) current() (*token, *span, error) {
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

func (p *parser) parseFilterExpr(exp expr) (*expr, error) {
loop:
	for {
		tkn, _, err := p.stream.current()
		if err != nil {
			return nil, err
		}
		switch tkn.kind {
		case tokenKindPipe:
			panic("not implemented")
		case tokenKindIdent:
			ident := tkn.data.(identTokenData)
			if ident == "is" {
				panic("not implemented")
			} else {
				break loop
			}
		default:
			break loop
		}
	}
	return &exp, nil
}

func (p *parser) parsePostfix(exp expr, spn span) (*expr, error) {
loop:
	for {
		nextSpan := p.stream.currentSpan()
		tkn, _, err := p.stream.current()
		if err != nil {
			return nil, err
		}
		switch tkn.kind {
		case tokenKindDot:
			if _, _, err := p.stream.next(); err != nil {
				return nil, err
			}
			if tkn2, _, err := p.expectToken(isTokenOfKind(tokenKindIdent), "identifier"); err != nil {
				return nil, err
			} else {
				name := tkn2.data.(identTokenData)
				exp = expr{
					kind: exprKindGetAttr,
					data: getAttrExprData{expr: exp, name: name},
					span: p.stream.expandSpan(spn),
				}
			}
		case tokenKindBracketOpen:
			if _, _, err := p.stream.next(); err != nil {
				return nil, err
			}

			start := option[expr]{}
			stop := option[expr]{}
			step := option[expr]{}
			isSlice := false

			if matched, err := p.matchesToken(isTokenOfKind(tokenKindColon)); err != nil {
				return nil, err
			} else if !matched {
				if exp, err := p.parseExpr(); err != nil {
					return nil, err
				} else {
					start = option[expr]{valid: true, data: *exp}
				}
			}
			if matched, err := p.skipToken(tokenKindColon); err != nil {
				return nil, err
			} else if matched {
				isSlice = true
				if matched, err := p.matchesToken(func(tkn token) bool {
					return tkn.kind == tokenKindBracketClose || tkn.kind == tokenKindColon
				}); err != nil {
					return nil, err
				} else if !matched {
					if exp, err := p.parseExpr(); err != nil {
						return nil, err
					} else {
						stop = option[expr]{valid: true, data: *exp}
					}
				}
				if matched, err := p.skipToken(tokenKindColon); err != nil {
					return nil, err
				} else if matched {
					if matched, err := p.matchesToken(isTokenOfKind(tokenKindBracketClose)); err != nil {
						return nil, err
					} else if !matched {
						if exp, err := p.parseExpr(); err != nil {
							return nil, err
						} else {
							step = option[expr]{valid: true, data: *exp}
						}
					}
				}
			}
			if _, _, err := p.expectToken(isTokenOfKind(tokenKindBracketClose), "`]`"); err != nil {
				return nil, err
			}

			if !isSlice {
				if !start.valid {
					return nil, syntaxError("empty subscript")
				}
				exp = expr{
					kind: exprKindGetItem,
					data: getItemExprData{
						expr:          exp,
						subscriptExpr: start.data,
					},
					span: p.stream.expandSpan(spn),
				}
			} else {
				exp = expr{
					kind: exprKindSlice,
					data: sliceExprData{
						expr:  exp,
						start: start,
						stop:  stop,
						step:  step,
					},
					span: p.stream.expandSpan(spn),
				}
			}

		case tokenKindParenOpen:
			panic("not implemented")
		default:
			break loop
		}
		spn = nextSpan
	}
	return &exp, nil
}

func (p *parser) parsePrimary() (*expr, error) {
	return withRecursionGuard[expr](p, p.parsePrimaryImpl)
}

func (p *parser) parsePrimaryImpl() (*expr, error) {
	tkn, spn, err := p.stream.next()
	if err != nil {
		return nil, err
	}
	if tkn == nil {
		return nil, unexpectedEOF("expression")
	}
	switch tkn.kind {
	case tokenKindIdent:
		data := tkn.data.(identTokenData)
		switch data {
		case "true", "True":
			return makeConst(value{kind: valueKindBool, data: true}, *spn), nil
		case "false", "False":
			return makeConst(value{kind: valueKindBool, data: false}, *spn), nil
		case "none", "None":
			return makeConst(value{kind: valueKindNone}, *spn), nil
		default:
			return &expr{kind: exprKindVar, data: varExprData{id: data}, span: *spn}, nil
		}
	case tokenKindString:
		data := tkn.data.(stringTokenData)
		return makeConst(value{kind: valueKindString, data: data}, *spn), nil
	case tokenKindInt:
		data := tkn.data.(intTokenData)
		return makeConst(value{kind: valueKindI64, data: data}, *spn), nil
	case tokenKindFloat:
		data := tkn.data.(floatTokenData)
		return makeConst(value{kind: valueKindF64, data: data}, *spn), nil
	case tokenKindParenOpen:
		return p.parseTupleOrExpression(*spn)
	case tokenKindBracketOpen:
		return p.parseListExpr(*spn)
	case tokenKindBraceOpen:
		return p.parseMapExpr(*spn)
	default:
		return nil, syntaxError(fmt.Sprintf("expected %v", *tkn))
	}
}

func (p *parser) parseListExpr(spn span) (*expr, error) {
	var items []expr
	for {
		if matched, err := p.skipToken(tokenKindBracketClose); err != nil {
			return nil, err
		} else if matched {
			break
		}
		if len(items) > 0 {
			if _, _, err := p.expectToken(isTokenOfKind(tokenKindComma), "`,`"); err != nil {
				return nil, err
			}
			if matched, err := p.skipToken(tokenKindBracketClose); err != nil {
				return nil, err
			} else if matched {
				break
			}
		}
		if item, err := p.parseExpr(); err != nil {
			return nil, err
		} else {
			items = append(items, *item)
		}
	}
	return &expr{
		kind: exprKindList,
		data: listExprData{items: items},
		span: p.stream.expandSpan(spn),
	}, nil
}

func (p *parser) parseMapExpr(spn span) (*expr, error) {
	panic("not implemented")
}

func (p *parser) parseTupleOrExpression(spn span) (*expr, error) {
	// MiniJinja does not really have tuples, but it treats the tuple
	// syntax the same as lists.
	panic("not implemented")
}

func (p *parser) parseUnaryOnly() (*expr, error) {
	spn := p.stream.currentSpan()
	exp, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	exp, err = p.parsePostfix(*exp, spn)
	if err != nil {
		return nil, err
	}
	return p.parseFilterExpr(*exp)
}

func (p *parser) parseUnary() (*expr, error) {
	return p.parseUnaryOnly()
}

func (p *parser) parsePow() (*expr, error) {
	return p.parseUnary()
}

func (p *parser) parseMath2() (*expr, error) {
	return p.parsePow()
}

func (p *parser) parseConcat() (*expr, error) {
	return p.parseMath2()
}

func (p *parser) parseMath1() (*expr, error) {
	return p.parseConcat()
}

func (p *parser) parseCompare() (*expr, error) {
	return p.parseMath1()
}

func (p *parser) parseNot() (*expr, error) {
	return p.parseCompare()
}

func (p *parser) parseAnd() (*expr, error) {
	return p.parseNot()
}

func (p *parser) parseOr() (*expr, error) {
	return p.parseAnd()
}

func (p *parser) parseIfExpr() (*expr, error) {
	exp, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	// not implemented
	return exp, nil
}

func (p *parser) parseExpr() (*expr, error) {
	return withRecursionGuard[expr](p, p.parseIfExpr)
}

func (p *parser) expectToken(f func(tkn token) bool, expected string) (token, span, error) {
	tkn, spn, err := p.stream.next()
	if err != nil {
		return token{}, span{}, err
	}
	if tkn == nil {
		return token{}, span{}, unexpectedEOF(expected)
	}
	if f(*tkn) {
		return *tkn, *spn, nil
	}
	return token{}, span{}, unexpected(tkn, expected)
}

func (p *parser) matchesToken(f func(tkn token) bool) (bool, error) {
	tkn, _, err := p.stream.current()
	if err != nil {
		return false, err
	}
	if tkn == nil {
		return false, nil
	}
	return f(*tkn), nil
}

func (p *parser) skipToken(k tokenKind) (matched bool, err error) {
	if err = p.stream.curErr; err != nil {
		return false, err
	}
	if p.stream.curToken.kind == k {
		p.stream.next()
		return true, nil
	}
	return false, nil
}

const parseMaxRecursion = 150

func withRecursionGuard[T any](p *parser, f func() (*T, error)) (*T, error) {
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

func makeConst(v value, spn span) *expr {
	return &expr{
		kind: exprKindConst,
		data: constExprData{value: v},
		span: spn,
	}
}

func syntaxError(msg string) error {
	return &Error{
		kind:   SyntaxError,
		detail: option[string]{valid: true, data: msg},
	}
}

func (p *parser) subparse(endCheck func(*token) bool) ([]stmt, error) {
	var rv []stmt
	for {
		tkn, spn, err := p.stream.next()
		if err != nil {
			return nil, err
		}
		if tkn == nil {
			break
		}
		switch tkn.kind {
		case tokenKindTemplateData:
			raw := tkn.data.(templateDataTokenData)
			rv = append(rv, stmt{
				kind: stmtKindEmitRaw,
				data: emitRawStmtData{raw: raw},
				span: *spn,
			})
		case tokenKindVariableStart:
			exp, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			rv = append(rv, stmt{
				kind: stmtKindEmitExpr,
				data: emitExprStmtData{expr: *exp},
				span: p.stream.expandSpan(*spn),
			})
			if _, _, err := p.expectToken(isTokenOfKind(tokenKindVariableEnd), "end of variable block"); err != nil {
				return nil, err
			}
		case tokenKindBlockStart:
			panic("not implemented")
		default:
			panic("lexer produced garbage")
		}
	}
	return rv, nil
}

func (p *parser) parse() (*stmt, error) {
	spn := p.stream.lastSpan
	ss, err := p.subparse(func(*token) bool { return false })
	if err != nil {
		return nil, err
	}
	return &stmt{
		kind: stmtKindTemplate,
		data: templateStmtData{children: ss},
		span: p.stream.expandSpan(spn),
	}, nil
}

func parse(source, filename string) (*stmt, error) {
	return parseWithSyntax(source, filename, DefaultSyntaxConfig)
}

func parseWithSyntax(source, filename string, syntax SyntaxConfig) (*stmt, error) {
	// we want to chop off a single newline at the end.  This means that a template
	// by default does not end in a newline which is a useful property to allow
	// inline templates to work.  If someone wants a trailing newline the expectation
	// is that the user adds it themselves for achieve consistency.
	source = strings.TrimSuffix(source, "\n")
	source = strings.TrimSuffix(source, "\r")

	parser := newParser(source, false, &syntax)
	return parser.parse()
}

func isTokenOfKind(k tokenKind) func(tkn token) bool {
	return func(tkn token) bool {
		return tkn.kind == k
	}
}
