package main

import (
	"reflect"
	"testing"
)

func TestDecodeBackSlash(t *testing.T) {
	testDecodeBackSlash(t, `a\nb\rc\td\"e\\ activies`, "a\nb\rc\td\"e\\", "activies")
	testDecodeBackSlash(t, `\x41`, "A", "")
	testDecodeBackSlash(t, `\x41\x42`, "AB", "")
	testDecodeBackSlash(t, `\101`, "A", "")
	testDecodeBackSlash(t, `\101\102`, "AB", "")
	testDecodeBackSlash(t, `\x41\102abc`, "ABabc", "")
	testDecodeBackSlash(t, `\x41\102abc def`, "ABabc", "def")
}

func TestDecodeBackSlashMalformed(t *testing.T) {
	// TODO Result from malformed input is undefined, but should probably mimic varnish
	testDecodeBackSlash(t, `\`, "\\", "")
	testDecodeBackSlash(t, `\a`, "\\a", "")
	testDecodeBackSlash(t, `\x4`, "\x04", "")
	testDecodeBackSlash(t, `\x`, "\\x", "")
	testDecodeBackSlash(t, `\02ABC`, "\x02BC", "")
}

func TestTokenizeRequest(t *testing.T) {
	input := `"auth" "2049dfd74a49800f06c28137df6c8224a56f6335f277b5fc773ac5831e5dcf07s"`
	tokensActual := tokenizeRequest(input)
	tokensExpected := []string{`auth`, `2049dfd74a49800f06c28137df6c8224a56f6335f277b5fc773ac5831e5dcf07`}

	if !reflect.DeepEqual(tokensActual, tokensExpected) {
		t.Errorf("Expected token of %#v to be %#v but was %#v.", input, tokensExpected, tokensActual)
	}
}

func testDecodeBackSlash(t *testing.T, input string, tokenExpected string, tailExpected string) {
	tokenActual, tailActual, _ := readSingleToken(input)
	if tokenActual != tokenExpected {
		t.Errorf("Expected token of %#v to be %#v but was %#v.", input, tokenExpected, tokenActual)
	}
	if tailActual != tailExpected {
		t.Errorf("Expected tail of %#v to be %#v but was %#v.", input, tailExpected, tailActual)
	}
}
