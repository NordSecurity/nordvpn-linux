package session_test

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccessTokenSessionStore_Renew_NotExpired(t *testing.T) {
	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "ab78bb36299d442fa0715fb53b5e3e57", // valid hex token
				RenewToken:  "renew",
				TokenExpiry: futureTime.Format(internal.ServerDateFormat),
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

	err := store.Renew()
	assert.NoError(t, err)
}

func TestAccessTokenSessionStore_Renew_NoTokenData(t *testing.T) {
	uid := int64(123)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData:      map[int64]config.TokenData{},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

	err := store.Renew()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no token data")
}

func TestAccessTokenSessionStore_Renew_ConfigLoadError(t *testing.T) {
	cfgManager := &mock.ConfigManager{
		LoadErr: errors.New("config load error"),
	}

	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

	err := store.Renew()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config load error")
}

func TestAccessTokenSessionStore_Renew_ExternalValidatorError(t *testing.T) {
	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	validHexToken := "ab78bb36299d442fa0715fb53b5e3e57"

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       validHexToken,
				RenewToken:  "renew",
				TokenExpiry: futureTime.Format(internal.ServerDateFormat),
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewalCalled := false
	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		renewalCalled = true
		return &session.AccessTokenResponse{
			Token:      "new-token",
			RenewToken: "new-renew",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	externalValidator := func(token string) error {
		return errors.New("external validation failed")
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, externalValidator)
	err := store.Renew()

	assert.NoError(t, err)
	assert.True(t, renewalCalled, "Should renew because external validation failed")
}

func TestAccessTokenSessionStore_Renew_ExpiredToken(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().UTC().Add(-24 * time.Hour)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		assert.Equal(t, "old-token", token)
		assert.Equal(t, idempotencyKey, key)
		return &session.AccessTokenResponse{
			Token:      "new-token",
			RenewToken: "new-renew",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.NoError(t, err)
	assert.Equal(t, "new-token", cfgManager.Cfg.TokensData[uid].Token)
	assert.Equal(t, "new-renew", cfgManager.Cfg.TokensData[uid].RenewToken)
	assert.Equal(t, futureTime.Format(internal.ServerDateFormat), cfgManager.Cfg.TokensData[uid].TokenExpiry)
}

func TestAccessTokenSessionStore_Renew_SetIdempotencyKey(t *testing.T) {
	uid := int64(123)
	pastTime := time.Now().UTC().Add(-24 * time.Hour)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: nil, // No idempotency key
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return &session.AccessTokenResponse{
			Token:      "new-token",
			RenewToken: "new-renew",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.NoError(t, err)
	assert.NotNil(t, cfgManager.Cfg.TokensData[uid].IdempotencyKey)
}

func TestAccessTokenSessionStore_Renew_APIErrorWithHandler(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().UTC().Add(-24 * time.Hour)
	apiError := errors.New("api-error")

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	handlerCalled := false
	errorRegistry.Add(func(reason error) {
		handlerCalled = true
	}, apiError)

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return nil, apiError
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_Renew_APIErrorNoHandler(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return nil, errors.New("unhandled-api-error")
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error: unhandled-api-error")
}

func TestAccessTokenSessionStore_Renew_NilAPIResponse(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return nil, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "renewal API returned nil response")
}

func TestAccessTokenSessionStore_Renew_InvalidExpiryFormat(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return &session.AccessTokenResponse{
			Token:      "new-token",
			RenewToken: "new-renew",
			ExpiresAt:  "invalid-date-format",
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing expiry time")
}

func TestAccessTokenSessionStore_Renew_ExternalValidatorSuccess(t *testing.T) {
	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "de62575eaaa54ca8bd9416d98bdc9c1c",
				RenewToken:  "renew",
				TokenExpiry: futureTime.Format(internal.ServerDateFormat),
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	validatorCalled := false
	externalValidator := func(token string) error {
		validatorCalled = true
		assert.Equal(t, "de62575eaaa54ca8bd9416d98bdc9c1c", token)
		return nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, externalValidator)
	err := store.Renew()

	assert.NoError(t, err)
	assert.True(t, validatorCalled)
}

func TestAccessTokenSessionStore_Renew_ExternalValidatorFailure(t *testing.T) {
	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "de62575eaaa54ca8bd9416d98bdc9c1c",
				RenewToken:  "renew",
				TokenExpiry: futureTime.Format(internal.ServerDateFormat),
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	externalValidator := func(token string) error {
		return errors.New("token validation failed")
	}

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return &session.AccessTokenResponse{
			Token:      "new-token",
			RenewToken: "new-renew",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, externalValidator)
	err := store.Renew()

	assert.NoError(t, err)
	assert.Equal(t, "new-token", cfgManager.Cfg.TokensData[uid].Token)
}

func TestAccessTokenSessionStore_Renew_ErrNotFoundWithHandler(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	handlerCalled := false
	errorRegistry.Add(func(reason error) {
		handlerCalled = true
	}, core.ErrNotFound)

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return nil, core.ErrNotFound
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_Renew_ErrBadRequestWithHandler(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	handlerCalled := false
	errorRegistry.Add(func(reason error) {
		handlerCalled = true
	}, core.ErrBadRequest)

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return nil, core.ErrBadRequest
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_Renew_ErrNotFoundNoHandler(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return nil, core.ErrNotFound
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrNotFound)
	assert.Contains(t, err.Error(), "handling session error")
}

func TestAccessTokenSessionStore_Renew_ErrBadRequestNoHandler(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    pastTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return nil, core.ErrBadRequest
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrBadRequest)
	assert.Contains(t, err.Error(), "handling session error")
}

func TestAccessTokenSessionStore_HandleError_WithHandler(t *testing.T) {
	testError := errors.New("test error")
	
	cfg := &config.Config{
		TokensData: map[int64]config.TokenData{
			123: {Token: "token1"},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	handlerCalled := false
	errorRegistry.Add(func(reason error) {
		handlerCalled = true
	}, testError)

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)
	err := store.HandleError(testError)

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_HandleError_NoHandler(t *testing.T) {
	testError := errors.New("test error")
	
	cfg := &config.Config{
		TokensData: map[int64]config.TokenData{
			123: {Token: "token1"},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)
	err := store.HandleError(testError)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
}

func TestAccessTokenSessionStore_GetToken(t *testing.T) {
	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "test-token",
				RenewToken:  "test-renew",
				TokenExpiry: futureTime.Format(internal.ServerDateFormat),
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

	token := store.GetToken()
	assert.Equal(t, "test-token", token)
}

func TestAccessTokenSessionStore_GetToken_NoData(t *testing.T) {
	uid := int64(123)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData:      map[int64]config.TokenData{},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

	token := store.GetToken()
	assert.Equal(t, "", token)
}

func TestAccessTokenSessionStore_GetToken_ConfigError(t *testing.T) {
	cfgManager := &mock.ConfigManager{
		LoadErr: errors.New("config error"),
	}
	
	store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

	token := store.GetToken()
	assert.Equal(t, "", token)
}