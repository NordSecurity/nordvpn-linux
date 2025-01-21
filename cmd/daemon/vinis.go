//go:build vinis

package main

import (
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/request/vinis"
)

func getPinningTransport(inner http.RoundTripper) http.RoundTripper {
	return vinis.New(inner)
}
