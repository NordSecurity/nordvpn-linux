package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/kernel"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/request"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

const (
	netCoreRmemMaxKey    = "net.core.rmem_max"
	netCoreWmemMaxKey    = "net.core.wmem_max"
	netCoreMemMaxValue   = 7500000
	envHTTPTransportsKey = "HTTP_TRANSPORTS"
)

// SetBufferSizeForHTTP3 increase receive buffer size to roughly 7.5 MB, as recommended for quic-go library.
// see: https://github.com/quic-go/quic-go/wiki/UDP-Receive-Buffer-Size
func SetBufferSizeForHTTP3() error {
	if err := kernel.SetParameter(netCoreRmemMaxKey, netCoreMemMaxValue); err != nil {
		return fmt.Errorf("setting receive buffer: %w", err)
	}
	if err := kernel.SetParameter(netCoreWmemMaxKey, netCoreMemMaxValue); err != nil {
		return fmt.Errorf("setting write buffer: %w", err)
	}
	return nil
}

func createH1Transport(resolver network.DNSResolver, fwmark uint32) func() http.RoundTripper {
	return func() http.RoundTripper {
		var operr error
		fwmark := func(fd uintptr) {
			operr = syscall.SetsockoptInt(
				int(fd),
				unix.SOL_SOCKET,
				unix.SO_MARK,
				int(fwmark),
			)
		}
		dialer := &net.Dialer{
			Control: func(network, address string, conn syscall.RawConn) error {
				if err := conn.Control(fwmark); err != nil {
					return err
				}
				return operr
			},
			Timeout: request.DefaultTimeout,
		}
		return &http.Transport{
			DialContext: func(ctx context.Context, netw, addr string) (net.Conn, error) {
				domain, _, ok := strings.Cut(addr, ":")
				if !ok {
					return nil, fmt.Errorf("malformed address: %s", addr)
				}

				ips, err := resolver.Resolve(domain)
				if err != nil {
					return nil, err
				}

				var newAddr string
				if ip := ips[0]; ip.Is6() {
					newAddr = fmt.Sprintf("[%s]", ip.String())
				} else {
					newAddr = ip.String()
				}
				return dialer.DialContext(
					ctx,
					netw,
					strings.ReplaceAll(addr, domain, newAddr),
				)
			},
			TLSHandshakeTimeout: request.TransportTimeout,
		}
	}
}

func createH3Transport() *http3.Transport {
	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}

	// as of quic-go 0.40.1, GSO handling causes race conditions
	_ = os.Setenv("QUIC_GO_DISABLE_GSO", "1")
	// #nosec G402 -- minimum tls version is controlled by the standard library
	return &http3.Transport{
		QUICConfig: &quic.Config{
			MaxIdleTimeout: request.TransportTimeout,
		},
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
}

var validTransportTypes = []string{"http1", "http3"}

func validateHTTPTransportsString(val string) []string {
	if val == "" {
		return validTransportTypes
	}
	finalVal := []string{}
	val = strings.ToLower(val)
	for _, item := range strings.Split(val, ",") {
		if slices.Contains(validTransportTypes, item) {
			finalVal = append(finalVal, item)
		} else {
			log.Println(internal.WarningPrefix, "invalid http transport type value:", item, "; valid values:", validTransportTypes)
		}
	}

	if len(finalVal) == 0 {
		finalVal = validTransportTypes
	}
	return finalVal
}

// createTimedOutTransports provides transports to APIs' client
func createTimedOutTransport(
	resolver network.DNSResolver,
	fwmark uint32,
	httpCallsSubject events.Publisher[events.DataRequestAPI],
	connectSubject events.PublishSubcriber[events.DataConnect],
) http.RoundTripper {
	transportsStr := os.Getenv(envHTTPTransportsKey)
	log.Println(internal.InfoPrefix, "http transports to use (environment):", transportsStr)
	transportTypes := validateHTTPTransportsString(transportsStr)
	log.Println(internal.InfoPrefix, "http transports to use (after validation):", transportTypes)

	containsH1 := slices.Contains(transportTypes, "http1")
	containsH3 := slices.Contains(transportTypes, "http3")

	var h1Transport *request.HTTPReTransport
	var h3Transport *request.QuicTransport
	if containsH1 {
		h1Transport = request.NewHTTPReTransport(createH1Transport(resolver, fwmark))
		connectSubject.Subscribe(h1Transport.NotifyConnect)
		if !containsH3 {
			return request.NewPublishingRoundTripper(
				h1Transport,
				httpCallsSubject,
			)
		}
	}
	if containsH3 {
		if err := SetBufferSizeForHTTP3(); err != nil {
			log.Println(internal.WarningPrefix, "failed to set buffer size for HTTP/3:", err)
		}
		h3Transport = request.NewQuicTransport(createH3Transport)
		connectSubject.Subscribe(h3Transport.NotifyConnect)
		if !containsH1 {
			return request.NewPublishingRoundTripper(
				h3Transport,
				httpCallsSubject,
			)
		}
	}
	// This should never happen as validation makes sure of that but it is here for nil panics
	if h1Transport == nil || h3Transport == nil {
		log.Println(internal.ErrorPrefix, "Unexpected transport configuration, using default")
		// http.Client handles nil transport
		return nil
	}

	rotatingRoundTriper := request.NewRotatingRoundTripper(h1Transport, h3Transport, time.Hour)
	return request.NewPublishingRoundTripper(rotatingRoundTriper, httpCallsSubject)
}
