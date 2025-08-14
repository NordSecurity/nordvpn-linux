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

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
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

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		assert.Equal(t, "old-token", token)
		assert.Equal(t, idempotencyKey, key)
		return &AccessTokenResponse{
			Token:      "ab78bb36299d442fa0715fb53b5e3e58",
			RenewToken: "deadbeef1234567890abcdef1234567890",
			ExpiresAt:  futureTime.Format(internal.ServerDateFormat),
		}, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
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

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew(SilentRenewal())

	assert.NoError(t, err)
	// Renewal should succeed and update the invalid token
	assert.Equal(t, "ab78bb36299d442fa0715fb53b5e3e58", cfg.TokensData[userID].Token)

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_InvalidRenewTokenWithForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "ab78bb36299d442fa0715fb53b5e3e57",
				RenewToken:     "INVALID-RENEW-TOKEN",
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

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		// This should not be called because validation fails even with force renewal
		t.Fatal("Renewal API should not be called when renew token validation fails")
		return nil, nil
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	// Even with ForceRenewal, invalid token format causes validation error
	err := store.Renew(ForceRenewal(), SilentRenewal())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validating renew token")

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

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
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

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
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

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
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

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRenewToken, err)
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_InvalidTokenFormatWithForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:          "invalid", // Invalid token format
				RenewToken:     "deadbeef",
				TokenExpiry:    futureTime.Format(internal.ServerDateFormat), // Not expired
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
		// This should not be called because validation fails even with force renewal
		t.Fatal("Renewal API should not be called when token format validation fails")
		return nil, nil
	}

	externalValidator := func(token string) error {
		// External validator won't be called because token format validation fails first
		t.Fatal("External validator should not be called when token format is invalid")
		return errors.New("external validation failed")
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, externalValidator)
	// Even with ForceRenewal, invalid token format causes validation error
	err := store.Renew(ForceRenewal(), SilentRenewal())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validating access token format")

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestAccessTokenSessionStore_Renew_SilentRenewal_ForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	idempotencyKey := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)
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

	renewAPICall := func(token string, key uuid.UUID) (*AccessTokenResponse, error) {
		return nil, errors.New("api-error")
	}

	store := NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew(ForceRenewal(), SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, "api-error", err.Error())

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}
