package mjingo

import "testing"

func TestFindStartMaker(t *testing.T) {
	testCases := []struct {
		input  string
		pos    uint
		hyphen bool
		found  bool
	}{
		{input: "{", pos: 0, hyphen: false, found: false},
		{input: "foo", pos: 0, hyphen: false, found: false},
		{input: "foo {", pos: 0, hyphen: false, found: false},
		{input: "foo {{", pos: 4, hyphen: false, found: true},
		{input: "foo {{-", pos: 4, hyphen: true, found: true},
	}
	for _, tc := range testCases {
		pos, hyphen, found := findStartMarker(tc.input, nil)
		if pos != tc.pos || hyphen != tc.hyphen || found != tc.found {
			t.Errorf("got pos=%d hyphen=%t found=%t, want pos=%d hyphen=%t found=%t for input=%q",
				pos, hyphen, found, tc.pos, tc.hyphen, tc.found, tc.input)
		}
	}
}

func TestSkipBasicTag(t *testing.T) {
	testCases := []struct {
		blockStr, name, blockEnd string
		raw                      uint
		trim                     bool
		ok                       bool
	}{
		{blockStr: " raw %}", name: "raw", blockEnd: "%}", raw: 7, trim: false, ok: true},
		{blockStr: " raw %}", name: "endraw", blockEnd: "%}", raw: 0, trim: false, ok: false},
		{blockStr: "  raw  %}", name: "raw", blockEnd: "%}", raw: 9, trim: false, ok: true},
		{blockStr: "-  raw  -%}", name: "raw", blockEnd: "%}", raw: 11, trim: true, ok: true},
	}
	for _, tc := range testCases {
		raw, trim, ok := skipBasicTag(tc.blockStr, tc.name, tc.blockEnd)
		if raw != tc.raw || trim != tc.trim || ok != tc.ok {
			t.Errorf("got raw=%d trim=%t ok=%t, want raw=%d trim=%t ok=%t for blockStr=%q name=%q blockEnd=%q",
				raw, trim, ok, tc.raw, tc.trim, tc.ok, tc.blockStr, tc.name, tc.blockEnd)
		}
	}
}
