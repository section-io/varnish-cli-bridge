package main

import "testing"
import (
	"bytes"
	"strconv"
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

func TestAuthenticationChallengeIsRemembered(t *testing.T) {
	mockWriter := new(bytes.Buffer)
	mockSession := &varnishCliSession{mockWriter, false, true, ""}
	writeVarnishCliAuthenticationChallenge(mockSession)
	response := mockWriter.String()
	fields := strings.Fields(response)
	if len(fields) < 3 {
		t.Errorf("Expected response to contain at least 3 fields but had %v. Raw: %#v", len(fields), response)
	}
	if code, err := strconv.ParseInt(fields[0], 10, 16); err != nil || code != CLIS_AUTH {
		t.Errorf("Expected response code 107 (auth) but was %#v. Raw: %#v", fields[0], response)
	}
	if len(fields[2]) < 32 {
		t.Errorf("Expected challenge to be at least 32 characters but was %v. Raw: %#v", len(fields[2]), response)
	}
	responseChallenge := fields[2][:32]
	if responseChallenge != mockSession.AuthChallenge {
		t.Errorf("Expected remembered challenge %#v to match response challenge %#v", mockSession.AuthChallenge, responseChallenge)
	}
}

func TestCalculatedAuthenticator(t *testing.T) {
	// example values from https://www.varnish-cache.org/docs/trunk/reference/varnish-cli.html
	authChallenge := "ixslvvxrgkjptxmcgnnsdxsvdmvfympg"
	secretBytes := []byte("foo\n")
	authenticator := "455ce847f0073c7ab3b1465f74507b75d3dc064c1e7de3b71e00de9092fdc89a"

	mockWriter := new(bytes.Buffer)
	mockSession := &varnishCliSession{mockWriter, false, true, authChallenge}

	handleVarnishCliAuthenticationAttemptInternal(authenticator, mockSession, secretBytes)
	response := mockWriter.String()

	if !strings.HasPrefix(response, "200 ") {
		t.Errorf("Expected response to being with '200 ' but was '%#v'.", response)
	}
	if !mockSession.HasAuthenticated {
		t.Errorf("Expected session to become authenticated but was not.")
	}
}
