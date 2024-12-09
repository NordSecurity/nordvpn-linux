package core

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"

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
				http.DefaultClient,
				response.NoopValidator{},
			)
			_, err := api.CurrentUser("refresh me")
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestDefaultAPI_TokenRenew(t *testing.T) {
	category.Set(t, category.Integration)

	idempotencyKey := uuid.New()
	tests := []testCase{
		testNewCase(t, http.StatusCreated, TokenRenewURL, "core_token_renew", nil),
		testNewCase(t, http.StatusBadRequest, TokenRenewURL, "core_token_renew", ErrBadRequest),
		testNewCase(t, http.StatusNotFound, TokenRenewURL, "core_token_renew", ErrNotFound),
		testNewCase(t, http.StatusInternalServerError, TokenRenewURL, "", ErrServerInternal),

		{
			name: "Idempotent request",
			handler: func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, idempotencyKey.String(), r.Header.Get("Idempotency-Key"))
				data, err := os.ReadFile(fmt.Sprintf("testdata/%s_%d.json", "core_token_renew", http.StatusCreated))
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusCreated)
				rw.Write(data)
			},
			err: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				http.DefaultClient,
				response.NoopValidator{},
			)
			_, err := api.TokenRenew("refresh me", idempotencyKey)
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
				http.DefaultClient,
				response.NoopValidator{},
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
				http.DefaultClient,
				response.NoopValidator{},
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
				http.DefaultClient,
				response.NoopValidator{},
			)
			_, err := api.ServiceCredentials("refresh me")
			assert.True(t, errors.Is(err, test.err))
		})
	}
}

func TestMaxBytes(t *testing.T) {

	randomBytes := func(size int64) []byte {
		data := make([]byte, size+1)
		for i := int64(0); i < size; i++ {
			data[size] = byte(rand.Intn(255))
		}
		return data
	}

	t.Run("too big input", func(t *testing.T) {
		input := randomBytes(maxBytesLimit + 1024)
		rc := bytes.NewReader(input)
		_, err := MaxBytesReadAll(rc)

		assert.Error(t, err)
	})

	t.Run("input with okay size", func(t *testing.T) {
		input := randomBytes(maxBytesLimit - 1024)
		rc := bytes.NewReader(input)
		output, err := MaxBytesReadAll(rc)

		assert.Nil(t, err)
		assert.Equal(t, input, output)
	})
}
