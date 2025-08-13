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

const (
	// testUserID is the default user ID used in tests
	testUserID = int64(123)
	// testExpiredDuration represents a duration in the past for expired credentials
	testExpiredDuration = -24 * time.Hour
	// testValidDuration represents a duration in the future for valid credentials
	testValidDuration = 24 * time.Hour
)

func TestNCCredentialsSessionStore_Renew_NotExpired(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID
	futureTime := time.Now().UTC().Add(testValidDuration)

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)

	err := store.Renew()
	assert.NoError(t, err)
	assert.False(t, renewCalled, "Renew API should not be called when credentials are valid")
}

func TestNCCredentialsSessionStore_Renew_NoTokenData(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)

	err := store.Renew()
	// When validate() fails due to config load error, Renew() continues and calls the renewal API
	assert.NoError(t, err)
	assert.True(t, renewCalled, "Renew API should be called even when config load fails")

	// Verify the credentials were saved
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[cfgManager.Cfg.AutoConnectData.ID].NCData.Username)
}

func TestNCCredentialsSessionStore_Renew_ExpiredCredentials(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID
	pastTime := time.Now().UTC().Add(testExpiredDuration)

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
			ExpiresIn: 86400 * time.Second, // 24 hours as Duration
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)

	beforeRenew := time.Now().UTC()
	err := store.Renew()
	afterRenew := time.Now().UTC()

	assert.NoError(t, err)
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[uid].NCData.Username)
	assert.Equal(t, "newpass", cfgManager.Cfg.TokensData[uid].NCData.Password)
	assert.Equal(t, "https://new.example.com", cfgManager.Cfg.TokensData[uid].NCData.Endpoint)

	// Verify the expiration date is approximately 24 hours from now
	actualExpiration := cfgManager.Cfg.TokensData[uid].NCData.ExpirationDate

	// The expiration should be set to approximately 24 hours from when the renewal happened
	assert.True(t, actualExpiration.After(beforeRenew.Add(23*time.Hour)),
		"Expiration should be at least 23 hours from before renewal. Got: %v, Expected after: %v",
		actualExpiration, beforeRenew.Add(23*time.Hour))
	assert.True(t, actualExpiration.Before(afterRenew.Add(25*time.Hour)),
		"Expiration should be at most 25 hours from after renewal. Got: %v, Expected before: %v",
		actualExpiration, afterRenew.Add(25*time.Hour))
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
			uid := testUserID
			pastTime := time.Now().UTC().Add(testExpiredDuration)

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

			store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
			err := store.Renew()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNCCredentialsSessionStore_Renew_APIError(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID
	pastTime := time.Now().UTC().Add(testExpiredDuration)
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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api-error")
	assert.True(t, handlerCalled)
}

func TestNCCredentialsSessionStore_Renew_SaveError(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID
	pastTime := time.Now().UTC().Add(testExpiredDuration)

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
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
			wantErr:         false,
			wantErrContains: "",
			wantHandlerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TokensData: map[int64]config.TokenData{
					testUserID: {},
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

			store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, nil, nil)
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
			uid := testUserID

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

			store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
			err := store.Renew()

			assert.NoError(t, err)
			assert.Equal(t, !tt.wantValid, renewCalled, "Renew should be called when credentials are invalid")
		})
	}
}

func TestNCCredentialsSessionStore_GetConfig(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID
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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)

	// When getConfig fails due to load error, validate() fails, but Renew() continues
	err := store.Renew()
	assert.NoError(t, err)

	// Verify the credentials were saved (creates new entry)
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[cfgManager.Cfg.AutoConnectData.ID].NCData.Username)
}

func TestNCCredentialsSessionStore_GetConfig_NoTokenData(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)

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

	uid := testUserID
	pastTime := time.Now().UTC().Add(testExpiredDuration)

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, nil, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "renewal API not configured")
}

func TestNCCredentialsSessionStore_Renew_NilResponse(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID
	pastTime := time.Now().UTC().Add(testExpiredDuration)

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

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, nil)
	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "renewal API returned nil response")
}

func TestNCCredentialsSessionStore_ExternalValidator(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		ncData            config.NCData
		externalValidator session.NCCredentialsExternalValidator
		wantRenewCalled   bool
		wantErr           bool
		wantErrContains   string
	}{
		{
			name: "valid credentials with passing external validator",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "testpass",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(time.Hour),
			},
			externalValidator: func(username, password, endpoint string) error {
				// Validator passes
				return nil
			},
			wantRenewCalled: false,
			wantErr:         false,
		},
		{
			name: "valid credentials with failing external validator",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "testpass",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(time.Hour),
			},
			externalValidator: func(username, password, endpoint string) error {
				return errors.New("external validation failed")
			},
			wantRenewCalled: true,
			wantErr:         false,
		},
		{
			name: "valid credentials with nil external validator",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "testpass",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(time.Hour),
			},
			externalValidator: nil,
			wantRenewCalled:   false,
			wantErr:           false,
		},
		{
			name: "expired credentials with external validator",
			ncData: config.NCData{
				Username:       "testuser",
				Password:       "testpass",
				Endpoint:       "https://api.example.com",
				ExpirationDate: time.Now().UTC().Add(-time.Hour),
			},
			externalValidator: func(username, password, endpoint string) error {
				// Should not be called for expired credentials
				return errors.New("should not be called")
			},
			wantRenewCalled: true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid := testUserID

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
					ExpiresIn: 86400 * time.Second,
				}, nil
			}

			store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, tt.externalValidator)
			err := store.Renew()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantRenewCalled, renewCalled,
				"Renew API called = %v, want %v", renewCalled, tt.wantRenewCalled)
		})
	}
}

func TestNCCredentialsSessionStore_ExternalValidator_VerifyParameters(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID
	expectedUsername := "testuser"
	expectedPassword := "testpass"
	expectedEndpoint := "https://api.example.com"

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       expectedUsername,
					Password:       expectedPassword,
					Endpoint:       expectedEndpoint,
					ExpirationDate: time.Now().UTC().Add(time.Hour),
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	validatorCalled := false
	var actualUsername, actualPassword, actualEndpoint string

	externalValidator := func(username, password, endpoint string) error {
		validatorCalled = true
		actualUsername = username
		actualPassword = password
		actualEndpoint = endpoint
		return nil
	}

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, externalValidator)
	err := store.Renew()

	assert.NoError(t, err)
	assert.True(t, validatorCalled, "External validator should have been called")
	assert.Equal(t, expectedUsername, actualUsername, "Username passed to validator")
	assert.Equal(t, expectedPassword, actualPassword, "Password passed to validator")
	assert.Equal(t, expectedEndpoint, actualEndpoint, "Endpoint passed to validator")
}

func TestNCCredentialsSessionStore_ExternalValidator_WithRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	uid := testUserID

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				NCData: config.NCData{
					Username:       "olduser",
					Password:       "oldpass",
					Endpoint:       "https://old.example.com",
					ExpirationDate: time.Now().UTC().Add(time.Hour),
				},
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()

	// External validator that fails, triggering renewal
	externalValidator := func(username, password, endpoint string) error {
		return errors.New("credentials invalid")
	}

	renewAPICall := func() (*session.NCCredentialsResponse, error) {
		return &session.NCCredentialsResponse{
			Username:  "newuser",
			Password:  "newpass",
			Endpoint:  "https://new.example.com",
			ExpiresIn: 86400 * time.Second,
		}, nil
	}

	store := session.NewNCCredentialsSessionStore(cfgManager, errorRegistry, renewAPICall, externalValidator)
	err := store.Renew()

	assert.NoError(t, err)
	// Verify the credentials were renewed
	assert.Equal(t, "newuser", cfgManager.Cfg.TokensData[uid].NCData.Username)
	assert.Equal(t, "newpass", cfgManager.Cfg.TokensData[uid].NCData.Password)
	assert.Equal(t, "https://new.example.com", cfgManager.Cfg.TokensData[uid].NCData.Endpoint)
}
