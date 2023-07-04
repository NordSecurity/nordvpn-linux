package core

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name    string
	handler http.HandlerFunc
	err     error
}

func testNewCase(t *testing.T, httpStatus int, url, fixture string, err error) testCase {
	t.Helper()

	var handler http.HandlerFunc
	switch fixture {
	case "":
		handler = func(rw http.ResponseWriter, r *http.Request) {
			assert.Equal(t, url, r.URL.String())
			rw.WriteHeader(httpStatus)
		}
	default:
		handler = func(rw http.ResponseWriter, r *http.Request) {
			assert.Equal(t, url, r.URL.String())
			data, err := os.ReadFile(fmt.Sprintf("testdata/%s_%d.json", fixture, httpStatus))
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			rw.WriteHeader(httpStatus)
			rw.Write(data)
		}
	}

	return testCase{name: http.StatusText(httpStatus), handler: handler, err: err}
}

func TestCreateUser(t *testing.T) {
	category.Set(t, category.Integration)

	email := "email@one.lt"

	api := testNewDefaultAPI(GeneralInfo)
	resp, err := api.CreateUser(email, "securepswd")
	assert.NoError(t, err)
	assert.Equal(t, email, resp.Email)
	assert.Equal(t, email, resp.Username)
}

func TestCreateUser_Error(t *testing.T) {
	category.Set(t, category.Integration)

	api := testNewDefaultAPI(InvalidInfo)
	_, err := api.CreateUser("email@one.lt", "securepswd")
	assert.Error(t, err)
}

func TestPlans(t *testing.T) {
	category.Set(t, category.Integration)

	api := testNewDefaultAPI(GeneralInfo)
	_, err := api.Plans()
	assert.NoError(t, err)
}

func TestPlans_Error(t *testing.T) {
	category.Set(t, category.Integration)

	api := testNewDefaultAPI(InvalidInfo)
	_, err := api.Plans()
	assert.Error(t, err)
}

func TestDefaultAPI_CurrentUser(t *testing.T) {
	category.Set(t, category.Integration)
	tests := []testCase{
		testNewCase(t, http.StatusOK, CurrentUserURL, "core_current_user", nil),
		testNewCase(t, http.StatusInternalServerError, CurrentUserURL, "", ErrServerInternal),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.CurrentUser("refresh me")
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestDefaultAPI_TokenRenew(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []testCase{
		testNewCase(t, http.StatusCreated, TokenRenewURL, "core_token_renew", nil),
		testNewCase(t, http.StatusBadRequest, TokenRenewURL, "core_token_renew", ErrBadRequest),
		testNewCase(t, http.StatusNotFound, TokenRenewURL, "core_token_renew", ErrNotFound),
		testNewCase(t, http.StatusInternalServerError, TokenRenewURL, "", ErrServerInternal),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.TokenRenew("refresh me")
			assert.True(t, errors.Is(err, test.err))
		})
	}
}

func TestDefaultAPI_Servers(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []testCase{
		testNewCase(t, http.StatusOK, ServersURL+ServersURLConnectQuery, "core_servers", nil),
		testNewCase(t, http.StatusInternalServerError, ServersURL+ServersURLConnectQuery, "", ErrServerInternal),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, _, err := api.Servers()
			assert.True(t, errors.Is(err, test.err))
		})
	}
}

func TestDefaultAPI_Services(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []testCase{
		testNewCase(t, http.StatusOK, ServicesURL, "core_services", nil),
		testNewCase(t, http.StatusInternalServerError, ServicesURL, "", ErrServerInternal),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.Services("refresh me")
			assert.True(t, errors.Is(err, test.err))
		})
	}
}

func TestDefaultAPI_ServiceCredentials(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []testCase{
		testNewCase(t, http.StatusOK, CredentialsURL, "core_service_credentials", nil),
		testNewCase(t, http.StatusInternalServerError, CredentialsURL, "", ErrServerInternal),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.ServiceCredentials("refresh me")
			assert.True(t, errors.Is(err, test.err))
		})
	}
}
