# Varnish CLI Bridge

This program will listen for incoming connections from `varnishadm` and other
Varnish Cache CLI clients and relay `ban` requests to all Varnish Cache
instances for a given section.io-hosted Varnish by forwarding the request
to the section.io API.

[![Build status image](https://travis-ci.org/section-io/varnish-cli-bridge.svg?branch=master)](https://travis-ci.org/section-io/varnish-cli-bridge)

## Required Configuration

The Varnish CLI Bridge has some mandatory configuration requirements:

* API endpoint: The absolute section.io application URL with account
and application IDs. Can be configured via the `SECTION_IO_API_ENDPOINT`
environment variable or the `-api-endpoint` command line argument, with the
latter taking precedence. The URL must contain the account ID and application
ID to target. To support multiple section.io applications, run multiple
instances of the bridge. Example URL for account `1`, application `2`:
https://aperture.section.io/api/v1/account/1/application/2/

* API username: The username with which to authenticate to the section.io API.
Can be configured via the `SECTION_IO_USERNAME` environment variable or the
`-username` command line argument, with the latter taking precedence.

* API password: The password corresponding to the username above. Can only be
configured via the `SECTION_IO_PASSWORD` environment variable. Providing
passwords on the command line is not recommended in general and not supported
by the Varnish CLI Bridge.

## Optional Configuration

The Varnish CLI Bridge also has some optional configuration that can be
specified if the defaults are not suitable:

* API proxy name: The name of the proxy in the section.io stack to target.
Can be specified via the `SECTION_IO_PROXY_NAME` environment variable or the
`-proxy-name` command line argument, with the latter taking precedence.
Defaults to `varnish` if not provided.

* Varnish CLI secret file: The path to the file containing the pre-shared
secret used to authenticate connections to the Varnish CLI Bridge.
Can be specified via the `VARNISH_CLI_BRIDGE_SECRET_FILE` environment variable
or the `-secret-file` command line argument, with the latter taking precedence.
If specified the ile contents must be identical to the secret file passed
to `varnishadm` via its `-S` argument. If left blank the bridge will not send
the authentication challenge and so the client should not specify a
secret file or value either.

* Listen address: The TCP port and optional interface IP address on which the
Varnish CLI Bridge should listen for incoming connections. Can be specified
via the `VARNISH_CLI_BRIDGE_LISTEN_ADDRESS` environment variable or the
`-listen-address` command line argument, with the latter taking precedence.
Format is `[IP]:PORT` and the default is `127.0.0.1:6082` if not provided.
Omitting the IP results in binding to all interfaces (ie `INADDR_ANY`).

* Varnish version: The protocol version to simulate. Also sets the default
version reported in the protocol banner response unless overridden.
Can be specified via the `VARNISH_CLI_BRIDGE_VARNISH_VERSION` ennvironment variable
or the `-varnish-version` command line argument, with the latter taking precedence.
The only supported values are `3.0` and `4.0`. The default value is `3.0`.

* Varnish banner: The text to include in the protocol banner response
denoting the Varnish version. Can be specified via the
`VARNISH_CLI_BRIDGE_BANNER_VERSION` ennvironment variable or the
`-banner-version` command line argument, with the latter taking precedence.
The recommended format is `varnish-[MAJOR].[MINOR].[BUILD] revision [REVISION]`
but is not enforced. The default value is `varnish-3.0.0 revision 0000000`.

## Supported commands

The Varnish CLI Bridge does not implement every command yet and some are not
planned to be implemented.

### Implemented now:

* `auth`
* `ban`
* `ban.url` (via automatic rewriting to `ban`)
* `banner`
* `help`
* `ping`
* `param.show` (currently only `esi_syntax` and `cli_buffer`)
* `vcl.inline`
* `vcl.use`

### May be implemented later (in no particular order):

* `backend.list`
* `ban.list`
* `quit`
* `status`
* `vcl.list`
* `vcl.show`

### Implementation not planned:

* `backend.set_health`
* `param.set`
* `panic.clear`
* `panic.show`
* `start`
* `stop`
* `storage.list`
* `vcl.discard`
* `vcl.load`

Read more about the CLI commands here:
https://www.varnish-cache.org/docs/trunk/reference/varnish-cli.html
