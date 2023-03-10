package request

import (
	"net"
	"net/http"
	"time"
)

const (
	DefaultTimeout   = 15 * time.Second
	TransportTimeout = 5 * time.Second
)

// StdOpt allows configuring standard library's http client.
type StdOpt func(*http.Client)

// NewStdHTTP returns standard library's http client with opts.
func NewStdHTTP(opts ...StdOpt) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: TransportTimeout}).DialContext,
			TLSHandshakeTimeout: TransportTimeout,
		},
		Timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}
