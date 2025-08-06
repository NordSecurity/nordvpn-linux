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

// testValidate is a test helper that tests the validation logic of TrustedPassSessionStore
// without going through the full Renew flow
func testValidate(store SessionStore) error {
	// Type assert to access the unexported validate method
	tpStore, ok := store.(*TrustedPassSessionStore)
	if !ok {
		return errors.New("not a TrustedPassSessionStore")
	}
	return tpStore.validate()
}

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
			expiry:  time.Now().UTC().Add(time.Hour),
			wantErr: nil,
		},
		{
			name:    "empty token",
			token:   "",
			ownerID: "nordvpn",
			expiry:  time.Now().UTC().Add(time.Hour),
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
			expiry:  time.Now().UTC().Add(time.Hour),
			wantErr: ErrInvalidOwnerID,
		},
		{
			name:    "external validator success",
			token:   "valid-token",
			ownerID: "nordvpn",
			expiry:  time.Now().UTC().Add(time.Hour),
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
			expiry:  time.Now().UTC().Add(time.Hour),
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

			// Test validation directly using the test helper
			err := testValidate(store)

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
			}

			store := NewTrustedPassSessionStore(cfgManager, errRegistry, nil, nil)
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

func TestTrustedPassSessionStore_ValidateWithInvalidExpiryFormat(t *testing.T) {
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

	err := testValidate(store)

	assert.Error(t, err)
	assert.Equal(t, ErrSessionExpired, err)
}

func TestTrustedPassSessionStore_Renew(t *testing.T) {
	userID := int64(123)

	tests := []struct {
		name            string
		tokenData       config.TokenData
		renewAPICall    TrustedPassRenewalAPICall
		wantErr         bool
		wantErrContains string
		checkRenewCall  func(t *testing.T, renewCalled bool)
	}{
		{
			name: "valid session does not renew",
			tokenData: config.TokenData{
				TrustedPassToken:       "valid-token",
				TrustedPassOwnerID:     "nordvpn",
				TrustedPassTokenExpiry: time.Now().UTC().Add(time.Hour).Format(internal.ServerDateFormat),
				IsOAuth:                true,
			},
			renewAPICall: func(token string) (*TrustedPassAccessTokenResponse, error) {
				assert.Fail(t, "Renew API should not be called for valid session")
				return nil, nil
			},
			wantErr: false,
			checkRenewCall: func(t *testing.T, renewCalled bool) {
				assert.False(t, renewCalled, "Renew API should not be called for valid session")
			},
		},
		{
			name: "invalid session triggers renewal",
			tokenData: config.TokenData{
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "nordvpn",
				TrustedPassTokenExpiry: time.Now().UTC().Add(time.Hour).Format(internal.ServerDateFormat),
				IsOAuth:                true,
			},
			renewAPICall: func(token string) (*TrustedPassAccessTokenResponse, error) {
				return &TrustedPassAccessTokenResponse{
					Token:   "new-token",
					OwnerID: "nordvpn",
				}, nil
			},
			wantErr: false,
			checkRenewCall: func(t *testing.T, renewCalled bool) {
				assert.True(t, renewCalled, "Renew API should be called for invalid session")
			},
		},
		{
			name: "nil renewal API",
			tokenData: config.TokenData{
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "nordvpn",
				TrustedPassTokenExpiry: time.Now().UTC().Add(time.Hour).Format(internal.ServerDateFormat),
				IsOAuth:                true,
			},
			renewAPICall: nil,
			wantErr:      false, // HandleError returns nil when no handlers are registered
		},
		{
			name: "renewal API returns nil response",
			tokenData: config.TokenData{
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "nordvpn",
				TrustedPassTokenExpiry: time.Now().UTC().Add(time.Hour).Format(internal.ServerDateFormat),
				IsOAuth:                true,
			},
			renewAPICall: func(token string) (*TrustedPassAccessTokenResponse, error) {
				return nil, nil
			},
			wantErr: false, // HandleError returns nil when no handlers are registered
		},
		{
			name: "renewal API returns empty token",
			tokenData: config.TokenData{
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "nordvpn",
				TrustedPassTokenExpiry: time.Now().UTC().Add(time.Hour).Format(internal.ServerDateFormat),
				IsOAuth:                true,
			},
			renewAPICall: func(token string) (*TrustedPassAccessTokenResponse, error) {
				return &TrustedPassAccessTokenResponse{
					Token:   "",
					OwnerID: "nordvpn",
				}, nil
			},
			wantErr: false, // HandleError returns nil when no handlers are registered
		},
		{
			name: "renewal API error",
			tokenData: config.TokenData{
				TrustedPassToken:       "",
				TrustedPassOwnerID:     "nordvpn",
				TrustedPassTokenExpiry: time.Now().UTC().Add(time.Hour).Format(internal.ServerDateFormat),
				IsOAuth:                true,
			},
			renewAPICall: func(token string) (*TrustedPassAccessTokenResponse, error) {
				return nil, errors.New("API error")
			},
			wantErr: false, // HandleError returns nil when no handlers are registered
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
			var renewAPICall TrustedPassRenewalAPICall
			if tt.renewAPICall != nil {
				originalCall := tt.renewAPICall
				renewAPICall = func(token string) (*TrustedPassAccessTokenResponse, error) {
					renewCalled = true
					return originalCall(token)
				}
			}

			store := NewTrustedPassSessionStore(cfgManager, errRegistry, renewAPICall, nil)

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
		})
	}
}
