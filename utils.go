package mjingo

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

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
