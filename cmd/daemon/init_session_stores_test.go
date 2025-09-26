package main

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	mockauth "github.com/NordSecurity/nordvpn-linux/test/mock/auth"
	mocknetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedToken      = "ab78bb36299d442fa0715fb53b5e3e57"
	expectedRenewToken = "cd89cc47300e553fb1826fc64c6f4f68"
)

type mockRawClientAPI struct {
	tokenRenewFunc              func(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error)
	serviceCredentialsFunc      func(token string) (*core.CredentialsResponse, error)
	trustedPassTokenFunc        func(token string) (*core.TrustedPassTokenResponse, error)
	notificationCredentialsFunc func(token, appUserID string) (core.NotificationCredentialsResponse, error)
}

func (m *mockRawClientAPI) TokenRenew(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
	if m.tokenRenewFunc != nil {
		return m.tokenRenewFunc(token, idempotencyKey)
	}
	return &core.TokenRenewResponse{
		Token:      "new-token",
		RenewToken: "new-renew-token",
		ExpiresAt:  time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
	}, nil
}

func (m *mockRawClientAPI) ServiceCredentials(token string) (*core.CredentialsResponse, error) {
	if m.serviceCredentialsFunc != nil {
		return m.serviceCredentialsFunc(token)
	}
	return &core.CredentialsResponse{
		Username:           "vpn-user",
		Password:           "vpn-pass",
		NordlynxPrivateKey: "nordlynx-key",
	}, nil
}

func (m *mockRawClientAPI) TrustedPassToken(token string) (*core.TrustedPassTokenResponse, error) {
	if m.trustedPassTokenFunc != nil {
		return m.trustedPassTokenFunc(token)
	}
	return &core.TrustedPassTokenResponse{
		Token:   "trusted-pass-token",
		OwnerID: "nordvpn",
	}, nil
}

func (m *mockRawClientAPI) NotificationCredentials(token, appUserID string) (core.NotificationCredentialsResponse, error) {
	if m.notificationCredentialsFunc != nil {
		return m.notificationCredentialsFunc(token, appUserID)
	}
	return core.NotificationCredentialsResponse{
		Username:  "nc-user",
		Password:  "nc-pass",
		Endpoint:  "wss://nc.example.com",
		ExpiresIn: 3600,
	}, nil
}

func (m *mockRawClientAPI) NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error) {
	return core.NotificationCredentialsRevokeResponse{}, nil
}
func (m *mockRawClientAPI) Services(token string) (core.ServicesResponse, error) {
	return nil, nil
}
func (m *mockRawClientAPI) CurrentUser(token string) (*core.CurrentUserResponse, error) {
	return &core.CurrentUserResponse{}, nil
}
func (m *mockRawClientAPI) DeleteToken(token string) error {
	return nil
}
func (m *mockRawClientAPI) MultifactorAuthStatus(token string) (*core.MultifactorAuthStatusResponse, error) {
	return &core.MultifactorAuthStatusResponse{}, nil
}
func (m *mockRawClientAPI) Logout(token string) error {
	return nil
}

func (m *mockRawClientAPI) Insights() (*core.Insights, error) {
	return &core.Insights{}, nil
}

func (m *mockRawClientAPI) Servers() (core.Servers, http.Header, error) {
	return core.Servers{}, nil, nil
}
func (m *mockRawClientAPI) RecommendedServers(filter core.ServersFilter, longitude, latitude float64) (core.Servers, http.Header, error) {
	return core.Servers{}, nil, nil
}
func (m *mockRawClientAPI) Server(id int64) (*core.Server, error) {
	return &core.Server{}, nil
}
func (m *mockRawClientAPI) ServersCountries() (core.Countries, http.Header, error) {
	return core.Countries{}, nil, nil
}

func (m *mockRawClientAPI) Base() string {
	return "https://api.test.com"
}

func (m *mockRawClientAPI) Plans() (*core.Plans, error) {
	return &core.Plans{}, nil
}
func (m *mockRawClientAPI) CreateUser(email, password string) (*core.UserCreateResponse, error) {
	return &core.UserCreateResponse{}, nil
}
func (m *mockRawClientAPI) Orders(token string) ([]core.Order, error) {
	return []core.Order{}, nil
}
func (m *mockRawClientAPI) Payments(token string) ([]core.PaymentResponse, error) {
	return []core.PaymentResponse{}, nil
}

func (m *mockRawClientAPI) Register(token string, peer mesh.Machine) (*mesh.Machine, error) {
	return &mesh.Machine{}, nil
}
func (m *mockRawClientAPI) Update(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
	return nil
}
func (m *mockRawClientAPI) Configure(token string, id uuid.UUID, peerID uuid.UUID, peerUpdateInfo mesh.PeerUpdateRequest) error {
	return nil
}
func (m *mockRawClientAPI) Unregister(token string, self uuid.UUID) error {
	return nil
}
func (m *mockRawClientAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	return &mesh.MachineMap{}, nil
}
func (m *mockRawClientAPI) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
	return nil
}
func (m *mockRawClientAPI) Invite(token string, self uuid.UUID, email string, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool) error {
	return nil
}
func (m *mockRawClientAPI) Received(token string, self uuid.UUID) (mesh.Invitations, error) {
	return mesh.Invitations{}, nil
}
func (m *mockRawClientAPI) Sent(token string, self uuid.UUID) (mesh.Invitations, error) {
	return mesh.Invitations{}, nil
}
func (m *mockRawClientAPI) Accept(token string, self uuid.UUID, invitation uuid.UUID, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool) error {
	return nil
}
func (m *mockRawClientAPI) Reject(token string, self uuid.UUID, invitation uuid.UUID) error {
	return nil
}
func (m *mockRawClientAPI) Revoke(token string, self uuid.UUID, invitation uuid.UUID) error {
	return nil
}
func (m *mockRawClientAPI) NotifyNewTransfer(token string, self uuid.UUID, peer uuid.UUID, fileName string, fileCount int, transferID string) error {
	return nil
}

type mockClientAPI struct {
	*mockRawClientAPI
}

func (m *mockClientAPI) NotificationCredentials(appUserID string) (core.NotificationCredentialsResponse, error) {
	return m.mockRawClientAPI.NotificationCredentials("", appUserID)
}
func (m *mockClientAPI) NotificationCredentialsRevoke(appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error) {
	return m.mockRawClientAPI.NotificationCredentialsRevoke("", appUserID, purgeSession)
}
func (m *mockClientAPI) Services() (core.ServicesResponse, error) {
	return m.mockRawClientAPI.Services("")
}
func (m *mockClientAPI) CurrentUser() (*core.CurrentUserResponse, error) {
	return m.mockRawClientAPI.CurrentUser("")
}
func (m *mockClientAPI) DeleteToken() error {
	return m.mockRawClientAPI.DeleteToken("")
}
func (m *mockClientAPI) TrustedPassToken() (*core.TrustedPassTokenResponse, error) {
	return m.mockRawClientAPI.TrustedPassToken("")
}
func (m *mockClientAPI) MultifactorAuthStatus() (*core.MultifactorAuthStatusResponse, error) {
	return m.mockRawClientAPI.MultifactorAuthStatus("")
}
func (m *mockClientAPI) Logout() error {
	return m.mockRawClientAPI.Logout("")
}
func (m *mockClientAPI) Orders() ([]core.Order, error) {
	return m.mockRawClientAPI.Orders("")
}
func (m *mockClientAPI) Payments() ([]core.PaymentResponse, error) {
	return m.mockRawClientAPI.Payments("")
}

func setupTestConfig() *config.Config {
	uid := int64(123)
	idempotencyKey := uuid.New()
	userID := uuid.New()

	return &config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          expectedToken,
				RenewToken:     expectedRenewToken,
				TokenExpiry:    time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
				IdempotencyKey: &idempotencyKey,
				NCData: config.NCData{
					UserID: userID,
				},
			},
		},
	}
}

func TestNewSessionStoresBuilder(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = setupTestConfig()
	builder := NewSessionStoresBuilder(cfgManager)

	assert.NotNil(t, builder)
	assert.Equal(t, cfgManager, builder.confman)
	assert.NotNil(t, builder.registries.accessToken)
	assert.NotNil(t, builder.registries.vpnCreds)
	assert.NotNil(t, builder.registries.trustedPass)
	assert.NotNil(t, builder.registries.ncCreds)
}

func TestBuildAccessTokenStore(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = setupTestConfig()
	builder := NewSessionStoresBuilder(cfgManager)
	mockAPI := &mockRawClientAPI{}

	store := builder.BuildAccessTokenStore(mockAPI)

	assert.NotNil(t, store)
	assert.Equal(t, store, builder.stores.accessToken)
}

func TestBuildVPNCredsStore(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = setupTestConfig()
	builder := NewSessionStoresBuilder(cfgManager)
	mockAPI := &mockClientAPI{mockRawClientAPI: &mockRawClientAPI{}}

	store := builder.BuildVPNCredsStore(mockAPI)

	assert.NotNil(t, store)
	assert.Equal(t, store, builder.stores.vpnCreds)
}

func TestBuildTrustedPassStore(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = setupTestConfig()
	builder := NewSessionStoresBuilder(cfgManager)
	mockAPI := &mockClientAPI{mockRawClientAPI: &mockRawClientAPI{}}

	store := builder.BuildTrustedPassStore(mockAPI)

	assert.NotNil(t, store)
	assert.Equal(t, store, builder.stores.trustedPass)
}

func TestBuildNCCredsStore(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = setupTestConfig()
	builder := NewSessionStoresBuilder(cfgManager)
	mockAPI := &mockClientAPI{mockRawClientAPI: &mockRawClientAPI{}}

	store := builder.BuildNCCredsStore(mockAPI)

	assert.NotNil(t, store)
	assert.Equal(t, store, builder.stores.ncCreds)
}

func TestGetStores(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = setupTestConfig()
	builder := NewSessionStoresBuilder(cfgManager)
	mockRawAPI := &mockRawClientAPI{}
	mockAPI := &mockClientAPI{mockRawClientAPI: mockRawAPI}

	// Build all stores
	builder.BuildAccessTokenStore(mockRawAPI)
	builder.BuildVPNCredsStore(mockAPI)
	builder.BuildTrustedPassStore(mockAPI)
	builder.BuildNCCredsStore(mockAPI)

	stores := builder.GetStores()

	assert.Len(t, stores, 4)
	assert.Contains(t, stores, builder.stores.accessToken)
	assert.Contains(t, stores, builder.stores.vpnCreds)
	assert.Contains(t, stores, builder.stores.trustedPass)
	assert.Contains(t, stores, builder.stores.ncCreds)
}

func TestAccessTokenStore_SuccessfulRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := setupTestConfig()
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = cfg
	builder := NewSessionStoresBuilder(cfgManager)

	mockAPI := &mockRawClientAPI{
		tokenRenewFunc: func(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
			return &core.TokenRenewResponse{
				Token:      "renewed-token",
				RenewToken: "renewed-renew-token",
				ExpiresAt:  time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
			}, nil
		},
	}

	store := builder.BuildAccessTokenStore(mockAPI)
	err := store.Renew()

	assert.NoError(t, err)
}

func TestVPNCredsStore_SuccessfulRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := setupTestConfig()
	uid := cfg.AutoConnectData.ID
	data := cfg.TokensData[uid]
	data.OpenVPNUsername = ""
	data.OpenVPNPassword = ""
	cfg.TokensData[uid] = data

	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = cfg
	builder := NewSessionStoresBuilder(cfgManager)

	mockRawAPI := &mockRawClientAPI{}
	mockAPI := &mockClientAPI{
		mockRawClientAPI: &mockRawClientAPI{
			serviceCredentialsFunc: func(token string) (*core.CredentialsResponse, error) {
				return &core.CredentialsResponse{
					Username:           "vpn-user",
					Password:           "vpn-pass",
					NordlynxPrivateKey: "nordlynx-key",
				}, nil
			},
		},
	}

	builder.BuildAccessTokenStore(mockRawAPI)
	store := builder.BuildVPNCredsStore(mockAPI)
	err := store.Renew()

	assert.NoError(t, err)
}

func TestRenewalFunctions_AccessToken(t *testing.T) {
	category.Set(t, category.Unit)

	mockAPI := &mockRawClientAPI{
		tokenRenewFunc: func(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
			assert.Equal(t, expectedToken, token)
			assert.NotEqual(t, uuid.Nil, idempotencyKey)

			return &core.TokenRenewResponse{
				Token:      "renewed-token",
				RenewToken: "renewed-renew-token",
				ExpiresAt:  "2025-01-01 00:00:00",
			}, nil
		},
	}

	renewFunc := renewAccessToken(mockAPI)

	key := uuid.New()
	resp, err := renewFunc(expectedToken, key)

	require.NoError(t, err)
	assert.Equal(t, "renewed-token", resp.Token)
	assert.Equal(t, "renewed-renew-token", resp.RenewToken)
	assert.Equal(t, "2025-01-01 00:00:00", resp.ExpiresAt)
}

func TestRenewalFunctions_TrustedPass(t *testing.T) {
	category.Set(t, category.Unit)

	mockAPI := &mockClientAPI{
		mockRawClientAPI: &mockRawClientAPI{
			trustedPassTokenFunc: func(token string) (*core.TrustedPassTokenResponse, error) {
				return &core.TrustedPassTokenResponse{
					Token:   "trusted-token",
					OwnerID: "nordvpn",
				}, nil
			},
		},
	}

	renewFunc := renewTrustedPass(mockAPI)
	resp, err := renewFunc("test-token")

	require.NoError(t, err)
	assert.Equal(t, "trusted-token", resp.Token)
	assert.Equal(t, "nordvpn", resp.OwnerID)
}

func TestRenewalFunctions_VPNCredentials(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := setupTestConfig()
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = cfg

	mockAPI := &mockClientAPI{
		mockRawClientAPI: &mockRawClientAPI{
			serviceCredentialsFunc: func(token string) (*core.CredentialsResponse, error) {
				assert.Equal(t, expectedToken, token)

				return &core.CredentialsResponse{
					Username:           "vpn-user",
					Password:           "vpn-pass",
					NordlynxPrivateKey: "nordlynx-key",
				}, nil
			},
		},
	}

	renewFunc := renewVPNCredentials(cfgManager, mockAPI)
	resp, err := renewFunc()

	require.NoError(t, err)
	assert.Equal(t, "vpn-user", resp.Username)
	assert.Equal(t, "vpn-pass", resp.Password)
	assert.Equal(t, "nordlynx-key", resp.NordLynxPrivateKey)
}

func TestRenewalFunctions_NCCredentials(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := setupTestConfig()
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = cfg

	mockAPI := &mockClientAPI{
		mockRawClientAPI: &mockRawClientAPI{
			notificationCredentialsFunc: func(token, appUserID string) (core.NotificationCredentialsResponse, error) {
				assert.Equal(t, cfg.TokensData[cfg.AutoConnectData.ID].NCData.UserID.String(), appUserID)

				return core.NotificationCredentialsResponse{
					Username:  "nc-user",
					Password:  "nc-pass",
					Endpoint:  "wss://nc.example.com",
					ExpiresIn: 3600,
				}, nil
			},
		},
	}

	renewFunc := renewNCCredentials(cfgManager, mockAPI)
	resp, err := renewFunc()

	require.NoError(t, err)
	assert.Equal(t, "nc-user", resp.Username)
	assert.Equal(t, "nc-pass", resp.Password)
	assert.Equal(t, "wss://nc.example.com", resp.Endpoint)
	assert.Equal(t, time.Duration(3600), resp.ExpiresIn)
}

func TestGetTokenData_Success(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := setupTestConfig()
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = cfg

	data, err := getTokenData(cfgManager)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, expectedToken, data.Token)
	assert.Equal(t, expectedRenewToken, data.RenewToken)
}

func TestGetTokenData_NoTokenData(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: 999},
		TokensData:      map[int64]config.TokenData{},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = cfg

	data, err := getTokenData(cfgManager)

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "token data not found")
}

func TestGetTokenData_ConfigLoadError(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := mock.NewMockConfigManager()
	cfgManager.LoadErr = errors.New("config load failed")

	data, err := getTokenData(cfgManager)

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "config load failed")
}

func TestLogoutReasonCodeSelectionWithProductionCode(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name               string
		sessionStore       string
		apiError           error
		expectedReasonCode events.ReasonCode
	}{
		{
			name:               "access token - not found error",
			sessionStore:       "accessToken",
			apiError:           core.ErrNotFound,
			expectedReasonCode: events.ReasonTokenMissing,
		},
		{
			name:               "access token - bad request error",
			sessionStore:       "accessToken",
			apiError:           core.ErrBadRequest,
			expectedReasonCode: events.ReasonAuthTokenBad,
		},
		{
			name:               "access token - invalid renew token",
			sessionStore:       "accessToken",
			apiError:           session.ErrInvalidRenewToken,
			expectedReasonCode: events.ReasonTokenCorrupted,
		},
		{
			name:               "access token - session invalidated",
			sessionStore:       "accessToken",
			apiError:           session.ErrSessionInvalidated,
			expectedReasonCode: events.ReasonAuthTokenInvalidated,
		},
		{
			name:               "access token - missing access token response",
			sessionStore:       "accessToken",
			apiError:           session.ErrMissingAccessTokenResponse,
			expectedReasonCode: events.ReasonTokenMissing,
		},
		{
			name:               "access token - unauthorized error",
			sessionStore:       "accessToken",
			apiError:           core.ErrUnauthorized,
			expectedReasonCode: events.ReasonNotSpecified,
		},
		{
			name:               "vpn creds - missing vpn credentials",
			sessionStore:       "vpnCreds",
			apiError:           session.ErrMissingVPNCredentials,
			expectedReasonCode: events.ReasonCorruptedVPNCreds,
		},
		{
			name:               "vpn creds - missing nordlynx key",
			sessionStore:       "vpnCreds",
			apiError:           session.ErrMissingNordLynxPrivateKey,
			expectedReasonCode: events.ReasonCorruptedVPNCreds,
		},
		{
			name:               "vpn creds - missing vpn creds response",
			sessionStore:       "vpnCreds",
			apiError:           session.ErrMissingVPNCredsResponse,
			expectedReasonCode: events.ReasonCorruptedVPNCreds,
		},
		{
			name:               "vpn creds - bad request error",
			sessionStore:       "vpnCreds",
			apiError:           core.ErrBadRequest,
			expectedReasonCode: events.ReasonCorruptedVPNCredsAuthBad,
		},
		{
			name:               "vpn creds - unauthorized error",
			sessionStore:       "vpnCreds",
			apiError:           core.ErrUnauthorized,
			expectedReasonCode: events.ReasonCorruptedVPNCreds,
		},
		{
			name:               "trusted pass - bad request error",
			sessionStore:       "trustedPass",
			apiError:           core.ErrBadRequest,
			expectedReasonCode: events.ReasonNotSpecified,
		},
		{
			name:               "trusted pass - unauthorized error",
			sessionStore:       "trustedPass",
			apiError:           core.ErrUnauthorized,
			expectedReasonCode: events.ReasonNotSpecified,
		},
		{
			name:               "trusted pass - not found error",
			sessionStore:       "trustedPass",
			apiError:           core.ErrNotFound,
			expectedReasonCode: events.ReasonNotSpecified,
		},
		{
			name:               "nc creds - bad request error",
			sessionStore:       "ncCreds",
			apiError:           core.ErrBadRequest,
			expectedReasonCode: events.ReasonNotSpecified,
		},
		{
			name:               "nc creds - unauthorized error",
			sessionStore:       "ncCreds",
			apiError:           core.ErrUnauthorized,
			expectedReasonCode: events.ReasonNotSpecified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := setupTestConfig()
			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = cfg

			builder := NewSessionStoresBuilder(cfgManager)

			// Set up mock API with default successful token renewal
			mockRawAPI := &mockRawClientAPI{
				tokenRenewFunc: func(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
					return &core.TokenRenewResponse{
						Token:      "ab78bb36299d442fa0715fb53b5e3e58",
						RenewToken: "cd89cc47300e553fb1826fc64c6f4f69",
						ExpiresAt:  time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
					}, nil
				},
			}
			mockAPI := &mockClientAPI{mockRawClientAPI: mockRawAPI}

			builder.BuildAccessTokenStore(mockRawAPI)
			builder.BuildVPNCredsStore(mockAPI)
			builder.BuildTrustedPassStore(mockAPI)
			builder.BuildNCCredsStore(mockAPI)

			var capturedReasonCode events.ReasonCode

			mockNetworker := &mocknetworker.Mock{}
			mockConfigManager := mock.NewMockConfigManager()
			mockConfigManager.Cfg = setupTestConfig()

			logoutHandler := daemon.NewLogoutHandler(daemon.LogoutHandlerDependencies{
				PublishLogoutEventFunc: captureLogoutReason(&capturedReasonCode),
				Networker:              mockNetworker,
				ConfigManager:          mockConfigManager,
				NotificationClient:     &mockNotificationClient{},
				AuthChecker:            &mockauth.AuthCheckerMock{},
				PublishDisconnectFunc:  func(events.DataDisconnect) {},
				DebugPublisherFunc:     func(string) {},
			})

			switch tt.sessionStore {
			case "accessToken":
				builder.registerAccessTokenHandlers(logoutHandler)
			case "vpnCreds":
				builder.registerAccessTokenHandlers(logoutHandler)
				builder.registerVPNCredsHandlers(logoutHandler)
			case "trustedPass":
				builder.registerTrustedPassHandlers(logoutHandler)
			case "ncCreds":
				builder.registerNCCredsHandlers(logoutHandler)
			}

			var registry *internal.ErrorHandlingRegistry[error]
			switch tt.sessionStore {
			case "accessToken":
				registry = builder.registries.accessToken
			case "vpnCreds":
				registry = builder.registries.vpnCreds
			case "trustedPass":
				registry = builder.registries.trustedPass
			case "ncCreds":
				registry = builder.registries.ncCreds
			}

			handlers := registry.GetHandlers(tt.apiError)
			if len(handlers) > 0 {
				handlers[0](tt.apiError)
			}

			assert.Equal(t, tt.expectedReasonCode, capturedReasonCode,
				"Expected reason code %v but got %v", tt.expectedReasonCode, capturedReasonCode)
		})
	}
}

func TestVPNCredsReasonCodeWithAccessTokenRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                    string
		vpnCredsError           error
		accessTokenRenewalError error
		expectedReasonCode      events.ReasonCode
	}{
		{
			name:                    "vpn creds missing response with failed access token renewal",
			vpnCredsError:           session.ErrMissingVPNCredsResponse,
			accessTokenRenewalError: core.ErrBadRequest,
			expectedReasonCode:      events.ReasonCorruptedVPNCreds,
		},
		{
			name:                    "vpn creds missing credentials with failed access token renewal",
			vpnCredsError:           session.ErrMissingVPNCredentials,
			accessTokenRenewalError: core.ErrNotFound,
			expectedReasonCode:      events.ReasonCorruptedVPNCreds,
		},
		{
			name:                    "vpn creds missing nordlynx with failed access token renewal",
			vpnCredsError:           session.ErrMissingNordLynxPrivateKey,
			accessTokenRenewalError: core.ErrNotFound,
			expectedReasonCode:      events.ReasonCorruptedVPNCreds,
		},
		{
			name:                    "vpn creds missing response with successful access token",
			vpnCredsError:           session.ErrMissingVPNCredsResponse,
			accessTokenRenewalError: nil,
			expectedReasonCode:      events.ReasonCorruptedVPNCreds,
		},
		{
			name:                    "vpn creds missing response with other error",
			vpnCredsError:           session.ErrMissingVPNCredsResponse,
			accessTokenRenewalError: errors.New("other error"),
			expectedReasonCode:      events.ReasonCorruptedVPNCreds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := setupTestConfig()
			uid := cfg.AutoConnectData.ID
			data := cfg.TokensData[uid]
			data.TokenExpiry = time.Now().Add(-24 * time.Hour).Format(internal.ServerDateFormat)
			cfg.TokensData[uid] = data

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = cfg

			builder := NewSessionStoresBuilder(cfgManager)

			mockRawAPI := &mockRawClientAPI{
				tokenRenewFunc: func(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
					if tt.accessTokenRenewalError != nil {
						return nil, tt.accessTokenRenewalError
					}
					return &core.TokenRenewResponse{
						Token:      "new-token",
						RenewToken: "new-renew-token",
						ExpiresAt:  time.Now().Add(24 * time.Hour).Format(internal.ServerDateFormat),
					}, nil
				},
			}
			mockAPI := &mockClientAPI{mockRawClientAPI: mockRawAPI}

			builder.BuildAccessTokenStore(mockRawAPI)
			builder.BuildVPNCredsStore(mockAPI)

			var capturedReasonCode events.ReasonCode

			mockNetworker := &mocknetworker.Mock{}
			mockConfigManager := mock.NewMockConfigManager()
			mockConfigManager.Cfg = setupTestConfig()

			logoutHandler := daemon.NewLogoutHandler(daemon.LogoutHandlerDependencies{
				PublishLogoutEventFunc: captureLogoutReason(&capturedReasonCode),
				Networker:              mockNetworker,
				ConfigManager:          mockConfigManager,
				NotificationClient:     &mockNotificationClient{},
				AuthChecker:            &mockauth.AuthCheckerMock{},
				PublishDisconnectFunc:  func(events.DataDisconnect) {},
				DebugPublisherFunc:     func(string) {},
			})

			builder.registerAccessTokenHandlers(logoutHandler)
			builder.registerVPNCredsHandlers(logoutHandler)

			handlers := builder.registries.vpnCreds.GetHandlers(tt.vpnCredsError)
			if len(handlers) > 0 {
				handlers[0](tt.vpnCredsError)
			} else {
				if errors.Is(tt.vpnCredsError, session.ErrMissingVPNCredsResponse) ||
					errors.Is(tt.vpnCredsError, session.ErrMissingVPNCredentials) ||
					errors.Is(tt.vpnCredsError, session.ErrMissingNordLynxPrivateKey) {
					t.Errorf("No handler registered for VPN creds error: %v", tt.vpnCredsError)
				}
			}

			assert.Equal(t, tt.expectedReasonCode, capturedReasonCode,
				"Expected reason code %v but got %v", tt.expectedReasonCode, capturedReasonCode)
		})
	}
}

func captureLogoutReason(capturedReason *events.ReasonCode) func(events.DataAuthorization) {
	return func(data events.DataAuthorization) {
		*capturedReason = data.Reason
	}
}

type mockNotificationClient struct{}

func (m *mockNotificationClient) Start() error { return nil }
func (m *mockNotificationClient) Stop() error  { return nil }
func (m *mockNotificationClient) Revoke() bool { return true }

func TestProductionErrorHandlerIntegration(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := setupTestConfig()
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = cfg

	builder := NewSessionStoresBuilder(cfgManager)

	mockRawAPI := &mockRawClientAPI{}
	mockAPI := &mockClientAPI{mockRawClientAPI: mockRawAPI}

	builder.BuildAccessTokenStore(mockRawAPI)
	builder.BuildVPNCredsStore(mockAPI)
	builder.BuildTrustedPassStore(mockAPI)
	builder.BuildNCCredsStore(mockAPI)

	logoutHandler := daemon.NewLogoutHandler(daemon.LogoutHandlerDependencies{})

	builder.ConfigureErrorHandlers(logoutHandler)

	assert.NotNil(t, builder.registries.accessToken, "Access token registry should be initialized")
	assert.NotNil(t, builder.registries.vpnCreds, "VPN creds registry should be initialized")
	assert.NotNil(t, builder.registries.trustedPass, "Trusted pass registry should be initialized")
	assert.NotNil(t, builder.registries.ncCreds, "NC creds registry should be initialized")

	testCases := []struct {
		registry *internal.ErrorHandlingRegistry[error]
		error    error
		name     string
	}{
		{builder.registries.accessToken, core.ErrBadRequest, "access token bad request"},
		{builder.registries.accessToken, core.ErrNotFound, "access token not found"},
		{builder.registries.accessToken, core.ErrUnauthorized, "access token unauthorized"},
		{builder.registries.accessToken, session.ErrInvalidRenewToken, "access token invalid renew token"},
		{builder.registries.accessToken, session.ErrSessionInvalidated, "access token session invalidated"},
		{builder.registries.accessToken, session.ErrMissingAccessTokenResponse, "access token missing response"},
		{builder.registries.vpnCreds, core.ErrBadRequest, "vpn creds bad request"},
		{builder.registries.vpnCreds, core.ErrUnauthorized, "vpn creds unauthorized"},
		{builder.registries.vpnCreds, session.ErrMissingVPNCredsResponse, "vpn creds missing response"},
		{builder.registries.vpnCreds, session.ErrMissingVPNCredentials, "vpn creds missing credentials"},
		{builder.registries.vpnCreds, session.ErrMissingNordLynxPrivateKey, "vpn creds missing nordlynx key"},
		{builder.registries.trustedPass, core.ErrBadRequest, "trusted pass bad request"},
		{builder.registries.trustedPass, core.ErrUnauthorized, "trusted pass unauthorized"},
		{builder.registries.trustedPass, core.ErrNotFound, "trusted pass not found"},
		{builder.registries.ncCreds, core.ErrBadRequest, "nc creds bad request"},
		{builder.registries.ncCreds, core.ErrUnauthorized, "nc creds unauthorized"},
	}

	for _, tc := range testCases {
		handlers := tc.registry.GetHandlers(tc.error)
		assert.Greater(t, len(handlers), 0, "Should have handler registered for %s", tc.name)
	}
}
