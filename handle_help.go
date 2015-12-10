package main

import (
	"fmt"
	"io"
)

func handleVarnishCliHelpRequest(arg string, writer io.Writer) {

	if arg == "" {
		writeVarnishCliResponse(writer, CLIS_OK, `help [command]
ping [timestamp]
auth response
`+ /*
			quit
			*/
			`banner
`+ /*
			status
			start
			stop
			vcl.load <configname> <filename>
			*/
			`vcl.inline <configname> <quoted_VCLstring>
vcl.use <configname>
`+ /*
			vcl.discard <configname>
			vcl.list
			vcl.show <configname>
			*/
			`param.show [-l] [<param>]
`+ /*
			param.set <param> <value>
			panic.show
			panic.clear
			storage.list
			backend.list
			backend.set_health matcher state
			*/
			`ban.url <regexp>
ban <field> <operator> <arg> [&& <field> <oper> <arg>]...
`+ /*
			ban.list
			*/
			``)
		return
	}

	switch arg {
	default:
		writeVarnishCliResponse(writer, CLIS_PARAM,
			fmt.Sprintf("Unknown parameter \"%s\".", arg))
	}

}
