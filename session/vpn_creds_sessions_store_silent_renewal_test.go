package session

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestVPNCredentialsSessionStore_Renew_SilentRenewal_NoHandlerInvocation(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				OpenVPNUsername:    "",
				OpenVPNPassword:    "",
				NordLynxPrivateKey: "",
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

	renewAPICall := func() (*VPNCredentialsResponse, error) {
		return nil, errors.New("api-error")
	}

	store := NewVPNCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.Error(t, err)
	assert.Equal(t, "api-error", err.Error())

	assert.False(t, handlerCalled, "Error handler should not be invoked when using Renew with SilentRenewal")
}

func TestVPNCredentialsSessionStore_Renew_SilentRenewal_Success(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				OpenVPNUsername:    "",
				OpenVPNPassword:    "",
				NordLynxPrivateKey: "",
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func() (*VPNCredentialsResponse, error) {
		return &VPNCredentialsResponse{
			Username:           "new-user",
			Password:           "new-pass",
			NordLynxPrivateKey: "new-key",
		}, nil
	}

	store := NewVPNCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew(SilentRenewal())

	assert.NoError(t, err)

	savedData := cfg.TokensData[userID]
	assert.Equal(t, "new-user", savedData.OpenVPNUsername)
	assert.Equal(t, "new-pass", savedData.OpenVPNPassword)
	assert.Equal(t, "new-key", savedData.NordLynxPrivateKey)
}
