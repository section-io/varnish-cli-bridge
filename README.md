# Varnish CLI Bridge

This program will listen for incoming connections from `varnishadm` and other
Varnish Cache CLI clients and relay `ban` requests to all Varnish Cache
instances for a given section.io-hosted Varnish by forwarding the request
to the section.io API.

## Required Configuration

The Varnish CLI Bridge has some mandatory configuration requirements:

* API endpoint: The absolute section.io application state URL with account
and application IDs. Can be configured via the `SECTION_IO_API_ENDPOINT`
environment variable or the `-api-endpoint` command line argument, with the
latter taking precedence. The URL must contain the account ID and application
ID to target. To support multiple section.io applications, run multiple
instances of the bridge. Example URL for account `1`, application `2`:
https://aperture.section.io/api/v1/account/1/application/2/state

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
Defaults to `/etc/varnish/secret` if not provided. File contents must be
identical to the secret file passed to `varnishadm` via its `-S` argument.

* Listen address: The TCP port and optional interface IP address on which the
Varnish CLI Bridge should listen for incoming connections. Format is
`[IP]:PORT` and the default is `:6082` if not provided. Omitting the IP results
in binding to all interfaces (ie `INADDR_ANY`).

## Supported commands

The Varnish CLI Bridge does not implement every command yet and some are not
planned to be implemented.

### Implemented now:

* `auth`
* `ban`
* `ban.url` (via automatic rewriting to `ban`)
* `ping`

### May be implemented later (in no particular order):

* `banner`
* `backend.list`
* `ban.list`
* `help`
* `param.show`
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
* `vcl.inline`
* `vcl.load`
* `vcl.use`

Read more about the CLI commands here:
https://www.varnish-cache.org/docs/trunk/reference/varnish-cli.html