package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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

var listenPort = 6083 // normally 6082
var apiEndpoint = "http://httpbin.org/"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//port := flag.String("port", "6082", "Default listen port")
	//flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on port %d", listenPort)

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

func writeVarnishCliAuthenticationChallenge(writer io.Writer) {
	randomChallenge := "abcdefghijabcdefghijabcdefghijkl" // TODO randomise
	writeVarnishCliResponse(writer, CLIS_AUTH, randomChallenge)
}

func handleVarnishCliAuthenticationAttempt(args string, writer io.Writer) {

	log.Printf("Auth attempt '%s'", args)

	// TODO verify secret, if auth failed, resend challenge

	writeVarnishCliResponse(writer, CLIS_OK, "Welcome")

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
