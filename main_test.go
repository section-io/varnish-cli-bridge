package main

import "testing"
import (
	"bytes"
	"strings"
)

func TestCliResponseStatusLineLengthFieldIsLeftAligned(t *testing.T) {
	expected := "200 4       \n"
	mockWriter := new(bytes.Buffer)
	writeVarnishCliResponse(mockWriter, CLIS_OK, "four")
	actual := mockWriter.String()
	if !strings.HasPrefix(actual, expected) {
		t.Errorf("Expected response to being with '%#v' but was '%#v'.", expected, actual)
	}
}
