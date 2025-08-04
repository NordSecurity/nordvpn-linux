package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mocksession "github.com/NordSecurity/nordvpn-linux/test/mock/session"
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

func TestTokenRenewWithBadConnection(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	expiredDate := time.Now().Truncate(24 * time.Hour)
	validDate := time.Now().Add(24 * time.Hour)

	rt := mockRoundTripper{
		expectedIdempotencyKey: idempotencyKey,
	}

	uid := int64(1)
	cm := memoryConfigManager{
		c: config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:              "someExpiredToken",
					TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
					RenewToken:         "renew-token",
					IdempotencyKey:     &idempotencyKey,
					NordLynxPrivateKey: "nordlynx-pkey",
					OpenVPNUsername:    "openvpn-username",
					OpenVPNPassword:    "openvpn-password",
				},
			},
		},
	}

	resetExpiryDateToOriginal := func() {
		dt := cm.c.TokensData[cm.c.AutoConnectData.ID]
		dt.TokenExpiry = time.Now().UTC().Add(-12 * time.Hour).Format(internal.ServerDateFormat)
		cm.c.TokensData[cm.c.AutoConnectData.ID] = dt
	}

	client := request.NewStdHTTP()
	client.Transport = &rt
	simpleApi := core.NewSimpleAPI("", "", client, response.NoopValidator{})
	api := &mockApi{ClientAPI: core.NewSmartClientAPI(simpleApi, &mocksession.MockSessionStore{})}

	expirationChecker := NewTokenExpirationChecker()

	errRegistry := internal.NewErrorHandlingRegistry[error]()

	rc := RenewingChecker{
		cm:         &cm,
		creds:      api,
		expChecker: expirationChecker,
		sessionStores: []session.SessionStore{session.NewAccessTokenSessionStore(
			&cm,
			errRegistry,
			func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
				resp, err := api.TokenRenew(token, idempotencyKey)
				if err != nil {
					return nil, err
				}
				return &session.AccessTokenResponse{
					Token:      resp.Token,
					RenewToken: resp.RenewToken,
					ExpiresAt:  resp.ExpiresAt,
				}, nil
			},
			nil,
		)},
	}

	// valid token renewal request
	// make sure initial test data is correct
	resetExpiryDateToOriginal()
	assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[cm.c.AutoConnectData.ID].TokenExpiry))
	rt.resp = &core.TokenRenewResponse{
		Token:      uuid.New().String(),
		RenewToken: "renewed-token",
		ExpiresAt:  validDate.Format(internal.ServerDateFormat),
	}
	isLoggedIn, err := rc.IsLoggedIn()
	assert.NoError(t, err)
	assert.True(t, isLoggedIn, "user should be logged in")
	assert.Equal(t, rt.resp.Token, cm.c.TokensData[uid].Token, "token should be updated in the configuration")
	assert.Equal(t, rt.resp.RenewToken, cm.c.TokensData[uid].RenewToken, "renew-token should be updated in the configuration")

	cm.c.TokensData[0] = config.TokenData{
		Token:              "expired-token",
		RenewToken:         "expired-renew-token",
		TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
		IdempotencyKey:     &idempotencyKey,
		NordLynxPrivateKey: "nordlynx-pkey",
		OpenVPNUsername:    "openvpn-username",
		OpenVPNPassword:    "openvpn-password",
	}

	badConnErr := errors.New("we pretend that the connection failed")
	badConnErrHandled := false
	errRegistry.Add(func(err error) {
		badConnErrHandled = true
	}, badConnErr)

	// make sure initial test data is correct
	resetExpiryDateToOriginal()
	assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[cm.c.AutoConnectData.ID].TokenExpiry))

	lastExpiredToken := strings.Clone(cm.c.TokensData[uid].Token)
	lastExpiredRenewToken := strings.Clone(cm.c.TokensData[uid].RenewToken)
	rt.resp = nil
	rt.respError = badConnErr
	isLoggedIn, err = rc.IsLoggedIn()
	assert.False(t, isLoggedIn, "user should be logged out when error has handlers")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, badConnErrHandled)
	assert.Equal(t, lastExpiredToken, cm.c.TokensData[uid].Token, "token should not be updated in the configuration after a failed request")
	assert.Equal(t, lastExpiredRenewToken, cm.c.TokensData[uid].RenewToken, "renew-token should not be updated in the configuration after a failed request")

	cm.c.TokensData[0] = config.TokenData{
		Token:              "expired-token",
		RenewToken:         "expired-renew-token",
		TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
		IdempotencyKey:     &idempotencyKey,
		NordLynxPrivateKey: "nordlynx-pkey",
		OpenVPNUsername:    "openvpn-username",
		OpenVPNPassword:    "openvpn-password",
	}

	badConnErrHandled = false
	// make sure initial test data is correct
	resetExpiryDateToOriginal()
	assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[cm.c.AutoConnectData.ID].TokenExpiry))

	rt.resp = nil
	rt.respError = badConnErr
	rc.IsLoggedIn()

	rt.resp = &core.TokenRenewResponse{
		Token:      uuid.New().String(),
		RenewToken: "renew-token",
		ExpiresAt:  validDate.Format(internal.ServerDateFormat),
	}
	isLoggedIn, _ = rc.IsLoggedIn()
	assert.True(t, isLoggedIn, "user should be logged in")
	assert.Equal(t, rt.resp.Token, cm.c.TokensData[uid].Token, "token should be updated in the configuration")
	assert.Equal(t, rt.resp.RenewToken, cm.c.TokensData[uid].RenewToken, "renew-token should be updated in the configuration")
}

func Test_TokenRenewForcesUserLogout(t *testing.T) {
	category.Set(t, category.Unit)

	idempotencyKey := uuid.New()
	expiredDate := time.Now().Truncate(24 * time.Hour)
	rt := mockRoundTripper{expectedIdempotencyKey: idempotencyKey}
	uid := int64(1)

	cm := memoryConfigManager{
		c: config.Config{
			TokensData: map[int64]config.TokenData{},
		},
	}

	client := request.NewStdHTTP()
	client.Transport = &rt
	simpleApi := core.NewSimpleAPI("", "", client, response.NoopValidator{})
	api := &mockApi{ClientAPI: core.NewSmartClientAPI(simpleApi, &mocksession.MockSessionStore{})}

	expirationChecker := NewTokenExpirationChecker()
	errRegistry := internal.NewErrorHandlingRegistry[error]()
	errRegistry.Add(func(s error) {
		delete(cm.c.TokensData, uid)
	}, core.ErrBadRequest, core.ErrNotFound)

	rc := RenewingChecker{
		cm:         &cm,
		creds:      api,
		expChecker: expirationChecker,
		sessionStores: []session.SessionStore{session.NewAccessTokenSessionStore(
			&cm,
			errRegistry,
			func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
				resp, err := api.TokenRenew(token, idempotencyKey)
				if err != nil {
					return nil, err
				}
				return &session.AccessTokenResponse{
					Token:      resp.Token,
					RenewToken: resp.RenewToken,
					ExpiresAt:  resp.ExpiresAt,
				}, nil
			},
			nil,
		)},
	}

	resetExpiryDateToOriginal := func() {
		dt := cm.c.TokensData[cm.c.AutoConnectData.ID]
		dt.TokenExpiry = time.Now().UTC().Add(-12 * time.Hour).Format(internal.ServerDateFormat)
		cm.c.TokensData[cm.c.AutoConnectData.ID] = dt
	}

	randomErrs := []error{
		errors.New("very naughty error"),
		errors.New("never happens"),
		errors.New("worked on my machine"),
	}

	badConnErrHandled := false
	errRegistry.Add(func(err error) {
		badConnErrHandled = true
		delete(cm.c.TokensData, uid)
	}, randomErrs...)

	for _, exptectedErr := range randomErrs {
		badConnErrHandled = false
		// replace the token in the config with one that is expired.
		// so the next IsLoggedIn() request should attempt a token renewal
		cm.c.AutoConnectData.ID = uid
		cm.c.TokensData[uid] = config.TokenData{
			Token:              "expired-token",
			RenewToken:         "expired-renew-token",
			TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
			IdempotencyKey:     &idempotencyKey,
			NordLynxPrivateKey: "nordlynx-pkey",
			OpenVPNUsername:    "openvpn-username",
			OpenVPNPassword:    "openvpn-password",
		}

		// make sure initial test data is correct
		resetExpiryDateToOriginal()
		assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[uid].TokenExpiry))

		rt.resp = nil
		rt.respError = exptectedErr
		isLoggedIn, err := rc.IsLoggedIn()

		assert.False(t, isLoggedIn, "user should be logged out")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handling session error")
		assert.True(t, badConnErrHandled)
		assert.Empty(t, cm.c.TokensData[uid].Token, "token should be removed from the configuration")
		assert.Empty(t, cm.c.TokensData[uid].RenewToken, "renew-token should be removed from the configuration")
	}

	badConnErrHandled = false
	cm.c.TokensData[uid] = config.TokenData{
		Token:              "expired-token",
		RenewToken:         "expired-renew-token",
		TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
		IdempotencyKey:     &idempotencyKey,
		NordLynxPrivateKey: "nordlynx-pkey",
		OpenVPNUsername:    "openvpn-username",
		OpenVPNPassword:    "openvpn-password",
	}

	// make sure initial test data is correct
	resetExpiryDateToOriginal()
	assert.True(t, expirationChecker.IsExpired(cm.c.TokensData[uid].TokenExpiry))

	rt.resp = nil
	rt.respError = errors.New("the new bad")
	isLoggedIn, err := rc.IsLoggedIn()

	assert.True(t, isLoggedIn, "user should remain logged in when error has no handler")
	assert.NoError(t, err, "no error should be returned when no handler is registered")
	assert.False(t, badConnErrHandled)
	assert.NotEmpty(t, cm.c.TokensData[uid].Token, "token should not be removed when no handler is registered")
	assert.NotEmpty(t, cm.c.TokensData[uid].RenewToken, "renew-token should not be removed when no handler is registered")
}
