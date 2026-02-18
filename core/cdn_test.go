package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestCdnApi(t *testing.T) {
	category.Set(t, category.Integration)

	cdnApi, cancel := setupCdnApi()
	assert.NotNil(t, cdnApi)

	nameservers, err := cdnApi.ThreatProtectionLite()
	assert.NoError(t, err)
	assert.NotNil(t, nameservers)

	_, fileBytes, err := cdnApi.ConfigTemplate(false, http.MethodGet)
	assert.NoError(t, err)
	assert.NotZero(t, len(fileBytes))

	fileBytes, err = cdnApi.GetRemoteFile("/configs/templates/ovpn/1.0/template.xslt")
	assert.NoError(t, err)
	assert.NotZero(t, len(fileBytes))

	cancel()
}

func setupCdnApi() (*CDNAPI, context.CancelFunc) {
	Environment := "dev"
	Version := "3.3.3"
	httpCallsSubject := &subs.Subject[events.DataRequestAPI]{}

	// API
	var err error
	var validator response.Validator
	if !internal.IsProdEnv(Environment) {
		validator = response.NoopValidator{}
	} else {
		validator, err = response.NewNordValidator()
		if err != nil {
			log.Fatalln("Error on creating validator:", err)
		}
	}

	userAgent, err := request.GetUserAgentValue(Version, sysinfo.GetHostOSPrettyName)
	if err != nil {
		userAgent = fmt.Sprintf("%s/%s (unknown)", request.AppName, Version)
		log.Printf("Error while constructing UA value: %s. Falls back to default: %s\n", err, userAgent)
	}

	httpGlobalCtx, httpCancel := context.WithCancel(context.Background())

	// simple standard http client with dialer wrapped inside
	httpClientSimple := request.NewStdHTTP()
	httpClientSimple.Transport = request.NewHTTPReTransport(
		1, 1, "HTTP/1.1", func() http.RoundTripper {
			return request.NewPublishingRoundTripper(
				request.NewContextRoundTripper(request.NewStdTransport(), httpGlobalCtx),
				httpCallsSubject,
			)
		}, nil)

	cdnAPI := NewCDNAPI(
		userAgent,
		CDNURL,
		httpClientSimple,
		validator,
	)

	return cdnAPI, httpCancel
}

// failingValidator is a test validator that always fails
type failingValidator struct{}

func (failingValidator) Validate(int, http.Header, []byte) error {
	return errors.New("validation failed")
}

func TestCDNAPI_Request_Success(t *testing.T) {
	category.Set(t, category.Unit)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "test"}`))
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	resp, err := api.GetRemoteFile("/test")

	assert.NoError(t, err)
	assert.Equal(t, `{"data": "test"}`, string(resp))
}

func TestCDNAPI_Request_GzipDecompression(t *testing.T) {
	category.Set(t, category.Unit)

	originalData := []byte(`{"data": "compressed content"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/json")

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, _ = gz.Write(originalData)
		_ = gz.Close()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	resp, err := api.GetRemoteFile("/test")

	assert.NoError(t, err)
	assert.Equal(t, string(originalData), string(resp))
}

func TestCDNAPI_Request_InvalidGzipData(t *testing.T) {
	category.Set(t, category.Unit)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		// Write invalid gzip data
		_, _ = w.Write([]byte("not valid gzip data"))
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	_, err := api.GetRemoteFile("/test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gzip")
}

func TestCDNAPI_Request_CorruptedGzipData(t *testing.T) {
	category.Set(t, category.Unit)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		// Write valid gzip header but corrupted content
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, _ = gz.Write([]byte("some data"))
		_ = gz.Close()
		// Corrupt the data by truncating
		_, _ = w.Write(buf.Bytes()[:10])
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	_, err := api.GetRemoteFile("/test")

	assert.Error(t, err)
}

func TestCDNAPI_Request_OversizedResponse(t *testing.T) {
	category.Set(t, category.Unit)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		// Write more than MaxBytesLimit
		data := make([]byte, internal.MaxBytesLimit+1000)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	_, err := api.GetRemoteFile("/test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max limit")
}

func TestCDNAPI_Request_ValidationFailure(t *testing.T) {
	category.Set(t, category.Unit)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "test"}`))
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, failingValidator{})
	_, err := api.GetRemoteFile("/test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cdn api")
	assert.Contains(t, err.Error(), "validation failed")
}

func TestCDNAPI_Request_HTTPError(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		statusCode int
	}{
		{"BadRequest", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"Forbidden", http.StatusForbidden},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"errors": {"code": 123, "message": "error"}}`))
			}))
			defer server.Close()

			api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
			_, err := api.GetRemoteFile("/test")

			assert.Error(t, err)
		})
	}
}

func TestCDNAPI_Request_HeadMissingHeaders(t *testing.T) {
	category.Set(t, category.Unit)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't set mandatory headers
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	_, _, err := api.ConfigTemplate(false, http.MethodHead)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mandatory response headers")
}

func TestCDNAPI_Request_HeadWithMandatoryHeaders(t *testing.T) {
	category.Set(t, category.Unit)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Authorization", "auth")
		w.Header().Set("X-Digest", "digest")
		w.Header().Set("X-Accept-Before", "before")
		w.Header().Set("X-Signature", "sig")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	headers, body, err := api.ConfigTemplate(false, http.MethodHead)

	assert.NoError(t, err)
	assert.NotNil(t, headers)
	assert.Empty(t, body)
}

func TestCDNAPI_Request_NetworkError(t *testing.T) {
	category.Set(t, category.Unit)

	// Use invalid URL to trigger network error
	api := NewCDNAPI("test-agent", "http://localhost:1", &http.Client{}, response.NoopValidator{})
	_, err := api.GetRemoteFile("/test")

	assert.Error(t, err)
}

func TestCDNAPI_Request_SizeLimitAppliesToCompressedData(t *testing.T) {
	category.Set(t, category.Unit)

	// Create data that compresses well - large when decompressed, small when compressed
	largeData := make([]byte, 1024*1024) // 1MB of repetitive data
	for i := range largeData {
		largeData[i] = 'X'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/octet-stream")

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, _ = gz.Write(largeData)
		_ = gz.Close()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
	defer server.Close()

	api := NewCDNAPI("test-agent", server.URL, &http.Client{}, response.NoopValidator{})
	resp, err := api.GetRemoteFile("/test")

	// Should succeed because compressed size is small
	assert.NoError(t, err)
	assert.Equal(t, len(largeData), len(resp))
}
