package session_test

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

type mockValidator struct {
	validateFunc  func(store interface{}) error
	validateCalls int
}

func (m *mockValidator) Validate(store interface{}) error {
	m.validateCalls++
	return m.validateFunc(store)
}

func TestAccessTokenSessionStore_Renew(t *testing.T) {
	t.Run("should return nil when validation succeeds", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData: map[int64]config.TokenData{
				123: {Token: "token", RenewToken: "renew", TokenExpiry: "2025-01-01 10:10:10"},
			},
		}

		cfgManager := &mockConfigManager{
			config: testCfg,
		}

		validator := &mockValidator{
			validateFunc: func(store interface{}) error {
				return nil
			},
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		store := session.NewAccessTokenSessionStore(cfgManager, validator, errorRegistry, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.Equal(t, 1, validator.validateCalls)
	})

	t.Run("should return error when validation fails with ErrAccessTokenRevoked", func(t *testing.T) {
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData: map[int64]config.TokenData{
				123: {Token: "token", RenewToken: "renew", TokenExpiry: "2025-01-01 10:10:10"},
			},
		}

		cfgManager := &mockConfigManager{config: testCfg}

		validator := &mockValidator{
			validateFunc: func(store interface{}) error {
				return session.ErrAccessTokenRevoked
			},
		}

		store := session.NewAccessTokenSessionStore(
			cfgManager,
			validator,
			internal.NewErrorHandlingRegistry[error](),
			func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
				return &session.AccessTokenResponse{}, nil
			},
		)

		err := store.Renew()

		assert.ErrorIs(t, err, session.ErrAccessTokenRevoked)
		assert.Equal(t, 1, validator.validateCalls)
	})

	t.Run("should renew token when validation fails with unhandled error", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:          "old-token",
					RenewToken:     "old-renew",
					TokenExpiry:    "2025-01-01 10:10:10",
					IdempotencyKey: &idempotencyKey,
				},
			},
		}

		cfgManager := &mockConfigManager{
			config: testCfg,
		}

		validator := &mockValidator{
			validateFunc: func(store interface{}) error {
				return errors.New("validation error")
			},
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
				ExpiresAt:  "2025-02-01 10:10:10",
			}, nil
		}

		store := session.NewAccessTokenSessionStore(cfgManager, validator, errorRegistry, mockRenewAPICall)

		err := store.Renew()

		assert.NoError(t, err)
		assert.Equal(t, 1, validator.validateCalls)
		assert.True(t, renewCalled)
		assert.Equal(t, "old-token", passedToken)
		assert.Equal(t, idempotencyKey, passedKey)

		assert.Equal(t, "new-token", cfgManager.config.TokensData[uid].Token)
		assert.Equal(t, "new-renew", cfgManager.config.TokensData[uid].RenewToken)
		assert.Equal(t, "2025-02-01 10:10:10", cfgManager.config.TokensData[uid].TokenExpiry)
	})

	t.Run("should set idempotency key if not present", func(t *testing.T) {
		uid := int64(123)
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:          "old-token",
					RenewToken:     "old-renew",
					TokenExpiry:    "2025-01-01 10:10:10",
					IdempotencyKey: nil,
				},
			},
		}

		cfgManager := &mockConfigManager{
			config: testCfg,
		}

		validator := &mockValidator{
			validateFunc: func(store interface{}) error {
				return errors.New("validation error")
			},
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		mockRenewAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
			return &session.AccessTokenResponse{
				Token:      "new-token",
				RenewToken: "new-renew",
				ExpiresAt:  "2025-02-01 10:10:10",
			}, nil
		}

		store := session.NewAccessTokenSessionStore(cfgManager, validator, errorRegistry, mockRenewAPICall)

		err := store.Renew()

		assert.NoError(t, err)
		assert.Equal(t, 1, validator.validateCalls)
		assert.NotNil(t, cfgManager.config.TokensData[uid].IdempotencyKey)
	})

	t.Run("should handle not-found error during renewal", func(t *testing.T) {
		uid := int64(123)
		idempotencyKey := uuid.New()
		testCfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: uid},
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:          "old-token",
					RenewToken:     "old-renew",
					TokenExpiry:    "2025-01-01 10:10:10",
					IdempotencyKey: &idempotencyKey,
				},
			},
		}

		cfgManager := &mockConfigManager{
			config: testCfg,
		}

		validator := &mockValidator{
			validateFunc: func(store interface{}) error {
				return errors.New("validation error")
			},
		}

		someError := errors.New("not-found")
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

		store := session.NewAccessTokenSessionStore(cfgManager, validator, errorRegistry, mockRenewAPICall)

		err := store.Renew()

		assert.NoError(t, err)
		assert.Equal(t, 1, validator.validateCalls)
		assert.True(t, handlerCalled)
		assert.Equal(t, someError, handledError)

		// external handler did not remove any configuration
		_, exists := cfgManager.config.TokensData[uid]
		assert.True(t, exists)
	})
}

func TestAccessTokenSessionStore_Invalidate(t *testing.T) {
	t.Run("should call error handlers for all users", func(t *testing.T) {
		testCfg := config.Config{
			TokensData: map[int64]config.TokenData{
				123: {Token: "token1"},
			},
		}

		cfgManager := &mockConfigManager{
			config: testCfg,
		}

		validator := &mockValidator{
			validateFunc: func(store interface{}) error {
				return nil
			},
		}

		errorRegistry := internal.NewErrorHandlingRegistry[error]()

		handledIds := make(map[error]bool)
		testError := errors.New("test error")

		errorRegistry.Add(func(reason error) {
			handledIds[reason] = true
		}, testError)

		store := session.NewAccessTokenSessionStore(cfgManager, validator, errorRegistry, nil)

		err := store.Invalidate(testError)

		assert.NoError(t, err)
		assert.True(t, handledIds[testError])
	})
}
