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

func TestAccessTokenSessionStore_Renew_SimpleScenarios(t *testing.T) {
	uid := int64(123)

	tests := []struct {
		name            string
		setupConfig     func() *config.Config
		configLoadErr   error
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "should return nil when token is not expired",
			setupConfig: func() *config.Config {
				return &config.Config{
					AutoConnectData: config.AutoConnectData{ID: uid},
					TokensData: map[int64]config.TokenData{
						uid: {
							Token:       "token",
							RenewToken:  "renew",
							TokenExpiry: time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
						},
					},
				}
			},
			wantErr: false,
		},
		{
			name: "should handle no token data",
			setupConfig: func() *config.Config {
				return &config.Config{
					AutoConnectData: config.AutoConnectData{ID: uid},
					TokensData:      map[int64]config.TokenData{},
				}
			},
			wantErr:         true,
			wantErrContains: "non existing data",
		},
		{
			name:            "should handle config load error",
			configLoadErr:   errors.New("config load error"),
			wantErr:         true,
			wantErrContains: "config load error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *config.Config
			if tt.setupConfig != nil {
				cfg = tt.setupConfig()
			}

			cfgManager := &mock.ConfigManager{
				Cfg:     cfg,
				LoadErr: tt.configLoadErr,
			}

			errorRegistry := internal.NewErrorHandlingRegistry[error]()
			store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

			err := store.Renew()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAccessTokenSessionStore_Renew(t *testing.T) {

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

}

func TestAccessTokenSessionStore_Renew_ErrorHandlingWithoutHandlers(t *testing.T) {
	uid := int64(123)
	idempotencyKey := uuid.New()
	pastTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name            string
		renewError      error
		wantErr         error
		wantErrContains string
	}{
		{
			name:            "should return error from invalidate when no handlers for ErrNotFound",
			renewError:      core.ErrNotFound,
			wantErr:         core.ErrNotFound,
			wantErrContains: "invalidating session",
		},
		{
			name:            "should return error from invalidate when no handlers for ErrBadRequest",
			renewError:      core.ErrBadRequest,
			wantErr:         core.ErrBadRequest,
			wantErrContains: "invalidating session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				return nil, tt.renewError
			}

			store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, mockRenewAPICall, nil)

			err := store.Renew()

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}

func TestAccessTokenSessionStore_Invalidate(t *testing.T) {
	tests := []struct {
		name            string
		setupRegistry   func() *internal.ErrorHandlingRegistry[error]
		testError       error
		wantErr         bool
		wantErrContains string
		checkHandler    func(t *testing.T, handlerCalled bool)
	}{
		{
			name: "should call error handlers",
			setupRegistry: func() *internal.ErrorHandlingRegistry[error] {
				registry := internal.NewErrorHandlingRegistry[error]()
				registry.Add(func(reason error) {}, errors.New("test error"))
				return registry
			},
			testError: errors.New("test error"),
			wantErr:   false,
			checkHandler: func(t *testing.T, handlerCalled bool) {
				assert.True(t, handlerCalled)
			},
		},
		{
			name: "should return error when no handlers registered",
			setupRegistry: func() *internal.ErrorHandlingRegistry[error] {
				return internal.NewErrorHandlingRegistry[error]()
			},
			testError:       errors.New("test error"),
			wantErr:         true,
			wantErrContains: "invalidating session",
			checkHandler: func(t *testing.T, handlerCalled bool) {
				assert.False(t, handlerCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCfg := config.Config{
				TokensData: map[int64]config.TokenData{
					123: {Token: "token1"},
				},
			}

			cfgManager := &mock.ConfigManager{
				Cfg: &testCfg,
			}

			handlerCalled := false
			errorRegistry := tt.setupRegistry()

			// Override the handler to track if it was called
			if tt.name == "should call error handlers" {
				errorRegistry = internal.NewErrorHandlingRegistry[error]()
				errorRegistry.Add(func(reason error) {
					handlerCalled = true
				}, tt.testError)
			}

			store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, nil, nil)

			err := store.Invalidate(tt.testError)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
			}

			tt.checkHandler(t, handlerCalled)
		})
	}
}

func TestAccessTokenSessionStore_GetToken(t *testing.T) {
	uid := int64(123)
	futureTime := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name      string
		cfg       *config.Config
		loadErr   error
		wantToken string
	}{
		{
			name: "GetToken",
			cfg: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: uid},
				TokensData: map[int64]config.TokenData{
					uid: {
						Token:       "test-token",
						RenewToken:  "test-renew",
						TokenExpiry: futureTime.Format(internal.ServerDateFormat),
					},
				},
			},
			wantToken: "test-token",
		},
		{
			name: "GetToken with no data",
			cfg: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: uid},
				TokensData:      map[int64]config.TokenData{},
			},
			wantToken: "",
		},
		{
			name:      "GetToken with config error",
			loadErr:   errors.New("config error"),
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgManager := &mock.ConfigManager{
				Cfg:     tt.cfg,
				LoadErr: tt.loadErr,
			}
			store := session.NewAccessTokenSessionStore(cfgManager, nil, nil, nil)

			token := store.GetToken()
			assert.Equal(t, tt.wantToken, token)
		})
	}
}
