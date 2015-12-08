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

func jsonPost(writer io.Writer, url string, activityFriendlyName string, postValues interface{}) (result map[string]interface{}, err error) {

	log.Printf("postValues: %+v", postValues)

	requestBody, err := json.Marshal(postValues)
	if err != nil {
		log.Printf("Error serialising %s request to JSON: %v", activityFriendlyName, err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to serialise the "+activityFriendlyName+" API request.")
		return nil, err
	}

	log.Printf("url: %s", url)
	log.Printf("requestBody: %s", string(requestBody))
	request, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	if err != nil {
		log.Printf("Error composing %s request: %v", activityFriendlyName, err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to compose the "+activityFriendlyName+" API request.")
		return nil, err
	}

	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(sectionioUsername, sectionioPassword)
	//log.Printf("sectionioUsername, sectionioPassword %s %s", sectionioUsername, sectionioPassword)

	localResponse, err := httpClient.Do(request)
	if err != nil {
		log.Printf("Error posting %s request: %v", activityFriendlyName, err)
		writeVarnishCliResponse(writer, CLIS_CANT, "Failed to post the "+activityFriendlyName+".")
		return nil, err
	}

	var responseBodyText string
	var responseBodyBytes []byte
	func() {
		defer localResponse.Body.Close()
		responseBodyBytes, err = ioutil.ReadAll(localResponse.Body)
		if err != nil {
			log.Printf("Error reading %s API response: %v", activityFriendlyName, err)
			writeVarnishCliResponse(writer, CLIS_CANT, "Failed to parse the "+activityFriendlyName+" API response.")
			return
		}
		responseBodyText = string(responseBodyBytes)
	}()
	if err != nil {
		return nil, err
	}

	log.Printf("responseBodyText: %s", responseBodyText)

	if localResponse.StatusCode != 200 {
		// TODO parse response body as JSON, expect:
		writeVarnishCliResponse(writer, CLIS_OK, "Call for "+activityFriendlyName+" failed.")
		log.Printf("Unexpected API response status: %d, body: %v", localResponse.StatusCode, responseBodyText)
		return nil, fmt.Errorf("Unexpected HTTP status code: %d", localResponse.StatusCode)
	}

	err = json.Unmarshal(responseBodyBytes, &result)
	if err != nil {
		log.Printf("Unable to decode JSON API response: %s", responseBodyText)
		return nil, err
	}
	return
}
