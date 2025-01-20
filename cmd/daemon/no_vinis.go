//go:build !vinis

package main

import "net/http"

func getPinningTransport(inner http.RoundTripper) http.RoundTripper {
	return inner
}
