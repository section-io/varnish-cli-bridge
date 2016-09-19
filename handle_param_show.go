package main

import (
	"fmt"
	"io"
)

func handleVarnishCliParamShowRequest(arg string, writer io.Writer) {
	switch arg {
	case "esi_syntax":
		writeVarnishCliResponse(writer, CLIS_OK, `esi_syntax                  2 [bitmap]
                            Default is 0
                            Bitmap controlling ESI parsing code:
                              0x00000001 - Don't check if it looks like XML
                              0x00000002 - Ignore non-esi elements
                              0x00000004 - Emit parsing debug records
                              0x00000008 - Force-split parser input
                            (debugging)
                            Use 0x notation and do the bitor in your head :-)
`)

	case `cli_buffer`:
		if varnishVersion == "4.0" {
			writeVarnishCliResponse(writer, CLIS_OK, `cli_buffer
        Value is: 32k [bytes]
        Default is: 8k
        Minimum is: 4k

        Size of buffer for CLI command input.
        You may need to increase this if you have big VCL files and use
        the vcl.inline CLI command.
        NB: Must be specified with -p to have effect.
`)
		} else {
			writeVarnishCliResponse(writer, CLIS_OK, `cli_buffer                  32768 [bytes]
                            Default is 8192
                            Size of buffer for CLI input.
                            You may need to increase this if you have big VCL
                            files and use the vcl.inline CLI command.
                            NB: Must be specified with -p to have effect.
`)
		}
	default:
		writeVarnishCliResponse(writer, CLIS_PARAM,
			fmt.Sprintf("Unknown parameter \"%s\".", arg))
	}
}
