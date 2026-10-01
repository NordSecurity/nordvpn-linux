package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/kernel"
	"github.com/NordSecurity/nordvpn-linux/log"
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

// resolverWithBackoff wraps a network.DNSResolver with a self-managing backoff
// mechanism for internal DNS resolution.
type resolverWithBackoff struct {
	resolver               network.DNSResolver
	mu                     sync.Mutex
	nextInternalDNSAttempt time.Time
	backoff                time.Duration
}

func newResolverWithBackoff(resolver network.DNSResolver) *resolverWithBackoff {
	return &resolverWithBackoff{
		resolver: resolver,
	}
}

func (r *resolverWithBackoff) isInBackoffModeThreadSafe() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.isInBackoffMode()
}

func (r *resolverWithBackoff) isInBackoffMode() bool {
	return time.Now().Before(r.nextInternalDNSAttempt)
}

// setBackoff sets the new backoff based on the previous backoff:
//
//  1. initial backoff is 5 minutes long
//
//  2. subsequent backoff is 30 minutes long
//
//  3. all backoffs after that are 60 minutes long
//
// Backoff is set only if backoff is not currently enabled, to prevent backoff saturation.
// Returns true if new backoff was set.
func (r *resolverWithBackoff) setBackoff() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isInBackoffMode() {
		return false
	}

	const (
		initialBackoff = 5 * time.Minute
		secondBackoff  = 30 * time.Minute
		maxBackoff     = 60 * time.Minute
	)

	//nolint:exhaustive // time.Duration is not an enum; default covers all other values
	switch r.backoff {
	case 0:
		r.backoff = initialBackoff
	case initialBackoff:
		r.backoff = secondBackoff
	default:
		r.backoff = maxBackoff
	}

	log.DNS.Info("backing off from internal DNS resolution for", r.backoff)
	r.nextInternalDNSAttempt = time.Now().Add(r.backoff)

	return true
}

func (r *resolverWithBackoff) unsetBackoff() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.backoff != 0 {
		log.DNS.Info("unsetting internal DNS resolution backoff")
	}

	r.backoff = 0
	r.nextInternalDNSAttempt = time.Time{}
}

// resolveDomainName resolves domain via the internal resolver, applying the
// backoff and killswitch rules:
//   - a backoff of 5 => 30 => 60 will be set after DNS resolution failures
//   - if killswitch is on, internal resolver will always be used
//   - if killswitch is off and the internal resolver fails or backoff is on, domain will be returned as is to be
//     resolved by the OS resolver
func (r *resolverWithBackoff) resolveDomainName(ctx context.Context, domain string) (string, error) {
	inBackoff := r.isInBackoffModeThreadSafe()

	if inBackoff {
		return domain, nil
	}

	addr, err := r.resolver.Resolve(ctx, domain)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return "", fmt.Errorf("resolving DNS: %w", err)
		}

		// only set backoff if it was not set(or the previous backoff has expired) so that it won't be saturated by
		// multiple failed DNS calls.
		if r.setBackoff() {
			log.DNS.Warn("failed to resolve domain name with internal resolver, enabling backoff:", err)
		}

		return domain, nil
	}

	r.unsetBackoff()

	if len(addr) == 0 {
		return "", fmt.Errorf("no resolved addresses")
	}

	var resolvedAddress string
	if ip := addr[0]; ip.Is6() {
		resolvedAddress = fmt.Sprintf("[%s]", ip.String())
	} else {
		resolvedAddress = ip.String()
	}

	return resolvedAddress, nil
}

func createH1Transport(
	resolver network.DNSResolver,
	fwmark uint32,
	environment string,
) func() http.RoundTripper {
	resolverWrapper := newResolverWithBackoff(resolver)

	return func() http.RoundTripper {
		dialer := &net.Dialer{
			Control: network.NewFwmarkControlFn(fwmark),
			Timeout: request.DefaultTimeout,
		}

		transport := &http.Transport{
			DialContext: func(ctx context.Context, netw, addr string) (net.Conn, error) {
				domain, _, ok := strings.Cut(addr, ":")
				if !ok {
					return nil, fmt.Errorf("malformed address: %s", addr)
				}

				// resolverWrapper will return unresolved address if DNS resolution fails and killswitch is off.
				// In such cases this address will be resolved by the OS resolver when it's passed on to the dialer.
				resolvedAddr, err := resolverWrapper.resolveDomainName(ctx, domain)
				if err != nil {
					return nil, err
				}

				return dialer.DialContext(
					ctx,
					netw,
					strings.ReplaceAll(addr, domain, resolvedAddr),
				)
			},
			TLSHandshakeTimeout: request.TransportTimeout,
		}

		if internal.IsDevEnv(environment) {
			transport.Proxy = http.ProxyFromEnvironment
		}

		return transport
	}
}

func createSimpleH1Transport(environment string) func() http.RoundTripper {
	return func() http.RoundTripper {
		t := &http.Transport{
			DialContext:         (&net.Dialer{Timeout: request.TransportTimeout}).DialContext,
			TLSHandshakeTimeout: request.TransportTimeout,
		}

		if internal.IsDevEnv(environment) {
			t.Proxy = http.ProxyFromEnvironment
		}

		return t
	}
}

type h3Wrapper struct {
	rt  http.RoundTripper
	err error
}

func (h *h3Wrapper) RoundTrip(r *http.Request) (*http.Response, error) {
	if h.err != nil {
		return nil, h.err
	}
	return h.rt.RoundTrip(r)
}

func createH3Transport(resolver network.DNSResolver, fwmark uint32) func() http.RoundTripper {
	return func() http.RoundTripper {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return &h3Wrapper{nil, err}
		}
		// as of quic-go 0.40.1, GSO handling causes race conditions
		_ = os.Setenv("QUIC_GO_DISABLE_GSO", "true")
		udpConn, err := NewMarkedUDPConn(fwmark)
		if err != nil {
			return &h3Wrapper{nil, err}
		}
		quicTransport := &quic.Transport{Conn: udpConn}
		// #nosec G402 -- minimum tls version is controlled by the standard library
		return &h3Wrapper{
			&http3.Transport{
				QUICConfig: &quic.Config{
					MaxIdleTimeout: request.TransportTimeout,
				},
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
				// Custom dial is needed to resolve domain names with fwmarked resolver as well as
				// dialing with fwmarked connection
				Dial: func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
					domain, portStr, ok := strings.Cut(addr, ":")
					if !ok {
						return nil, fmt.Errorf("malformed address: %s", addr)
					}
					port, err := strconv.ParseUint(portStr, 10, 16)
					if err != nil {
						return nil, fmt.Errorf("port conversion failed: %s", portStr)
					}
					ips, err := resolver.Resolve(ctx, domain)
					if err != nil {
						return nil, err
					}
					if !ips[0].IsValid() {
						return nil, fmt.Errorf("invalid IP resolved: %s", ips[0])
					}
					udpAddrPort := netip.AddrPortFrom(ips[0], uint16(port))
					udpAddr := net.UDPAddrFromAddrPort(udpAddrPort)
					c, err := quicTransport.DialEarly(ctx, udpAddr, tlsCfg, cfg)
					if err != nil {
						return nil, err
					}
					return c, nil
				},
			},
			nil,
		}
	}
}

func NewMarkedUDPConn(fwmark uint32) (*net.UDPConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var sockErr error
			err := c.Control(func(fd uintptr) {
				sockErr = unix.SetsockoptInt(
					int(fd), unix.SOL_SOCKET, unix.SO_MARK, int(fwmark),
				)
			})
			if err != nil {
				return err
			}
			return sockErr
		},
	}

	pc, err := lc.ListenPacket(context.Background(), "udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("listen udp: %w", err)
	}
	return pc.(*net.UDPConn), nil
}

// TODO: Remove the code in the future - LVPN-11168
var validTransportTypes = []string{"http1"}

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
			log.Warn("invalid http transport type value:", item, "; valid values:", validTransportTypes)
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
	ctx context.Context,
	environment string,
) http.RoundTripper {
	transportsStr := os.Getenv(envHTTPTransportsKey)
	log.Info("http transports to use (environment):", transportsStr)
	transportTypes := validateHTTPTransportsString(transportsStr)
	log.Info("http transports to use (after validation):", transportTypes)

	containsH1 := slices.Contains(transportTypes, "http1")
	containsH3 := slices.Contains(transportTypes, "http3")

	var h1Transport *request.HTTPReTransport
	var h3Transport *request.HTTPReTransport
	transportNeedsRecreate := false
	if containsH1 {
		h1Transport = request.NewHTTPReTransport(
			1,
			1,
			"HTTP/1.1",
			createH1Transport(resolver, fwmark, environment),
			nil,
			transportNeedsRecreate,
		)
		connectSubject.Subscribe(h1Transport.NotifyConnect)
		if !containsH3 {
			return request.NewPublishingRoundTripper(
				request.NewContextRoundTripper(h1Transport, ctx),
				httpCallsSubject,
			)
		}
	}
	if containsH3 {
		if err := SetBufferSizeForHTTP3(); err != nil {
			log.Warn("failed to set buffer size for HTTP/3:", err)
		}
		h3Transport = request.NewHTTPReTransport(
			3,
			0,
			"HTTP/3",
			createH3Transport(resolver, fwmark),
			shouldRetryHTTP3,
			transportNeedsRecreate,
		)
		connectSubject.Subscribe(h3Transport.NotifyConnect)
		if !containsH1 {
			return request.NewPublishingRoundTripper(
				request.NewContextRoundTripper(h3Transport, ctx),
				httpCallsSubject,
			)
		}
	}
	// This should never happen as validation makes sure of that but it is here for nil panics
	if h1Transport == nil || h3Transport == nil {
		log.Error("Unexpected transport configuration, using default")
		// http.Client handles nil transport
		return nil
	}

	rotatingRoundTriper := request.NewRotatingRoundTripper(h1Transport, h3Transport, time.Hour)
	return request.NewPublishingRoundTripper(rotatingRoundTriper, httpCallsSubject)
}

func shouldRetryHTTP3(err error) bool {
	return err != nil &&
		(strings.Contains(err.Error(), "Application error 0x100") ||
			strings.Contains(err.Error(), "no recent network activity") ||
			strings.Contains(err.Error(), "Timeout exceeded while awaiting headers"))
}
