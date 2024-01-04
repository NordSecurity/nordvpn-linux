package request

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
