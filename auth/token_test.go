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
		return nil, fmt.Errorf("we pretend that the connection failed ")
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

func (cm *memoryConfigManager) Reset(bool) error {
	cm.c = config.Config{}
	return nil
}

type mockApi struct {
	core.DefaultAPI
}

func (m *mockApi) TokenRenew(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
	return m.DefaultAPI.TokenRenew(token, idempotencyKey)
}

func (m *mockApi) NotificationCredentials(token string, appUserID string) (core.NotificationCredentialsResponse, error) {
	return core.NotificationCredentialsResponse{
		Endpoint: "",
		Username: "",
		Password: "",
	}, nil
}

func TestTokenRenewWithBadConnection(t *testing.T) {
	idempotencyKey := uuid.New()
	expiredDate := time.Now().Truncate(time.Hour)
	validDate := time.Now().Add(time.Hour)

	category.Set(t, category.Unit)

	rt := mockRoundTripper{
		expectedIdempotencyKey: idempotencyKey,
	}
	client := request.NewStdHTTP()
	client.Transport = &rt
	api := mockApi{DefaultAPI: *core.NewDefaultAPI("", "", client, response.NoopValidator{})}

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

	rc := RenewingChecker{
		cm:         &cm,
		creds:      &api,
		expChecker: systemTimeExpirationChecker{},
	}

	t.Run("valid token renewal request", func(t *testing.T) {
		rt.resp = &core.TokenRenewResponse{
			Token:      uuid.New().String(),
			RenewToken: "renew-token",
			ExpiresAt:  validDate.String(),
		}
		isLoggedIn := rc.IsLoggedIn()
		assert.True(t, isLoggedIn, "user should be logged in")
		assert.Equal(t, rt.resp.Token, cm.c.TokensData[0].Token, "token should be updated in the configuration")
	})

	t.Run("token renewal attempt with expected failure on a HTTP level", func(t *testing.T) {
		// replace the token in the config with one that is expired.
		// so the next IsLoggedIn() request should attempt a token renewal
		cm.c.TokensData[0] = config.TokenData{
			Token:              rt.resp.Token,
			RenewToken:         rt.resp.RenewToken,
			TokenExpiry:        expiredDate.String(),
			IdempotencyKey:     &idempotencyKey,
			NordLynxPrivateKey: "nordlynx-pkey",
			OpenVPNUsername:    "openvpn-username",
			OpenVPNPassword:    "openvpn-password",
		}

		// next request is a failure from our custom roundtripper,
		// make sure that the token in the configuration itself has not been changed, thus the client didn't log out
		lastToken := strings.Clone(cm.c.TokensData[0].Token)
		rt.resp = nil // setting the resp to nil means that the request will fail
		isLoggedIn := rc.IsLoggedIn()
		assert.True(t, isLoggedIn, "user should be logged in, even after a failed request")
		assert.Equal(t, lastToken, cm.c.TokensData[0].Token, "token should not be updated in the configuration after a failed request")
	})

	t.Run("valid token renewal request after a failure", func(t *testing.T) {
		rt.resp = &core.TokenRenewResponse{
			Token:      uuid.New().String(),
			RenewToken: "renew-token",
			ExpiresAt:  validDate.String(),
		}
		isLoggedIn := rc.IsLoggedIn()
		assert.True(t, isLoggedIn, "user should be logged in")
		assert.Equal(t, rt.resp.Token, cm.c.TokensData[0].Token, "token should be updated in the configuration")
	})
}
