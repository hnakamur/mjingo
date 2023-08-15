package mjingo

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

type UndefinedBehavior uint

const (
	// The default, somewhat lenient undefined behavior.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorLenient UndefinedBehavior = iota

	// Like `Lenient`, but also allows chaining of undefined lookups.
	//
	// * **printing:** allowed (returns empty string)
	// * **iteration:** allowed (returns empty array)
	// * **attribute access of undefined values:** allowed (returns [`undefined`](Value::UNDEFINED))
	UndefinedBehaviorChainable

	// Complains very quickly about undefined values.
	//
	// * **printing:** fails
	// * **iteration:** fails
	// * **attribute access of undefined values:** fails
	UndefinedBehaviorStrict

	UndefinedBehaviorDefault = UndefinedBehaviorLenient
)

func (b UndefinedBehavior) handleUndefined(parentWasUndefined bool) (option[value], error) {
	switch {
	case (b == UndefinedBehaviorLenient && !parentWasUndefined) || b == UndefinedBehaviorChainable:
		return option[value]{valid: true, data: valueUndefined}, nil
	case (b == UndefinedBehaviorLenient && parentWasUndefined) || b == UndefinedBehaviorStrict:
		return option[value]{}, &Error{kind: UndefinedError}
	default:
		panic("unreachable")
	}
}

// Un-escape a string, following JSON rules.
func unescape(s string) (string, error) {
	return (&unescaper{}).unescape(s)
}

type unescaper struct {
	out              []byte
	pendingSurrogate rune
}

func (u *unescaper) unescape(s string) (string, error) {
	rest := s
	for len(rest) > 0 {
		r, size := utf8.DecodeRuneInString(rest)
		if r == utf8.RuneError {
			if size != 0 {
				return "", &Error{kind: BadEscape}
			}
			break
		}
		rest = rest[size:]

		if r == '\\' {
			r, size = utf8.DecodeRuneInString(rest)
			if r == utf8.RuneError {
				return "", &Error{kind: BadEscape}
			}
			rest = rest[size:]

			switch r {
			case '"', '\\', '/', '\'':
				if err := u.pushChar(r); err != nil {
					return "", err
				}
			case 'b':
				if err := u.pushChar('\x08'); err != nil {
					return "", err
				}
			case 'f':
				if err := u.pushChar('\x0C'); err != nil {
					return "", err
				}
			case 'n':
				if err := u.pushChar('\n'); err != nil {
					return "", err
				}
			case 'r':
				if err := u.pushChar('\r'); err != nil {
					return "", err
				}
			case 't':
				if err := u.pushChar('\t'); err != nil {
					return "", err
				}
			case 'u':
				var val rune
				var err error
				val, rest, err = u.parseU16(rest)
				if err != nil {
					return "", err
				}
				if err := u.pushU16(val); err != nil {
					return "", err
				}
			}
		} else {
			if err := u.pushChar(r); err != nil {
				return "", err
			}
		}
	}
	if u.pendingSurrogate != 0 {
		return "", &Error{kind: BadEscape}
	}
	return string(u.out), nil
}

func (u *unescaper) parseU16(s string) (r rune, rest string, err error) {
	count := 0
	i := strings.IndexFunc(s, func(r rune) bool {
		count++
		return count > 4 || !unicode.Is(unicode.ASCII_Hex_Digit, r)
	})
	if i == -1 {
		i = len(s)
	}
	val, err := strconv.ParseUint(s[:i], 16, 16)
	if err != nil {
		return 0, "", &Error{kind: BadEscape}
	}
	return rune(val), s[i:], nil
}

func (u *unescaper) pushU16(c rune) error {
	if u.pendingSurrogate == 0 {
		if utf16.IsSurrogate(c) {
			u.pendingSurrogate = c
			return nil
		}
		return u.pushChar(c)
	}

	r := utf16.DecodeRune(u.pendingSurrogate, c)
	const replacementChar = '\ufffd'
	if r == replacementChar {
		return &Error{kind: BadEscape}
	}
	u.pendingSurrogate = 0
	return u.pushChar(r)
}

func (u *unescaper) pushChar(r rune) error {
	if u.pendingSurrogate != 0 {
		return &Error{kind: BadEscape}
	}
	u.out = utf8.AppendRune(u.out, r)
	return nil
}

type option[T any] struct {
	data  T
	valid bool
}

type stack[T any] struct {
	elems []T
}

func (s *stack[T]) push(elem T) {
	s.elems = append(s.elems, elem)
}

func (s *stack[T]) empty() bool {
	return len(s.elems) == 0
}

func (s *stack[T]) pop() *T {
	if s.empty() {
		return nil
	}
	st := s.elems[len(s.elems)-1]
	s.elems = s.elems[:len(s.elems)-1]
	return &st
}

func (s *stack[T]) peek() *T {
	if s.empty() {
		return nil
	}
	return &s.elems[len(s.elems)-1]
}

func untrustedSizeHint(value uint) uint {
	return min(value, 1024)
}
