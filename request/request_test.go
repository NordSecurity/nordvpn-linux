package request

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type mockRotator struct {
	url string
}

func (m *mockRotator) Rotate() (MetaTransport, error) {
	return MetaTransport{}, ErrNothingMoreToRotate
}

func (mockRotator) Restart() MetaTransport {
	return MetaTransport{}
}

type failingRotator struct{}

func (failingRotator) Rotate() (MetaTransport, error) {
	return MetaTransport{}, errors.New("failing rotator")
}

func (m *failingRotator) Restart() MetaTransport {
	return MetaTransport{}
}

func TestHTTPClient_DoRequest(t *testing.T) {
	category.Set(t, category.Unit)

	// Start a local HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			// Send response to be tested
			rw.Write([]byte(`Success!`))
		}
	}))
	// Close the server when test finishes
	defer testServer.Close()

	type fields struct {
		Client  *http.Client
		BaseURL string
		Path    string
		rotator CompleteRotator
	}
	tests := []struct {
		name     string
		fields   fields
		expected []byte
		hasError bool
	}{
		{
			name: "right url with nil rotator",
			fields: fields{
				Client:  testServer.Client(),
				BaseURL: testServer.URL,
				rotator: nil,
			},
			expected: []byte("Success!"),
			hasError: false,
		},
		{
			name: "right url with rotator",
			fields: fields{
				Client:  testServer.Client(),
				BaseURL: testServer.URL,
				rotator: &mockRotator{url: testServer.URL},
			},
			expected: []byte("Success!"),
			hasError: false,
		},
		{
			name: "right url with rotator and wrong server response",
			fields: fields{
				Client:  testServer.Client(),
				BaseURL: testServer.URL,
				Path:    "/wrong",
				rotator: &mockRotator{url: testServer.URL},
			},
			expected: []byte{},
			hasError: false,
		},
		{
			name: "right url with wrong rotator",
			fields: fields{
				Client:  testServer.Client(),
				BaseURL: testServer.URL,
				rotator: &mockRotator{url: "this_is_not_a_url"},
			},
			expected: []byte("Success!"),
			hasError: false,
		},
		{
			name: "right url with failing rotator",
			fields: fields{
				Client:  testServer.Client(),
				BaseURL: testServer.URL,
				rotator: &failingRotator{},
			},
			expected: []byte("Success!"),
			hasError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewHTTPClient(test.fields.Client, test.fields.rotator, nil)
			req, _ := http.NewRequest(http.MethodGet, test.fields.BaseURL+test.fields.Path, nil)
			got, err := c.DoRequest(req)
			if test.hasError {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)

			res, err := ioutil.ReadAll(got.Body)
			got.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			assert.Equal(t, test.expected, res)
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		base        string
		path        string
		agent       string
		contentType string
		contentLen  string
		encoding    string
		hasError    bool
	}{
		{
			name:        "invalid url",
			method:      http.MethodGet,
			base:        "::",
			contentType: "application/json",
			encoding:    "gzip",
			contentLen:  "666",
			hasError:    true,
		},
		{
			name:        "no agent",
			method:      http.MethodGet,
			contentType: "application/json",
			encoding:    "gzip",
			contentLen:  "666",
		},
		{
			name:        "no encoding",
			method:      http.MethodGet,
			agent:       "aliens",
			contentType: "application/json",
			contentLen:  "666",
		},
		{
			name:       "no content type",
			method:     http.MethodGet,
			agent:      "aliens",
			encoding:   "gzip",
			contentLen: "666",
		},
		{
			name:        "no content length",
			method:      http.MethodGet,
			agent:       "aliens",
			contentType: "application/json",
			encoding:    "gzip",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := NewRequest(
				test.method,
				test.agent,
				test.base,
				test.path,
				test.contentType,
				test.contentLen,
				test.encoding,
				nil,
			)
			assert.Equal(t, test.hasError, err != nil)
			if !test.hasError {
				assert.Equal(t, test.method, req.Method)
				assert.Equal(t, test.base, req.URL.Host)
				assert.Equal(t, test.path, req.URL.Path)
				assert.Equal(t, test.agent, req.Header.Get("User-Agent"))
				assert.Equal(t, test.contentType, req.Header.Get("Content-Type"))
				assert.Equal(t, test.contentLen, req.Header.Get("Content-Length"))
				assert.Equal(t, test.encoding, req.Header.Get("Accept-Encoding"))
			}
		})
	}
}

func TestNewRequestWithBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		path     string
		headers  string
		auth     *BasicAuth
		hasError bool
	}{
		{
			name:     "invalid url",
			base:     "::",
			hasError: true,
		},
		{
			name:     "auth is nil",
			hasError: true,
		},
		{
			name:     "auth is empty",
			auth:     &BasicAuth{},
			hasError: true,
		},
		{
			name:     "empty username",
			auth:     &BasicAuth{Password: "password"},
			hasError: true,
		},
		{
			name:     "empty password",
			auth:     &BasicAuth{Username: "username"},
			hasError: true,
		},
		{
			name:    "no error",
			auth:    &BasicAuth{Username: "username", Password: "password"},
			headers: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := NewRequestWithBasicAuth(
				"",
				"",
				test.base,
				test.path,
				"",
				"",
				"",
				nil,
				test.auth,
			)
			assert.Equal(t, test.hasError, err != nil)
			if !test.hasError {
				assert.Equal(t, test.base, req.URL.Host)
				assert.Equal(t, test.path, req.URL.Path)
				assert.Equal(t, test.headers, req.Header.Get("Authorization"))
			}
		})
	}
}

func TestNewRequestWithBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		path     string
		token    string
		hasError bool
	}{
		{
			name:     "invalid url",
			base:     "::",
			hasError: true,
		},
		{
			name:     "empty token",
			hasError: true,
		},
		{
			name:  "no error",
			token: "valid-token",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := NewRequestWithBearerToken(
				"",
				"",
				test.base,
				test.path,
				"",
				"",
				"",
				nil,
				test.token,
			)
			assert.Equal(t, test.hasError, err != nil)
			if !test.hasError {
				assert.Equal(t, test.base, req.URL.Host)
				assert.Equal(t, test.path, req.URL.Path)
				assert.Equal(t, fmt.Sprintf("Bearer token:%s", test.token), req.Header.Get("Authorization"))
			}
		})
	}
}
