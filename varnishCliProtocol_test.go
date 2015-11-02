package main

import (
	"testing"
	"strings"
)

func TestDecodeBackSlash(t *testing.T) {
	testDecodeBackSlash(t,`a\nb\rc\td\"e\\ activies`,"a\nb\rc\td\"e\\"," activies")
	testDecodeBackSlash(t, `\x41`, "A", "")
	testDecodeBackSlash(t, `\x41\x42`, "AB", "")
	testDecodeBackSlash(t, `\101`, "A", "")
	testDecodeBackSlash(t, `\101\102`, "AB", "")
	testDecodeBackSlash(t, `\x41\102abc`, "ABabc", "")
	testDecodeBackSlash(t, `\x41\102abc def`, "ABabc", " def")
}

func TestDecodeBackSlashMalformed(t *testing.T) {
	// TODO Result from malformed input is undefined, but should probably mimic varnish
	testDecodeBackSlash(t, `\`, "\\", "")
	testDecodeBackSlash(t, `\a`, "\\a", "")
	testDecodeBackSlash(t, `\x4`, "\x04", "")
	testDecodeBackSlash(t, `\x`, "\\x", "")
	testDecodeBackSlash(t, `\02ABC`, "\x02BC", "")
}

func testDecodeBackSlash(t *testing.T, input string, tokenExpected string, tailExpected string) {
	tokenActual, tailActual, _  := readSingleToken(input)
	if !strings.HasPrefix(tokenActual, tokenExpected) {
		t.Errorf("Expected token of %#v to be %#v but was %#v.", input, tokenExpected, tokenActual)
	}
	if !strings.HasPrefix(tailActual, tailExpected) {
		t.Errorf("Expected tail of %#v to be %#v but was %#v.", input, tailExpected, tailActual)
	}
}
