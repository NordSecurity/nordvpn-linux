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

	"github.com/quic-go/quic-go/http3"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

const (
	netCoreRmemMaxKey    = "net.core.rmem_max"
	netCodeRmemMaxValue  = 2500000
	envHTTPTransportsKey = "HTTP_TRANSPORTS"
)

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

func createH3Transport() *http3.RoundTripper {
	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}
	return &http3.RoundTripper{
		// #nosec G402 -- minimum tls version is controlled by the standard library
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

	var h1Transport http.RoundTripper
	var h3Transport http.RoundTripper
	if containsH1 {
		h1ReTransport := request.NewHTTPReTransport(createH1Transport(resolver, fwmark))
		connectSubject.Subscribe(h1ReTransport.NotifyConnect)
		h1Transport = request.NewPublishingRoundTripper(
			h1ReTransport,
			httpCallsSubject,
		)
		if !containsH3 {
			return h1Transport
		}
	}
	if containsH3 {
		// For quic-go need to increase receive buffer size
		// This command will increase the maximum receive buffer size to roughly 2.5 MB
		// see: https://github.com/quic-go/quic-go/wiki/UDP-Receive-Buffer-Size
		if err := kernel.SetParameter(netCoreRmemMaxKey, netCodeRmemMaxValue); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
		h3ReTransport := request.NewQuicTransport(createH3Transport)
		connectSubject.Subscribe(h3ReTransport.NotifyConnect)
		h3Transport = request.NewPublishingRoundTripper(
			h3ReTransport,
			httpCallsSubject,
		)
		if !containsH1 {
			return h3Transport
		}
	}
	// This should never happen as validation makes sure of that but it is here for nil panics
	if h1Transport == nil || h3Transport == nil {
		log.Println(internal.ErrorPrefix, "Unexpected transport configuration, using default")
		// http.Client handles nil transport
		return nil
	}

	return request.NewRotatingRoundTripper(h1Transport, h3Transport, time.Hour)
}
