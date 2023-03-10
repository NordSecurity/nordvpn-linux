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
	return m.inner.RoundTrip(req)
}

func (m *HTTPReTransport) NotifyConnect(events.DataConnect) error {
	m.inner = m.createFn()
	return nil
}
