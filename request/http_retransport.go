package request

import (
	"io"
	"net/http"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// HTTPReTransport is modified/enhanced RoundTripper.
// Thread safe.
type HTTPReTransport struct {
	mu              sync.RWMutex
	inner           http.RoundTripper
	createFn        RoundTripperCreateFunc
	protoMajor      int
	protoMinor      int
	proto           string
	shouldRetryFunc ShouldRetryFunc
	// counter provides a mechanism to check whether the inner RoundTripper was re-created
	// in the background during this round trip. Integer overflow is not important here as only
	// inequality operator is used.
	counter int
}

// RoundTripperCreateFunc is a function used to create a new instance of round tripper when it needs
// to be recreated.
type RoundTripperCreateFunc func() http.RoundTripper

// ShouldRetryFunc defines a function that determines whether an HTTP request should be retried
// based on the given error.
type ShouldRetryFunc func(err error) bool

// NewHTTPReTransport is a default constructor function for an HTTPReTransport.
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
	m.mu.RLock()
	counter := m.counter
	m.mu.RUnlock()
	m.recreateRoundTrip(counter)
	return nil
}

func (m *HTTPReTransport) executeRequest(req *http.Request) (*http.Response, error) {
	m.mu.RLock()
	inner := m.inner
	counter := m.counter
	m.mu.RUnlock()

	response, err := inner.RoundTrip(req)

	if err != nil {
		// Check whether inner RoundTripper was updated while request was executed.
		m.recreateRoundTrip(counter)

		// Check the errors if inner RoundTripper needs to be recreated.
		if m.shouldRetryFunc != nil && m.shouldRetryFunc(err) {
			m.mu.RLock()
			inner := m.inner
			m.mu.RUnlock()
			response, err = inner.RoundTrip(req)
		}
	}

	return response, err
}

func (m *HTTPReTransport) recreateRoundTrip(counter int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if counter != m.counter {
		return
	}
	if inner, ok := m.inner.(io.Closer); ok {
		_ = inner.Close()
	}
	m.counter++
	m.inner = m.createFn()
}
