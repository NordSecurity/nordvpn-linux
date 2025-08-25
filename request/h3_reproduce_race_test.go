//go:build race_repro
// +build race_repro

package request

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// TestReproduceHTTP3DataRace reproduces the original data race that occurs
// when using http3.Transport directly without our wrapper.
//
// Run this test with: go test -race -tags=race_repro ./request -run TestReproduceHTTP3DataRace
//
// This test should FAIL with a data race when run with the race detector,
// demonstrating the problem that existed before our fix.
func TestReproduceHTTP3DataRace(t *testing.T) {
	t.Log("WARNING: This test is expected to fail with a data race!")
	t.Log("It demonstrates the problem that existed before our fix.")

	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("Failed to get system cert pool: %v", err)
	}

	h3Transport := &http3.Transport{
		QUICConfig: &quic.Config{
			MaxIdleTimeout: 30 * time.Second,
		},
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
	defer h3Transport.Close()

	createTransport := func() http.RoundTripper {
		return NewHTTPReTransport(
			3,
			0,
			"HTTP/3",
			func() http.RoundTripper { return h3Transport },
			nil,
		)
	}

	cdnTransport := createTransport()
	apiTransport := createTransport()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// simulate concurrent requests from CDN client (like remote config loader)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			// note: no need for real url, we are testing initialization logic only
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://rc.nordvpn.com", nil)
			// this will trigger the race in http3.Transport.dial() at line 308 (in v0.48.2)
			_, _ = cdnTransport.RoundTrip(req)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// simulate concurrent requests from API client (like insights job)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.nordvpn.com", nil)
			// this will trigger the race in http3.Transport.dial() at line 313 (in v0.48.2)
			_, _ = apiTransport.RoundTrip(req)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// add more concurrent goroutines to increase the chance of hitting the race
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.nordvpn.com", nil)
				_, _ = h3Transport.RoundTrip(req)
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	t.Log("If you see this message without a data race, try running the test multiple times")
	t.Log("The race condition is timing-dependent and may not always trigger")
}

// TestDirectHTTP3TransportRace shows the race even more directly
// by using the raw http3.Transport without any wrappers
func TestDirectHTTP3TransportRace(t *testing.T) {
	t.Log("Direct test of http3.Transport race condition")

	transport := &http3.Transport{
		QUICConfig: &quic.Config{
			MaxIdleTimeout: 30 * time.Second,
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	defer transport.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// use different hosts to trigger new dial operations
			hosts := []string{
				"host1.nordpnv.com",
				"host2.nordvpn.com",
				"host3.nordvpn.com",
				"host4.nordvpn.com",
			}
			host := hosts[id%len(hosts)]

			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/test", nil)

			// this is where the race occurs - multiple goroutines calling
			// transport.RoundTrip() which internally calls dial() and accesses
			// t.transport without proper synchronization
			_, _ = transport.RoundTrip(req)
		}(i)
	}

	wg.Wait()

	t.Log("Expected data race output:")
	t.Log("  Read at 0x... by goroutine X:")
	t.Log("    github.com/quic-go/quic-go/http3.(*Transport).dial()")
	t.Log("      .../http3/transport.go:308")
	t.Log("  Previous write at 0x... by goroutine Y:")
	t.Log("    github.com/quic-go/quic-go/http3.(*Transport).dial()")
	t.Log("      .../http3/transport.go:313")
}
