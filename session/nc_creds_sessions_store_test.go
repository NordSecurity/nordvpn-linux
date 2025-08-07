package session_test

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestNCCredentialsSessionStore_Renew_NotExpired(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "testuser",
					Password:       "testpass",
					Endpoint:       "https://api.example.com",
					ExpirationDate: futureTime,
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewCalled := false
	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		renewCalled = true
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)

	err := store.Renew()
	assert.NoError(t, err)
	assert.False(t, renewCalled, "Renew API should not be called when credentials are valid")
}

func TestNCCredentialsSessionStore_Renew_NoTokenData(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData:      map[int64]config.TokenData{},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewCalled := false
	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		renewCalled = true
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)

	err := store.Renew()
	assert.NoError(t, err)
	assert.True(t, renewCalled, "Renew API should be called when there's no token data")

	// Verify the data was created
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[uid].NCData.Username)
	assert.Equal(t, "newpass", cfgManager.Cfg.TokensData[uid].NCData.Password)
	assert.Equal(t, "https://new.example.com", cfgManager.Cfg.TokensData[uid].NCData.Endpoint)
}

func TestNCCredentialsSessionStore_Renew_ConfigLoadError(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := &mock.ConfigManager{
		LoadErr: errors.New("config load error"),
	}

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewCalled := false
	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		renewCalled = true
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)

	err := store.Renew()
	// When validate() fails due to config load error, Renew() continues and calls the renewal API
	assert.NoError(t, err)
	assert.True(t, renewCalled, "Renew API should be called even when config load fails")

	// Verify the credentials were saved
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[cfgManager.Cfg.AutoConnectData.ID].NCData.Username)
}

func TestNCCredentialsSessionStore_Renew_ExpiredCredentials(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	pastTime := time.Now().UTC().Add(-24 * time.Hour)
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "olduser",
					Password:       "oldpass",
					Endpoint:       "https://old.example.com",
					ExpirationDate: pastTime,
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400, // 24 hours
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew()

	assert.NoError(t, err)
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[uid].NCData.Username)
	assert.Equal(t, "newpass", cfgManager.Cfg.TokensData[uid].NCData.Password)
	assert.Equal(t, "https://new.example.com", cfgManager.Cfg.TokensData[uid].NCData.Endpoint)
	assert.True(t, cfgManager.Cfg.TokensData[uid].NCData.ExpirationDate.After(futureTime.Add(-time.Hour)))
}

func TestNCCredentialsSessionStore_Renew_InvalidCredentials(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		response *session.NCCredentialsResponse
		wantErr  string
	}{
		{
			name: "empty username",
			response: &session.NCCredentialsResponse{
				Username:  "",
				Password:  "pass",
				Endpoint:  "https://api.example.com",
				ExpiresIn: 3600,
			},
			wantErr: "missing nc credentials",
		},
		{
			name: "empty password",
			response: &session.NCCredentialsResponse{
				Username:  "user",
				Password:  "",
				Endpoint:  "https://api.example.com",
				ExpiresIn: 3600,
			},
			wantErr: "missing nc credentials",
		},
		{
			name: "empty endpoint",
			response: &session.NCCredentialsResponse{
				Username:  "user",
				Password:  "pass",
				Endpoint:  "",
				ExpiresIn: 3600,
			},
			wantErr: "invalid endpoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid := int64(123)
			pastTime := time.Now().UTC().Add(-24 * time.Hour)

			cfg := &config.Config{
				AutoConnectData: config.AutoConnectData{ID: uid},
				TokensData: map[int64]config.TokenData{
					uid: {
						NCData: config.NCData{
							Username:       "olduser",
							Password:       "oldpass",
							Endpoint:       "https://old.example.com",
							ExpirationDate: pastTime,
						},
					},
				},
			}

			cfgManager := &mock.ConfigManager{Cfg: cfg}
			errorRegistry := internal.NewErrorHandlingRegistry[error]()

			renewAPICall := func() (*session.NCCredentialsResponse, error) {
				return tt.response, nil
			}

			store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
			err := store.Renew()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNCCredentialsSessionStore_Renew_APIError(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	pastTime := time.Now().UTC().Add(-24 * time.Hour)
	apiError := errors.New("api-error")

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "olduser",
					Password:       "oldpass",
					Endpoint:       "https://old.example.com",
					ExpirationDate: pastTime,
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	handlerCalled := false
	errorRegistry.Add(func(reason error) {
		handlerCalled = true
	}, apiError)

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return nil, apiError
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api-error")
	assert.True(t, handlerCalled)
}

func TestNCCredentialsSessionStore_Renew_SaveError(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "olduser",
					Password:       "oldpass",
					Endpoint:       "https://old.example.com",
					ExpirationDate: pastTime,
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{
		Cfg:     cfg,
		SaveErr: errors.New("save error"),
	}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "saving renewed nc creds")
}

func TestNCCredentialsSessionStore_HandleError(t *testing.T) {
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
			wantErrContains: "test error",
			wantHandlerCall: true,
		},
		{
			name:            "no registered handler",
			testError:       errors.New("unhandled error"),
			setupHandler:    false,
			wantErr:         true,
			wantErrContains: "handling NC credentials error",
			wantHandlerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TokensData: map[int64]config.TokenData{
					123: {},
				},
			}

			cfgManager := &mock.ConfigManager{Cfg: cfg}
			errorRegistry := internal.NewErrorHandlingRegistry[error]()

			handlerCalled := false
			if tt.setupHandler {
				errorRegistry.Add(func(reason error) {
					handlerCalled = true
				}, tt.testError)
			}

			store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, nil)
			err := store.HandleError(tt.testError)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantHandlerCall, handlerCalled)
		})
	}
}

func TestNCCredentialsSessionStore_Validate(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		ncData    config.NCData
		wantValid bool
	}{
		{
			name: "valid credentials",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "testpass",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(time.Hour),
			},
			wantValid: true,
		},
		{
			name: "expired credentials",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "testpass",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(-time.Hour),
			},
			wantValid: false,
		},
		{
			name: "empty username",
			ncData: config.NCData{
				Username:       "",
				Password:       "testpass",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(time.Hour),
			},
			wantValid: false,
		},
		{
			name: "empty password",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(time.Hour),
			},
			wantValid: false,
		},
		{
			name: "empty endpoint",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "testpass",
				Endpoint:       "",
				ExpirationDate: time.Now().UTC().Add(time.Hour),
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid := int64(123)

			cfg := &config.Config{
				AutoConnectData: config.AutoConnectData{ID: uid},
				TokensData: map[int64]config.TokenData{
					uid: {
						NCData: tt.ncData,
					},
				},
			}

			cfgManager := &mock.ConfigManager{Cfg: cfg}
			errorRegistry := internal.NewErrorHandlingRegistry[error]()

			renewCalled := false
			renewAPICall := func() (*session.NCCredentialsResponse, error) {
				renewCalled = true
				return &session.NCCredentialsResponse{
					Username:  "newuser",
					Password:  "newpass",
					Endpoint:  "https://new.example.com",
					ExpiresIn: 86400,
				}, nil
			}

			store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
			err := store.Renew()

			assert.NoError(t, err)
			assert.Equal(t, !tt.wantValid, renewCalled, "Renew should be called when credentials are invalid")
		})
	}
}

func TestNCCredentialsSessionStore_GetConfig(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	expectedTime := time.Now().UTC().Add(time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "testuser",
					Password:       "testpass",
					Endpoint:       "https://api.example.com",
					ExpirationDate: expectedTime,
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewCalled := false
	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		renewCalled = true
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)

	// We can't directly test getConfig since it's unexported, but we can verify
	// its behavior through the Renew method when credentials are valid
	err := store.Renew()
	assert.NoError(t, err)
	assert.False(t, renewCalled, "Renew should not be called when credentials are valid")
}

func TestNCCredentialsSessionStore_GetConfig_LoadError(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := &mock.ConfigManager{
		LoadErr: errors.New("load error"),
	}

	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)

	// When getConfig fails due to load error, validate() fails, but Renew() continues
	err := store.Renew()
	assert.NoError(t, err)

	// Verify the credentials were saved (creates new entry)
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[cfgManager.Cfg.AutoConnectData.ID].NCData.Username)
}

func TestNCCredentialsSessionStore_GetConfig_NoTokenData(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData:      map[int64]config.TokenData{}, // Empty map
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)

	// When there's no token data, getConfig returns error, but Renew() continues
	err := store.Renew()
	assert.NoError(t, err)

	// Verify new credentials were created
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[uid].NCData.Username)
	assert.Equal(t, "newpass", cfgManager.Cfg.TokensData[uid].NCData.Password)
	assert.Equal(t, "https://new.example.com", cfgManager.Cfg.TokensData[uid].NCData.Endpoint)
}

func TestNCCredentialsSessionStore_Renew_NilRenewalAPI(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "olduser",
					Password:       "oldpass",
					Endpoint:       "https://old.example.com",
					ExpirationDate: pastTime,
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "renewal API not configured")
}

func TestNCCredentialsSessionStore_Renew_NilResponse(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	pastTime := time.Now().UTC().Add(-24 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "olduser",
					Password:       "oldpass",
					Endpoint:       "https://old.example.com",
					ExpirationDate: pastTime,
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return nil, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "renewal API returned nil response")
}
