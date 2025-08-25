package request

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

func TestHTTP3TransportRaceCondition(t *testing.T) {
	category.Set(t, category.Unit)

	// this test simulates the exact scenario from the race detector output:
	// multiple goroutines making concurrent HTTP/3 requests through the same transport

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

	wrapper := NewH3TransportWrapper(h3Transport)
	defer wrapper.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	numGoroutines := 10
	numRequestsPerGoroutine := 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numRequestsPerGoroutine; j++ {
				// create requests to different hosts to trigger the dial race
				host := "example.com"
				if goroutineID%2 == 0 {
					host = "api.example.com"
				}

				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodGet,
					"https://"+host+"/test",
					nil,
				)
				if err != nil {
					t.Logf("Failed to create request: %v", err)
					continue
				}

				// this would trigger the race if no fix applied
				rsp, err := wrapper.RoundTrip(req)
				// we expect connection errors since these are not real servers
				// the important thing is no data race occurs
				if err != nil {
					continue
				}
				rsp.Body.Close()
			}
		}(i)
	}

	wg.Wait()

	t.Log("Concurrent HTTP/3 requests completed without data races")
}
