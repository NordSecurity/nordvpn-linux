package session

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestTrustedPassSessionStore_Renew_SilentRenewal_NoHandlerInvocation(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:                  "test-token",
				IsOAuth:                true,
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "",
				TrustedPassTokenExpiry: time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
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

	renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
		return nil, errors.New("api-error")
	}

	store := NewTrustedPassSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, "api-error", err.Error())

	assert.False(t, handlerCalled)
}

func TestTrustedPassSessionStore_Renew_SilentRenewal_Success(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:                  "test-token",
				IsOAuth:                true,
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "",
				TrustedPassTokenExpiry: time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
		assert.Equal(t, "test-token", token)
		return &TrustedPassAccessTokenResponse{
			Token:   "new-trusted-pass-token",
			OwnerID: "nordvpn",
		}, nil
	}

	store := NewTrustedPassSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.NoError(t, err)

	savedData := cfg.TokensData[userID]
	assert.Equal(t, "new-trusted-pass-token", savedData.TrustedPassToken)
	assert.Equal(t, "nordvpn", savedData.TrustedPassOwnerID)

	expiryTime, err := time.Parse(internal.ServerDateFormat, savedData.TrustedPassTokenExpiry)
	assert.NoError(t, err)
	assert.True(t, expiryTime.After(time.Now()))
}

func TestTrustedPassSessionStore_Renew_SilentRenewal_ValidationError(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				Token:                  "test-token",
				IsOAuth:                true,
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "",
				TrustedPassTokenExpiry: time.Now().Add(-1 * time.Hour).Format(internal.ServerDateFormat),
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		handlerCalled = true
	}, ErrMissingTrustedPassResponse)

	renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
		return &TrustedPassAccessTokenResponse{
			Token:   "new-trusted-pass-token",
			OwnerID: "invalid-owner",
		}, nil
	}

	store := NewTrustedPassSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, ErrMissingTrustedPassResponse, err)

	assert.False(t, handlerCalled)
}
