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

func TestAccessTokenSessionStore_Renew(t *testing.T) {
	t.Run("should return nil when token is not expired", func(t *testing.T) {
		uid := int64(123)
		futureTime := time.Now().Add(24 * time.Hour)
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "token",
					RenewToken:  "renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

		err := store.Renew()

		assert.NoError(t, err)
	})

	t.Run("should return error when token is revoked", func(t *testing.T) {
		uid := int64(123)
		validHexToken := "ab78bb36299d442fa0715fb53b5e3e57"
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       validHexToken,
					RenewToken:  "renew",
					TokenExpiry: session.ManualAccessTokenExpiryDate.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()
		handlerCalled := false
		errorRegistry.Add(func(err error) {
			handlerCalled = true
		}, session.ErrAccessTokenRevoked)

		store := session.NewAccessTokenSessionStore(
			cfgManager,
			errorRegistry,
			nil,
			func(token string) error {
				return session.ErrAccessTokenRevoked
			},
		)

		err := store.Renew()

		assert.ErrorIs(t, err, session.ErrAccessTokenRevoked)
		assert.True(t, handlerCalled)
	})

	t.Run("should renew token when expired", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		futureTime := time.Now().Add(24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		var renewCalled bool
		var passedToken string
		var passedKey uuid.UUID

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			renewCalled = true
			passedToken = token
			passedKey = key

			return &session.AccessTokenResponse{
				Token:      "new-token",
				RenewToken: "new-renew",
				ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
			}, nil
		}

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.True(t, renewCalled)
		assert.Equal(t, "old-token", passedToken)
		assert.Equal(t, idempotencyKey, passedKey)

		assert.Equal(t, "new-token", cfgManager.Cfg.TokensData[uid].Token)
		assert.Equal(t, "new-renew", cfgManager.Cfg.TokensData[uid].RenewToken)
		assert.Equal(t, futureTime.Format(internal.ServerDateFormat), cfgManager.Cfg.TokensData[uid].TokenExpiry)
	})

	t.Run("should set idempotency key if not present", func(t *testing.T) {
		uid := int64(123)
		pastTime := time.Now().Add(-24 * time.Hour)
		futureTime := time.Now().Add(24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return &session.AccessTokenResponse{
				Token:      "new-token",
				RenewToken: "new-renew",
				ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
			}, nil
		}

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.NotNil(t, cfgManager.Cfg.TokensData[uid].IdempotencyKey)
	})

	t.Run("should handle API error during renewal", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		someError := errors.New("api-error")
		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return nil, someError
		}

		var handlerCalled bool
		var handledError error

		errorRegistry.Add(func(reason error) {
			handlerCalled = true
			handledError = reason
		}, someError)

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, someError, handledError)

		_, exists := cfgManager.Cfg.TokensData[uid]
		assert.True(t, exists)
	})

	t.Run("should handle nil renewal API response", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return nil, nil
		}

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "renewal API returned nil response")
	})

	t.Run("should handle invalid expiry date format", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return &session.AccessTokenResponse{
				Token:      "new-token",
				RenewToken: "new-renew",
				ExpiresAt:  "invalid-date-format",
			}, nil
		}

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.Error(t, err)
	})

	t.Run("should validate manual access token", func(t *testing.T) {
		uid := int64(123)
		validHexToken := "de62575eaaa54ca8bd9416d98bdc9c1c"
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       validHexToken,
					RenewToken:  "renew",
					TokenExpiry: session.ManualAccessTokenExpiryDate.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		var externalValidatorCalled bool
		var passedToken string

		externalValidator := func(token string) error {
			externalValidatorCalled = true
			passedToken = token
			return nil
		}

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return &session.AccessTokenResponse{
				Token:      "new-token",
				RenewToken: "new-renew",
				ExpiresAt:  time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
			}, nil
		}

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, externalValidator)

		err := store.Renew()

		assert.NoError(t, err)
		assert.True(t, externalValidatorCalled)
		assert.Equal(t, validHexToken, passedToken)
	})

	t.Run("should handle no token data", func(t *testing.T) {
		uid := int64(123)
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData:      map[int64]config.TokenData{},
		}

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non existing data")
	})

	t.Run("should handle config load error", func(t *testing.T) {
		cfgManager := &mock.ConfigManager{
			LoadErr: errors.New("config load error"),
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config load error")
	})

	t.Run("should handle ErrNotFound during renewal and invalidate", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return nil, core.ErrNotFound
		}

		var handlerCalled bool
		var handledError error

		errorRegistry.Add(func(reason error) {
			handlerCalled = true
			handledError = reason
		}, core.ErrNotFound)

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.ErrorIs(t, handledError, core.ErrNotFound)
	})

	t.Run("should handle ErrBadData during renewal and invalidate", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return nil, core.ErrBadRequest
		}

		var handlerCalled bool
		var handledError error

		errorRegistry.Add(func(reason error) {
			handlerCalled = true
			handledError = reason
		}, core.ErrBadRequest)

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.ErrorIs(t, handledError, core.ErrBadRequest)
	})

	t.Run("should return error from invalidate when no handlers for ErrNotFound", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return nil, core.ErrNotFound
		}

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrNotFound)
		assert.Contains(t, err.Error(), "invalidating session")
	})

	t.Run("should return error from invalidate when no handlers for ErrBadData", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		pastTime := time.Now().Add(-24 * time.Hour)
		testCfg := config.Config{
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

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return nil, core.ErrBadRequest
		}

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.ErrorIs(t, err, core.ErrBadRequest)
		assert.Contains(t, err.Error(), "invalidating session")
	})
}

func TestAccessTokenSessionStore_Invalidate(t *testing.T) {
	t.Run("should call error handlers", func(t *testing.T) {
		testCfg := config.Config{
			TokensData: map[int64]config.TokenData{
				123: {Token: "token1"},
			},
		}

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		handledIds := make(map[error]bool)
		testError := errors.New("test error")

		errorRegistry.Add(func(reason error) {
			handledIds[reason] = true
		}, testError)

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

		err := store.Invalidate(testError)

		assert.NoError(t, err)
		assert.True(t, handledIds[testError])
	})

	t.Run("should return error when no handlers registered", func(t *testing.T) {
		testCfg := config.Config{}

		cfgManager := &mock.ConfigManager{
			Cfg: &testCfg,
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()
		testError := errors.New("test error")

		store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

		err := store.Invalidate(testError)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalidating session")
	})
}

func TestAccessTokenSessionStore_GettersAndSetters(t *testing.T) {
	uid := int64(123)
	futureTime := time.Now().Add(24 * time.Hour)

	t.Run("GetToken", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "test-token",
					RenewToken:  "test-renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		token := store.GetToken()
		assert.Equal(t, "test-token", token)
	})

	t.Run("GetRenewalToken", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "test-token",
					RenewToken:  "test-renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		renewToken := store.GetRenewalToken()
		assert.Equal(t, "test-renew", renewToken)
	})

	t.Run("GetExpiry", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "test-token",
					RenewToken:  "test-renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		expiry := store.GetExpiry()
		expectedTime, _ := time.Parse(internal.ServerDateFormat, futureTime.Format(internal.ServerDateFormat))
		assert.Equal(t, expectedTime, expiry)
	})

	t.Run("IsExpired", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "test-token",
					RenewToken:  "test-renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		assert.False(t, store.IsExpired())

		pastTime := time.Now().Add(-24 * time.Hour)
		cfgManager.Cfg.TokensData[uid] = config.TokenData{
			Token:       "test-token",
			RenewToken:  "test-renew",
			TokenExpiry: pastTime.Format(internal.ServerDateFormat),
		}

		assert.True(t, store.IsExpired())
	})

	t.Run("SetToken", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "old-token",
					RenewToken:  "test-renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		err := store.SetToken("new-token")
		assert.NoError(t, err)
		assert.Equal(t, "new-token", cfgManager.Cfg.TokensData[uid].Token)
	})

	t.Run("SetRenewToken", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "test-token",
					RenewToken:  "old-renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		err := store.SetRenewToken("new-renew")
		assert.NoError(t, err)
		assert.Equal(t, "new-renew", cfgManager.Cfg.TokensData[uid].RenewToken)
	})

	t.Run("SetExpiry", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:       "test-token",
					RenewToken:  "test-renew",
					TokenExpiry: futureTime.Format(internal.ServerDateFormat),
				},
			},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		newExpiry := time.Now().Add(48 * time.Hour)
		err := store.SetExpiry(newExpiry)
		assert.NoError(t, err)
		assert.Equal(t, newExpiry.Format(internal.ServerDateFormat), cfgManager.Cfg.TokensData[uid].TokenExpiry)
	})

	t.Run("Getters with no data", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData:      map[int64]config.TokenData{},
		}

		cfgManager := &mock.ConfigManager{Cfg: &testCfg}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		assert.Equal(t, "", store.GetToken())
		assert.Equal(t, "", store.GetRenewalToken())
		assert.Equal(t, time.Time{}, store.GetExpiry())
		assert.True(t, store.IsExpired())
	})

	t.Run("Getters with config error", func(t *testing.T) {
		cfgManager := &mock.ConfigManager{
			LoadErr: errors.New("config error"),
		}
		store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

		assert.Equal(t, "", store.GetToken())
		assert.Equal(t, "", store.GetRenewalToken())
		assert.Equal(t, time.Time{}, store.GetExpiry())
		assert.True(t, store.IsExpired())
	})
}
