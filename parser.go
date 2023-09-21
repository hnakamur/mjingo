package mjingo

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hnakamur/mjingo/internal/datast/hashset"
	"github.com/hnakamur/mjingo/option"
)

type tokenStream struct {
	iter     *tokenizeIterator
	curToken token
	curSpan  *span
	curErr   error
	lastSpan span
}

func newTokenStream(source string, inExpr bool, syntax *syntaxConfig) *tokenStream {
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

func (s *tokenStream) expandSpan(spn span) span {
	return span{
		StartLine:   spn.StartLine,
		StartCol:    spn.StartCol,
		StartOffset: spn.StartOffset,
		EndLine:     s.lastSpan.EndLine,
		EndCol:      s.lastSpan.EndCol,
		EndOffset:   s.lastSpan.EndOffset,
	}
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
	blocks  *hashset.StrHashSet
	depth   uint
}

func newParser(source string, inExpr bool, syntax *syntaxConfig) *parser {
	return &parser{
		stream: newTokenStream(source, inExpr, syntax),
		blocks: hashset.NewStrHashSet(),
	}
}

func (p *parser) parseFilterExpr(expr astExpr) (astExpr, error) {
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
			var spn span
			if tkIdent, spn, err = p.expectToken(isTokenOfType[identToken], "identifier"); err != nil {
				return nil, err
			}
			name := tkIdent.(identToken).ident
			var args []astExpr
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
				var spn span
				if tk, sp, err := p.expectToken(isTokenOfType[identToken], "`(`"); err != nil {
					return nil, err
				} else {
					name = tk.(identToken).ident
					spn = sp
				}
				args := []astExpr{}
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

func (p *parser) parseArgs() ([]astExpr, error) {
	args := []astExpr{}
	firstSpan := option.None[span]()
	kwargs := []kwargExpr{}

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
		getVarInVarAssign := func(expr astExpr) (option.Option[varExpr], error) {
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
		} else if optVarExp.IsSome() {
			varExp := optVarExp.Unwrap()
			if firstSpan.IsNone() {
				firstSpan = option.Some(varExp.span)
			}
			arg, err := p.parseExprNoIf()
			if err != nil {
				return nil, err
			}
			kwargs = append(kwargs, kwargExpr{key: varExp.id, arg: arg})
		} else if len(kwargs) != 0 {
			return nil, syntaxError("non-keyword arg after keyword arg")
		} else {
			args = append(args, expr)
		}
	}

	if len(kwargs) != 0 {
		args = append(args, kwargsExpr{
			pairs: kwargs,
			span:  p.stream.expandSpan(firstSpan.Unwrap()),
		})
	}

	return args, nil
}

func (p *parser) parsePostfix(exp astExpr, spn span) (astExpr, error) {
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

			start := option.None[astExpr]()
			stop := option.None[astExpr]()
			step := option.None[astExpr]()
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
				if start.IsNone() {
					return nil, syntaxError("empty subscript")
				}
				exp = getItemExpr{
					expr:          exp,
					subscriptExpr: start.Unwrap(),
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
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			exp = callExpr{
				call: call{
					expr: exp,
					args: args,
				},
				span: p.stream.expandSpan(spn),
			}
		default:
			break loop
		}
		spn = nextSpan
	}
	return exp, nil
}

func (p *parser) parsePrimary() (astExpr, error) {
	return p.withRecursionGuardExpr(p.parsePrimaryImpl)
}

func (p *parser) parsePrimaryImpl() (astExpr, error) {
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
			return makeConst(valueFromBool(true), *spn), nil
		case "false", "False":
			return makeConst(valueFromBool(false), *spn), nil
		case "none", "None":
			return makeConst(none, *spn), nil
		default:
			return varExpr{id: tkn.ident, span: *spn}, nil
		}
	case strToken:
		return makeConst(valueFromString(tkn.s), *spn), nil
	case intToken:
		return makeConst(valueFromU64(tkn.n), *spn), nil
	case int128Token:
		return makeConst(valueFromU128(tkn.n), *spn), nil
	case floatToken:
		return makeConst(valueFromF64(tkn.f), *spn), nil
	case parenOpenToken:
		return p.parseTupleOrExpression(*spn)
	case bracketOpenToken:
		return p.parseListExpr(*spn)
	case braceOpenToken:
		return p.parseMapExpr(*spn)
	default:
		return nil, syntaxError(fmt.Sprintf("unexpected %v", tkn))
	}
}

func (p *parser) parseListExpr(spn span) (astExpr, error) {
	var items []astExpr
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

func (p *parser) parseMapExpr(spn span) (astExpr, error) {
	var keys, values []astExpr
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

func (p *parser) parseTupleOrExpression(spn span) (astExpr, error) {
	// MiniJinja does not really have tuples, but it treats the tuple
	// syntax the same as lists.
	if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
		return nil, err
	} else if matched {
		return listExpr{
			items: []astExpr{},
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
		items := []astExpr{expr}
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

func (p *parser) parseUnaryOnly() (astExpr, error) {
	return p.unaryop(p.parseUnaryOnly, p.parsePrimary,
		func(tkn token) option.Option[unaryOpType] {
			if _, ok := tkn.(minusToken); ok {
				return option.Some(unaryOpTypeNeg)
			}
			return option.None[unaryOpType]()
		})
}

func (p *parser) parseUnary() (astExpr, error) {
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

func (p *parser) parsePow() (astExpr, error) {
	return p.binop(p.parseUnary, func(tkn token) option.Option[binOpType] {
		if _, ok := tkn.(powToken); ok {
			return option.Some(binOpTypePow)
		}
		return option.None[binOpType]()
	})
}

func (p *parser) parseMath2() (astExpr, error) {
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

func (p *parser) parseConcat() (astExpr, error) {
	return p.binop(p.parseMath2, func(tkn token) option.Option[binOpType] {
		if _, ok := tkn.(tildeToken); ok {
			return option.Some(binOpTypeConcat)
		}
		return option.None[binOpType]()
	})
}

func (p *parser) parseMath1() (astExpr, error) {
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

func (p *parser) parseCompare() (astExpr, error) {
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

func (p *parser) parseNot() (astExpr, error) {
	return p.unaryop(p.parseNot, p.parseCompare,
		func(tkn token) option.Option[unaryOpType] {
			if isIdentTokenWithName("not")(tkn) {
				return option.Some(unaryOpTypeNot)
			}
			return option.None[unaryOpType]()
		})
}

func (p *parser) parseAnd() (astExpr, error) {
	return p.binop(p.parseNot, func(tkn token) option.Option[binOpType] {
		if isIdentTokenWithName("and")(tkn) {
			return option.Some(binOpTypeScAnd)
		}
		return option.None[binOpType]()
	})
}

func (p *parser) parseOr() (astExpr, error) {
	return p.binop(p.parseAnd, func(tkn token) option.Option[binOpType] {
		if isIdentTokenWithName("or")(tkn) {
			return option.Some(binOpTypeScOr)
		}
		return option.None[binOpType]()
	})
}

func (p *parser) parseIfExpr() (astExpr, error) {
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
			exp3 := option.None[astExpr]()
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

func (p *parser) parseExpr() (astExpr, error) {
	return p.withRecursionGuardExpr(p.parseIfExpr)
}

func (p *parser) parseExprNoIf() (astExpr, error) {
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

func (p *parser) withRecursionGuardExpr(f func() (astExpr, error)) (astExpr, error) {
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
	return NewError(SyntaxError,
		fmt.Sprintf("unexpected %v, expected %s", unexpected, expected))
}

func unexpectedEOF(expected string) error {
	return unexpected("end of input", expected)
}

func makeConst(v Value, spn span) astExpr {
	return constExpr{val: v, span: spn}
}

func syntaxError(msg string) error {
	return NewError(SyntaxError, msg)
}

func (p *parser) parseMacroArgsAndDefaults(args, defaults *[]astExpr) error {
	for {
		if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
			return err
		} else if matched {
			break
		}
		if len(*args) != 0 {
			if _, _, err := p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
				return err
			}
			if matched, err := p.skipToken(isTokenOfType[parenCloseToken]); err != nil {
				return err
			} else if matched {
				break
			}
		}
		arg, err := p.parseAssignName()
		if err != nil {
			return err
		}
		*args = append(*args, arg)
		if matched, err := p.skipToken(isTokenOfType[assignToken]); err != nil {
			return err
		} else if matched {
			expr, err := p.parseExpr()
			if err != nil {
				return err
			}
			*defaults = append(*defaults, expr)
		} else if len(*defaults) != 0 {
			if _, _, err := p.expectToken(isTokenOfType[assignToken], "`=`"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *parser) parseMacroOrCallBlockBody(args, defaults []astExpr, name option.Option[string]) (macroStmt, error) {
	if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
		return macroStmt{}, err
	}
	oldInMacro := p.inMacro
	p.inMacro = true
	body, err := p.subparse(func(tkn token) bool {
		tk, ok := tkn.(identToken)
		return ok && ((tk.ident == "endmacro" && name.IsSome()) ||
			(tk.ident == "endcall" && name.IsNone()))
	})
	if err != nil {
		return macroStmt{}, err
	}
	p.inMacro = oldInMacro
	if _, _, err := p.stream.next(); err != nil {
		return macroStmt{}, err
	}
	return macroStmt{
		name:     name.UnwrapOr("caller"),
		args:     args,
		defaults: defaults,
		body:     body,
	}, nil
}

func (p *parser) parseMacro() (macroStmt, error) {
	tkn, _, err := p.expectToken(isTokenOfType[identToken], "identifier")
	if err != nil {
		return macroStmt{}, err
	}
	name := tkn.(identToken).ident
	if _, _, err := p.expectToken(isTokenOfType[parenOpenToken], "`(`"); err != nil {
		return macroStmt{}, err
	}
	var args, defaults []astExpr
	if err := p.parseMacroArgsAndDefaults(&args, &defaults); err != nil {
		return macroStmt{}, err
	}
	return p.parseMacroOrCallBlockBody(args, defaults, option.Some(name))
}

func (p *parser) parseCallBlock() (callBlockStmt, error) {
	spn := p.stream.lastSpan
	var args, defaults []astExpr
	if matched, err := p.skipToken(isTokenOfType[parenOpenToken]); err != nil {
		return callBlockStmt{}, err
	} else if matched {
		if err := p.parseMacroArgsAndDefaults(&args, &defaults); err != nil {
			return callBlockStmt{}, err
		}
	}
	expr, err := p.parseExpr()
	if err != nil {
		return callBlockStmt{}, err
	}
	callExp, ok := expr.(callExpr)
	if !ok {
		return callBlockStmt{}, syntaxError(fmt.Sprintf("expected call expression in call block, got %s", expr.typ().Description()))
	}
	macroDecl, err := p.parseMacroOrCallBlockBody(args, defaults, option.None[string]())
	if err != nil {
		return callBlockStmt{}, err
	}
	macroDecl.span = p.stream.expandSpan(spn)
	return callBlockStmt{
		call:      callExp.call,
		macroDecl: macroDecl,
	}, nil
}

func (p *parser) parseDo() (doStmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return doStmt{}, err
	}
	callExp, ok := expr.(callExpr)
	if !ok {
		return doStmt{}, syntaxError(fmt.Sprintf("expected call expression in call block, got %s", expr.typ().Description()))
	}
	return doStmt{
		call: callExp.call,
	}, nil
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
	case "block":
		st, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "extends":
		st, err := p.parseExtends()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "include":
		st, err := p.parseInclude()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "import":
		st, err := p.parseImport()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "from":
		st, err := p.parseFromImport()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "macro":
		st, err := p.parseMacro()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "call":
		st, err := p.parseCallBlock()
		if err != nil {
			return nil, err
		}
		st.span = p.stream.expandSpan(spn)
		return st, nil
	case "do":
		st, err := p.parseDo()
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

func (p *parser) parseAssignName() (astExpr, error) {
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

func (p *parser) parseAssignment() (astExpr, error) {
	spn := p.stream.currentSpan()
	var items []astExpr
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

		var item astExpr
		if matched, err := p.skipToken(isTokenOfType[parenOpenToken]); err != nil {
			return nil, err
		} else if matched {
			if rv, err := p.parseAssignment(); err != nil {
				return nil, err
			} else {
				if _, _, err := p.expectToken(isTokenOfType[parenCloseToken], "`)`"); err != nil {
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
	filterExpr := option.None[astExpr]()
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

		if len(assignments) != 0 {
			if _, _, err := p.expectToken(isTokenOfType[commaToken], "comma"); err != nil {
				return withBlockStmt{}, err
			}
		}
		var target astExpr
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
	var target astExpr
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
		filter := option.None[astExpr]()
		if matched, err := p.skipToken(isTokenOfType[pipeToken]); err != nil {
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

func (p *parser) parseBlock() (blockStmt, error) {
	if p.inMacro {
		return blockStmt{}, syntaxError("block tags in macros are not allowed")
	}

	tkn, _, err := p.expectToken(isTokenOfType[identToken], "identifier")
	if err != nil {
		return blockStmt{}, err
	}
	name := tkn.(identToken).ident
	if !p.blocks.Add(name) {
		return blockStmt{}, syntaxError(fmt.Sprintf("block '%s' defined twice", name))
	}

	if _, _, err := p.expectToken(isTokenOfType[blockEndToken], "end of block"); err != nil {
		return blockStmt{}, err
	}
	body, err := p.subparse(isIdentTokenWithName("endblock"))
	if err != nil {
		return blockStmt{}, err
	}
	if _, _, err := p.stream.next(); err != nil {
		return blockStmt{}, err
	}

	tkn, _, err = p.stream.current()
	if err != nil {
		return blockStmt{}, err
	}
	if tknIdent, ok := tkn.(identToken); ok {
		trailingName := tknIdent.ident
		if trailingName != name {
			return blockStmt{},
				syntaxError(fmt.Sprintf("mismatching name on block. Got `%s`, expected `%s`", trailingName, name))
		}
		if _, _, err := p.stream.next(); err != nil {
			return blockStmt{}, err
		}
	}

	return blockStmt{name: name, body: body}, nil
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

func (p *parser) parseFilterChain() (astExpr, error) {
	filter := option.None[astExpr]()

	for {
		if matched, err := p.matchesToken(isTokenOfType[blockEndToken]); err != nil {
			return nil, err
		} else if matched {
			break
		}
		if filter.IsSome() {
			if _, _, err := p.expectToken(isTokenOfType[pipeToken], "`|`"); err != nil {
				return nil, err
			}
		}
		var name string
		var spn span
		if tkn, s, err := p.expectToken(isTokenOfType[identToken], "identifier"); err != nil {
			return nil, err
		} else {
			name = tkn.(identToken).ident
			spn = s
		}
		args := []astExpr{}
		if matched, err := p.matchesToken(isTokenOfType[parenOpenToken]); err != nil {
			return nil, err
		} else if matched {
			if a, err := p.parseArgs(); err != nil {
				return nil, err
			} else {
				args = a
			}
		}
		filter = option.Some[astExpr](filterExpr{
			name: name,
			expr: filter,
			args: args,
			span: p.stream.expandSpan(spn),
		})
	}
	if filter.IsSome() {
		return filter.Unwrap(), nil
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

func (p *parser) parseExtends() (extendsStmt, error) {
	name, err := p.parseExpr()
	if err != nil {
		return extendsStmt{}, err
	}
	return extendsStmt{name: name}, nil
}

func (p *parser) parseInclude() (includeStmt, error) {
	name, err := p.parseExpr()
	if err != nil {
		return includeStmt{}, err
	}

	// with/without context is without meaning in MiniJinja, but for syntax
	// compatibility it's supported.
	if matched, err := p.skipToken(func(tkn token) bool {
		if identTkn, ok := tkn.(identToken); ok {
			return identTkn.ident == "without" || identTkn.ident == "with"
		}
		return false
	}); err != nil {
		return includeStmt{}, err
	} else if matched {
		if _, _, err := p.expectToken(isIdentTokenWithName("context"), "missing keyword"); err != nil {
			return includeStmt{}, err
		}
	}

	ignoreMissing := false
	if matched, err := p.skipToken(isIdentTokenWithName("ignore")); err != nil {
		return includeStmt{}, err
	} else if matched {
		ignoreMissing = true
		if _, _, err := p.expectToken(isIdentTokenWithName("missing"), "missing keyword"); err != nil {
			return includeStmt{}, err
		}
		if matched, err := p.skipToken(func(tkn token) bool {
			if identTkn, ok := tkn.(identToken); ok {
				return identTkn.ident == "without" || identTkn.ident == "with"
			}
			return false
		}); err != nil {
			return includeStmt{}, err
		} else if matched {
			if _, _, err := p.expectToken(isIdentTokenWithName("context"), "missing keyword"); err != nil {
				return includeStmt{}, err
			}
		}

	}

	return includeStmt{name: name, ignoreMissing: ignoreMissing}, nil
}

func (p *parser) parseImport() (importStmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return importStmt{}, err
	}
	if _, _, err = p.expectToken(isIdentTokenWithName("as"), "as"); err != nil {
		return importStmt{}, err
	}
	name, err := p.parseExpr()
	if err != nil {
		return importStmt{}, err
	}
	return importStmt{expr: expr, name: name}, nil
}

func (p *parser) parseFromImport() (fromImportStmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return fromImportStmt{}, err
	}
	if _, _, err = p.expectToken(isIdentTokenWithName("import"), "import"); err != nil {
		return fromImportStmt{}, err
	}
	var names []importName
	for {
		if matched, err := p.matchesToken(isTokenOfType[blockEndToken]); err != nil {
			return fromImportStmt{}, err
		} else if matched {
			break
		}
		if len(names) > 0 {
			if _, _, err = p.expectToken(isTokenOfType[commaToken], "`,`"); err != nil {
				return fromImportStmt{}, err
			}
		}
		if matched, err := p.matchesToken(isTokenOfType[blockEndToken]); err != nil {
			return fromImportStmt{}, err
		} else if matched {
			break
		}
		name, err := p.parseAssignName()
		if err != nil {
			return fromImportStmt{}, err
		}
		optAlias := option.None[astExpr]()
		if matched, err := p.skipToken(isIdentTokenWithName("as")); err != nil {
			return fromImportStmt{}, err
		} else if matched {
			alias, err := p.parseAssignName()
			if err != nil {
				return fromImportStmt{}, err
			}
			optAlias = option.Some(alias)
		}
		names = append(names, importName{name: name, as: optAlias})
	}
	return fromImportStmt{expr: expr, names: names}, nil
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
	return parseWithSyntax(source, filename, defaultSyntaxConfig, false)
}

func parseWithSyntax(source, filename string, syntax syntaxConfig, keepTrailingNewline bool) (statement, error) {
	// we want to chop off a single newline at the end.  This means that a template
	// by default does not end in a newline which is a useful property to allow
	// inline templates to work.  If someone wants a trailing newline the expectation
	// is that the user adds it themselves for achieve consistency.
	if !keepTrailingNewline {
		source = strings.TrimSuffix(source, "\n")
		source = strings.TrimSuffix(source, "\r")
	}

	parser := newParser(source, false, &syntax)
	stmt, err := parser.parse()
	if err != nil {
		var merr *Error
		if errors.As(err, &merr) {
			if merr.lineno.IsNone() {
				merr.setFilenameAndSpan(filename, parser.stream.lastSpan)
			}
		}
		return nil, err
	}
	return stmt, nil
}

func (p *parser) binop(next func() (astExpr, error), matchFn func(tkn token) option.Option[binOpType]) (astExpr, error) {
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
		if opType.IsNone() {
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
			op:    opType.Unwrap(),
			left:  left,
			right: right,
			span:  p.stream.expandSpan(spn),
		}
	}
	return left, nil
}

func (p *parser) unaryop(opFn, next func() (astExpr, error), matchFn func(tkn token) option.Option[unaryOpType]) (astExpr, error) {
	spn := p.stream.currentSpan()
	tkn, _, err := p.stream.current()
	if err != nil {
		return nil, err
	}
	if tkn == nil {
		return next()
	}
	opType := matchFn(tkn)
	if opType.IsNone() {
		return next()
	}
	if _, _, err := p.stream.next(); err != nil {
		return nil, err
	}
	exp, err := opFn()
	if err != nil {
		return nil, err
	}
	return unaryOpExpr{op: opType.Unwrap(), expr: exp, span: p.stream.expandSpan(spn)}, nil
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

func parseExpr(source string, syntax syntaxConfig) (astExpr, error) {
	parser := newParser(source, true, &syntax)
	expr, err := parser.parseExpr()
	if err == nil {
		if tkn, _, _ := parser.stream.next(); tkn != nil {
			err = syntaxError("unexpected input after expression")
		} else {
			return expr, nil
		}
	}
	var err2 *Error
	if errors.As(err, &err2) && err2.lineno.IsNone() {
		err2.setFilenameAndSpan("<expression>", parser.stream.lastSpan)
	}
	return nil, err
}
