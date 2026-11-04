package request

import (
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/log"
)

const NordAgentHeader = "X-Nord-Agent"

type NordAgentRoundTripper struct {
	inner http.RoundTripper
	value string
}

func NewNordAgentRoundTripper(inner http.RoundTripper, value string) *NordAgentRoundTripper {
	if inner == nil {
		log.Warn("nil RoundTripper passed, falling back to default transport")
		inner = http.DefaultTransport
	}
	return &NordAgentRoundTripper{
		inner: inner,
		value: value,
	}
}

func (rt *NordAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.value == "" {
		return rt.inner.RoundTrip(req)
	}

	reqWithAgent := req.Clone(req.Context())
	reqWithAgent.Header.Set(NordAgentHeader, rt.value)
	return rt.inner.RoundTrip(reqWithAgent)
}
