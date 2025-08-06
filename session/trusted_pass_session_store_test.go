package session

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
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
