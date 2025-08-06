package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	defaultTokenValue     = "valid-token"
	renewedTokenValue     = "new-token-value"
	testUsername          = "testuser"
	testEmail             = "test@example.com"
	defaultExpirationDays = 24 * time.Hour
)

type mockConfigManager struct {
	config    config.Config
	loadError error
	saveError error
}

func (m *mockConfigManager) Load(cfg *config.Config) error {
	if m.loadError != nil {
		return m.loadError
	}
	*cfg = m.config
	return nil
}

func (m *mockConfigManager) SaveWith(f config.SaveFunc) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.config = f(m.config)
	return nil
}

func (m *mockConfigManager) Reset(preserveLoginData bool, disableKillswitch bool) error {
	return nil
}

func (m *mockConfigManager) SetTokenExpiry(value time.Time) {
	dt := m.config.TokensData[m.config.AutoConnectData.ID]
	dt.TokenExpiry = value.Format(internal.ServerDateFormat)
	m.config.TokensData[m.config.AutoConnectData.ID] = dt
}

type mockRoundTripperWithExpiration struct {
	renewResp              *core.TokenRenewResponse
	respError              error
	expectedIdempotencyKey uuid.UUID
	createdAt              time.Time
	currentUserResp        *core.CurrentUserResponse
	RoundTripFunc          func(req *http.Request) (*http.Response, error)
}

func (rt *mockRoundTripperWithExpiration) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.RoundTripFunc(req)
}

func (rt *mockRoundTripperWithExpiration) ForceExpiration() {
	rt.createdAt = time.Now().Add(-12 * defaultExpirationDays)
}

func (rt *mockRoundTripperWithExpiration) ResetExpiration() {
	rt.createdAt = time.Now()
}

func createUnauthorizedResponse() (*http.Response, error) {
	errorBody := `{
        "errors": {
            "code": 101301,
            "message": "Unauthorized"
        }
    }`

	return &http.Response{
		Status:     http.StatusText(http.StatusUnauthorized),
		StatusCode: http.StatusUnauthorized,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:          io.NopCloser(bytes.NewReader([]byte(errorBody))),
		ContentLength: int64(len(errorBody)),
	}, nil
}

func createDefaultRoundTripper(rt *mockRoundTripperWithExpiration) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v1/users/tokens/renew" && time.Since(rt.createdAt) > defaultExpirationDays {
			return createUnauthorizedResponse()
		}

		switch req.URL.Path {
		case "/v1/users/tokens/renew":
			idempotencyKey, ok := req.Header["Idempotency-Key"]
			if !ok {
				panic("missing Idempotency-Key header in the request")
			} else if idempotencyKey[0] != rt.expectedIdempotencyKey.String() {
				panic(fmt.Errorf("expected idempotency key %s, got %s", rt.expectedIdempotencyKey, idempotencyKey[0]))
			}

			if rt.renewResp == nil {
				return nil, rt.respError
			} else {
				rt.ResetExpiration()

				body, err := json.Marshal(*rt.renewResp)
				if err != nil {
					panic(err)
				}
				return &http.Response{
					Status:     http.StatusText(http.StatusCreated),
					StatusCode: http.StatusCreated,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body:          io.NopCloser(bytes.NewReader(body)),
					ContentLength: int64(len(body)),
				}, nil
			}

		case "/v1/users/current":
			if rt.currentUserResp == nil {
				return nil, rt.respError
			} else {
				body, err := json.Marshal(rt.currentUserResp)
				if err != nil {
					panic(err)
				}
				return &http.Response{
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body:          io.NopCloser(bytes.NewReader(body)),
					ContentLength: int64(len(body)),
				}, nil
			}

		default:
			panic(fmt.Errorf("requested on an unexpected path: `%s`", req.URL.Path))
		}
	}
}

func setupTestEnvironment(
	idempotencyKey uuid.UUID,
	tokenExpiry string,
	saveError error,
) (core.ClientAPI, *mockRoundTripperWithExpiration, *mockConfigManager) {
	renewResp := &core.TokenRenewResponse{
		Token:     renewedTokenValue,
		ExpiresAt: time.Now().UTC().Add(defaultExpirationDays).Format(internal.ServerDateFormat),
	}

	currentUserResp := &core.CurrentUserResponse{
		Username: testUsername,
		Email:    testEmail,
	}

	rt := &mockRoundTripperWithExpiration{
		renewResp:              renewResp,
		currentUserResp:        currentUserResp,
		respError:              nil,
		expectedIdempotencyKey: idempotencyKey,
		createdAt:              time.Now(),
	}
	//nolint:bodyclose
	rt.RoundTripFunc = createDefaultRoundTripper(rt)

	mockCfg := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 1},
			TokensData: map[int64]config.TokenData{
				1: {
					Token:          defaultTokenValue,
					TokenExpiry:    tokenExpiry,
					IdempotencyKey: &idempotencyKey,
				},
			},
		},
		saveError: saveError,
	}

	client := request.NewStdHTTP()
	client.Transport = rt

	simpleApi := core.NewSimpleAPI("", "", client, response.NoopValidator{})
	sessionStore := session.NewAccessTokenSessionStore(
		mockCfg,
		internal.NewErrorHandlingRegistry[error](),
		func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
			resp, err := simpleApi.TokenRenew(token, idempotencyKey)
			if err == nil {
				return &session.AccessTokenResponse{
					Token:      resp.Token,
					RenewToken: resp.RenewToken,
					ExpiresAt:  resp.ExpiresAt,
				}, nil
			}

			return nil, err
		},
		nil, // no external validator needed for this test
	)

	return core.NewSmartClientAPI(simpleApi, sessionStore), rt, mockCfg
}

func Test_TokenExpiration(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	futureExpiry := time.Now().UTC().Add(defaultExpirationDays).Format(internal.ServerDateFormat)

	api, mockRT, mockCfg := setupTestEnvironment(idempotencyKey, futureExpiry, nil)

	resp, err := api.CurrentUser()
	assert.NoError(t, err)
	assert.Equal(t, testUsername, resp.Username)
	assert.Equal(t, testEmail, resp.Email)

	mockRT.ForceExpiration()
	mockCfg.SetTokenExpiry(time.Now().UTC().Add(-12 * time.Hour))
	resp, err = api.CurrentUser()
	assert.NoError(t, err)
	assert.Equal(t, testUsername, resp.Username)
	assert.Equal(t, testEmail, resp.Email)
}

func Test_TokenRenewalWithNetworkError(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	pastExpiry := time.Now().UTC().Add(-12 * time.Hour).Format(internal.ServerDateFormat)

	// Create a custom round tripper that returns network error for renewal
	rt := &mockRoundTripperWithExpiration{
		renewResp:              nil,
		currentUserResp:        nil,
		respError:              errors.New("network error"),
		expectedIdempotencyKey: idempotencyKey,
		createdAt:              time.Now().Add(-2 * defaultExpirationDays), // Already expired
	}

	rt.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v1/users/tokens/renew":
			// Return network error for renewal attempt
			return nil, rt.respError
		case "/v1/users/current":
			// First call should return unauthorized to trigger renewal
			return createUnauthorizedResponse()
		default:
			panic(fmt.Errorf("unexpected path: %s", req.URL.Path))
		}
	}

	mockCfg := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 1},
			TokensData: map[int64]config.TokenData{
				1: {
					Token:          defaultTokenValue,
					TokenExpiry:    pastExpiry,
					IdempotencyKey: &idempotencyKey,
				},
			},
		},
	}

	client := request.NewStdHTTP()
	client.Transport = rt

	simpleApi := core.NewSimpleAPI("", "", client, response.NoopValidator{})
	sessionStore := session.NewAccessTokenSessionStore(
		mockCfg,
		internal.NewErrorHandlingRegistry[error](),
		func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
			resp, err := simpleApi.TokenRenew(token, idempotencyKey)
			if err == nil {
				return &session.AccessTokenResponse{
					Token:      resp.Token,
					RenewToken: resp.RenewToken,
					ExpiresAt:  resp.ExpiresAt,
				}, nil
			}
			return nil, err
		},
		nil,
	)

	api := core.NewSmartClientAPI(simpleApi, sessionStore)

	resp, err := api.CurrentUser()
	assert.Error(t, err)
	assert.ErrorIs(t, err, rt.respError)
	assert.Nil(t, resp)
}

func Test_TokenRenewalWithInvalidToken(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	pastExpiry := time.Now().UTC().Add(-12 * time.Hour).Format(internal.ServerDateFormat)

	api, mockRT, _ := setupTestEnvironment(idempotencyKey, pastExpiry, nil)

	mockRT.renewResp = nil
	mockRT.respError = core.ErrNotFound
	mockRT.ForceExpiration()

	resp, err := api.CurrentUser()
	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrNotFound)
	assert.Nil(t, resp)
}

func Test_ConfigSaveErrorDuringRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	pastExpiry := time.Now().UTC().Add(-12 * time.Hour).Format(internal.ServerDateFormat)

	api, mockRT, _ := setupTestEnvironment(idempotencyKey, pastExpiry, errors.New("failed to save config"))
	mockRT.ForceExpiration()

	resp, err := api.CurrentUser()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save config")
	assert.Nil(t, resp)
}
