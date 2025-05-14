package request

import (
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// HTTPReTransport is modified/enhanced RoundTripper.
// Thread safe.
type HTTPReTransport struct {
	// for protects the access to inner and to shouldRecreate. This will allow multiple
	// requests in the same time for inner, but only one can recreate it
	mu sync.RWMutex
	// when executing the requests set this to true on error and when inner is recreated is set
	// to false. This is used to ensure only one failed request will recreate inner value on
	// error.
	inner           http.RoundTripper
	createFn        RoundTripperCreateFunc
	protoMajor      int
	protoMinor      int
	proto           string
	shouldRetryFunc ShouldRetryFunc
}

// RoundTripperCreateFunc is a function used to create a new instance of round tripper when it needs
// to be recreated.
type RoundTripperCreateFunc func() http.RoundTripper

// ShouldRetryFunc defines a function that determines whether an HTTP request should be retried
// based on the given error.
type ShouldRetryFunc func(err error) bool

// NewHTTPReTransport is a default constructor function for a HTTPReTransport.
func NewHTTPReTransport(
	protoMajor int,
	protoMinor int,
	proto string,
	createFn RoundTripperCreateFunc,
	shouldRetryFn ShouldRetryFunc,
) *HTTPReTransport {
	return &HTTPReTransport{
		mu:              sync.RWMutex{},
		inner:           createFn(),
		createFn:        createFn,
		shouldRetryFunc: shouldRetryFn,
	}
}

// RoundTrip sets the appropriate protocol for the request and executs it with retry and re-create
// logic. HTTP request will only be retried once if error matches the given criteria defined in
// ShouldRetryFunc.
func (m *HTTPReTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.ProtoMajor = m.protoMajor
	req.ProtoMinor = m.protoMinor
	req.Proto = m.proto
	resp, err := m.executeRequest(req)
	return resp, err
}

// NotifyConnect initiates re-creating the inner round tripper when called.
func (m *HTTPReTransport) NotifyConnect(events.DataConnect) error {
	m.recreateRoundTrip()
	return nil
}

func (m *HTTPReTransport) executeRequest(req *http.Request) (*http.Response, error) {
	m.mu.RLock()
	inner := m.inner
	m.mu.RUnlock()
	response, err := inner.RoundTrip(req)

	if err != nil {
		// Re-create the RoundTripper inner and retry one more time to execute the same
		// request.
		m.recreateRoundTrip()
	}

	// Check the errors if inner RoundTripper needs to be recreated.
	if m.shouldRetryFunc != nil && m.shouldRetryFunc(err) {
		log.Println(internal.InfoPrefix, "recreate RoundTripper on error", err)
		m.mu.RLock()
		inner := m.inner
		m.mu.RUnlock()
		response, err = inner.RoundTrip(req)
	}

	return response, err
}

func (m *HTTPReTransport) recreateRoundTrip() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inner, ok := m.inner.(io.Closer); ok {
		_ = inner.Close()
	}
	m.inner = m.createFn()
}
