package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func getVarnishSecret() ([]byte, error) {
	file, err := os.Open(secretFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open secret file '%s':\n%#v", secretFile, err)
	}
	defer file.Close()
	secretBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read secret file '%s':\n%#v", secretFile, err)
	}
	return secretBytes, nil
}

func writeVarnishCliAuthenticationChallenge(session *varnishCliSession) {
	const challengeSize = 32

	challengeBytes := make([]byte, challengeSize)
	bytesRead, err := rand.Read(challengeBytes)
	if err != nil || bytesRead != challengeSize {
		writeVarnishCliResponse(session.Writer, CLIS_CANT, "Failed to generate an authentication challenge.")
		return
	}
	for index, value := range challengeBytes {
		challengeBytes[index] = byte('a') + value%26
	}
	session.AuthChallenge = string(challengeBytes)

	writeVarnishCliResponse(session.Writer, CLIS_AUTH, session.AuthChallenge)
}

func handleVarnishCliAuthenticationAttempt(args string, session *varnishCliSession) {
	secretBytes, err := getVarnishSecret()
	if err != nil {
		log.Printf("Cannot get secret: %#v", err)
		writeVarnishCliResponse(session.Writer, CLIS_CANT, "Secret not available.")
		return
	}
	handleVarnishCliAuthenticationAttemptInternal(args, session, secretBytes)
}

func handleVarnishCliAuthenticationAttemptInternal(args string, session *varnishCliSession, secretBytes []byte) {

	if len(session.AuthChallenge) == 0 {
		writeVarnishCliResponse(session.Writer, CLIS_CANT, "Authentication challenge not initialised.")
		return
	}

	hash := sha256.New()
	hash.Write([]byte(session.AuthChallenge + "\n"))
	hash.Write(secretBytes)
	hash.Write([]byte(session.AuthChallenge + "\n"))

	expectedAuthResponse := hex.EncodeToString(hash.Sum(nil))
	log.Printf("expectedAuthResponse: %s", expectedAuthResponse)

	// TODO allow whitespace-trimmed and case-insensitive compare of hex
	if strings.ToLower(args) == expectedAuthResponse {
		session.HasAuthenticated = true
		writeVarnishCliBanner(session.Writer)
	} else {
		log.Print("Failed to authenticate")
		writeVarnishCliAuthenticationChallenge(session)
	}
}
