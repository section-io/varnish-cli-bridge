package main // import "github.com/section-io/varnish-cli-bridge"

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type VarnishCliResponseStatus int

const (
	CLIS_SYNTAX    VarnishCliResponseStatus = 100
	CLIS_UNKNOWN                            = 101
	CLIS_UNIMPL                             = 102
	CLIS_TOOFEW                             = 104
	CLIS_TOOMANY                            = 105
	CLIS_PARAM                              = 106
	CLIS_AUTH                               = 107
	CLIS_OK                                 = 200
	CLIS_TRUNCATED                          = 201
	CLIS_CANT                               = 300
	CLIS_COMMS                              = 400
	CLIS_CLOSE                              = 500
)

type JsonBanRequest struct {
	Proxy string `json:"proxy"`
	Ban   string `json:"ban"`
}

type VarnishCliSession struct {
	Writer           io.Writer
	HasAuthenticated bool
	AuthChallenge    string
}

var (
	sectionioApiEndpointRx = regexp.MustCompile("^https?://")
	httpClient             = &http.Client{
		Timeout: time.Minute,
	}

	listenAddress = ":6082"
	secretFile    = "/etc/varnish/secret"

	// nexcess/magento-turpentine checks the banner text to determine the ban syntax
	// TODO make version configurable, or query from section.io API
	bannerVarnishVersion = "varnish-3.0.0 revision 0000000"

	// eg "https://aperture.section.io/api/v1/account/1/application/1/environment/Production/proxy/varnish/state"
	sectionioApiEndpoint string
	sectionioUsername    string
	sectionioPassword    string
)

func configure() {
	const cliEnvKeyPrefix = "VARNISH_CLI_BRIDGE_"
	const sectionioEnvKeyPrefix = "SECTION_IO_"

	envListenAddress := os.Getenv(cliEnvKeyPrefix + "LISTEN_ADDRESS")
	if envListenAddress != "" {
		listenAddress = envListenAddress
	}
	flag.StringVar(&listenAddress, "listen-address", listenAddress,
		"Address and port to listen for inbound Varnish CLI connections.")

	envSecretFile := os.Getenv(cliEnvKeyPrefix + "SECRET_FILE")
	if envSecretFile != "" {
		secretFile = envSecretFile
	}
	flag.StringVar(&secretFile, "secret-file", secretFile,
		"Path to file containing the Varnish CLI authentication secret.")

	envApiEndpoint := os.Getenv(sectionioEnvKeyPrefix + "API_ENDPOINT")
	if envApiEndpoint != "" {
		if sectionioApiEndpointRx.MatchString(envApiEndpoint) {
			sectionioApiEndpoint = envApiEndpoint
		} else {
			log.Fatal(sectionioEnvKeyPrefix + "API_ENDPOINT variable is invalid.")
		}
	}
	flag.StringVar(&sectionioApiEndpoint, "api-endpoint", sectionioApiEndpoint,
		"The absolute section.io application state POST url with account and application IDs.")

	envUsername := os.Getenv(sectionioEnvKeyPrefix + "USERNAME")
	if envUsername != "" {
		sectionioUsername = envUsername
	}
	flag.StringVar(&sectionioUsername, "username", "",
		"The section.io username to use for API requests.")

	sectionioPassword = os.Getenv(sectionioEnvKeyPrefix + "PASSWORD")
	if sectionioPassword == "" {
		log.Fatal(sectionioEnvKeyPrefix + "PASSWORD environment variable is required.")
	}

	help := flag.Bool("help", false, "Display this help.")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !sectionioApiEndpointRx.MatchString(sectionioApiEndpoint) {
		log.Fatal("api-endpoint argument is invalid.")
	}
	if sectionioUsername == "" {
		log.Fatal("section.io username is required.")
	}

	log.Printf("Using Varnish CLI secret file '%s'.", secretFile)
	log.Printf("Using API endpoint '%s'.", sectionioApiEndpoint)
	log.Printf("Using API username '%s'.", sectionioUsername)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configure()

	log.Printf("Listening on '%s'.", listenAddress)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err) // panic instead?
		}
		go handleConnection(connection)
	}
	err = listener.Close() // defer instead
	if err != nil {
		log.Fatal(err)
	}
}

func writeVarnishCliResponse(writer io.Writer, status VarnishCliResponseStatus, body string) {
	responseLength := len(body) // NOTE len() returns byte count, not character count
	statusLine := fmt.Sprintf("%3d %-8d\n", status, responseLength)
	response := statusLine + body
	buffer := []byte(response + "\n")

	_, err := writer.Write(buffer)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Sent response %#v", response)
}

func writeVarnishCliAuthenticationChallenge(session *VarnishCliSession) {
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

func writeVarnishCliBanner(writer io.Writer) {
	// emulate the normal banner Varnish for client-compatibility.
	bannerFormat := `-----------------------------
Varnish Cache CLI Bridge
-----------------------------
https://github.com/section-io/varnish-cli-bridge
%s

Type 'help' for command list.
Type 'quit' to close CLI session.`

	writeVarnishCliResponse(writer, CLIS_OK, fmt.Sprintf(bannerFormat, bannerVarnishVersion))
}

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

func handleVarnishCliAuthenticationAttempt(args string, session *VarnishCliSession, secretBytes []byte) {

	if len(session.AuthChallenge) == 0 {
		writeVarnishCliResponse(session.Writer, CLIS_CANT, "Authentication challenge not initialised.")
		return
	}

	hash := sha256.New()
	hash.Write([]byte(session.AuthChallenge + "\n"))
	hash.Write(secretBytes)
	hash.Write([]byte(session.AuthChallenge + "\n"))

	expectedAuthResponse := hex.EncodeToString(hash.Sum(nil))

	// TODO allow whitespace-trimmed and case-insensitive compare of hex
	if strings.ToLower(args) == expectedAuthResponse {
		session.HasAuthenticated = true
		writeVarnishCliBanner(session.Writer)
	} else {
		writeVarnishCliAuthenticationChallenge(session)
	}

}

func handleVarnishCliPingRequest(writer io.Writer) {
	response := fmt.Sprintf("PONG %d 1.0", time.Now().Unix())
	writeVarnishCliResponse(writer, CLIS_OK, response)
}

func handleVarnishCliBanRequest(args string, writer io.Writer) {

	requestURL, err := url.Parse(sectionioApiEndpoint)
	if err != nil {
		log.Printf("Error parsing url: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to parse API URL.")
		return
	}
	q := requestURL.Query()
	q.Set("banExpression", args)
	requestURL.RawQuery = q.Encode()

	request, err := http.NewRequest("POST", requestURL.String(), bytes.NewReader([]byte("")))
	if err != nil {
		log.Printf("Error composing ban request: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to compose the API request.")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(sectionioUsername, sectionioPassword)

	response, err := httpClient.Do(request)
	if err != nil {
		log.Printf("Error posting ban request '%s': %v", args, err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to forward the ban.")
		return
	}
	var responseBodyText string
	func() {
		defer response.Body.Close()
		responseBodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("Error reading ban API response: %v", err)
			writeVarnishCliResponse(writer, CLIS_CANT, "Failed to parse the API response.")
			return
		}
		responseBodyText = string(responseBodyBytes)
	}()

	if response.StatusCode == 200 {
		// TODO parse response body as JSON, expect:
		// {"success":true,"description":"Ban applied"}
		writeVarnishCliResponse(writer, CLIS_OK, "Ban forwarded.")
		return
	}

	log.Printf("Unexpected API response status: %d, body: %v",
		response.StatusCode,
		responseBodyText)

	writeVarnishCliResponse(writer, CLIS_CANT,
		fmt.Sprintf("API responded with status %d.", response.StatusCode))
}

func handleRequest(requestLine string, session *VarnishCliSession) {
	log.Printf("Received request %#v", requestLine)
	requestLine = strings.TrimLeft(requestLine, " ")

	commandAndArgs := tokenizeRequest(requestLine)
	command := commandAndArgs[0]
	if command != strings.ToLower(command) {
		writeVarnishCliResponse(session.Writer, CLIS_UNKNOWN, "all commands are in lower-case.")
		return
	}

	switch command {
	case "banner":
		writeVarnishCliBanner(session.Writer)
		return
	case "auth":
		secretBytes, err := getVarnishSecret()
		if err != nil {
			log.Printf("Cannot get secret: %#v", err)
			writeVarnishCliResponse(session.Writer, CLIS_CANT, "Secret not available.")
			return
		}
		handleVarnishCliAuthenticationAttempt(commandAndArgs[1], session, secretBytes)
		return
	case "ping":
		handleVarnishCliPingRequest(session.Writer)
		return
	}

	if !session.HasAuthenticated {
		writeVarnishCliAuthenticationChallenge(session)
		return
	}

	switch command {
	case "ban":
		handleVarnishCliBanRequest(varnishQuoteArgs(commandAndArgs[1:]), session.Writer)
		return
	case "ban.url":
		handleVarnishCliBanRequest("req.url ~ "+commandAndArgs[1], session.Writer)
		return
	}

	log.Printf("Unrecognised command '%s'.", command)
	writeVarnishCliResponse(session.Writer, CLIS_UNIMPL, "Unimplemented")
}

func handleConnection(connection net.Conn) {
	defer connection.Close()
	scanner := bufio.NewScanner(connection)

	session := &VarnishCliSession{connection, false, ""}

	writeVarnishCliAuthenticationChallenge(session)

	for {
		if scanner.Scan() {
			handleRequest(scanner.Text(), session)
		} else {
			break
		}
	}

	err := scanner.Err()
	if err != nil {
		log.Print(err)
	}
}
