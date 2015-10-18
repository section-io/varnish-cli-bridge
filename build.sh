#!/bin/sh

CGO_ENABLED=0 GOOS=linux go build -a -tags netgo github.com/section-io/varnish-cli-bridge

#CGO_ENABLED=0 GOOS=windows go build -a -tags netgo github.com/section-io/varnish-cli-bridge

file varnish-cli-bridge | grep --color 'statically linked' || echo 'Failed to build static-linked binary.'
