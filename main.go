package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
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

var listenAddress = ":6082"
var apiEndpoint = "http://httpbin.org/"
var secretFile = "/etc/varnish/secret"

func configure() {
	const envKeyPrefix = "VARNISH_CLI_BRIDGE_" // maybe allow override via command-line?

	envListenAddress := os.Getenv(envKeyPrefix + "LISTEN_ADDRESS")
	if envListenAddress != "" {
		listenAddress = envListenAddress
	}
	flag.StringVar(&listenAddress, "listen-address", listenAddress,
		"Address and port to listen for inbound Varnish CLI connections.")

	envSecretFile := os.Getenv(envKeyPrefix + "SECRET_FILE")
	if envSecretFile != "" {
		secretFile = envSecretFile
	}
	flag.StringVar(&secretFile, "secret-file", secretFile,
		"Path to file containing the Varnish CLI authentication secret.")

	help := flag.Bool("help", false, "Display this help.")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("Using Varnish CLI secret file '%s'.", secretFile)
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

	bytesWritten, err := writer.Write(buffer)
	if err != nil {
		log.Panic(err)
	} else {
		log.Printf("Wrote %d bytes", bytesWritten)
	}
}

const randomChallenge = "abcdefghijabcdefghijabcdefghijkl" // TODO randomise on each attempt

func writeVarnishCliAuthenticationChallenge(writer io.Writer) {
	writeVarnishCliResponse(writer, CLIS_AUTH, randomChallenge)
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
		writeVarnishCliResponse(writer, CLIS_OK, "Welcome")
	} else {
		writeVarnishCliAuthenticationChallenge(writer)
	}

}

func handleVarnishCliPingRequest(writer io.Writer) {
	response := fmt.Sprintf("PONG %d 1.0", time.Now().Unix())
	writeVarnishCliResponse(writer, CLIS_OK, response)
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
	case "auth":
		handleVarnishCliAuthenticationAttempt(commandAndArgs[1], writer)
		return
	case "ping":
		handleVarnishCliPingRequest(writer)
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

	response, err := http.Get(apiEndpoint)
	if err != nil {
		log.Print(err)
		return
	}
	defer response.Body.Close()
	_, err = ioutil.ReadAll(response.Body)
}
