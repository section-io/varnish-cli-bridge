package main

// VarnishCliResponseStatus is the varnish status code
type VarnishCliResponseStatus int

// VarnishCliResponseStatus codes
const (
	CLIS_SYNTAX    VarnishCliResponseStatus = 100
	CLIS_UNKNOWN                            = 101
	CLIS_UNIMPL                             = 102
	CLIS_TOOFEW                             = 104
	CLIS_TOOMANY                            = 105
	CLIS_PARAM                              = 106
	CLIS_AUTH                               = 107
	CLIS_OK                                 = 200
	CLIS_TRUNCATED                          = 201
	CLIS_CANT                               = 300
	CLIS_COMMS                              = 400
	CLIS_CLOSE                              = 500
)
