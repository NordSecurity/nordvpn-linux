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

func TestNCCredentialsSessionStore_Renew_SilentRenewal_NoHandlerInvocation(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				NCData: config.NCData{
					Username:       "",
					Password:       "",
					Endpoint:       "",
					ExpirationDate: time.Now().Add(-1 * time.Hour),
				},
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

	renewAPICall := func() (*NCCredentialsResponse, error) {
		return nil, errors.New("api-error")
	}

	store := NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, "api-error", err.Error())

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestNCCredentialsSessionStore_Renew_SilentRenewal_Success(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				NCData: config.NCData{
					Username:       "",
					Password:       "",
					Endpoint:       "",
					ExpirationDate: time.Now().Add(-1 * time.Hour),
				},
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func() (*NCCredentialsResponse, error) {
		return &NCCredentialsResponse{
			Username:  "nc-user",
			Password:  "nc-pass",
			Endpoint:  "wss://nc.example.com",
			ExpiresIn: 3600 * time.Second,
		}, nil
	}

	store := NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew(SilentRenewal())

	assert.NoError(t, err)

	savedData := cfg.TokensData[userID]
	assert.Equal(t, "nc-user", savedData.NCData.Username)
	assert.Equal(t, "nc-pass", savedData.NCData.Password)
	assert.Equal(t, "wss://nc.example.com", savedData.NCData.Endpoint)
	assert.True(t, savedData.NCData.ExpirationDate.After(time.Now()))
}
