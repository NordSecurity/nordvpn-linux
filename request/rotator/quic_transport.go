package rotator

import (
	"net/http"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// QuicTransport is modified/enhanced RoundTripper
type QuicTransport struct {
	inner    http.RoundTripper
	createFn func() http.RoundTripper
}

func NewQuicTransport(fn func() http.RoundTripper) *QuicTransport {
	return &QuicTransport{
		inner:    fn(),
		createFn: fn,
	}
}

func (m *QuicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := m.inner.RoundTrip(req)
	if err != nil &&
		(strings.Contains(err.Error(), "Application error 0x100") ||
			strings.Contains(err.Error(), "no recent network activity") ||
			strings.Contains(err.Error(), "Timeout exceeded while awaiting headers")) {
		// connection closed, need to reconnect
		m.inner = m.createFn()
		resp, err = m.inner.RoundTrip(req)
	}
	return resp, err
}

func (m *QuicTransport) NotifyConnect(events.DataConnect) error {
	m.inner = m.createFn()
	return nil
}
