package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	resp                   *core.TokenRenewResponse
	respError              error
	expectedIdempotencyKey uuid.UUID
}

func (rt *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path != "/v1/users/tokens/renew" {
		panic(fmt.Errorf("requested on an unexpected path: `%s`", req.URL.Path))
	}

	idempotencyKey, ok := req.Header["Idempotency-Key"]
	if !ok {
		panic("missing Idempotency-Key header in the request")
	} else if idempotencyKey[0] != rt.expectedIdempotencyKey.String() {
		panic(fmt.Errorf("expected idempotency key %s, got %s", rt.expectedIdempotencyKey, idempotencyKey[0]))
	}

	if rt.resp == nil {
		return nil, rt.respError
	} else {
		body, err := json.Marshal(*rt.resp)
		if err != nil {
			panic(err)
		}
		httpResponse := &http.Response{
			Status:     http.StatusText(http.StatusCreated),
			StatusCode: http.StatusCreated,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:          io.NopCloser(bytes.NewReader(body)),
			ContentLength: int64(len(body)),
		}

		return httpResponse, nil
	}
}

type memoryConfigManager struct {
	c config.Config
}

func (cm *memoryConfigManager) SaveWith(f config.SaveFunc) error {
	cm.c = f(cm.c)
	return nil
}

func (cm *memoryConfigManager) Load(c *config.Config) error {
	*c = cm.c
	return nil
}

func (cm *memoryConfigManager) Reset(bool, bool) error {
	cm.c = config.Config{}
	return nil
}

type mockApi struct {
	core.ClientAPI
}

func (m *mockApi) TokenRenew(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
	return m.ClientAPI.TokenRenew(token, idempotencyKey)
}

func (m *mockApi) NotificationCredentials(appUserID string) (core.NotificationCredentialsResponse, error) {
	return core.NotificationCredentialsResponse{
		Endpoint: "",
		Username: "",
		Password: "",
	}, nil
}

type mockLoginTokenManager struct {
	core.TokenManager
}

func TestTokenRenewWithBadConnection(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	expiredDate := time.Now().Truncate(time.Hour)
	validDate := time.Now().Add(time.Hour)

	rt := mockRoundTripper{
		expectedIdempotencyKey: idempotencyKey,
	}

	cm := memoryConfigManager{
		c: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 1},
			TokensData: map[int64]config.TokenData{
				0: {
					Token:              "someExpiredToken",
					TokenExpiry:        expiredDate.String(),
					RenewToken:         "renew-token",
					IdempotencyKey:     &idempotencyKey,
					NordLynxPrivateKey: "nordlynx-pkey",
					OpenVPNUsername:    "openvpn-username",
					OpenVPNPassword:    "openvpn-password",
				},
			},
		},
	}

	client := request.NewStdHTTP()
	client.Transport = &rt
	simpleApi := core.NewSimpleAPI("", "", client, response.NoopValidator{})
	api := &mockApi{ClientAPI: core.NewSmartClientAPI(simpleApi, &mockLoginTokenManager{})}

	expirationChecker := NewTokenExpirationChecker()

	rc := RenewingChecker{
		cm:         &cm,
		creds:      api,
		expChecker: expirationChecker,
		loginTokenManager: core.NewLoginTokenManager(
			&cm,
			api.TokenRenew,
			core.NewErrorHandlingRegistry[int64](),
			core.NewLoginTokenValidator(simpleApi, expirationChecker),
		),
	}

	t.Run("valid token renewal request", func(t *testing.T) {
		// make sure initial test data is correct
		assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[cm.c.AutoConnectData.ID].TokenExpiry))
		rt.resp = &core.TokenRenewResponse{
			Token:      uuid.New().String(),
			RenewToken: "renewed-token",
			ExpiresAt:  validDate.String(),
		}
		isLoggedIn, err := rc.IsLoggedIn()
		assert.NoError(t, err)
		assert.True(t, isLoggedIn, "user should be logged in")
		assert.Equal(t, rt.resp.Token, cm.c.TokensData[0].Token, "token should be updated in the configuration")
		assert.Equal(t, rt.resp.RenewToken, cm.c.TokensData[0].RenewToken, "renew-token should be updated in the configuration")
	})

	t.Run("token renewal attempt with expected failure on a HTTP level", func(t *testing.T) {
		// replace the token in the config with one that is expired.
		// so the next IsLoggedIn() request should attempt a token renewal
		cm.c.TokensData[0] = config.TokenData{
			Token:              "expired-token",
			RenewToken:         "expired-renew-token",
			TokenExpiry:        expiredDate.String(),
			IdempotencyKey:     &idempotencyKey,
			NordLynxPrivateKey: "nordlynx-pkey",
			OpenVPNUsername:    "openvpn-username",
			OpenVPNPassword:    "openvpn-password",
		}

		// make sure initial test data is correct
		assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[cm.c.AutoConnectData.ID].TokenExpiry))

		// next request is a failure from our custom roundtripper,
		// make sure that the token in the configuration itself has not been changed, thus the client didn't log out
		lastExpiredToken := strings.Clone(cm.c.TokensData[0].Token)
		lastExpiredRenewToken := strings.Clone(cm.c.TokensData[0].RenewToken)
		rt.resp = nil // setting the resp to nil means that the request will fail
		rt.respError = fmt.Errorf("we pretend that the connection failed")
		isLoggedIn, _ := rc.IsLoggedIn()
		assert.True(t, isLoggedIn, "user should be logged in, even after a failed request")
		assert.Equal(t, lastExpiredToken, cm.c.TokensData[0].Token, "token should not be updated in the configuration after a failed request")
		assert.Equal(t, lastExpiredRenewToken, cm.c.TokensData[0].RenewToken, "renew-token should not be updated in the configuration after a failed request")
	})

	t.Run("valid token renewal request after a failure", func(t *testing.T) {
		// replace the token in the config with one that is expired.
		// so the next IsLoggedIn() request should attempt a token renewal
		cm.c.TokensData[0] = config.TokenData{
			Token:              "expired-token",
			RenewToken:         "expired-renew-token",
			TokenExpiry:        expiredDate.String(),
			IdempotencyKey:     &idempotencyKey,
			NordLynxPrivateKey: "nordlynx-pkey",
			OpenVPNUsername:    "openvpn-username",
			OpenVPNPassword:    "openvpn-password",
		}

		// make sure initial test data is correct
		assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[cm.c.AutoConnectData.ID].TokenExpiry))

		// next request is a failure from our custom roundtripper
		rt.resp = nil
		rt.respError = fmt.Errorf("we pretend that the connection failed")
		rc.IsLoggedIn()

		rt.resp = &core.TokenRenewResponse{
			Token:      uuid.New().String(),
			RenewToken: "renew-token",
			ExpiresAt:  validDate.String(),
		}
		isLoggedIn, _ := rc.IsLoggedIn()
		assert.True(t, isLoggedIn, "user should be logged in")
		assert.Equal(t, rt.resp.Token, cm.c.TokensData[0].Token, "token should be updated in the configuration")
		assert.Equal(t, rt.resp.RenewToken, cm.c.TokensData[0].RenewToken, "renew-token should be updated in the configuration")
	})
}

func Test_TokenRenewForcesUserLogout(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	expiredDate := time.Now().Truncate(time.Hour)
	rt := mockRoundTripper{expectedIdempotencyKey: idempotencyKey}

	cm := memoryConfigManager{
		c: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 1},
			TokensData: map[int64]config.TokenData{
				0: {
					Token:              "someExpiredToken",
					TokenExpiry:        expiredDate.String(),
					RenewToken:         "renew-token",
					IdempotencyKey:     &idempotencyKey,
					NordLynxPrivateKey: "nordlynx-pkey",
					OpenVPNUsername:    "openvpn-username",
					OpenVPNPassword:    "openvpn-password",
				},
			},
		},
	}

	client := request.NewStdHTTP()
	client.Transport = &rt
	simpleApi := core.NewSimpleAPI("", "", client, response.NoopValidator{})
	api := &mockApi{ClientAPI: core.NewSmartClientAPI(simpleApi, &mockLoginTokenManager{})}

	expirationChecker := NewTokenExpirationChecker()

	rc := RenewingChecker{
		cm:         &cm,
		creds:      api,
		expChecker: expirationChecker,
		loginTokenManager: core.NewLoginTokenManager(
			&cm,
			api.TokenRenew,
			core.NewErrorHandlingRegistry[int64](),
			core.NewLoginTokenValidator(simpleApi, expirationChecker),
		),
	}

	t.Run("token renewal attempt with log-out invoking errors on a HTTP level", func(t *testing.T) {
		errs := []error{core.ErrUnauthorized, core.ErrNotFound, core.ErrBadRequest}
		for _, exptectedErr := range errs {
			// replace the token in the config with one that is expired.
			// so the next IsLoggedIn() request should attempt a token renewal
			cm.c.TokensData[0] = config.TokenData{
				Token:              "expired-token",
				RenewToken:         "expired-renew-token",
				TokenExpiry:        expiredDate.String(),
				IdempotencyKey:     &idempotencyKey,
				NordLynxPrivateKey: "nordlynx-pkey",
				OpenVPNUsername:    "openvpn-username",
				OpenVPNPassword:    "openvpn-password",
			}

			// make sure initial test data is correct
			assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[cm.c.AutoConnectData.ID].TokenExpiry))

			// next request is a failure from our custom roundtripper
			rt.resp = nil
			rt.respError = exptectedErr
			isLoggedIn, _ := rc.IsLoggedIn()

			assert.False(t, isLoggedIn, "user should be logged out")
			assert.Empty(t, cm.c.TokensData[0].Token, "token should be removed from the configuration")
			assert.Empty(t, cm.c.TokensData[0].RenewToken, "renew-token should be removed from the configuration")
		}
	})
}
