package core

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2_Login(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		regularLogin bool
		expectedURL  string
		hasError     bool
	}{
		{
			name: http.StatusText(http.StatusOK),
			handler: func(rw http.ResponseWriter, r *http.Request) {
				url, err := url.Parse(r.URL.String())
				assert.NoError(t, err)
				assert.Equal(t, url.Path, urlOAuth2Login)
				assert.True(t, url.Query().Get("preferred_flow") == "login")
				data, err := os.ReadFile("testdata/login_200.json")
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write(data)
			},
			regularLogin: true,
			expectedURL:  "https://api.nordvpn.com/v1/users/oauth/login-redirect?attempt=bfeb71e5-6c9b-459b-bc50-40d0d0186bc4",
		},
		{
			name: http.StatusText(http.StatusBadRequest),
			handler: func(rw http.ResponseWriter, r *http.Request) {
				url, err := url.Parse(r.URL.String())
				assert.NoError(t, err)
				assert.Equal(t, url.Path, urlOAuth2Login)
				assert.True(t, url.Query().Get("preferred_flow") == "login")
				data, err := os.ReadFile("testdata/login_400.json")
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write(data)
			},
			regularLogin: true,
			hasError:     true,
		},
		{
			name: http.StatusText(http.StatusOK),
			handler: func(rw http.ResponseWriter, r *http.Request) {
				url, err := url.Parse(r.URL.String())
				assert.NoError(t, err)
				assert.Equal(t, url.Path, urlOAuth2Login)
				assert.True(t, url.Query().Get("preferred_flow") == "registration")
				data, err := os.ReadFile("testdata/login_200.json")
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write(data)
			},
			regularLogin: false,
			expectedURL:  "https://api.nordvpn.com/v1/users/oauth/login-redirect?attempt=bfeb71e5-6c9b-459b-bc50-40d0d0186bc4",
			hasError:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewOAuth2(http.DefaultClient, server.URL)
			url, err := api.Login(test.regularLogin)
			assert.Equal(t, test.hasError, err != nil)
			if test.hasError {
				assert.True(t, strings.Contains(err.Error(), test.name))
			}
			assert.Equal(t, test.expectedURL, url)
		})
	}
}

func TestOAuth2_Token(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		handler  http.HandlerFunc
		hasError bool
	}{
		{
			name: http.StatusText(http.StatusOK),
			handler: func(rw http.ResponseWriter, r *http.Request) {
				assert.True(t, strings.HasPrefix(r.URL.String(), urlOAuth2Token))
				data, err := os.ReadFile("testdata/token_200.json")
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write(data)
			},
		},
		{
			name: http.StatusText(http.StatusBadRequest),
			handler: func(rw http.ResponseWriter, r *http.Request) {
				assert.True(t, strings.HasPrefix(r.URL.String(), urlOAuth2Token))
				data, err := os.ReadFile("testdata/token_400.json")
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write(data)
			},
			hasError: true,
		},
		{
			name: http.StatusText(http.StatusNotFound),
			handler: func(rw http.ResponseWriter, r *http.Request) {
				assert.True(t, strings.HasPrefix(r.URL.String(), urlOAuth2Token))
				data, err := os.ReadFile("testdata/token_404.json")
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusNotFound)
				rw.Write(data)
			},
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewOAuth2(http.DefaultClient, server.URL)
			resp, err := api.Token("exchange")
			assert.Equal(t, test.hasError, err != nil)
			if test.hasError {
				assert.True(t, strings.Contains(err.Error(), test.name))
			} else {
				assert.NotEmpty(t, resp.Token)
				assert.NotEmpty(t, resp.RenewToken)
			}
		})
	}
}
