package session_test

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccessTokenSessionStore_Renew_NotExpired(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "ab78bb36299d442fa0715fb53b5e3e57",
				RenewToken:  "deadbeef1234567890abcdef1234567890abcdef",
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
	category.Set(t, category.Unit)

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
	category.Set(t, category.Unit)

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
	category.Set(t, category.Unit)

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
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "ab78bb36299d442fa0715fb53b5e3e59",
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
	category.Set(t, category.Unit)

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
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "ab78bb36299d442fa0715fb53b5e3e59",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.NoError(t, err)
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e58", cfgManager.Cfg.TokensData[uid].Token)
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e59", cfgManager.Cfg.TokensData[uid].RenewToken)
	assert.Equal(t, futureTime.Format(internal.ServerDateFormat), cfgManager.Cfg.TokensData[uid].TokenExpiry)
}

func TestAccessTokenSessionStore_Renew_SetIdempotencyKey(t *testing.T) {
	category.Set(t, category.Unit)

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
				IdempotencyKey: nil,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		return &session.AccessTokenResponse{
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "ab78bb36299d442fa0715fb53b5e3e59",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.NoError(t, err)
	assert.NotNil(t, cfgManager.Cfg.TokensData[uid].IdempotencyKey)
}

func TestAccessTokenSessionStore_Renew_APIErrorWithHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_Renew_APIErrorNoHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.NoError(t, err)
}

func TestAccessTokenSessionStore_Renew_NilAPIResponse(t *testing.T) {
	category.Set(t, category.Unit)

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
	category.Set(t, category.Unit)

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
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "ab78bb36299d442fa0715fb53b5e3e59",
			ExpiresAt:  "invalid-date-format",
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing expiry time")
}

func TestAccessTokenSessionStore_Renew_ExternalValidatorSuccess(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "de62575eaaa54ca8bd9416d98bdc9c1c",
				RenewToken:  "abcdef1234567890abcdef1234567890",
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
	category.Set(t, category.Unit)

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
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "ab78bb36299d442fa0715fb53b5e3e59",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, externalValidator)
	err := store.Renew()

	assert.NoError(t, err)
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e58", cfgManager.Cfg.TokensData[uid].Token)
}

func TestAccessTokenSessionStore_Renew_ErrNotFoundWithHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_Renew_ErrBadRequestWithHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_Renew_ErrNotFoundNoHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.NoError(t, err)
}

func TestAccessTokenSessionStore_Renew_ErrBadRequestNoHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.NoError(t, err)
}

func TestAccessTokenSessionStore_HandleError_WithHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, handlerCalled)
}

func TestAccessTokenSessionStore_HandleError_NoHandler(t *testing.T) {
	category.Set(t, category.Unit)

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

	assert.NoError(t, err)
}

func TestAccessTokenSessionStore_GetToken(t *testing.T) {
	category.Set(t, category.Unit)

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
	category.Set(t, category.Unit)

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
	category.Set(t, category.Unit)

	cfgManager := &mock.ConfigManager{
		LoadErr: errors.New("config error"),
	}

	store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

	token := store.GetToken()
	assert.Equal(t, "", token)
}

func TestAccessTokenSessionStore_Renew_ForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name             string
		setupConfig      func() *config.Config
		renewAPIResponse *session.AccessTokenResponse
		renewAPIError    error
		expectRenewal    bool
		expectError      bool
		useForceRenewal  bool
	}{
		{
			name: "valid token without force renewal - no renewal",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					AutoConnectData: config.AutoConnectData{ID: 123},
					TokensData: map[int64]config.TokenData{
						123: {
							Token:          "abc123def456789012345678901234567890",
							RenewToken:     "def456abc123789012345678901234567890",
							TokenExpiry:    time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
							IdempotencyKey: &uuid.Nil,
						},
					},
				}
				return cfg
			},
			expectRenewal:   false,
			useForceRenewal: false,
		},
		{
			name: "valid token with force renewal - triggers renewal",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					AutoConnectData: config.AutoConnectData{ID: 123},
					TokensData: map[int64]config.TokenData{
						123: {
							Token:          "abc123def456789012345678901234567890",
							RenewToken:     "def456abc123789012345678901234567890",
							TokenExpiry:    time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
							IdempotencyKey: &uuid.Nil,
						},
					},
				}
				return cfg
			},
			renewAPIResponse: &session.AccessTokenResponse{
				Token:      "1234567890abcdef1234567890abcdef",
				RenewToken: "abcdef1234567890abcdef1234567890",
				ExpiresAt:  time.Now().Add(48 * time.Hour).Format(internal.ServerDateFormat),
			},
			expectRenewal:   true,
			useForceRenewal: true,
		},
		{
			name: "expired token without force renewal - triggers renewal",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					AutoConnectData: config.AutoConnectData{ID: 123},
					TokensData: map[int64]config.TokenData{
						123: {
							Token:          "abc123def456789012345678901234567890",
							RenewToken:     "def456abc123789012345678901234567890",
							TokenExpiry:    time.Now().Add(-24 * time.Hour).Format(internal.ServerDateFormat),
							IdempotencyKey: &uuid.Nil,
						},
					},
				}
				return cfg
			},
			renewAPIResponse: &session.AccessTokenResponse{
				Token:      "1234567890abcdef1234567890abcdef",
				RenewToken: "abcdef1234567890abcdef1234567890",
				ExpiresAt:  time.Now().Add(48 * time.Hour).Format(internal.ServerDateFormat),
			},
			expectRenewal:   true,
			useForceRenewal: false,
		},
		{
			name: "force renewal with API error - no handler",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					AutoConnectData: config.AutoConnectData{ID: 123},
					TokensData: map[int64]config.TokenData{
						123: {
							Token:          "abc123def456789012345678901234567890",
							RenewToken:     "def456abc123789012345678901234567890",
							TokenExpiry:    time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
							IdempotencyKey: &uuid.Nil,
						},
					},
				}
				return cfg
			},
			renewAPIError:   core.ErrBadRequest,
			expectRenewal:   true,
			useForceRenewal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			cfgManager := &mock.ConfigManager{Cfg: cfg}

			renewCalled := false
			renewAPICall := func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
				renewCalled = true
				if tt.renewAPIError != nil {
					return nil, tt.renewAPIError
				}
				return tt.renewAPIResponse, nil
			}

			registry := internal.NewErrorHandlingRegistry[error]()
			store := session.NewAccessTokenSessionStore(cfgManager, registry, renewAPICall, nil)

			var err error
			if tt.useForceRenewal {
				err = store.Renew(session.ForceRenewal())
			} else {
				err = store.Renew()
			}

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectRenewal, renewCalled, "Renewal call expectation mismatch")

			if tt.expectRenewal && !tt.expectError && tt.renewAPIResponse != nil {
				savedCfg := cfgManager.Cfg
				tokenData := savedCfg.TokensData[savedCfg.AutoConnectData.ID]
				assert.Equal(t, tt.renewAPIResponse.Token, tokenData.Token)
				assert.Equal(t, tt.renewAPIResponse.RenewToken, tokenData.RenewToken)
			}
		})
	}
}
