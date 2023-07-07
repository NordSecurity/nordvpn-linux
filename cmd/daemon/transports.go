package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/request"

	"github.com/quic-go/quic-go/http3"
	"golang.org/x/sys/unix"
)

const (
	netCoreRmemMaxKey   = "net.core.rmem_max"
	netCodeRmemMaxValue = 2500000
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

func createH3Transport() http.RoundTripper {
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
