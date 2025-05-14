package request

import (
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

type roundTripper struct {
	closeCh chan error
	errCh   chan error
	wg      *sync.WaitGroup
}

func (rt *roundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	rt.wg.Done()
	defer rt.wg.Done()
	select {
	case err := <-rt.closeCh:
		return nil, err
	case err := <-rt.errCh:
		return nil, err
	}
}

func (rt *roundTripper) Close() error {
	// Write to the channel while there are readers (close all active connections).
	for {
		select {
		case rt.closeCh <- io.ErrClosedPipe:
		default:
			return nil
		}
	}
}

func TestHTTPReTransport_RoundTrip(t *testing.T) {
	category.Set(t, category.Unit)
	iterations := 5
	closeCh := make(chan error)
	errCh := make(chan error)
	wg := &sync.WaitGroup{}
	transport := NewHTTPReTransport(1, 1, "HTTP/1.1", func() http.RoundTripper {
		return &roundTripper{closeCh: closeCh, errCh: errCh, wg: wg}
	}, nil)

	// Wrap access to `errs` with mutex to avoid modifying from same value twice.
	mu := sync.Mutex{}
	var errs []error
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			//nolint:bodyclose
			_, err := transport.RoundTrip(&http.Request{})
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			wg.Done()
		}()
	}
	// Wait until all of RoundTrip calls are started.
	wg.Wait()

	wg.Add(iterations * 2)
	errCh <- io.ErrUnexpectedEOF
	// Wait until all of RoundTrip calls are done and errors are collected.
	wg.Wait()

	assert.Equal(t, []error{
		io.ErrUnexpectedEOF,
		io.ErrClosedPipe,
		io.ErrClosedPipe,
		io.ErrClosedPipe,
		io.ErrClosedPipe,
	}, errs)

	// Check if transports were recreated only once.
	assert.Equal(t, 1, transport.counter)
}
