#!/bin/sh -e

go test github.com/section-io/varnish-cli-bridge || { echo 'Failed to compile and test.' ; exit 1; }

CGO_ENABLED=0 GOOS=linux go build -a -tags netgo github.com/section-io/varnish-cli-bridge || { echo 'Failed to build.' ; exit 1; }

#CGO_ENABLED=0 GOOS=windows go build -a -tags netgo github.com/section-io/varnish-cli-bridge

file varnish-cli-bridge | grep --color 'statically linked' || { echo 'Failed to produce static-linked binary.' ; exit 1; }
