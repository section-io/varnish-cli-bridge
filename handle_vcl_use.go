package main

import (
	"fmt"
	"io"
	"os"
)

func handleVarnishCliVclUse(configname string, writer io.Writer) {
	if configname != inlineConfigName {
		writeVarnishCliResponse(writer, CLIS_PARAM, fmt.Sprintf(`No configuration named %s known.`, configname))
		return
	}

	os.Remove("./vcl.vcl")
	f, _ := os.Create("./vcl.vcl")
	f.WriteString(inlineQuotedVCLstring)
	f.Close()

	//Post it to the API

	//Varnishd actually returns a 200 & zero byte reponse (a problem to match since we add a trailing /n in writeVarnishCliResponse)
	writeVarnishCliResponse(writer, CLIS_OK, ``)
}
