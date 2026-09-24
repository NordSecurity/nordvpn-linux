package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"github.com/stretchr/testify/assert"
)

const (
	serverListLargeURL string = "https://api.nordvpn.com/v1/servers?limit=1073741824"
	serverListSmallURL string = "https://api.nordvpn.com/v1/servers?limit=1"
	nonH3serverURL     string = "https://nordsec.com"
)

type workingResolver struct {
	IP string
}

func (w workingResolver) Resolve(string, context.Context) ([]netip.Addr, error) {
	if w.IP != "" {
		return []netip.Addr{netip.MustParseAddr(w.IP)}, nil
	}
	return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
}

func queryAPI(url string, transp http.RoundTripper) error {
	fmt.Printf("Query API url: %s\n\n", url)

	hclient := &http.Client{
		Transport: transp,
	}

	rsp, err := hclient.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	//fmt.Printf("Got response for %s: %#v\n\n", url, rsp)

	body := &bytes.Buffer{}
	_, err = io.Copy(body, rsp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Response Body: %d bytes\n\n", body.Len())

	return nil
}

func TestTransports(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []struct {
		comment     string
		inputURL    string
		transport   http.RoundTripper
		expectError bool
	}{
		{
			comment:     "test older transport small req/resp",
			inputURL:    serverListSmallURL,
			transport:   createH1Transport(workingResolver{}, 0, "", mock.NewMockConfigManager())(),
			expectError: false,
		},
		{
			comment:     "test older transport large resp",
			inputURL:    serverListLargeURL,
			transport:   createH1Transport(workingResolver{}, 0, "", mock.NewMockConfigManager())(),
			expectError: false,
		},
		{
			comment:     "test quic transport small req/resp",
			inputURL:    serverListSmallURL,
			transport:   createH3Transport(workingResolver{}, 0)(),
			expectError: false,
		},
		// { Fix in LVPN-6886
		// 	comment:     "test quic transport large resp",
		// 	inputURL:    serverListLargeURL,
		// 	transport:   createH3Transport(),
		// 	expectError: false,
		// },
		{
			comment:     "test non quic/H3 url with H1 transport",
			inputURL:    nonH3serverURL,
			transport:   createH1Transport(workingResolver{}, 0, "", mock.NewMockConfigManager())(),
			expectError: false,
		},
		{
			comment:     "test non quic/H3 url with H3 transport",
			inputURL:    nonH3serverURL,
			transport:   createH3Transport(workingResolver{}, 0)(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			err := queryAPI(tt.inputURL, tt.transport)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestH1Transport_RoundTrip(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []struct {
		ip string
	}{
		{ip: "127.0.0.1"},
		{ip: "::1"},
	}

	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			transport := createH1Transport(workingResolver{IP: test.ip}, 0, "", mock.NewMockConfigManager())()
			req, err := http.NewRequest(http.MethodGet, serverListSmallURL, nil)
			assert.NoError(t, err)
			resp, err := transport.RoundTrip(req)
			if err == nil {
				defer resp.Body.Close()
			}
			assert.Contains(t, err.Error(), "connection refused")
		})
	}
}

func Test_validateHttpTransportsString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		value         string
		expectedValue []string
	}{
		{value: "http1", expectedValue: []string{"http1"}},
		{value: "AAhttp1", expectedValue: validTransportTypes},
		{value: "http1AA", expectedValue: validTransportTypes},
		{value: "http3,http1", expectedValue: []string{"http1"}},
		{value: "http2,http1", expectedValue: []string{"http1"}},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			assert.Equal(t, test.expectedValue, validateHTTPTransportsString(test.value))
		})
	}
}

// mockDNSResolver lets each test case control the resolver's return values.
type mockDNSResolver struct {
	addrs []netip.Addr
	err   error
}

func (m mockDNSResolver) Resolve(domain string, ctx context.Context) ([]netip.Addr, error) {
	return m.addrs, m.err
}

func TestResolverWrapper_ResolveDomainName(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                   string
		resolverAddrs          []netip.Addr
		resolverErr            error
		killSwitch             bool
		loadErr                error
		initialBackoffMinutes  int64
		initialNextAttemptUnix int64
		domain                 string
		expectedAddress        string
		errorIsExpected        bool
		expectedBackoffMins    int64
		expectedBackoffSet     bool
	}{
		{
			name:            "successful IPv4 resolution",
			resolverAddrs:   []netip.Addr{netip.MustParseAddr("1.2.3.4")},
			domain:          "example.com",
			expectedAddress: "1.2.3.4",
		},
		{
			name:                  "success clears existing backoff",
			resolverAddrs:         []netip.Addr{netip.MustParseAddr("1.2.3.4")},
			initialBackoffMinutes: 60,
			domain:                "example.com",
			expectedAddress:       "1.2.3.4",
			expectedBackoffMins:   0,
			expectedBackoffSet:    false,
		},
		{
			name:                "resolve error with kill switch off returns raw domain",
			resolverErr:         errors.New("dns failure"),
			killSwitch:          false,
			domain:              "example.com",
			expectedAddress:     "example.com",
			expectedBackoffMins: 5,
			expectedBackoffSet:  true,
		},
		{
			name:                "resolve error with kill switch on returns error",
			resolverErr:         errors.New("dns failure"),
			killSwitch:          true,
			domain:              "example.com",
			errorIsExpected:     true,
			expectedBackoffMins: 5,
			expectedBackoffSet:  true,
		},
		{
			name:                "resolve error with config load error returns error",
			resolverErr:         errors.New("dns failure"),
			loadErr:             errors.New("config load failure"),
			domain:              "example.com",
			errorIsExpected:     true,
			expectedBackoffMins: 5,
			expectedBackoffSet:  true,
		},
		{
			name:                  "backoff escalates 5 to 30 on failure",
			resolverErr:           errors.New("dns failure"),
			killSwitch:            true,
			initialBackoffMinutes: 5,
			domain:                "example.com",
			errorIsExpected:       true,
			expectedBackoffMins:   30,
			expectedBackoffSet:    true,
		},
		{
			name:                  "backoff caps at 60 minutes",
			resolverErr:           errors.New("dns failure"),
			killSwitch:            true,
			initialBackoffMinutes: 60,
			domain:                "example.com",
			errorIsExpected:       true,
			expectedBackoffMins:   60,
			expectedBackoffSet:    true,
		},
		{
			name:                   "in backoff mode returns raw domain without resolving",
			initialNextAttemptUnix: time.Now().Add(time.Hour).Unix(),
			domain:                 "example.com",
			expectedAddress:        "example.com",
			expectedBackoffSet:     true,
		},
		{
			name:            "empty resolver result returns error",
			resolverAddrs:   []netip.Addr{},
			domain:          "example.com",
			errorIsExpected: true,
		},
		{
			name:                   "killswitch is assumed to be on in case of config load error, doesn't change backoff when already in backoff mode",
			resolverAddrs:          []netip.Addr{},
			domain:                 "example.com",
			resolverErr:            errors.New("dns failure"),
			loadErr:                errors.New("config load failure"),
			initialNextAttemptUnix: time.Now().Add(time.Hour).Unix(),
			initialBackoffMinutes:  5,
			expectedBackoffMins:    5,
			errorIsExpected:        true,
			expectedBackoffSet:     true,
		},
		{
			name:                  "killswitch is assumed to be on in case of config load error, updates backoff when not in backoff mode",
			resolverAddrs:         []netip.Addr{},
			domain:                "example.com",
			resolverErr:           errors.New("dns failure"),
			loadErr:               errors.New("config load failure"),
			initialBackoffMinutes: 5,
			expectedBackoffMins:   30,
			errorIsExpected:       true,
			expectedBackoffSet:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg.KillSwitch = test.killSwitch
			cfgManager.LoadErr = test.loadErr

			resolver := mockDNSResolver{
				addrs: test.resolverAddrs,
				err:   test.resolverErr,
			}

			resolverWrapper := newResolverWrapper(resolver, cfgManager)
			resolverWrapper.internalDNSBackoffMinutes.Store(test.initialBackoffMinutes)
			resolverWrapper.nextInternalDNSAttemptUnix.Store(test.initialNextAttemptUnix)

			resolvedAddress, err := resolverWrapper.resolveDomainName(test.domain, context.Background())

			assert.Equal(t, test.expectedAddress, resolvedAddress, "Domain name was resolved to an unexpected address.")
			if test.errorIsExpected {
				assert.Error(t, err, "Expected error not returned by the resolver wrapper.")
			} else {
				assert.NoError(t, err, "Unexpected error returned by the resolver wrapper.")
			}
			assert.Equal(t, test.expectedBackoffMins, resolverWrapper.internalDNSBackoffMinutes.Load(),
				"Unexpected backoff value after DNS resolution attempt.")
			assert.Equal(t, test.expectedBackoffSet, resolverWrapper.isInBackoffMode(),
				"Backoff not set as expected after DNS resolution attempt.")
		})
	}
}
