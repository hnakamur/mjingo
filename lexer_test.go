package mjingo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	mustReadFile := func(filename string) string {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("cannot read file, filename=%s, err=%v", filename, err)
		}
		return string(data)
	}

	inputFilenames, err := filepath.Glob(filepath.Join(".", "tests", "lexer-inputs", "*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, inputFilename := range inputFilenames {
		inputContent := mustReadFile(inputFilename)
		iter := tokenize(inputContent, false, &defaultSyntaxConfig)
		var b strings.Builder
		for {
			tkn, spn, err := iter.Next()
			if err != nil {
				t.Fatal(err)
			}
			if tkn == nil {
				break
			}
			tokenSource := inputContent[spn.StartOffset:spn.EndOffset]
			fmt.Fprintf(&b, "%s\n", tkn.DebugString())
			fmt.Fprintf(&b, "  %q\n", tokenSource)
		}
		got := b.String()

		snapFilename := inputFilename + ".snap"
		want := mustReadFile(snapFilename)
		if got != want {
			t.Errorf("result mismatch, inputFilename=%s,\n got=%s\nwant=%s", inputFilename, got, want)
		}
	}
}

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
