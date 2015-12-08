package main

import (
	"fmt"
	"io"
	"log"
)

type jsonUpdateVclRequest struct {
	Personality string `json:"personality"`
	Message     string `json:"message"`
	Content     string `json:"content"`
}

func handleVarnishCliVclUse(configname string, writer io.Writer) {
	if configname != inlineConfigName {
		writeVarnishCliResponse(writer, CLIS_PARAM, fmt.Sprintf(`No configuration named %s known.`, configname))
		return
	}

	//Post it to the API
	postValues := jsonUpdateVclRequest{
		Personality: "MagentoTurpentine",
		Message:     "Update from varnish-cli-bridge",
		Content:     inlineQuotedVCLstring,
	}

	response, err := jsonPost(writer, sectionioApiEndpoint+"configuration", "configuration update", postValues)

	if err != nil {
		//CLI Response already written on non-200 or other error
		return
	}

	log.Printf("Update submitted. Response message: %s", response["message"])

	//Varnishd actually returns a 200 & zero byte reponse (a problem to match since we add a trailing /n in writeVarnishCliResponse)
	writeVarnishCliResponse(writer, CLIS_OK, ``)
}
