package mjingo

import (
	"testing"
)

func TestBasicIdentifiers(t *testing.T) {
	t.Run("ident", func(t *testing.T) {
		assertIdent := func(s string) {
			tk, _, err := tokenize(s, true, nil).Next()
			if err != nil {
				t.Fatal(err)
			}
			if tk, ok := tk.(identToken); ok {
				if got, want := tk.ident, s; got != want {
					t.Errorf("token StrData  mismatch, got=%q, want=%q, input=%q", got, want, s)
				}
			} else {
				t.Errorf("token type mismatch, got=%T, want=%v, input=%q", tk, tokenTypeIdent, s)
			}
		}

		assertIdent("foo_bar_baz")
		assertIdent("_foo_bar_baz")
		assertIdent("_42world")
		assertIdent("_world42")
		assertIdent("world42")
	})
	t.Run("notIdent", func(t *testing.T) {
		assertNotIdent := func(s string) {
			it := tokenize(s, true, nil)
			tk, _, _ := it.Next()
			if tk, ok := tk.(identToken); ok {
				t.Errorf("token should not be an identifier, got=%s, input=%q", tk.typ(), s)
			}
		}

		assertNotIdent("42world")
	})
}

func TestFindstartMarkerIndexRune(t *testing.T) {
	testCases := []struct {
		input  string
		pos    uint
		hyphen bool
		found  bool
	}{
		{input: "", pos: 0, hyphen: false, found: false},
		{input: "  {% if true %}", pos: 2, hyphen: false, found: true},
		{input: "  {%- if true %}", pos: 2, hyphen: true, found: true},
		{input: " {{ var }}", pos: 1, hyphen: false, found: true},
		{input: " {{- var }}", pos: 1, hyphen: true, found: true},
		{input: "   {# comment #}", pos: 3, hyphen: false, found: true},
		{input: "   {#- comment #}", pos: 3, hyphen: true, found: true},
	}
	for _, tc := range testCases {
		pos, hyphen, found := findStartMarkerIndexRune(tc.input)
		if pos != tc.pos || hyphen != tc.hyphen || found != tc.found {
			t.Errorf("expected pos=%d hyphen=%t found=%t, got pos=%d hyphen=%t found=%t", tc.pos, tc.hyphen, tc.found, pos, hyphen, found)
		}
	}
}
