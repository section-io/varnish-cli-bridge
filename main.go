package main

import (
	"fmt"
	"net/http"
	//"io"
	"io/ioutil"
	"log"
	"net"
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
			log.Fatal(err)
		}
		go handleConnection(connection)
	}
	err = listener.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func writeVarnishCliResponse(connection net.Conn, status int, body string) error {
	responseLength := len(body)
	statusLine := fmt.Sprintf("%3d %8d\n", status, responseLength)
	buffer := []byte(statusLine + body + "\n")

	bytesWritten, err := connection.Write(buffer)
	if err != nil {
		log.Print(err)
	} else {
		log.Printf("Wrote %d bytes", bytesWritten)
	}

	return err
}

func writeVarnishCliAuthenticationChallenge(connection net.Conn) error {
	randomChallenge := "abcdefghijabcdefghijabcdefghijkl" // TODO randomise
	err := writeVarnishCliResponse(connection, 107, randomChallenge)
	return err
}

func readVarnishCliAuthenticationAttempt(connection net.Conn) error {
	buffer := make([]byte, 512)
	bytesRead, err := connection.Read(buffer)
	if err != nil {
		log.Print(err)
	}
	log.Printf("read %d bytes", bytesRead)

	// TODO verify secret

}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	err := writeVarnishCliAuthenticationChallenge(connection)
	if err != nil {
		log.Print(err)
	}

	err = writeVarnishCliResponse(connection, 200, "Welcome")
	if err != nil {
		log.Print(err)
	}

	err = readVarnishCliAuthenticationAttempt(connection)
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
