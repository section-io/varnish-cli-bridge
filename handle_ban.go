package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type jsonBanRequest struct {
	Proxy string `json:"proxy"`
	Ban   string `json:"ban"`
}

func handleVarnishCliBanRequest(args string, writer io.Writer) {

	postValues := &jsonBanRequest{
		Proxy: sectionioProxyName,
		Ban:   args,
	}
	requestBody, err := json.Marshal(postValues)
	if err != nil {
		log.Printf("Error serialising ban request to JSON: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to serialise the API request.")
		return
	}

	log.Printf("requestBody: %s", requestBody)
	request, err := http.NewRequest("POST", sectionioApiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		log.Printf("Error composing ban request: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to compose the API request.")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(sectionioUsername, sectionioPassword)
	log.Printf("sectionioUsername, sectionioPassword %s %s", sectionioUsername, sectionioPassword)

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

	log.Printf("responseBodyText: %s", responseBodyText)

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
