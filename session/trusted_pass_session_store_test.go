package session

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedPassSessionStore_Validate(t *testing.T) {
	tests := []struct {
		name              string
		token             string
		ownerID           string
		expiry            time.Time
		externalValidator TrustedPassExternalValidator
		wantErr           error
	}{
		{
			name:    "valid session",
			token:   "valid-token",
			ownerID: "nordvpn",
			expiry:  time.Now().Add(time.Hour),
			wantErr: nil,
		},
		{
			name:    "empty token",
			token:   "",
			ownerID: "nordvpn",
			expiry:  time.Now().Add(time.Hour),
			wantErr: ErrInvalidToken,
		},
		{
			name:    "expired session",
			token:   "valid-token",
			ownerID: "nordvpn",
			expiry:  time.Now().UTC().Add(-time.Hour),
			wantErr: ErrSessionExpired,
		},
		{
			name:    "invalid owner ID",
			token:   "valid-token",
			ownerID: "invalid",
			expiry:  time.Now().Add(time.Hour),
			wantErr: ErrInvalidOwnerId,
		},
		{
			name:    "external validator success",
			token:   "valid-token",
			ownerID: "nordvpn",
			expiry:  time.Now().Add(time.Hour),
			externalValidator: func(token string, ownerID string) error {
				assert.Equal(t, "valid-token", token)
				assert.Equal(t, "nordvpn", ownerID)
				return nil
			},
			wantErr: nil,
		},
		{
			name:    "external validator failure",
			token:   "valid-token",
			ownerID: "nordvpn",
			expiry:  time.Now().Add(time.Hour),
			externalValidator: func(token string, ownerID string) error {
				return errors.New("external validation failed")
			},
			wantErr: errors.New("external validation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := int64(123)
			tokenData := config.TokenData{
				TrustedPassToken:       tt.token,
				TrustedPassOwnerID:     tt.ownerID,
				TrustedPassTokenExpiry: tt.expiry.Format(internal.ServerDateFormat),
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg
			errRegistry := internal.NewErrorHandlingRegistry[error]()

			store := NewTrustedPassSessionStore(
				cfgManager,
				errRegistry,
				nil,
				tt.externalValidator,
			)

			err := store.(*TrustedPassSessionStore).Validate()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrustedPassSessionStore_SetToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userID := int64(123)
		initialToken := "initial-token"
		newToken := "new-token"

		tokenData := config.TokenData{
			TrustedPassToken:       initialToken,
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.(*TrustedPassSessionStore).SetToken(newToken)

		require.NoError(t, err)
		assert.Equal(t, newToken, cfgManager.Cfg.TokensData[userID].TrustedPassToken)
	})

	t.Run("load error", func(t *testing.T) {
		cfgManager := mock.NewMockConfigManager()
		cfgManager.LoadErr = errors.New("load failed")
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.(*TrustedPassSessionStore).SetToken("new-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load failed")
	})

	t.Run("save error", func(t *testing.T) {
		userID := int64(123)
		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: {}},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		cfgManager.SaveErr = errors.New("save failed")
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.(*TrustedPassSessionStore).SetToken("new-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "save failed")
	})

	t.Run("missing token data", func(t *testing.T) {
		userID := int64(123)
		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.(*TrustedPassSessionStore).SetToken("new-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non existing data")
	})
}

func TestTrustedPassSessionStore_SetOwnerID(t *testing.T) {
	userID := int64(123)
	initialOwnerID := "initial-owner"
	newOwnerID := "new-owner"

	tokenData := config.TokenData{
		TrustedPassToken:       "token",
		TrustedPassOwnerID:     initialOwnerID,
		TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
	}

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData:      map[int64]config.TokenData{userID: tokenData},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg
	errRegistry := internal.NewErrorHandlingRegistry[error]()

	store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

	err := store.(*TrustedPassSessionStore).SetOwnerID(newOwnerID)

	require.NoError(t, err)
	assert.Equal(t, newOwnerID, cfgManager.Cfg.TokensData[userID].TrustedPassOwnerID)
}

func TestTrustedPassSessionStore_SetExpiry(t *testing.T) {
	userID := int64(123)
	newExpiry := time.Now().Add(2 * time.Hour)

	tokenData := config.TokenData{
		TrustedPassToken:       "token",
		TrustedPassOwnerID:     "nordvpn",
		TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
	}

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: userID},
		TokensData:      map[int64]config.TokenData{userID: tokenData},
	}

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg
	errRegistry := internal.NewErrorHandlingRegistry[error]()

	store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

	err := store.(*TrustedPassSessionStore).SetExpiry(newExpiry)

	require.NoError(t, err)
	assert.Equal(t, newExpiry.Format(internal.ServerDateFormat), cfgManager.Cfg.TokensData[userID].TrustedPassTokenExpiry)
}

func TestTrustedPassSessionStore_GetToken(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		loadErr   error
		wantToken string
	}{
		{
			name:      "success",
			token:     "test-token",
			loadErr:   nil,
			wantToken: "test-token",
		},
		{
			name:      "load error",
			token:     "test-token",
			loadErr:   errors.New("load failed"),
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := int64(123)
			tokenData := config.TokenData{
				TrustedPassToken: tt.token,
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg
			cfgManager.LoadErr = tt.loadErr
			errRegistry := internal.NewErrorHandlingRegistry[error]()

			store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

			token := store.(*TrustedPassSessionStore).GetToken()

			assert.Equal(t, tt.wantToken, token)
		})
	}
}

func TestTrustedPassSessionStore_Invalidate(t *testing.T) {
	t.Run("with registered handler", func(t *testing.T) {
		cfgManager := mock.NewMockConfigManager()
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		handlerCalled := false
		var handlerErr error

		testErr := errors.New("test error")
		errRegistry.Add(func(err error) {
			handlerCalled = true
			handlerErr = err
		}, testErr)

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.Invalidate(testErr)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, testErr, handlerErr)
	})

	t.Run("no registered handler", func(t *testing.T) {
		cfgManager := mock.NewMockConfigManager()
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		testErr := errors.New("unhandled error")
		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.Invalidate(testErr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalidating session: unhandled error")
	})
}

func TestTrustedPassSessionStore_EdgeCases(t *testing.T) {
	t.Run("validate with invalid expiry format", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassToken:       "valid-token",
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: "invalid-date-format",
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.(*TrustedPassSessionStore).Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrSessionExpired, err)
	})

	t.Run("GetExpiry with invalid date format", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassTokenExpiry: "invalid-date-format",
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		expiry := store.(*TrustedPassSessionStore).GetExpiry()

		assert.True(t, time.Now().After(expiry))
	})

	t.Run("IsExpired with invalid date format", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassTokenExpiry: "invalid-date-format",
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		expired := store.(*TrustedPassSessionStore).IsExpired()

		assert.True(t, expired)
	})
}

func TestTrustedPassSessionStore_GetOwnerID(t *testing.T) {
	tests := []struct {
		name        string
		ownerID     string
		loadErr     error
		wantOwnerID string
	}{
		{
			name:        "success",
			ownerID:     "nordvpn",
			loadErr:     nil,
			wantOwnerID: "nordvpn",
		},
		{
			name:        "load error",
			ownerID:     "nordvpn",
			loadErr:     errors.New("load failed"),
			wantOwnerID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := int64(123)
			tokenData := config.TokenData{
				TrustedPassOwnerID: tt.ownerID,
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg
			cfgManager.LoadErr = tt.loadErr
			errRegistry := internal.NewErrorHandlingRegistry[error]()

			store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

			ownerID := store.(*TrustedPassSessionStore).GetOwnerID()

			assert.Equal(t, tt.wantOwnerID, ownerID)
		})
	}
}

func TestTrustedPassSessionStore_GetExpiry(t *testing.T) {
	tests := []struct {
		name       string
		expiry     time.Time
		loadErr    error
		wantExpiry time.Time
	}{
		{
			name:       "success",
			expiry:     time.Now().UTC().Add(time.Hour).Truncate(time.Second),
			loadErr:    nil,
			wantExpiry: time.Now().UTC().Add(time.Hour).Truncate(time.Second),
		},
		{
			name:       "load error",
			expiry:     time.Now().Add(time.Hour),
			loadErr:    errors.New("load failed"),
			wantExpiry: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := int64(123)
			tokenData := config.TokenData{
				TrustedPassTokenExpiry: tt.expiry.Format(internal.ServerDateFormat),
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg
			cfgManager.LoadErr = tt.loadErr
			errRegistry := internal.NewErrorHandlingRegistry[error]()

			store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

			expiry := store.(*TrustedPassSessionStore).GetExpiry()

			assert.Equal(t, tt.wantExpiry.Unix(), expiry.Unix())
		})
	}
}

func TestTrustedPassSessionStore_IsExpired(t *testing.T) {
	tests := []struct {
		name        string
		expiry      time.Time
		loadErr     error
		wantExpired bool
	}{
		{
			name:        "not expired",
			expiry:      time.Now().Add(time.Hour),
			loadErr:     nil,
			wantExpired: false,
		},
		{
			name:        "expired",
			expiry:      time.Now().UTC().Add(-time.Hour),
			loadErr:     nil,
			wantExpired: true,
		},
		{
			name:        "load error",
			expiry:      time.Now().Add(time.Hour),
			loadErr:     errors.New("load failed"),
			wantExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := int64(123)
			tokenData := config.TokenData{
				TrustedPassTokenExpiry: tt.expiry.Format(internal.ServerDateFormat),
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: userID},
				TokensData:      map[int64]config.TokenData{userID: tokenData},
			}

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg
			cfgManager.LoadErr = tt.loadErr
			errRegistry := internal.NewErrorHandlingRegistry[error]()

			store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

			expired := store.(*TrustedPassSessionStore).IsExpired()

			assert.Equal(t, tt.wantExpired, expired)
		})
	}
}

func TestTrustedPassSessionStore_Renew(t *testing.T) {
	t.Run("valid session does not renew", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassToken:       "valid-token",
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
			IsOAuth:                true,
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		renewCalled := false
		renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
			renewCalled = true
			return nil, nil
		}

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, renewAPICall, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.False(t, renewCalled, "Renew API should not be called for valid session")
	})

	t.Run("invalid session triggers renewal", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassToken:       "",
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
			IsOAuth:                true,
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		renewCalled := false
		renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
			renewCalled = true
			return &TrustedPassAccessTokenResponse{
				Token:   "new-token",
				OwnerID: "nordvpn",
			}, nil
		}

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, renewAPICall, nil)

		err := store.Renew()

		assert.NoError(t, err)
		assert.True(t, renewCalled, "Renew API should be called for invalid session")
	})

	t.Run("nil renewal API", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassToken:       "",
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
			IsOAuth:                true,
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "renewal API call not configured")
	})

	t.Run("renewal API returns nil response", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassToken:       "",
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
			IsOAuth:                true,
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
			return nil, nil
		}

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, renewAPICall, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "renewal API returned nil response")
	})

	t.Run("renewal API returns empty token", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassToken:       "",
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
			IsOAuth:                true,
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
			return &TrustedPassAccessTokenResponse{
				Token:   "",
				OwnerID: "nordvpn",
			}, nil
		}

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, renewAPICall, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "renewal API returned empty token")
	})

	t.Run("renewal API error", func(t *testing.T) {
		userID := int64(123)
		tokenData := config.TokenData{
			TrustedPassToken:       "",
			TrustedPassOwnerID:     "nordvpn",
			TrustedPassTokenExpiry: time.Now().Add(time.Hour).Format(internal.ServerDateFormat),
			IsOAuth:                true,
		}

		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{ID: userID},
			TokensData:      map[int64]config.TokenData{userID: tokenData},
		}

		cfgManager := mock.NewMockConfigManager()
		cfgManager.Cfg = &cfg
		errRegistry := internal.NewErrorHandlingRegistry[error]()

		renewAPICall := func(token string) (*TrustedPassAccessTokenResponse, error) {
			return nil, errors.New("API error")
		}

		store := NewTrustedPassSessionStore(cfgManager, errRegistry, renewAPICall, nil)

		err := store.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})
}
