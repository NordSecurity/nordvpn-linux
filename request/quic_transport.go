package request

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/quic-go/quic-go/http3"
)

// QuicTransport is modified/enhanced RoundTripper.
// Thread safe.
type QuicTransport struct {
	// for protects the access to inner and to shouldRecreate.
	// This will allow multiple requests in the same time for inner, but only one can recreate it
	mu sync.RWMutex
	// when executing the requests set this to true on error and when inner is recreated is set to false.
	// This is used to ensure only one failed request will recreate inner value on error
	shouldRecreate atomic.Bool
	inner          *http3.RoundTripper
	createFn       func() *http3.RoundTripper
}

func NewQuicTransport(fn func() *http3.RoundTripper) *QuicTransport {
	return &QuicTransport{
		mu:       sync.RWMutex{},
		inner:    fn(),
		createFn: fn,
	}
}

func (m *QuicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.ProtoMajor = 3
	req.ProtoMinor = 0
	req.Proto = "HTTP/3"
	resp, err := m.executeRequest(req)
	return resp, err
}

func (m *QuicTransport) NotifyConnect(events.DataConnect) error {
	m.recreateRoundTrip(true)
	return nil
}

func (m *QuicTransport) executeRequest(req *http.Request) (*http.Response, error) {
	m.mu.RLock()
	inner := m.inner
	m.mu.RUnlock()
	response, err := inner.RoundTrip(req)

	// check the errors if inner needs to be recreated
	recreate := shouldRecreate(err)

	// mark that inner needs to be recreated while holding the read lock to be sure shouldRecreate is not replaced before recreation
	m.shouldRecreate.Store(recreate)

	if recreate {
		log.Println(internal.InfoPrefix, "recreate RoundTripper on error", err)
		// recreate the inner and retry one more time to execute the same request
		m.recreateRoundTrip(false)
		m.mu.RLock()
		inner := m.inner
		m.mu.RUnlock()
		response, err = inner.RoundTrip(req)
	}

	return response, err
}

func shouldRecreate(err error) bool {
	return err != nil &&
		(strings.Contains(err.Error(), "Application error 0x100") ||
			strings.Contains(err.Error(), "no recent network activity") ||
			strings.Contains(err.Error(), "Timeout exceeded while awaiting headers"))
}

func (m *QuicTransport) recreateRoundTrip(forceRecreate bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// if shouldRecreate is still true then create inner and set shouldRecreate to false
	if forceRecreate || m.shouldRecreate.Load() {
		m.shouldRecreate.Store(false)
		m.inner.Close()
		m.inner = m.createFn()
	}
}
