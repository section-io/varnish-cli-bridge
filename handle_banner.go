package main

import (
	"fmt"
	"io"
)

func writeVarnishCliBanner(writer io.Writer) {
	// emulate the normal banner Varnish for client-compatibility.
	bannerFormat := `-----------------------------
Varnish Cache CLI Bridge
-----------------------------
https://github.com/section-io/varnish-cli-bridge
%s

Type 'help' for command list.
Type 'quit' to close CLI session.`

	writeVarnishCliResponse(writer, CLIS_OK, fmt.Sprintf(bannerFormat, bannerVarnishVersion))
}
