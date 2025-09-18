package session

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccessTokenSessionStore_Renew_SilentRenewal_NoHandlerInvocation(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		handlerCalled = true
	}, errors.New("api-error"))

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		return nil, errors.New("api-error")
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, "api-error", err.Error())

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_Success(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(renewToken string, key uuid.UUID) (*AccessTokenResponse, error) {
		assert.Equal(t, "old-renew", renewToken)
		assert.Equal(t, idempotencyKey, key)
		return &AccessTokenResponse{
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "deadbeef1234567890abcdef1234567890",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.NoError(t, err)

	savedData := cfg.TokensData[userID]
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e58", savedData.Token)
	assert.Equal(t, "deadbeef1234567890abcdef1234567890", savedData.RenewToken)
	assert.Equal(t, futureTime.Format(internal.ServerDateFormat), savedData.TokenExpiry)
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_ValidationError(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "invalid-token-format", // Invalid format, but token is not expired
				RenewToken:     "validrenew",
				TokenExpiry:    futureTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		handlerCalled = true
	}, ErrInvalidToken)

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		// The API will be called because validation fails, triggering renewal
		return &AccessTokenResponse{
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "deadbeef",
			ExpiresAt:  futureTime.Add(24 * time.Hour).Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.NoError(t, err)
	// Renewal should succeed and update the invalid token
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e58", cfg.TokensData[userID].Token)

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_InvalidRenewTokenResponseWithForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "ab78bb36299d442fa0715fb53b5e3e57",
				RenewToken:     "deadbeef", // Valid format
				TokenExpiry:    futureTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		handlerCalled = true
	}, ErrInvalidRenewToken)

	renewAPICallCount := 0
	renewAPICall := func(renewToken string, key uuid.UUID) (*AccessTokenResponse, error) {
		renewAPICallCount++
		assert.Equal(t, "deadbeef", renewToken)
		assert.Equal(t, idempotencyKey, key)
		// API returns invalid token in response (checked before renew token)
		return &AccessTokenResponse{
			Token:      "INVALID-TOKEN-FORMAT",
			RenewToken: "INVALID-RENEW-TOKEN",
			ExpiresAt:  futureTime.Add(24 * time.Hour).Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	// With ForceRenewal, renewal is attempted but fails due to invalid response
	err := store.Renew(ForceRenewal(), SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err) // Token is validated before renew token
	assert.Equal(t, 1, renewAPICallCount, "Renewal API should be called with ForceRenewal")

	// Verify the token was NOT updated due to invalid response
	savedData := cfg.TokensData[userID]
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e57", savedData.Token)
	assert.Equal(t, "deadbeef", savedData.RenewToken)
	assert.Equal(t, futureTime.Format(internal.ServerDateFormat), savedData.TokenExpiry)

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_MissingTokenData(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData:      map[int64]config.TokenData{}, // No token data
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		// This should not be called
		t.Fatal("Renewal API should not be called when token data is missing")
		return nil, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no token data")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_NilResponse(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		handlerCalled = true
	}, ErrMissingAccessTokenResponse)

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		return nil, nil // Return nil response
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, ErrMissingAccessTokenResponse, err)

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_InvalidResponseToken(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		return &AccessTokenResponse{
			Token:      "INVALID-TOKEN-FORMAT", // Invalid format
			RenewToken: "deadbeef",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_InvalidResponseRenewToken(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "old-token",
				RenewToken:     "old-renew",
				TokenExpiry:    time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		return &AccessTokenResponse{
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "INVALID-RENEW-TOKEN", // Invalid format
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRenewToken, err)
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_InvalidTokenResponseWithForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "ab78bb36299d442fa0715fb53b5e3e57", // Valid token format
				RenewToken:     "deadbeef",
				TokenExpiry:    futureTime.Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		handlerCalled = true
	}, ErrInvalidToken)

	renewAPICallCount := 0
	renewAPICall := func(renewToken string, key uuid.UUID) (*AccessTokenResponse, error) {
		renewAPICallCount++
		assert.Equal(t, "deadbeef", renewToken)
		assert.Equal(t, idempotencyKey, key)
		// API returns invalid token in response
		return &AccessTokenResponse{
			Token:      "INVALID-TOKEN-FORMAT", // Invalid format in response
			RenewToken: "newrenew1234567890abcdef1234567890",
			ExpiresAt:  futureTime.Add(24 * time.Hour).Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	// With ForceRenewal, renewal is attempted but fails due to invalid response
	err := store.Renew(ForceRenewal(), SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Equal(t, 1, renewAPICallCount, "Renewal API should be called with ForceRenewal")

	// Verify the token was NOT updated due to invalid response
	savedData := cfg.TokensData[userID]
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e57", savedData.Token)
	assert.Equal(t, "deadbeef", savedData.RenewToken)
	assert.Equal(t, futureTime.Format(internal.ServerDateFormat), savedData.TokenExpiry)

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_ForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	newFutureTime := futureTime.Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "ab78bb36299d442fa0715fb53b5e3e57",
				RenewToken:     "deadbeef",
				TokenExpiry:    futureTime.Format(internal.ServerDateFormat), // Still valid
				IdempotencyKey: &idempotencyKey,
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		handlerCalled = true
	}, errors.New("api-error"))

	renewAPICallCount := 0
	renewAPICall := func(renewToken string, key uuid.UUID) (*AccessTokenResponse, error) {
		renewAPICallCount++
		assert.Equal(t, "deadbeef", renewToken)
		assert.Equal(t, idempotencyKey, key)
		return &AccessTokenResponse{
			Token:      "1234567890abcdef1234567890abcdef",
			RenewToken: "abcdef1234567890abcdef1234567890",
			ExpiresAt:  newFutureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(ForceRenewal(), SilentRenewal())

	// ForceRenewal should trigger renewal
	assert.NoError(t, err)
	assert.Equal(t, 1, renewAPICallCount, "Renewal API should be called with ForceRenewal")

	// Verify the token was updated
	savedData := cfg.TokensData[userID]
	assert.Equal(t, "1234567890abcdef1234567890abcdef", savedData.Token)
	assert.Equal(t, "abcdef1234567890abcdef1234567890", savedData.RenewToken)
	assert.Equal(t, newFutureTime.Format(internal.ServerDateFormat), savedData.TokenExpiry)

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}
