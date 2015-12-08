package main // import "github.com/section-io/varnish-cli-bridge"

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type varnishCliSession struct {
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

	// eg "https://aperture.section.io/api/v1/account/1/application/1/state"
	sectionioApiEndpoint string
	sectionioUsername    string
	sectionioPassword    string
	sectionioProxyName   = "varnish"

	version    string
	commitHash string
	userAgent  string
)

func configure() {
	const cliEnvKeyPrefix = "VARNISH_CLI_BRIDGE_"
	const sectionioEnvKeyPrefix = "SECTION_IO_"

	userAgent = fmt.Sprintf("section.io varnish-cli-bridge, version %s, commit %s", version, commitHash)
	log.Printf("varinsh-cli-bridge Version: %s, Commit: %s", version, commitHash)

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
	sectionioPassword = os.Getenv(sectionioEnvKeyPrefix + "PASSWORD")
	if sectionioPassword == "" {
		log.Fatal(sectionioEnvKeyPrefix + "PASSWORD environment variable is required.")
	}

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
	statusLine := fmt.Sprintf("%3d %-8d\n", status, responseLength)
	response := statusLine + body
	buffer := []byte(response + "\n")

	_, err := writer.Write(buffer)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Sent response %#v", response)
}

func handleRequest(requestLine string, session *varnishCliSession) {
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
		handleVarnishCliAuthenticationAttempt(commandAndArgs[1], session)
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
	case "param.show":
		handleVarnishCliParamShowRequest(commandAndArgs[1], session.Writer)
		return
	case "ban":
		handleVarnishCliBanRequest(varnishQuoteArgs(commandAndArgs[1:]), session.Writer)
		return
	case "ban.url":
		handleVarnishCliBanRequest("req.url ~ "+commandAndArgs[1], session.Writer)
		return
	case "vcl.inline":
		handleVarnishCliVclInline(commandAndArgs[1], commandAndArgs[2], session.Writer)
		return
	case "vcl.use":
		handleVarnishCliVclUse(commandAndArgs[1], session.Writer)
		return

	}

	log.Printf("Unrecognised command '%s'.", command)
	writeVarnishCliResponse(session.Writer, CLIS_UNIMPL, "Unimplemented")
}

func handleConnection(connection net.Conn) {
	defer connection.Close()
	scanner := bufio.NewScanner(connection)

	session := &varnishCliSession{connection, false, ""}

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
