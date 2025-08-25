package request

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// This is workaround solution for `quic-go` external library's
// issue: https://github.com/quic-go/quic-go/issues/5307
// TODO: remove this wrapper when `quic-go` issue is resolved.

// H3TransportWrapper provides a thread-safe wrapper around http3.Transport
// that ensures the internal QUIC transport is initialized only once,
// preventing data races during concurrent connection establishment.
type H3TransportWrapper struct {
	once      sync.Once
	transport *http3.Transport
	// quicTransport is initialized once and reused for all connections
	quicTransport *quic.Transport
	initErr       error
}

// NewH3TransportWrapper creates a new thread-safe wrapper for http3.Transport
func NewH3TransportWrapper(transport *http3.Transport) *H3TransportWrapper {
	return &H3TransportWrapper{
		transport: transport,
	}
}

// initializeQuicTransport ensures the QUIC transport is initialized only once
func (w *H3TransportWrapper) initializeQuicTransport() {
	w.once.Do(func() {
		udpConn, err := net.ListenUDP("udp", nil)
		if err != nil {
			w.initErr = err
			return
		}
		w.quicTransport = &quic.Transport{Conn: udpConn}
		w.transport.Dial = func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlyConnection, error) {
			udpAddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				return nil, err
			}
			return w.quicTransport.DialEarly(ctx, udpAddr, tlsCfg, cfg)
		}
	})
}

// RoundTrip implements the http.RoundTripper interface
func (w *H3TransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	w.initializeQuicTransport()
	if w.initErr != nil {
		return nil, w.initErr
	}
	return w.transport.RoundTrip(req)
}

// Close closes the underlying transports
func (w *H3TransportWrapper) Close() error {
	// First close the HTTP/3 transport
	if err := w.transport.Close(); err != nil {
		return err
	}
	// Then close our QUIC transport if it was initialized
	if w.quicTransport != nil {
		if err := w.quicTransport.Close(); err != nil {
			return err
		}
		if err := w.quicTransport.Conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
