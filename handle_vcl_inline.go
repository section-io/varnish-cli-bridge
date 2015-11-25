package main

import "io"

var inlineConfigName string
var inlineQuotedVCLstring string

func handleVarnishCliVclInline(configname string, quotedVCLstring string, writer io.Writer) {
	inlineConfigName = configname
	inlineQuotedVCLstring = quotedVCLstring
	writeVarnishCliResponse(writer, CLIS_OK, `VCL.compiled`)
}
