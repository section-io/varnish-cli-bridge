package main

import (
	"fmt"
	"strconv"
	"strings"
)

func readSingleToken(s string) (token string, tail string, err error) {
	token = ""
	i, n, quoted := 0, len(s), false

	loop:
	for ; i < n; i++ {
		char := s[i:i+1]

		switch char {
		case `"`:
			quoted =! quoted
		case ` `:
			if (!quoted) {
				i--; //Include in tail
				break loop
			}
			token += char
		case `\`:
			i++
			if (i<n) {
				char := s[i:i+1]

				backSlashloop:
				switch char {
				case `n`:
					token += "\n"
					break backSlashloop
				case `r`:
					token += "\r"
					break backSlashloop
				case `t`:
					token += "\t"
					break backSlashloop
				case `"`:
					token += "\""
					break backSlashloop
				case `\`:
					token += "\\"
					break backSlashloop
				case `0`,`1`,`2`,`3`,`4`,`5`,`6`,`7`: // Octal syntax \nnn
					// TODO: Handle malformed tokens
					var sscanfChar byte
					res, _ :=fmt.Sscanf(s[i:], "%03o", &sscanfChar)
					if (res == 1) {
						i += 2
						token += string(sscanfChar)
					} else {
						token += `\` + char
					}
					break backSlashloop
				case `x`: // Hexidecimcal syntax \xnn
					// TODO: Handle malformed tokens
					var sscanfChar byte
					res, _ := fmt.Sscanf(s[i+1:], "%02x", &sscanfChar)
					if (res == 1) {
						i += 2
						token +=  string(sscanfChar)
					} else {
						token += `\` + char
					}
					break backSlashloop
				default:
					//Unsupported escape
					token += `\` + char
					break backSlashloop
				}
			} else {
				token += `\`
			}
		default:
			token += char
		}
	}

	if (i >= n) {
		tail = ""
	} else {
		tail = s[i+1:]
	}
	return
}

func tokenizeRequest(input string) (result []string) {
	// We must decode the command
	// Requests are whitespace separated tokens terminated by a newline (NL) character.
	// - https://www.varnish-cache.org/trac/wiki/ManagementPort
	// - https://www.varnish-cache.org/docs/trunk/reference/varnish-cli.html - SH style syntax
	// - https://github.com/varnish/Varnish-Cache/blob/master/lib/libvarnish/vav.c

	//Naive implementation
	result = make([]string, 0)
	for len(input) > 0 {
		token, tail, _ := readSingleToken(input)
		//log.Printf("tokenizeRequest token: %s ", token)
		input = tail
		result = append(result, token)
	}
	return
}

func varnishQuoteArgs(input []string ) (result string) {
	quoted := []string{}

	for _, i := range input {
		quoted=append(quoted,strconv.Quote(i))
	}

	result = strings.Join(quoted, " ")
	return
}
