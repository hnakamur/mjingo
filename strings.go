package mjingo

// The code for trimLeftSpace and trimRightSpace are adapted from
// Go's strings.TrimSpace.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func trimSpacePrefix(s string) string {
	// Fast path for ASCII: look for the first ASCII non-space byte
	start := 0
	for ; start < len(s); start++ {
		c := s[start]
		if c >= utf8.RuneSelf {
			// If we run into a non-ASCII byte, fall back to the
			// slower unicode-aware method on the remaining bytes
			return strings.TrimLeftFunc(s[start:], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}

	// At this point s[start:] starts with an ASCII
	// non-space bytes, so we're done. Non-ASCII cases have already
	// been handled above.
	return s[start:]
}

func trimSpaceSuffix(s string) string {
	// Now look for the first ASCII non-space byte from the end
	stop := len(s)
	for ; stop > 0; stop-- {
		c := s[stop-1]
		if c >= utf8.RuneSelf {
			// start has been already trimmed above, should trim end only
			return strings.TrimRightFunc(s[:stop], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}

	// At this point s[:stop]  ends with an ASCII
	// non-space bytes, so we're done. Non-ASCII cases have already
	// been handled above.
	return s[:stop]
}
