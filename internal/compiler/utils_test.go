package compiler

import "testing"

func TestUnescape(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: `foo\u2603bar`, want: "foo\u2603bar"},
		{input: `\t\b\f\r\n\\\/`, want: "\t\x08\x0c\r\n\\/"},
		{input: `foobarbaz`, want: "foobarbaz"},
		{input: `\ud83d\udca9`, want: "💩"},
	}
	for _, tc := range testCases {
		got, err := unescape(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("unescape(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
