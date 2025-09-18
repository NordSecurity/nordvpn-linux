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

func testVPNCredsValidate(store SessionStore) error {
	vpnStore, ok := store.(*VPNCredentialsSessionStore)
	if !ok {
		return errors.New("not a VPNCredentialsSessionStore")
	}
	return vpnStore.validate()
}

func TestVPNCredentialsSessionStore_Validate(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		username    string
		password    string
		nordLynxKey string
		wantErr     error
	}{
		{
			name:        "valid credentials",
			username:    "testuser",
			password:    "testpass",
			nordLynxKey: "testkey",
			wantErr:     nil,
		},
		{
			name:        "empty username",
			username:    "",
			password:    "testpass",
			nordLynxKey: "testkey",
			wantErr:     ErrMissingVPNCredentials,
		},
		{
			name:        "empty password",
			username:    "testuser",
			password:    "",
			nordLynxKey: "testkey",
			wantErr:     ErrMissingVPNCredentials,
		},
		{
			name:        "empty nordlynx key",
			username:    "testuser",
			password:    "testpass",
			nordLynxKey: "",
			wantErr:     ErrMissingNordLynxPrivateKey,
		},
		{
			name:        "all empty",
			username:    "",
			password:    "",
			nordLynxKey: "",
			wantErr:     ErrMissingVPNCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := int64(123)
			tokenData := config.TokenData{
				OpenVPNUsername:    tt.username,
				OpenVPNPassword:    tt.password,
				NordLynxPrivateKey: tt.nordLynxKey,
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg
			errRegistry := internal.NewErrorHandlingRegistry[error]()

			store := NewVPNCredentialsSessionStore(
				cfgManager,
				errRegistry,
				nil,
			)

			err := testVPNCredsValidate(store)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVPNCredentialsSessionStore_GetConfig(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		setupConfig func() *config.Config
		wantErr     bool
		wantErrMsg  string
		wantCreds   vpnCredentialsConfig
	}{
		{
			name: "valid config",
			setupConfig: func() *config.Config {
				userID := int64(123)
				return &config.Config{
					AutoConnectData: config.AutoConnectData{ID: userID},
					TokensData: map[int64]config.TokenData{
						userID: {
							OpenVPNUsername:    "testuser",
							OpenVPNPassword:    "testpass",
							NordLynxPrivateKey: "testkey",
						},
					},
				}
			},
			wantErr: false,
			wantCreds: vpnCredentialsConfig{
				Username:           "testuser",
				Password:           "testpass",
				NordLynxPrivateKey: "testkey",
			},
		},
		{
			name: "missing token data for user",
			setupConfig: func() *config.Config {
				return &config.Config{
					AutoConnectData: config.AutoConnectData{ID: 123},
					TokensData:      map[int64]config.TokenData{},
				}
			},
			wantErr:    true,
			wantErrMsg: "non existing data",
		},
		{
			name: "config load error",
			setupConfig: func() *config.Config {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgManager := mock.NewMockConfigManager()
			if tt.setupConfig() != nil {
				cfgManager.Cfg = tt.setupConfig()
			} else {
				cfgManager.LoadErr = errors.New("config load error")
			}

			errRegistry := internal.NewErrorHandlingRegistry[error]()
			store := &VPNCredentialsSessionStore{
				cfgManager:         cfgManager,
				errHandlerRegistry: errRegistry,
			}

			creds, err := store.getConfig()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCreds, creds)
			}
		})
	}
}

func TestVPNCredentialsSessionStore_HandleError(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		testError       error
		setupHandler    bool
		wantErr         bool
		wantErrContains string
		wantHandlerCall bool
	}{
		{
			name:            "with registered handler",
			testError:       errors.New("test error"),
			setupHandler:    true,
			wantErr:         true,
			wantErrContains: "handling session error",
			wantHandlerCall: true,
		},
		{
			name:            "no registered handler",
			testError:       errors.New("unhandled error"),
			setupHandler:    false,
			wantErr:         false,
			wantHandlerCall: false,
		},
		{
			name:            "multiple handlers",
			testError:       errors.New("test error"),
			setupHandler:    true,
			wantErr:         true,
			wantErrContains: "handling session error",
			wantHandlerCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgManager := mock.NewMockConfigManager()
			errRegistry := internal.NewErrorHandlingRegistry[error]()

			handlerCalled := false
			var handlerErr error

			if tt.setupHandler {
				errRegistry.Add(func(err error) {
					handlerCalled = true
					handlerErr = err
				}, tt.testError)

				if tt.name == "multiple handlers" {
					errRegistry.Add(func(err error) {

					}, tt.testError)
				}
			}

			store := NewVPNCredentialsSessionStore(cfgManager, errRegistry, nil)
			err := store.HandleError(tt.testError)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantHandlerCall, handlerCalled)
			if tt.wantHandlerCall {
				assert.Equal(t, tt.testError, handlerErr)
			}
		})
	}
}

func TestVPNCredentialsSessionStore_Renew(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)

	tests := []struct {
		name            string
		tokenData       config.TokenData
		renewAPICall    VPNCredentialsRenewalAPICall
		wantErr         bool
		wantErrContains string
		checkRenewCall  func(t *testing.T, renewCalled bool)
		checkConfig     func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "valid credentials do not renew",
			tokenData: config.TokenData{
				OpenVPNUsername:    "testuser",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				assert.Fail(t, "Renew API should not be called for valid credentials")
				return nil, nil
			},
			wantErr: false,
			checkRenewCall: func(t *testing.T, renewCalled bool) {
				assert.False(t, renewCalled, "Renew API should not be called for valid credentials")
			},
		},
		{
			name: "invalid credentials trigger renewal",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return &VPNCredentialsResponse{
					Username:           "newuser",
					Password:           "newpass",
					NordLynxPrivateKey: "newkey",
				}, nil
			},
			wantErr: false,
			checkRenewCall: func(t *testing.T, renewCalled bool) {
				assert.True(t, renewCalled, "Renew API should be called for invalid credentials")
			},
			checkConfig: func(t *testing.T, cfg *config.Config) {
				data := cfg.TokensData[userID]
				assert.Equal(t, "newuser", data.OpenVPNUsername)
				assert.Equal(t, "newpass", data.OpenVPNPassword)
				assert.Equal(t, "newkey", data.NordLynxPrivateKey)
			},
		},
		{
			name: "nil renewal API",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall:    nil,
			wantErr:         true,
			wantErrContains: "renewal api call not configured",
		},
		{
			name: "renewal API returns nil response",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return nil, nil
			},
			wantErr:         false,
			wantErrContains: "",
		},
		{
			name: "renewal API returns empty username",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return &VPNCredentialsResponse{
					Username:           "",
					Password:           "newpass",
					NordLynxPrivateKey: "newkey",
				}, nil
			},
			wantErr:         false,
			wantErrContains: "",
		},
		{
			name: "renewal API returns empty password",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return &VPNCredentialsResponse{
					Username:           "newuser",
					Password:           "",
					NordLynxPrivateKey: "newkey",
				}, nil
			},
			wantErr:         false,
			wantErrContains: "",
		},
		{
			name: "renewal API error",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return nil, errors.New("api error")
			},
			wantErr: false,
		},
		{
			name: "config save error",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return &VPNCredentialsResponse{
					Username:           "newuser",
					Password:           "newpass",
					NordLynxPrivateKey: "newkey",
				}, nil
			},
			wantErr:         true,
			wantErrContains: "failed to save vpn credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tt.tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg

			if tt.name == "config save error" {
				cfgManager.SaveErr = errors.New("save error")
			}

			errRegistry := internal.NewErrorHandlingRegistry[error]()

			renewCalled := false
			var renewAPICall VPNCredentialsRenewalAPICall
			if tt.renewAPICall != nil {
				originalCall := tt.renewAPICall
				renewAPICall = func() (*VPNCredentialsResponse, error) {
					renewCalled = true
					return originalCall()
				}
			}

			store := NewVPNCredentialsSessionStore(cfgManager, errRegistry, renewAPICall)

			err := store.Renew()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.checkRenewCall != nil {
				tt.checkRenewCall(t, renewCalled)
			}

			if tt.checkConfig != nil {
				tt.checkConfig(t, &cfg)
			}
		})
	}
}

func TestVPNCredentialsSessionStore_RenewWithErrorHandler(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)
	tokenData := config.TokenData{
		OpenVPNUsername:    "",
		OpenVPNPassword:    "testpass",
		NordLynxPrivateKey: "testkey",
	}

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData:      map[int64]config.TokenData{userID: tokenData},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	errRegistry := internal.NewErrorHandlingRegistry[error]()

	apiError := errors.New("API error")
	handlerCalled := false

	errRegistry.Add(func(err error) {
		handlerCalled = true
		assert.Equal(t, apiError, err)
	}, apiError)

	renewAPICall := func() (*VPNCredentialsResponse, error) {
		return nil, apiError
	}

	store := NewVPNCredentialsSessionStore(cfgManager, errRegistry, renewAPICall)

	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, handlerCalled, "Error handler should be called")
}

func TestVPNCredentialsSessionStore_InterfaceCompliance(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	errRegistry := internal.NewErrorHandlingRegistry[error]()

	var store SessionStore = NewVPNCredentialsSessionStore(cfgManager, errRegistry, nil)

	assert.NotNil(t, store)
	assert.Implements(t, (*SessionStore)(nil), store)
}

func TestVPNCredentialsSessionStore_Renew_ForceRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)

	tests := []struct {
		name            string
		tokenData       config.TokenData
		renewAPICall    VPNCredentialsRenewalAPICall
		expectRenewal   bool
		expectError     bool
		useForceRenewal bool
	}{
		{
			name: "valid credentials without force renewal - no renewal",
			tokenData: config.TokenData{
				OpenVPNUsername:    "testuser",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				t.Fatal("Renew API should not be called for valid credentials without force")
				return nil, nil
			},
			expectRenewal:   false,
			useForceRenewal: false,
		},
		{
			name: "valid credentials with force renewal - triggers renewal",
			tokenData: config.TokenData{
				OpenVPNUsername:    "testuser",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return &VPNCredentialsResponse{
					Username:           "newuser",
					Password:           "newpass",
					NordLynxPrivateKey: "newkey",
				}, nil
			},
			expectRenewal:   true,
			useForceRenewal: true,
		},
		{
			name: "invalid credentials without force renewal - triggers renewal",
			tokenData: config.TokenData{
				OpenVPNUsername:    "",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return &VPNCredentialsResponse{
					Username:           "newuser",
					Password:           "newpass",
					NordLynxPrivateKey: "newkey",
				}, nil
			},
			expectRenewal:   true,
			useForceRenewal: false,
		},
		{
			name: "force renewal with API error",
			tokenData: config.TokenData{
				OpenVPNUsername:    "testuser",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
			renewAPICall: func() (*VPNCredentialsResponse, error) {
				return nil, errors.New("api error")
			},
			expectRenewal:   true,
			expectError:     false,
			useForceRenewal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tt.tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg

			errRegistry := internal.NewErrorHandlingRegistry[error]()

			renewCalled := false
			var renewAPICall VPNCredentialsRenewalAPICall
			if tt.renewAPICall != nil {
				originalCall := tt.renewAPICall
				renewAPICall = func() (*VPNCredentialsResponse, error) {
					renewCalled = true
					return originalCall()
				}
			}

			store := NewVPNCredentialsSessionStore(cfgManager, errRegistry, renewAPICall)

			var err error
			if tt.useForceRenewal {
				err = store.Renew(ForceRenewal())
			} else {
				err = store.Renew()
			}

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectRenewal, renewCalled, "Renewal call expectation mismatch")

			if tt.expectRenewal && !tt.expectError && renewCalled {
				savedCfg := cfgManager.Cfg
				tokenData := savedCfg.TokensData[userID]
				if tt.renewAPICall != nil {
					if tokenData.OpenVPNUsername == "newuser" {
						assert.Equal(t, "newuser", tokenData.OpenVPNUsername)
						assert.Equal(t, "newpass", tokenData.OpenVPNPassword)
						assert.Equal(t, "newkey", tokenData.NordLynxPrivateKey)
					}
				}
			}
		})
	}
}

func TestVPNCredentialsSessionStore_Renew_ForceRenewalWithSilentRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	userID := int64(123)

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData: map[int64]config.TokenData{
			userID: {
				OpenVPNUsername:    "testuser",
				OpenVPNPassword:    "testpass",
				NordLynxPrivateKey: "testkey",
			},
		},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	handlerCalled := false
	errRegistry := internal.NewErrorHandlingRegistry[error]()
	apiError := errors.New("API error")
	errRegistry.Add(func(err error) {
		handlerCalled = true
	}, apiError)

	renewAPICall := func() (*VPNCredentialsResponse, error) {
		return nil, apiError
	}

	store := NewVPNCredentialsSessionStore(cfgManager, errRegistry, renewAPICall)

	err := store.Renew(ForceRenewal(), SilentRenewal())
	assert.Error(t, err)
	assert.False(t, handlerCalled, "Handler should not be called with SilentRenewal")
	assert.Contains(t, err.Error(), "API error")
}
