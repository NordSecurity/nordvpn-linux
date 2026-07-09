package core

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, r.URL.Path, urlOAuth2Login)
				var body loginBody
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.NotEmpty(t, body.Challenge)
				assert.Equal(t, "login", body.PreferredFlow)
				assert.Equal(t, "default", body.RedirectFlow)
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
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, r.URL.Path, urlOAuth2Login)
				var body loginBody
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.NotEmpty(t, body.Challenge)
				assert.Equal(t, "login", body.PreferredFlow)
				assert.Equal(t, "default", body.RedirectFlow)
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
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, r.URL.Path, urlOAuth2Login)
				var body loginBody
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.NotEmpty(t, body.Challenge)
				assert.Equal(t, "registration", body.PreferredFlow)
				assert.Equal(t, "default", body.RedirectFlow)
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

			api := NewOAuth2(http.DefaultClient, server.URL, response.NoopValidator{})
			url, err := api.Login(test.regularLogin)
			assert.Equal(t, test.hasError, err != nil)
			if test.hasError {
				assert.True(t, strings.Contains(err.Error(), test.name))
			}
			assert.Equal(t, test.expectedURL, url)
		})
	}
}

func TestOAuth2_Login_Validation(t *testing.T) {
	category.Set(t, category.Unit)

	loginFixture, err := os.ReadFile("testdata/login_200.json")
	require.NoError(t, err)

	tests := []struct {
		name         string
		validatorErr error
		wantEmptyURI bool
		wantCode     int
		wantBody     []byte
		wantHeaders  http.Header
	}{
		{
			name:         "validator rejection blocks login",
			validatorErr: errors.New("signature mismatch"),
			wantEmptyURI: true,
		},
		{
			name:     "validator receives correct status code",
			wantCode: http.StatusOK,
		},
		{
			name:     "validator receives unmodified response body",
			wantBody: loginFixture,
		},
		{
			name:        "validator receives response headers",
			wantHeaders: http.Header{"Content-Type": []string{"application/json"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusOK)
				rw.Write(loginFixture)
			}))
			defer server.Close()

			spy := &spyValidator{err: tt.validatorErr}
			api := NewOAuth2(http.DefaultClient, server.URL, spy)

			uri, err := api.Login(true)

			if tt.validatorErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.validatorErr)
				assert.NotZero(t, spy.calledWith.code)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantEmptyURI {
				assert.Empty(t, uri)
			}

			if tt.wantCode != 0 {
				assert.Equal(t, tt.wantCode, spy.calledWith.code)
			}

			if tt.wantBody != nil {
				assert.JSONEq(t, string(tt.wantBody), string(spy.calledWith.body))
			}

			for key, vals := range tt.wantHeaders {
				assert.Equal(t, vals, spy.calledWith.headers[key])
			}
		})
	}
}

func TestOAuth2_Token_Validation(t *testing.T) {
	category.Set(t, category.Unit)

	tokenFixture, err := os.ReadFile("testdata/token_200.json")
	require.NoError(t, err)

	tests := []struct {
		name         string
		validatorErr error
		wantNilResp  bool
		wantCode     int
		wantBody     []byte
		wantHeaders  http.Header
	}{
		{
			name:         "validator rejection blocks token exchange",
			validatorErr: errors.New("signature mismatch"),
			wantNilResp:  true,
		},
		{
			name:     "validator receives correct status code",
			wantCode: http.StatusOK,
		},
		{
			name:     "validator receives unmodified response body",
			wantBody: tokenFixture,
		},
		{
			name:        "validator receives response headers",
			wantHeaders: http.Header{"Content-Type": []string{"application/json"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusOK)
				rw.Write(tokenFixture)
			}))
			defer server.Close()

			spy := &spyValidator{err: tt.validatorErr}
			api := NewOAuth2(http.DefaultClient, server.URL, spy)

			resp, err := api.Token("exchange")

			if tt.validatorErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.validatorErr)
				assert.NotZero(t, spy.calledWith.code)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNilResp {
				assert.Nil(t, resp)
			}

			if tt.wantCode != 0 {
				assert.Equal(t, tt.wantCode, spy.calledWith.code)
			}

			if tt.wantBody != nil {
				assert.JSONEq(t, string(tt.wantBody), string(spy.calledWith.body))
			}

			for key, vals := range tt.wantHeaders {
				assert.Equal(t, vals, spy.calledWith.headers[key])
			}
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
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, r.URL.Path, urlOAuth2Token)
				assert.Equal(t, "exchange", r.URL.Query().Get("exchange_token"))
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
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, r.URL.Path, urlOAuth2Token)
				assert.Equal(t, "exchange", r.URL.Query().Get("exchange_token"))
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
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, r.URL.Path, urlOAuth2Token)
				assert.Equal(t, "exchange", r.URL.Query().Get("exchange_token"))
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

			api := NewOAuth2(http.DefaultClient, server.URL, response.NoopValidator{})
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

type spyValidator struct {
	err        error
	calledWith calledWith
}

type calledWith struct {
	code    int
	headers http.Header
	body    []byte
}

func (v *spyValidator) Validate(code int, headers http.Header, body []byte) error {
	v.calledWith.code = code
	v.calledWith.headers = headers
	v.calledWith.body = body
	return v.err
}
