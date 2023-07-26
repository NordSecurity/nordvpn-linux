package request

import (
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// HTTPReTransport is std RoundTripper enhanced to reconnect after vpn connect is done
type HTTPReTransport struct {
	inner    http.RoundTripper
	createFn func() http.RoundTripper
}

func NewHTTPReTransport(fn func() http.RoundTripper) *HTTPReTransport {
	return &HTTPReTransport{
		inner:    fn(),
		createFn: fn,
	}
}

func (m *HTTPReTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Note: This assumes that it is using HTTP/1.1 In order to use it with other transport,
	// migrate Proto modifications to inner transport.
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Proto = "HTTP/1.1"
	return m.inner.RoundTrip(req)
}

func (m *HTTPReTransport) NotifyConnect(events.DataConnect) error {
	m.inner = m.createFn()
	return nil
}
