package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/test/category"

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

func (w workingResolver) Resolve(string) ([]netip.Addr, error) {
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
			transport:   createH1Transport(func() network.DNSResolver { return workingResolver{} }, 0)(),
			expectError: false,
		},
		{
			comment:     "test older transport large resp",
			inputURL:    serverListLargeURL,
			transport:   createH1Transport(func() network.DNSResolver { return workingResolver{} }, 0)(),
			expectError: false,
		},
		{
			comment:     "test quic transport small req/resp",
			inputURL:    serverListSmallURL,
			transport:   createH3Transport(),
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
			transport:   createH1Transport(func() network.DNSResolver { return workingResolver{} }, 0)(),
			expectError: false,
		},
		{
			comment:     "test non quic/H3 url with H3 transport",
			inputURL:    nonH3serverURL,
			transport:   createH3Transport(),
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
			transport := createH1Transport(func() network.DNSResolver { return workingResolver{IP: test.ip} }, 0)()
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
		{value: "http3,http1", expectedValue: []string{"http3", "http1"}},
		{value: "http2,http1", expectedValue: []string{"http1"}},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			assert.Equal(t, test.expectedValue, validateHTTPTransportsString(test.value))
		})
	}
}
