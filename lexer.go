package mjingo

import (
	"strconv"
	"strings"
	"unicode"
)

type lexerState int

const (
	lexerStateTemplate lexerState = iota
	lexerStateInVariable
	lexerStateInBlock
)

type startMarker int

const (
	startMarkerVariable startMarker = iota
	startMarkerBlock
	startMarkerComment
)

type tokenizerState struct {
	stack         lexerStateStack
	rest          string
	failed        bool
	currentLine   uint32
	currentCol    uint32
	currentOffset uint32
}

func (s *tokenizerState) loc() (line, col, offset uint32) {
	return s.currentLine, s.currentCol, s.currentOffset
}

func (s *tokenizerState) span(startLine, startCol, startOffset uint32) *span {
	return &span{
		startLine:   startLine,
		startCol:    startCol,
		startOffset: startOffset,
		endLine:     s.currentLine,
		endCol:      s.currentCol,
		endOffset:   s.currentOffset,
	}
}

func (s *tokenizerState) advance(bytes uint) string {
	skipped, newRest := s.rest[:bytes], s.rest[bytes:]
	for _, c := range skipped {
		if c == '\n' {
			s.currentLine++
			s.currentCol = 0
		} else {
			s.currentCol++
		}
	}
	s.currentOffset += uint32(bytes)
	s.rest = newRest
	return skipped
}

func (s *tokenizerState) eatNumber() (*token, *span, error) {
	type numberState int
	const (
		numberStateIntger numberState = iota
		numberStateFraction
		numberStateExponent
		numberStateExponentSign
	)

	oldLine, oldCol, oldOffset := s.loc()
	state := numberStateIntger
	numLen := prefixLenFunc(s.rest, isASCIIDigit)
	for _, c := range s.rest[numLen:] {
		switch c {
		case '.':
			if state == numberStateIntger {
				state = numberStateFraction
			} else {
				break
			}
		case 'e', 'E':
			if state == numberStateIntger || state == numberStateFraction {
				state = numberStateExponent
			} else {
				break
			}
		case '+', '-':
			if state == numberStateExponent {
				state = numberStateExponentSign
			} else {
				break
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if state == numberStateExponent {
				state = numberStateExponentSign
			}
		default:
			break
		}
		numLen++
	}

	num := s.advance(uint(numLen))
	if state != numberStateIntger {
		fVal, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return nil, nil, s.syntaxError("invalid float")
		}
		return &token{Type: tokenFloat, FloatData: fVal}, s.span(oldLine, oldCol, oldOffset), nil
	}
	iVal, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return nil, nil, s.syntaxError("invalid integer")
	}
	return &token{Type: tokenInt, IntData: iVal}, s.span(oldLine, oldCol, oldOffset), nil
}

func (s *tokenizerState) eatIdentifier() (*token, *span, error) {
	identLen := lexIndentifier(s.rest)
	if identLen == 0 {
		return nil, nil, s.syntaxError("unexpected character")
	}
	oldLine, oldCol, oldOffset := s.loc()
	ident := s.advance(identLen)
	return &token{Type: tokenIdent, StrData: ident}, s.span(oldLine, oldCol, oldOffset), nil
}

func lexIndentifier(s string) uint {
	first := true
	identLen := prefixLenFunc(s, func(r rune) bool {
		var ok bool
		if r == '_' {
			ok = true
		} else if first {
			ok = isASCIIAlpha(r)
		} else {
			ok = isASCIIAlphaNumeric(r)
		}
		first = false
		return ok
	})
	return identLen
}

func isASCIIAlpha(r rune) bool {
	return 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z'
}

func isASCIIAlphaNumeric(r rune) bool {
	return 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' || '0' <= r && r <= '9'
}

func (s *tokenizerState) eatString(delim rune) (*token, *span, error) {
	oldLine, oldCol, oldOffset := s.loc()
	escaped := false
	hasEscapes := false
	strLen := prefixLenFunc(s.rest[1:], func(r rune) bool {
		if escaped {
			escaped = false
			return true
		}
		if r == '\\' {
			escaped = true
			hasEscapes = true
			return true
		}
		return r != delim
	})
	if escaped || !strings.HasPrefix(s.rest[1+strLen:], string(delim)) {
		return nil, nil, s.syntaxError("unexpected end of string")
	}
	str := s.advance(strLen + 2)
	if hasEscapes {
		str, err := unescape(str[1 : len(str)-1])
		if err != nil {
			return nil, nil, err
		}
		return &token{Type: tokenString, StrData: str}, s.span(oldLine, oldCol, oldOffset), nil
	}
	return &token{Type: tokenString, StrData: str[1 : len(str)-1]}, s.span(oldLine, oldCol, oldOffset), nil
}

func (s *tokenizerState) skipWhitespace() {
	if skip := prefixLenFunc(s.rest, isWhitespace); skip > 0 {
		_ = s.advance(uint(skip))
	}
}

func isWhitespace(r rune) bool {
	return r == ' ' || ('\x09' <= r && r <= '\x0d') || (r > '\x7f' && unicode.IsSpace(r))
}

func isASCIIWhitespace(r rune) bool {
	switch r {
	case '\t', '\n', '\f', '\r', ' ':
		return true
	default:
		return false
	}
}

func (s *tokenizerState) syntaxError(msg string) error {
	s.failed = true
	return &Error{kind: SyntaxError, detail: msg}
}

type tokenizeIterator struct {
	state                 tokenizerState
	inExpr                bool
	syntaxConfig          *SyntaxConfig
	trimLeadingWhitespace bool
	variableEnd           string
	blockStart            string
	blockEnd              string
	commentEnd            string
}

type lexerStateStack struct {
	stack []lexerState
}

func (s *lexerStateStack) push(state lexerState) {
	s.stack = append(s.stack, state)
}

func (s *lexerStateStack) empty() bool {
	return len(s.stack) == 0
}

func (s *lexerStateStack) pop() lexerState {
	st := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return st
}

func (s *lexerStateStack) peek() lexerState {
	return s.stack[len(s.stack)-1]
}

func Tokenize(input string, inExpr bool, syntaxConfig *SyntaxConfig) *tokenizeIterator {
	if syntaxConfig == nil {
		syntaxConfig = &DefaultSyntaxConfig
	}
	return newTokenizeIterator(input, inExpr, syntaxConfig)
}

func newTokenizeIterator(input string, inExpr bool, syntaxConfig *SyntaxConfig) *tokenizeIterator {
	ls := lexerStateTemplate
	if inExpr {
		ls = lexerStateInVariable
	}
	stack := lexerStateStack{}
	stack.push(ls)
	state := tokenizerState{
		rest:          input,
		stack:         stack,
		failed:        false,
		currentLine:   1,
		currentCol:    0,
		currentOffset: 0,
	}

	return &tokenizeIterator{
		state:                 state,
		inExpr:                inExpr,
		syntaxConfig:          syntaxConfig,
		trimLeadingWhitespace: false,
		variableEnd:           syntaxConfig.syntax.VariableEnd,
		blockStart:            syntaxConfig.syntax.BlockStart,
		blockEnd:              syntaxConfig.syntax.BlockEnd,
		commentEnd:            syntaxConfig.syntax.CommentEnd,
	}
}

func (i *tokenizeIterator) Next() (*token, *span, error) {
	for i.state.rest != "" && !i.state.failed {
		oldLine, oldCol, oldOffset := i.state.loc()
		if i.state.stack.empty() {
			panic("empty lexer state")
		}
		switch i.state.stack.peek() {
		case lexerStateTemplate:
			if startMarker, skip, matched := matchStartMarker(i.state.rest, i.syntaxConfig); matched {
				switch startMarker {
				case startMarkerComment:
					if end := strings.Index(i.state.rest[skip:], i.commentEnd); end != -1 {
						if i.state.rest[skip+uint(end)-1] == '-' {
							i.trimLeadingWhitespace = true
						}
						_ = i.state.advance(uint(end) + skip + uint(len(i.commentEnd)))
						continue
					} else {
						return nil, nil, i.state.syntaxError("unexpected end of comment")
					}
				case startMarkerVariable:
					if strings.HasPrefix(i.state.rest[skip:], "-") {
						i.state.advance(skip + 1)
					} else {
						i.state.advance(skip)
					}
					i.state.stack.push(lexerStateInVariable)
					return &token{Type: tokenVariableStart}, i.state.span(oldLine, oldCol, oldOffset), nil
				case startMarkerBlock:
					// raw blocks require some special handling.  If we are at the beginning of a raw
					// block we want to skip everything until {% endraw %} completely ignoring iterior
					// syntax and emit the entire raw block as TemplateData.
				}
			}

			if i.trimLeadingWhitespace {
				i.trimLeadingWhitespace = false
				i.state.skipWhitespace()
			}
			oldLine, oldCol, oldOffset = i.state.loc()

			var lead string
			var spn *span
			start, hyphen, found := findStartMarker(i.state.rest, i.syntaxConfig)
			if found {
				if hyphen {
					peeked := i.state.rest[:start]
					trimmed := strings.TrimRightFunc(peeked, isWhitespace)
					lead = i.state.advance(uint(len(trimmed)))
					spn = i.state.span(oldLine, oldCol, oldOffset)
					i.state.advance(uint(len(peeked) - len(trimmed)))
				} else {
					lead = i.state.advance(start)
					spn = i.state.span(oldLine, oldCol, oldOffset)
				}
			} else {
				lead = i.state.advance(uint(len(i.state.rest)))
				spn = i.state.span(oldLine, oldCol, oldOffset)
			}
			if lead == "" {
				continue
			}
			return &token{Type: tokenTemplateData, StrData: lead}, spn, nil

		case lexerStateInBlock, lexerStateInVariable:
			// in blocks whitespace is generally ignored, skip it.
			if trimLen := prefixLenFunc(i.state.rest, isASCIIWhitespace); trimLen > 0 {
				_ = i.state.advance(trimLen)
				continue
			}

			// look out for the end of blocks
			if i.state.stack.peek() == lexerStateInBlock {
				if strings.HasPrefix(i.state.rest, "-") && strings.HasPrefix(i.state.rest[1:], i.blockEnd) {
					_ = i.state.stack.pop()
					i.trimLeadingWhitespace = true
					_ = i.state.advance(1 + uint(len(i.blockEnd)))
					return &token{Type: tokenBlockEnd}, i.state.span(oldLine, oldCol, oldOffset), nil
				}
				if strings.HasPrefix(i.state.rest, i.blockEnd) {
					_ = i.state.stack.pop()
					_ = i.state.advance(uint(len(i.blockEnd)))
					return &token{Type: tokenBlockEnd}, i.state.span(oldLine, oldCol, oldOffset), nil
				}
			} else {
				if strings.HasPrefix(i.state.rest, "-") && strings.HasPrefix(i.state.rest[1:], i.variableEnd) {
					_ = i.state.stack.pop()
					i.trimLeadingWhitespace = true
					_ = i.state.advance(1 + uint(len(i.variableEnd)))
					return &token{Type: tokenVariableEnd}, i.state.span(oldLine, oldCol, oldOffset), nil
				}
				if strings.HasPrefix(i.state.rest, i.variableEnd) {
					_ = i.state.stack.pop()
					_ = i.state.advance(uint(len(i.variableEnd)))
					return &token{Type: tokenVariableEnd}, i.state.span(oldLine, oldCol, oldOffset), nil
				}
			}

			// two character operators
			if len(i.state.rest) >= 2 {
				var tk *token
				switch i.state.rest[:2] {
				case "//":
					tk = &token{Type: tokenFloorDiv}
				case "**":
					tk = &token{Type: tokenPow}
				case "==":
					tk = &token{Type: tokenEq}
				case "!=":
					tk = &token{Type: tokenNe}
				case ">=":
					tk = &token{Type: tokenGte}
				case "<=":
					tk = &token{Type: tokenLte}
				}
				if tk != nil {
					_ = i.state.advance(2)
					return tk, i.state.span(oldLine, oldCol, oldOffset), nil
				}
			}

			if len(i.state.rest) >= 1 {
				var tk *token
				switch i.state.rest[0] {
				case '+':
					tk = &token{Type: tokenPlus}
				case '-':
					tk = &token{Type: tokenMinus}
				case '*':
					tk = &token{Type: tokenMul}
				case '/':
					tk = &token{Type: tokenDiv}
				case '%':
					tk = &token{Type: tokenMod}
				case '!':
					tk = &token{Type: tokenBang}
				case '.':
					tk = &token{Type: tokenDot}
				case ',':
					tk = &token{Type: tokenComma}
				case ':':
					tk = &token{Type: tokenColon}
				case '~':
					tk = &token{Type: tokenTilde}
				case '|':
					tk = &token{Type: tokenPipe}
				case '=':
					tk = &token{Type: tokenAssign}
				case '<':
					tk = &token{Type: tokenLt}
				case '>':
					tk = &token{Type: tokenGt}
				case '(':
					tk = &token{Type: tokenParenOpen}
				case ')':
					tk = &token{Type: tokenParenClose}
				case '[':
					tk = &token{Type: tokenBracketOpen}
				case ']':
					tk = &token{Type: tokenBracketClose}
				case '{':
					tk = &token{Type: tokenBraceOpen}
				case '}':
					tk = &token{Type: tokenBraceClose}
				case '\'':
					return i.state.eatString('\'')
				case '"':
					return i.state.eatString('"')
				default:
					if isASCIIDigit(rune(i.state.rest[0])) {
						return i.state.eatNumber()
					}
				}
				if tk != nil {
					_ = i.state.advance(1)
					return tk, i.state.span(oldLine, oldCol, oldOffset), nil
				}
			}

			return i.state.eatIdentifier()
		}
	}
	return nil, nil, nil
}

func skipBasicTag(blockStr, name, blockEnd string) (raw uint, trim bool, ok bool) {
	ptr := blockStr
	trim = false
	ptr = strings.TrimPrefix(ptr, "-")
	ptr = strings.TrimLeftFunc(ptr, isASCIIWhitespace)
	if strings.HasPrefix(ptr, name) {
		ptr = ptr[len(name):]
	} else {
		return 0, false, false
	}
	ptr = strings.TrimLeftFunc(ptr, isASCIIWhitespace)
	if strings.HasPrefix(ptr, "-") {
		ptr = ptr[1:]
		trim = true
	}
	if strings.HasPrefix(ptr, blockEnd) {
		ptr = ptr[len(blockEnd):]
	} else {
		return 0, false, false
	}
	return uint(len(blockStr) - len(ptr)), trim, true
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func matchStartMarker(rest string, syntaxConfig *SyntaxConfig) (startMarker, uint, bool) {
	return matchStartMarkerDefault(rest)
}

func matchStartMarkerDefault(rest string) (startMarker, uint, bool) {
	if strings.HasPrefix(rest, "{{") {
		return startMarkerVariable, 2, true
	}
	if strings.HasPrefix(rest, "{%") {
		return startMarkerBlock, 2, true
	}
	if strings.HasPrefix(rest, "{#") {
		return startMarkerComment, 2, true
	}
	return 0, 0, false
}

func findStartMarker(s string, syntaxConfig *SyntaxConfig) (pos uint, hyphen bool, found bool) {
	return findStartMarkerIndexRune(s)
}

func findStartMarkerIndexRune(s string) (pos uint, hyphen bool, found bool) {
	offset := 0
	for {
		idx := strings.IndexRune(s[offset:], '{')
		if idx == -1 {
			return 0, false, false
		}
		if j := offset + idx; j+1 < len(s) {
			switch s[j+1] {
			case '{', '%', '#':
				return uint(j), strings.HasPrefix(s[j+2:], "-"), true
			}
		}
		offset += idx + 1
	}
}

func prefixLenFunc(s string, f func(r rune) bool) uint {
	return uint(len(s) - len(strings.TrimLeftFunc(s, f)))
}
