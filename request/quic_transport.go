package request

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/quic-go/quic-go/http3"
)

// QuicTransport is modified/enhanced RoundTripper.
// Thread safe.
type QuicTransport struct {
	inner    *atomic.Pointer[http3.RoundTripper]
	createFn func() *http3.RoundTripper
}

func NewQuicTransport(fn func() *http3.RoundTripper) *QuicTransport {
	p := &atomic.Pointer[http3.RoundTripper]{}
	p.Store(fn())
	return &QuicTransport{
		inner:    p,
		createFn: fn,
	}
}

func (m *QuicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.ProtoMajor = 3
	req.ProtoMinor = 0
	req.Proto = "HTTP/3"
	inner := m.inner.Load()
	resp, err := inner.RoundTrip(req)
	if err != nil &&
		(strings.Contains(err.Error(), "Application error 0x100") ||
			strings.Contains(err.Error(), "no recent network activity") ||
			strings.Contains(err.Error(), "Timeout exceeded while awaiting headers")) {
		// connection closed, need to reconnect
		inner := m.createFn()
		m.inner.Store(inner)
		resp, err = inner.RoundTrip(req)
	}
	return resp, err
}

func (m *QuicTransport) NotifyConnect(events.DataConnect) error {
	m.inner.Store(m.createFn())
	return nil
}
