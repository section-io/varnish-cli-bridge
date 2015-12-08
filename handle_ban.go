package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type jsonBanRequest struct {
	Proxy string `json:"proxy"`
	Ban   string `json:"ban"`
}

func handleVarnishCliBanRequest(args string, writer io.Writer) {

	requestURL, err := url.Parse(sectionioApiEndpoint + "state")
	if err != nil {
		log.Printf("Error parsing url: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to parse API URL.")
		return
	}
	q := requestURL.Query()
	q.Set("banExpression", args)
	requestURL.RawQuery = q.Encode()

	request, err := http.NewRequest("POST", requestURL.String(), nil)

	if err != nil {
		log.Printf("Error composing ban request: %v", err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to compose the API request.")
		return
	}

	request.Header.Set("User-Agent", userAgent)
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
