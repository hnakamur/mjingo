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

func (s *tokenizerState) loc() loc {
	return loc{
		line:   s.currentLine,
		col:    s.currentCol,
		offset: s.currentOffset,
	}
}

func (s *tokenizerState) span(startLoc loc) *span {
	return &span{
		startLine:   startLoc.line,
		startCol:    startLoc.col,
		startOffset: startLoc.offset,
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

func (s *tokenizerState) eatNumber() (token, *span, error) {
	type numberState int
	const (
		numberStateIntger numberState = iota
		numberStateFraction
		numberStateExponent
		numberStateExponentSign
	)

	oldLoc := s.loc()
	state := numberStateIntger
	numLen := prefixLenFunc(s.rest, isASCIIDigit)
loop:
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
			break loop
		}
		numLen++
	}

	num := s.advance(uint(numLen))
	if state != numberStateIntger {
		fVal, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return nil, nil, s.syntaxError("invalid float")
		}
		return floatToken{f: fVal}, s.span(oldLoc), nil
	}
	iVal, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return nil, nil, s.syntaxError("invalid integer")
	}
	return intToken{n: iVal}, s.span(oldLoc), nil
}

func (s *tokenizerState) eatIdentifier() (token, *span, error) {
	identLen := lexIndentifier(s.rest)
	if identLen == 0 {
		return nil, nil, s.syntaxError("unexpected character")
	}
	oldLoc := s.loc()
	ident := s.advance(identLen)
	return identToken{ident: ident}, s.span(oldLoc), nil
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

func (s *tokenizerState) eatString(delim rune) (token, *span, error) {
	oldLoc := s.loc()
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
		return stringToken{s: str}, s.span(oldLoc), nil
	}
	return stringToken{s: str[1 : len(str)-1]}, s.span(oldLoc), nil
}

func (s *tokenizerState) skipWhitespace() {
	if skip := prefixLenFunc(s.rest, isWhitespace); skip > 0 {
		s.advance(uint(skip))
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
	return &Error{typ: SyntaxError, detail: option[string]{valid: true, data: msg}}
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

func (i *tokenizeIterator) Next() (token, *span, error) {
	for i.state.rest != "" && !i.state.failed {
		oldLoc := i.state.loc()
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
						i.state.advance(uint(end) + skip + uint(len(i.commentEnd)))
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
					return variableStartToken{}, i.state.span(oldLoc), nil
				case startMarkerBlock:
					// raw blocks require some special handling.  If we are at the beginning of a raw
					// block we want to skip everything until {% endraw %} completely ignoring iterior
					// syntax and emit the entire raw block as TemplateData.
					if raw, trimStart, ok := skipBasicTag(i.state.rest[skip:], "raw", i.blockEnd); ok {
						i.state.advance(raw + skip)
						ptr := 0
						for {
							block := strings.Index(i.state.rest[ptr:], i.blockStart)
							if block == -1 {
								break
							}
							ptr += block + len(i.blockStart)
							trimEnd := i.state.rest[ptr] == '-'
							if endRaw, trimNext, ok := skipBasicTag(i.state.rest[ptr:], "endraw", i.blockEnd); ok {
								result := i.state.rest[:ptr-len(i.blockStart)]
								if trimStart {
									result = trimSpacePrefix(result)
								}
								if trimEnd {
									result = trimSpaceSuffix(result)
								}
								i.state.advance(uint(ptr) + endRaw)
								i.trimLeadingWhitespace = trimNext
								return templateDataToken{s: result}, i.state.span(oldLoc), nil
							}
						}
						return nil, nil, i.state.syntaxError("unexpected end of raw block")
					}
					if strings.HasPrefix(i.state.rest[skip:], "-") {
						i.state.advance(skip + 1)
					} else {
						i.state.advance(skip)
					}

					i.state.stack.push(lexerStateInBlock)
					return blockStartToken{}, i.state.span(oldLoc), nil
				}
			}

			if i.trimLeadingWhitespace {
				i.trimLeadingWhitespace = false
				i.state.skipWhitespace()
			}
			oldLoc = i.state.loc()

			var lead string
			var spn *span
			start, hyphen, found := findStartMarker(i.state.rest, i.syntaxConfig)
			if found {
				if hyphen {
					peeked := i.state.rest[:start]
					trimmed := strings.TrimRightFunc(peeked, isWhitespace)
					lead = i.state.advance(uint(len(trimmed)))
					spn = i.state.span(oldLoc)
					i.state.advance(uint(len(peeked) - len(trimmed)))
				} else {
					lead = i.state.advance(start)
					spn = i.state.span(oldLoc)
				}
			} else {
				lead = i.state.advance(uint(len(i.state.rest)))
				spn = i.state.span(oldLoc)
			}
			if lead == "" {
				continue
			}
			return templateDataToken{s: lead}, spn, nil

		case lexerStateInBlock, lexerStateInVariable:
			// in blocks whitespace is generally ignored, skip it.
			if trimLen := prefixLenFunc(i.state.rest, isASCIIWhitespace); trimLen > 0 {
				i.state.advance(trimLen)
				continue
			}

			// look out for the end of blocks
			if i.state.stack.peek() == lexerStateInBlock {
				if strings.HasPrefix(i.state.rest, "-") && strings.HasPrefix(i.state.rest[1:], i.blockEnd) {
					i.state.stack.pop()
					i.trimLeadingWhitespace = true
					i.state.advance(1 + uint(len(i.blockEnd)))
					return blockEndToken{}, i.state.span(oldLoc), nil
				}
				if strings.HasPrefix(i.state.rest, i.blockEnd) {
					i.state.stack.pop()
					i.state.advance(uint(len(i.blockEnd)))
					return blockEndToken{}, i.state.span(oldLoc), nil
				}
			} else {
				if strings.HasPrefix(i.state.rest, "-") && strings.HasPrefix(i.state.rest[1:], i.variableEnd) {
					i.state.stack.pop()
					i.trimLeadingWhitespace = true
					i.state.advance(1 + uint(len(i.variableEnd)))
					return variableEndToken{}, i.state.span(oldLoc), nil
				}
				if strings.HasPrefix(i.state.rest, i.variableEnd) {
					i.state.stack.pop()
					i.state.advance(uint(len(i.variableEnd)))
					return variableEndToken{}, i.state.span(oldLoc), nil
				}
			}

			// two character operators
			if len(i.state.rest) >= 2 {
				var tk token
				switch i.state.rest[:2] {
				case "//":
					tk = floorDivToken{}
				case "**":
					tk = powToken{}
				case "==":
					tk = eqToken{}
				case "!=":
					tk = neToken{}
				case ">=":
					tk = gteToken{}
				case "<=":
					tk = lteToken{}
				}
				if tk != nil {
					i.state.advance(2)
					return tk, i.state.span(oldLoc), nil
				}
			}

			if len(i.state.rest) >= 1 {
				var tk token
				switch i.state.rest[0] {
				case '+':
					tk = plusToken{}
				case '-':
					tk = minusToken{}
				case '*':
					tk = mulToken{}
				case '/':
					tk = divToken{}
				case '%':
					tk = modToken{}
				case '!':
					tk = bangToken{}
				case '.':
					tk = dotToken{}
				case ',':
					tk = commaToken{}
				case ':':
					tk = colonToken{}
				case '~':
					tk = tildeToken{}
				case '|':
					tk = pipeToken{}
				case '=':
					tk = assignToken{}
				case '<':
					tk = ltToken{}
				case '>':
					tk = gtToken{}
				case '(':
					tk = parenOpenToken{}
				case ')':
					tk = parenCloseToken{}
				case '[':
					tk = bracketOpenToken{}
				case ']':
					tk = bracketCloseToken{}
				case '{':
					tk = braceOpenToken{}
				case '}':
					tk = braceCloseToken{}
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
					i.state.advance(1)
					return tk, i.state.span(oldLoc), nil
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
