package main // import "github.com/section-io/varnish-cli-bridge"

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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

	// eg "https://aperture.section.io/api/v1/account/1/application/1/state"
	sectionioApiEndpoint string
	sectionioUsername    string
	sectionioPassword    string
	sectionioProxyName   = "varnish"
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

	envProxyName := os.Getenv(sectionioEnvKeyPrefix + "PROXY_NAME")
	if envProxyName != "" {
		sectionioProxyName = envProxyName
	}
	flag.StringVar(&sectionioProxyName, "proxy-name", sectionioProxyName,
		"The section.io Varnish proxy name to target for API requests.")

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
	if sectionioProxyName == "" {
		log.Fatal("section.io proxy name is required.")
	}

	log.Printf("Using Varnish CLI secret file '%s'.", secretFile)
	log.Printf("Using API endpoint '%s'.", sectionioApiEndpoint)
	log.Printf("Using API proxy name '%s'.", sectionioProxyName)
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
	statusLine := fmt.Sprintf("%3d %8d\n", status, responseLength)
	buffer := []byte(statusLine + body + "\n")

	_, err := writer.Write(buffer)
	if err != nil {
		log.Panic(err)
	}
}

const randomChallenge = "abcdefghijabcdefghijabcdefghijkl" // TODO randomise on each attempt

func writeVarnishCliAuthenticationChallenge(writer io.Writer) {
	writeVarnishCliResponse(writer, CLIS_AUTH, randomChallenge)
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

func handleVarnishCliAuthenticationAttempt(args string, writer io.Writer) {

	log.Printf("Auth attempt '%s'", args)

	file, err := os.Open(secretFile)
	if err != nil {
		log.Panicf("Failed to open secret file '%s':\n%v", secretFile, err)
	}
	defer file.Close()
	secretBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panicf("Failed to read secret file '%s':\n%v", secretFile, err)
	}
	// TODO close file sooner, ie here

	hash := sha256.New()
	hash.Write([]byte(randomChallenge + "\n"))
	hash.Write(secretBytes)
	hash.Write([]byte(randomChallenge + "\n"))

	expectedAuthResponse := hex.EncodeToString(hash.Sum(nil))

	// TODO allow whitespace-trimmed and case-insensitive compare of hex
	if strings.ToLower(args) == expectedAuthResponse {
		writeVarnishCliBanner(writer)
	} else {
		writeVarnishCliAuthenticationChallenge(writer)
	}

}

func handleVarnishCliPingRequest(writer io.Writer) {
	response := fmt.Sprintf("PONG %d 1.0", time.Now().Unix())
	writeVarnishCliResponse(writer, CLIS_OK, response)
}

func handleVarnishCliBanRequest(args string, writer io.Writer) {

	postValues := &JsonBanRequest{
		Proxy: sectionioProxyName,
		Ban:   args,
	}
	requestBody, err := json.Marshal(postValues)
	if err != nil {
		log.Printf("Error serialising ban request to JSON: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to serialise the API request.")
		return
	}

	request, err := http.NewRequest("POST", sectionioApiEndpoint, bytes.NewReader(requestBody))
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
	defer response.Body.Close()
	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading ban API response: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to parse the API response.")
		return
	}

	if response.StatusCode == 200 {
		// TODO parse response body as JSON, expect:
		// {"success":true,"description":"Ban applied"}
		writeVarnishCliResponse(writer, CLIS_OK, "Ban forwarded.")
		return
	}

	log.Printf("Unexpected API response status: %d, body: %v",
		response.StatusCode,
		string(responseBodyBytes))

	writeVarnishCliResponse(writer, CLIS_CANT,
		fmt.Sprintf("API responded with status %d.", response.StatusCode))
}

func handleRequest(requestLine string, writer io.Writer) {
	requestLine = strings.TrimLeft(requestLine, " ")
	commandAndArgs := strings.SplitN(requestLine, " ", 2)
	command := commandAndArgs[0]
	if command != strings.ToLower(command) {
		writeVarnishCliResponse(writer, CLIS_UNKNOWN, "all commands are in lower-case.")
		return
	}

	switch command {
	case "banner":
		writeVarnishCliBanner(writer)
		return
	case "auth":
		handleVarnishCliAuthenticationAttempt(commandAndArgs[1], writer)
		return
	case "ping":
		handleVarnishCliPingRequest(writer)
		return
	case "ban":
		handleVarnishCliBanRequest(commandAndArgs[1], writer)
		return
	case "ban.url":
		handleVarnishCliBanRequest("req.url ~ "+commandAndArgs[1], writer)
		return
	}

	log.Printf("Unrecognised command '%s'.", command)
	writeVarnishCliResponse(writer, CLIS_UNIMPL, "Unimplemented")
}

func handleConnection(connection net.Conn) {
	defer connection.Close()
	scanner := bufio.NewScanner(connection)

	writeVarnishCliAuthenticationChallenge(connection)

	for {
		if scanner.Scan() {
			handleRequest(scanner.Text(), connection)
		} else {
			break
		}
	}

	err := scanner.Err()
	if err != nil {
		log.Print(err)
	}
}
