package mjingo

// The code in this file is adapted from Go's strings_test.go
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import "testing"

// Test case for any function which accepts and returns a single string.
type stringTest struct {
	in, out string
}

// Execute f on each test case.  funcName should be the name of f; it's used
// in failure reports.
func runStringTests(t *testing.T, f func(string) string, funcName string, testCases []stringTest) {
	for _, tc := range testCases {
		actual := f(tc.in)
		if actual != tc.out {
			t.Errorf("%s(%q) = %q; want %q", funcName, tc.in, actual, tc.out)
		}
	}
}

const space = "\t\v\r\f\n\u0085\u00a0\u2000\u3000"

var trimSpacePrefixTests = []stringTest{
	{"", ""},
	{"abc", "abc"},
	{space + "abc" + space, "abc" + space},
	{" ", ""},
	{" \t\r\n \t\t\r\r\n\n ", ""},
	{" \t\r\n x\t\t\r\r\n\n ", "x\t\t\r\r\n\n "},
	{" \u2000\t\r\n x\t\t\r\r\ny\n \u3000", "x\t\t\r\r\ny\n \u3000"},
	{"1 \t\r\n2", "1 \t\r\n2"},
	{" x\x80 ", "x\x80 "},
	{" x\xc0", "x\xc0"},
	{"x \xc0\xc0 ", "x \xc0\xc0 "},
	{"x \xc0", "x \xc0"},
	{"x \xc0 ", "x \xc0 "},
	{"x \xc0\xc0 ", "x \xc0\xc0 "},
	{"x ☺\xc0\xc0 ", "x ☺\xc0\xc0 "},
	{"x ☺ ", "x ☺ "},
}

var trimSpaceSuffixTests = []stringTest{
	{"", ""},
	{"abc", "abc"},
	{space + "abc" + space, space + "abc"},
	{" ", ""},
	{" \t\r\n \t\t\r\r\n\n ", ""},
	{" \t\r\n x\t\t\r\r\n\n ", " \t\r\n x"},
	{" \u2000\t\r\n x\t\t\r\r\ny\n \u3000", " \u2000\t\r\n x\t\t\r\r\ny"},
	{"1 \t\r\n2", "1 \t\r\n2"},
	{" x\x80", " x\x80"},
	{" x\xc0", " x\xc0"},
	{"x \xc0\xc0 ", "x \xc0\xc0"},
	{"x \xc0", "x \xc0"},
	{"x \xc0 ", "x \xc0"},
	{"x \xc0\xc0 ", "x \xc0\xc0"},
	{"x ☺\xc0\xc0 ", "x ☺\xc0\xc0"},
	{"x ☺ ", "x ☺"},
}

func TestTrimSpacePrefix(t *testing.T) {
	runStringTests(t, trimSpacePrefix, "trimSpacePrefix", trimSpacePrefixTests)
}

func TestTrimSpaceSuffix(t *testing.T) {
	runStringTests(t, trimSpaceSuffix, "trimSpaceSuffix", trimSpaceSuffixTests)
}
