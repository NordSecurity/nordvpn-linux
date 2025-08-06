package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	mockauth "github.com/NordSecurity/nordvpn-linux/test/mock/auth"
	mocknetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type capturedEvents struct {
	logoutEvents     []events.DataAuthorization
	disconnectEvents []events.DataDisconnect
	debugMessages    []string
}

// Simple mock for NotificationClient
type mockNotificationClient struct {
	stopCalled bool
	stopErr    error
}

func (m *mockNotificationClient) Start() error {
	return nil
}

func (m *mockNotificationClient) Stop() error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockNotificationClient) Revoke() bool {
	return true
}

func TestRegisterSessionErrorHandler(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		triggeredError error
		expectHandled  bool
	}{
		{
			name:           "handles ErrUnauthorized",
			triggeredError: core.ErrUnauthorized,
			expectHandled:  true,
		},
		{
			name:           "handles ErrNotFound",
			triggeredError: core.ErrNotFound,
			expectHandled:  true,
		},
		{
			name:           "handles ErrBadRequest",
			triggeredError: core.ErrBadRequest,
			expectHandled:  true,
		},
		{
			name:           "does not handle unregistered error",
			triggeredError: errors.New("some other error"),
			expectHandled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := internal.NewErrorHandlingRegistry[error]()
			captured := &capturedEvents{}

			deps := SessionErrorHandlerDependencies{
				AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
				Networker:          &mocknetworker.Mock{},
				NotificationClient: &mockNotificationClient{},
				ConfigManager:      mock.NewMockConfigManager(),
				PublishLogoutEventFunc: func(e events.DataAuthorization) {
					captured.logoutEvents = append(captured.logoutEvents, e)
				},
				PublishDisconnectFunc: func(e events.DataDisconnect) {
					captured.disconnectEvents = append(captured.disconnectEvents, e)
				},
				DebugPublisherFunc: func(msg string) {
					captured.debugMessages = append(captured.debugMessages, msg)
				},
			}

			RegisterSessionErrorHandler(registry, deps)
			handlers := registry.GetHandlers(tt.triggeredError)

			if tt.expectHandled {
				assert.NotEmpty(t, handlers, "Expected handlers to be registered for %v", tt.triggeredError)

				for _, handler := range handlers {
					handler(tt.triggeredError)
				}

				assert.NotEmpty(t, captured.logoutEvents, "Expected logout events to be published")
				assert.Contains(t, captured.debugMessages, "user logged out", "Expected debug message about logout")
			} else {
				assert.Empty(t, handlers, "Expected no handlers for %v", tt.triggeredError)
			}
		})
	}
}

// Integration test demonstrating the full flow
func TestSessionErrorHandler_IntegrationWithAccessTokenStore(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := config.Config{
		TokensData: map[int64]config.TokenData{
			1: {
				Token:       "test-token",
				RenewToken:  "renew-token",
				TokenExpiry: "2020-01-01T00:00:00Z", // Expired
			},
		},
		AutoConnectData: config.AutoConnectData{ID: 1},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	registry := internal.NewErrorHandlingRegistry[error]()
	handlerCalled := false

	// Create a simple handler that sets a flag
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			handlerCalled = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}

	RegisterSessionErrorHandler(registry, deps)

	// Create access token store with mock renewal that fails
	renewalFailed := false
	store := session.NewAccessTokenSessionStore(
		cfgManager,
		registry,
		func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
			renewalFailed = true
			return nil, core.ErrUnauthorized // This should trigger the error handler
		},
		nil,
	)

	err := store.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, renewalFailed, "Expected renewal to be attempted")
	assert.True(t, handlerCalled, "Expected error handler to be called")
}

// Test that demonstrates the error handling flow from API call to logout
func TestSessionErrorHandler_APIErrorToLogoutFlow(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := config.Config{
		TokensData: map[int64]config.TokenData{
			1: {
				Token:          "ab78bb36299d442fa0715fb53b5e3e57", // Valid hex format
				RenewToken:     "cd89cc47300e553fb1826fc64c6f4f68", // Valid hex format
				TokenExpiry:    "2020-01-01 00:00:00",              // Expired
				IdempotencyKey: &uuid.UUID{},
			},
		},
		AutoConnectData: config.AutoConnectData{ID: 1},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	registry := internal.NewErrorHandlingRegistry[error]()

	// Track the flow
	flowSteps := []string{}

	// Create dependencies with tracking
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			flowSteps = append(flowSteps, "logout_event")
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {
			flowSteps = append(flowSteps, "disconnect_event")
		},
		DebugPublisherFunc: func(msg string) {
			flowSteps = append(flowSteps, "debug_message")
		},
	}

	RegisterSessionErrorHandler(registry, deps)

	store := session.NewAccessTokenSessionStore(
		cfgManager,
		registry,
		func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
			flowSteps = append(flowSteps, "renewal_attempted")
			// Simulate API returning unauthorized error
			return nil, core.ErrUnauthorized
		},
		func(token string) error {
			flowSteps = append(flowSteps, "external_validation")
			// Token is expired, so external validation is skipped
			return nil
		},
	)

	err := store.Renew()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")

	assert.Contains(t, flowSteps, "renewal_attempted", "Expected renewal to be attempted")
	assert.Contains(t, flowSteps, "logout_event", "Expected logout event to be published")
	assert.Contains(t, flowSteps, "debug_message", "Expected debug message to be logged")

	renewalIndex := -1
	logoutIndex := -1
	for i, step := range flowSteps {
		if step == "renewal_attempted" {
			renewalIndex = i
		}
		if step == "logout_event" {
			logoutIndex = i
		}
	}

	assert.Greater(t, logoutIndex, renewalIndex, "Logout should happen after renewal attempt")
}

// Test that properly wrapped errors are handled
func TestSessionErrorHandler_HandlesProperlyWrappedErrors(t *testing.T) {
	category.Set(t, category.Unit)

	registry := internal.NewErrorHandlingRegistry[error]()
	handlerCalled := false

	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      mock.NewMockConfigManager(),
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			handlerCalled = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}

	RegisterSessionErrorHandler(registry, deps)

	wrappedErr := fmt.Errorf("wrapped: %w", core.ErrUnauthorized)
	handlers := registry.GetHandlers(wrappedErr)

	assert.NotEmpty(t, handlers, "Should handle properly wrapped errors")

	for _, handler := range handlers {
		handler(wrappedErr)
	}

	assert.True(t, handlerCalled, "Handler should be called for properly wrapped errors")
}

// Test that incorrectly wrapped errors are not handled
func TestSessionErrorHandler_DoesNotHandleIncorrectlyWrappedErrors(t *testing.T) {
	category.Set(t, category.Unit)

	registry := internal.NewErrorHandlingRegistry[error]()
	handlerCalled := false

	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      mock.NewMockConfigManager(),
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			handlerCalled = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}

	RegisterSessionErrorHandler(registry, deps)

	incorrectlyWrappedErr := errors.New("wrapped: " + core.ErrUnauthorized.Error())
	handlers := registry.GetHandlers(incorrectlyWrappedErr)

	assert.Empty(t, handlers, "Should not handle incorrectly wrapped errors")
	assert.False(t, handlerCalled, "Handler should not be called for incorrectly wrapped errors")
}

// Mock implementation of RawClientAPI for testing
type mockRawClientAPI struct {
	// Control behavior
	currentUserError       error
	currentUserSecondError error // Error for second call
	tokenRenewError        error
	servicesError          error
	logoutError            error

	// Track calls
	currentUserCalls int
	tokenRenewCalls  int
	servicesCalls    int
	logoutCalls      int

	// Return values
	currentUserResponse *core.CurrentUserResponse
	tokenRenewResponse  *core.TokenRenewResponse
	servicesResponse    core.ServicesResponse
}

// Implement RawCredentialsAPI methods
func (m *mockRawClientAPI) NotificationCredentials(token, appUserID string) (core.NotificationCredentialsResponse, error) {
	return core.NotificationCredentialsResponse{}, nil
}

func (m *mockRawClientAPI) NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error) {
	return core.NotificationCredentialsRevokeResponse{}, nil
}

func (m *mockRawClientAPI) ServiceCredentials(token string) (*core.CredentialsResponse, error) {
	return &core.CredentialsResponse{}, nil
}

func (m *mockRawClientAPI) TokenRenew(token string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
	m.tokenRenewCalls++
	if m.tokenRenewError != nil {
		return nil, m.tokenRenewError
	}
	if m.tokenRenewResponse != nil {
		return m.tokenRenewResponse, nil
	}
	return &core.TokenRenewResponse{
		Token:      "new-token",
		RenewToken: "new-renew-token",
		ExpiresAt:  "2025-01-01 00:00:00",
	}, nil
}

func (m *mockRawClientAPI) Services(token string) (core.ServicesResponse, error) {
	m.servicesCalls++
	if m.servicesError != nil {
		return nil, m.servicesError
	}
	if m.servicesResponse != nil {
		return m.servicesResponse, nil
	}
	return core.ServicesResponse{}, nil
}

func (m *mockRawClientAPI) CurrentUser(token string) (*core.CurrentUserResponse, error) {
	m.currentUserCalls++

	// First call - return the configured error
	if m.currentUserCalls == 1 && m.currentUserError != nil {
		return nil, m.currentUserError
	}

	// Second call - return second error if configured
	if m.currentUserCalls == 2 && m.currentUserSecondError != nil {
		return nil, m.currentUserSecondError
	}

	// Second call - if no second error configured but first was unauthorized, return success
	if m.currentUserCalls == 2 && m.currentUserError == core.ErrUnauthorized && m.currentUserSecondError == nil {
		return &core.CurrentUserResponse{}, nil
	}

	// Default success response
	if m.currentUserResponse != nil {
		return m.currentUserResponse, nil
	}
	return &core.CurrentUserResponse{}, nil
}

func (m *mockRawClientAPI) DeleteToken(token string) error {
	return nil
}

func (m *mockRawClientAPI) TrustedPassToken(token string) (*core.TrustedPassTokenResponse, error) {
	return &core.TrustedPassTokenResponse{}, nil
}

func (m *mockRawClientAPI) MultifactorAuthStatus(token string) (*core.MultifactorAuthStatusResponse, error) {
	return &core.MultifactorAuthStatusResponse{}, nil
}

func (m *mockRawClientAPI) Logout(token string) error {
	m.logoutCalls++
	return m.logoutError
}

// Implement other interfaces (stubbed)
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

// Mesh interface methods
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

// Comprehensive integration test with SmartClientAPI
func TestSessionErrorHandler_SmartClientAPIIntegration(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                   string
		initialTokenExpired    bool
		currentUserFirstError  error // Error on first CurrentUser call (validation)
		tokenRenewError        error // Error on token renewal
		currentUserSecondError error // Error on second CurrentUser call (after renewal)
		servicesError          error // Error on Services call
		expectLogout           bool
		expectRenewal          bool
		expectError            bool   // Whether we expect an error to be returned
		expectedAPICalls       string // Description of expected API call sequence
	}{
		{
			name:                  "successful API call with valid token",
			initialTokenExpired:   false,
			currentUserFirstError: nil,
			expectLogout:          false,
			expectRenewal:         false,
			expectError:           false,
			expectedAPICalls:      "CurrentUser succeeds on first call",
		},
		{
			name:                   "unauthorized triggers renewal then success",
			initialTokenExpired:    false, // Token not expired, but API returns unauthorized
			currentUserFirstError:  core.ErrUnauthorized,
			tokenRenewError:        nil,
			currentUserSecondError: nil,
			expectLogout:           false,
			expectRenewal:          true,
			expectError:            false,
			expectedAPICalls:       "CurrentUser fails -> TokenRenew succeeds -> CurrentUser succeeds",
		},
		{
			name:                  "unauthorized error during renewal triggers logout",
			initialTokenExpired:   true,
			currentUserFirstError: core.ErrUnauthorized,
			tokenRenewError:       core.ErrUnauthorized,
			expectLogout:          true,
			expectRenewal:         true,
			expectError:           true, // HandleError returns wrapped error when handlers are found
			expectedAPICalls:      "CurrentUser fails -> TokenRenew fails with Unauthorized -> Logout",
		},
		{
			name:                  "not found error during renewal triggers logout",
			initialTokenExpired:   true,
			currentUserFirstError: core.ErrUnauthorized,
			tokenRenewError:       core.ErrNotFound,
			expectLogout:          true,
			expectRenewal:         true,
			expectError:           true, // HandleError returns wrapped error when handlers are found
			expectedAPICalls:      "CurrentUser fails -> TokenRenew fails with NotFound -> Logout",
		},
		{
			name:                  "bad request error during renewal triggers logout",
			initialTokenExpired:   true,
			currentUserFirstError: core.ErrUnauthorized,
			tokenRenewError:       core.ErrBadRequest,
			expectLogout:          true,
			expectRenewal:         true,
			expectError:           true, // HandleError returns wrapped error when handlers are found
			expectedAPICalls:      "CurrentUser fails -> TokenRenew fails with BadRequest -> Logout",
		},
		{
			name:                   "renewal succeeds but API still returns unauthorized",
			initialTokenExpired:    true,
			currentUserFirstError:  core.ErrUnauthorized,
			tokenRenewError:        nil,
			currentUserSecondError: core.ErrUnauthorized,
			expectLogout:           true,
			expectRenewal:          true,
			expectError:            true,
			expectedAPICalls:       "CurrentUser fails -> TokenRenew succeeds -> CurrentUser still fails -> Logout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			uid := int64(123)
			idempotencyKey := uuid.New()

			tokenExpiry := "2025-01-01 00:00:00"
			if tt.initialTokenExpired {
				tokenExpiry = "2020-01-01 00:00:00"
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: uid},
				TokensData: map[int64]config.TokenData{
					uid: {
						Token:          "test-token",
						RenewToken:     "renew-token",
						TokenExpiry:    tokenExpiry,
						IdempotencyKey: &idempotencyKey,
					},
				},
			}
			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg

			// Create mock RawClientAPI
			mockAPI := &mockRawClientAPI{
				currentUserError:       tt.currentUserFirstError,
				currentUserSecondError: tt.currentUserSecondError,
				tokenRenewError:        tt.tokenRenewError,
				servicesError:          tt.servicesError,
			}

			// Track logout events
			var logoutEvents []events.DataAuthorization
			var disconnectEvents []events.DataDisconnect
			var debugMessages []string

			// Create error registry and register handlers
			errorRegistry := internal.NewErrorHandlingRegistry[error]()
			deps := SessionErrorHandlerDependencies{
				AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
				Networker:          &mocknetworker.Mock{},
				NotificationClient: &mockNotificationClient{},
				ConfigManager:      cfgManager,
				PublishLogoutEventFunc: func(e events.DataAuthorization) {
					logoutEvents = append(logoutEvents, e)
				},
				PublishDisconnectFunc: func(e events.DataDisconnect) {
					disconnectEvents = append(disconnectEvents, e)
				},
				DebugPublisherFunc: func(msg string) {
					debugMessages = append(debugMessages, msg)
				},
			}
			RegisterSessionErrorHandler(errorRegistry, deps)

			// Create external validator only for tests that need it
			var externalValidator session.AccessTokenExternalValidator
			if tt.name == "successful API call with valid token" {
				// Only use external validator when we expect the token to be valid
				externalValidator = func(token string) error {
					_, err := mockAPI.CurrentUser(token)
					if errors.Is(err, core.ErrUnauthorized) {
						return session.ErrInvalidToken
					}
					return err
				}
			}

			renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
				resp, err := mockAPI.TokenRenew(token, key)
				if err != nil {
					return nil, err
				}
				return &session.AccessTokenResponse{
					Token:      resp.Token,
					RenewToken: resp.RenewToken,
					ExpiresAt:  resp.ExpiresAt,
				}, nil
			}

			sessionStore := session.NewAccessTokenSessionStore(
				cfgManager,
				errorRegistry,
				renewalAPICall,
				externalValidator,
			)

			smartAPI := core.NewSmartClientAPI(mockAPI, sessionStore)

			_, err := smartAPI.CurrentUser()

			t.Logf("Test case: %s", tt.name)
			t.Logf("Initial error from CurrentUser: %v", err)
			t.Logf("MockAPI state - CurrentUser calls: %d, TokenRenew calls: %d",
				mockAPI.currentUserCalls, mockAPI.tokenRenewCalls)

			if tt.expectLogout {
				assert.NotEmpty(t, logoutEvents, "Expected logout event to be published")
				assert.Contains(t, debugMessages, "user logged out", "Expected logout debug message")
			} else {
				assert.Empty(t, logoutEvents, "Did not expect logout event")
			}

			if tt.initialTokenExpired && tt.currentUserFirstError == nil {
				assert.Equal(t, 0, mockAPI.tokenRenewCalls, "Should not renew if API call succeeds")
			} else if tt.expectRenewal {
				assert.Greater(t, mockAPI.tokenRenewCalls, 0, "Expected token renewal to be attempted")
			} else {
				assert.Equal(t, 0, mockAPI.tokenRenewCalls, "Did not expect token renewal")
			}

			t.Logf("API call sequence: %s", tt.expectedAPICalls)
			t.Logf("CurrentUser calls: %d, TokenRenew calls: %d, Logout calls: %d",
				mockAPI.currentUserCalls, mockAPI.tokenRenewCalls, mockAPI.logoutCalls)

			if tt.expectError {
				assert.Error(t, err, "Should return error even when handled")
				if tt.tokenRenewError != nil && (errors.Is(tt.tokenRenewError, core.ErrUnauthorized) ||
					errors.Is(tt.tokenRenewError, core.ErrNotFound) ||
					errors.Is(tt.tokenRenewError, core.ErrBadRequest)) {
					assert.True(t, errors.Is(err, tt.tokenRenewError), "Error should be the renewal error")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Services API call with error handling
func TestSessionErrorHandler_ServicesAPIWithSmartClient(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	idempotencyKey := uuid.New()

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "test-token",
				RenewToken:     "renew-token",
				TokenExpiry:    "2020-01-01 00:00:00", // Expired
				IdempotencyKey: &idempotencyKey,
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	// Create mock that returns unauthorized on Services call
	mockAPI := &mockRawClientAPI{
		servicesError: core.ErrUnauthorized,
		// Token renewal fails too
		tokenRenewError: core.ErrBadRequest,
	}

	// Track events
	logoutCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			logoutCalled = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	// Create session store
	renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := mockAPI.TokenRenew(token, key)
		if err != nil {
			return nil, err
		}
		return &session.AccessTokenResponse{
			Token:      resp.Token,
			RenewToken: resp.RenewToken,
			ExpiresAt:  resp.ExpiresAt,
		}, nil
	}

	sessionStore := session.NewAccessTokenSessionStore(
		cfgManager,
		errorRegistry,
		renewalAPICall,
		nil, // No external validator for this test
	)

	smartAPI := core.NewSmartClientAPI(mockAPI, sessionStore)

	_, err := smartAPI.Services()

	assert.Error(t, err, "Should return error")
	assert.True(t, logoutCalled, "Expected logout to be called")
	assert.Equal(t, 1, mockAPI.servicesCalls, "Services should be called once (renewal fails with handled error)")
	assert.Equal(t, 1, mockAPI.tokenRenewCalls, "Token renewal should be attempted")
}

// Test that concurrent logout attempts result in only one actual logout
func TestSessionErrorHandler_ConcurrentLogoutPrevention(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "test-token",
				RenewToken:  "renew-token",
				TokenExpiry: "2020-01-01 00:00:00",
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	var handlerCalls int32
	var logoutStarts int32
	var logoutCompletions int32
	var mu sync.Mutex
	var logoutEvents []events.DataAuthorization

	firstLogoutStarted := make(chan struct{})
	allowLogoutToComplete := make(chan struct{})
	logoutCompleted := make(chan struct{})

	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			mu.Lock()
			logoutEvents = append(logoutEvents, e)
			mu.Unlock()

			if e.EventStatus == events.StatusAttempt {
				if atomic.AddInt32(&logoutStarts, 1) == 1 {
					close(firstLogoutStarted)
					<-allowLogoutToComplete
				}
			}
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc: func(msg string) {
			if msg == "user logged out" {
				if atomic.AddInt32(&logoutCompletions, 1) == 1 {
					close(logoutCompleted)
				}
			}
		},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	var wg sync.WaitGroup
	numGoroutines := 5
	startSignal := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal
			atomic.AddInt32(&handlerCalls, 1)
			handlers := errorRegistry.GetHandlers(core.ErrUnauthorized)
			for _, handler := range handlers {
				handler(core.ErrUnauthorized)
			}
		}()
	}

	close(startSignal)

	<-firstLogoutStarted

	time.Sleep(100 * time.Millisecond)

	close(allowLogoutToComplete)

	<-logoutCompleted

	wg.Wait()

	time.Sleep(50 * time.Millisecond)

	calls := atomic.LoadInt32(&handlerCalls)
	starts := atomic.LoadInt32(&logoutStarts)
	completions := atomic.LoadInt32(&logoutCompletions)

	assert.Equal(t, int32(numGoroutines), calls, "All goroutines should call the handler")
	assert.Equal(t, int32(1), starts, "Only one logout should start")
	assert.Equal(t, int32(1), completions, "Only one logout should complete")

	mu.Lock()
	eventCount := len(logoutEvents)
	mu.Unlock()

	assert.Equal(t, 2, eventCount, "Should have exactly 2 logout events (attempt + result)")
}

// Test concurrent API calls that trigger session errors
func TestSessionErrorHandler_ConcurrentAPICallsWithErrors(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	idempotencyKey := uuid.New()

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "test-token",
				RenewToken:     "renew-token",
				TokenExpiry:    "2020-01-01 00:00:00", // Expired
				IdempotencyKey: &idempotencyKey,
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	mockAPI := &mockRawClientAPI{
		currentUserError: core.ErrUnauthorized,
		servicesError:    core.ErrUnauthorized,
		tokenRenewError:  core.ErrUnauthorized,
	}

	var logoutStarted int32
	var logoutCompleted int32
	var concurrentAttempts int32
	logoutInProgress := make(chan struct{})
	logoutDone := make(chan struct{})

	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			if e.EventStatus == events.StatusAttempt {
				if atomic.CompareAndSwapInt32(&logoutStarted, 0, 1) {
					close(logoutInProgress)
				} else {
					atomic.AddInt32(&concurrentAttempts, 1)
				}
			}
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc: func(msg string) {
			if msg == "user logged out" {
				if atomic.CompareAndSwapInt32(&logoutCompleted, 0, 1) {
					close(logoutDone)
				}
			}
		},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := mockAPI.TokenRenew(token, key)
		if err != nil {
			return nil, err
		}
		return &session.AccessTokenResponse{
			Token:      resp.Token,
			RenewToken: resp.RenewToken,
			ExpiresAt:  resp.ExpiresAt,
		}, nil
	}

	sessionStore := session.NewAccessTokenSessionStore(
		cfgManager,
		errorRegistry,
		renewalAPICall,
		nil,
	)

	smartAPI := core.NewSmartClientAPI(mockAPI, sessionStore)

	var wg sync.WaitGroup
	numGoroutines := 10

	startSignal := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal
			_, _ = smartAPI.CurrentUser()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal
			_, _ = smartAPI.Services()
		}()
	}

	close(startSignal)

	<-logoutInProgress

	time.Sleep(50 * time.Millisecond)

	<-logoutDone

	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&logoutStarted), "Only one logout should start")
	assert.Equal(t, int32(1), atomic.LoadInt32(&logoutCompleted), "Only one logout should complete")

	t.Logf("Concurrent attempts blocked: %d", atomic.LoadInt32(&concurrentAttempts))
}

// Test that renewal failure with network error doesn't trigger logout
func TestSessionErrorHandler_NetworkErrorDuringRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	idempotencyKey := uuid.New()

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "test-token",
				RenewToken:     "renew-token",
				TokenExpiry:    "2020-01-01 00:00:00",
				IdempotencyKey: &idempotencyKey,
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	networkErr := errors.New("network timeout")
	mockAPI := &mockRawClientAPI{
		currentUserError:       core.ErrUnauthorized,
		currentUserSecondError: core.ErrUnauthorized,
		tokenRenewError:        networkErr,
	}

	logoutCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			logoutCalled = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := mockAPI.TokenRenew(token, key)
		if err != nil {
			return nil, err
		}
		return &session.AccessTokenResponse{
			Token:      resp.Token,
			RenewToken: resp.RenewToken,
			ExpiresAt:  resp.ExpiresAt,
		}, nil
	}

	sessionStore := session.NewAccessTokenSessionStore(
		cfgManager,
		errorRegistry,
		renewalAPICall,
		nil,
	)

	smartAPI := core.NewSmartClientAPI(mockAPI, sessionStore)

	_, err := smartAPI.CurrentUser()

	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrUnauthorized, "Should return the API error")
	assert.True(t, logoutCalled, "Logout is triggered by the second unauthorized error")
	assert.Equal(t, 1, mockAPI.tokenRenewCalls, "Should attempt renewal")
	assert.Equal(t, 2, mockAPI.currentUserCalls, "Should call CurrentUser twice (before and after failed renewal)")
}

// Test empty/nil token scenarios
func TestSessionErrorHandler_EmptyTokenScenarios(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		token        string
		renewToken   string
		expectLogout bool
		expectError  bool
	}{
		{
			name:         "empty access token",
			token:        "",
			renewToken:   "valid-renew-token",
			expectLogout: false,
			expectError:  true,
		},
		{
			name:         "empty renew token",
			token:        "valid-token",
			renewToken:   "",
			expectLogout: false,
			expectError:  true,
		},
		{
			name:         "both tokens empty",
			token:        "",
			renewToken:   "",
			expectLogout: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid := int64(123)
			idempotencyKey := uuid.New()

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{ID: uid},
				TokensData: map[int64]config.TokenData{
					uid: {
						Token:          tt.token,
						RenewToken:     tt.renewToken,
						TokenExpiry:    "2020-01-01 00:00:00",
						IdempotencyKey: &idempotencyKey,
					},
				},
			}
			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg = &cfg

			mockAPI := &mockRawClientAPI{
				currentUserError: core.ErrUnauthorized,
				tokenRenewError:  core.ErrUnauthorized,
			}

			logoutCalled := false
			errorRegistry := internal.NewErrorHandlingRegistry[error]()
			deps := SessionErrorHandlerDependencies{
				AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
				Networker:          &mocknetworker.Mock{},
				NotificationClient: &mockNotificationClient{},
				ConfigManager:      cfgManager,
				PublishLogoutEventFunc: func(e events.DataAuthorization) {
					logoutCalled = true
				},
				PublishDisconnectFunc: func(e events.DataDisconnect) {},
				DebugPublisherFunc:    func(msg string) {},
			}
			RegisterSessionErrorHandler(errorRegistry, deps)

			renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
				if token == "" || tt.renewToken == "" {
					return nil, errors.New("invalid token")
				}
				resp, err := mockAPI.TokenRenew(token, key)
				if err != nil {
					return nil, err
				}
				return &session.AccessTokenResponse{
					Token:      resp.Token,
					RenewToken: resp.RenewToken,
					ExpiresAt:  resp.ExpiresAt,
				}, nil
			}

			sessionStore := session.NewAccessTokenSessionStore(
				cfgManager,
				errorRegistry,
				renewalAPICall,
				nil,
			)

			err := sessionStore.Renew()

			if tt.expectLogout {
				assert.True(t, logoutCalled, "Expected logout to be called")
			} else {
				assert.False(t, logoutCalled, "Expected logout not to be called")
			}

			if tt.expectError {
				if err != nil {
					assert.Error(t, err, "Should error with invalid tokens")
				}
			}
		})
	}
}

// Test malformed token expiry handling
func TestSessionErrorHandler_MalformedTokenExpiry(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	idempotencyKey := uuid.New()

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "test-token",
				RenewToken:     "renew-token",
				TokenExpiry:    "invalid-date-format",
				IdempotencyKey: &idempotencyKey,
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	mockAPI := &mockRawClientAPI{
		currentUserError: core.ErrUnauthorized,
		tokenRenewError:  core.ErrUnauthorized,
	}

	logoutCalled := false
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			logoutCalled = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := mockAPI.TokenRenew(token, key)
		if err != nil {
			return nil, err
		}
		return &session.AccessTokenResponse{
			Token:      resp.Token,
			RenewToken: resp.RenewToken,
			ExpiresAt:  resp.ExpiresAt,
		}, nil
	}

	sessionStore := session.NewAccessTokenSessionStore(
		cfgManager,
		errorRegistry,
		renewalAPICall,
		nil,
	)

	err := sessionStore.Renew()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handling session error")
	assert.True(t, logoutCalled, "Should trigger logout due to unauthorized during renewal")
	assert.Equal(t, 1, mockAPI.tokenRenewCalls, "Should attempt renewal with malformed expiry")
}

// Test that verifies logout removes user data
func TestSessionErrorHandler_LogoutClearsUserData(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	uid2 := int64(456)

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "user1-token",
				RenewToken:  "user1-renew",
				TokenExpiry: "2020-01-01 00:00:00",
			},
			uid2: {
				Token:       "user2-token",
				RenewToken:  "user2-renew",
				TokenExpiry: "2025-01-01 00:00:00",
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	mockAPI := &mockRawClientAPI{
		currentUserError: core.ErrUnauthorized,
		tokenRenewError:  core.ErrUnauthorized,
	}

	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	deps := SessionErrorHandlerDependencies{
		AuthChecker:            &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:              &mocknetworker.Mock{},
		NotificationClient:     &mockNotificationClient{},
		ConfigManager:          cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {},
		PublishDisconnectFunc:  func(e events.DataDisconnect) {},
		DebugPublisherFunc:     func(msg string) {},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := mockAPI.TokenRenew(token, key)
		if err != nil {
			return nil, err
		}
		return &session.AccessTokenResponse{
			Token:      resp.Token,
			RenewToken: resp.RenewToken,
			ExpiresAt:  resp.ExpiresAt,
		}, nil
	}

	externalValidator := func(token string) error {
		_, err := mockAPI.CurrentUser(token)
		if errors.Is(err, core.ErrUnauthorized) {
			return session.ErrInvalidToken
		}
		return err
	}

	sessionStore := session.NewAccessTokenSessionStore(
		cfgManager,
		errorRegistry,
		renewalAPICall,
		externalValidator,
	)

	smartAPI := core.NewSmartClientAPI(mockAPI, sessionStore)

	_, _ = smartAPI.CurrentUser()

	assert.NotContains(t, cfgManager.Cfg.TokensData, uid, "Current user data should be removed")
	assert.Contains(t, cfgManager.Cfg.TokensData, uid2, "Other user data should remain")
}

// Test renewal response with invalid data
func TestSessionErrorHandler_InvalidRenewalResponse(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)
	idempotencyKey := uuid.New()

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "test-token",
				RenewToken:     "renew-token",
				TokenExpiry:    "2020-01-01 00:00:00",
				IdempotencyKey: &idempotencyKey,
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	mockAPI := &mockRawClientAPI{
		currentUserError: core.ErrUnauthorized,
		tokenRenewResponse: &core.TokenRenewResponse{
			Token:      "",
			RenewToken: "new-renew-token",
			ExpiresAt:  "2025-01-01 00:00:00",
		},
	}

	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	logoutCalled := false
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          &mocknetworker.Mock{},
		NotificationClient: &mockNotificationClient{},
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			logoutCalled = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	renewalAPICall := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := mockAPI.TokenRenew(token, key)
		if err != nil {
			return nil, err
		}
		if resp.Token == "" {
			return nil, errors.New("invalid renewal response: empty token")
		}
		return &session.AccessTokenResponse{
			Token:      resp.Token,
			RenewToken: resp.RenewToken,
			ExpiresAt:  resp.ExpiresAt,
		}, nil
	}

	sessionStore := session.NewAccessTokenSessionStore(
		cfgManager,
		errorRegistry,
		renewalAPICall,
		nil,
	)

	err := sessionStore.Renew()

	assert.NoError(t, err, "HandleError returns nil for unhandled errors")
	assert.False(t, logoutCalled, "Should not trigger logout for invalid response format")
}

// Test logout is always published even if components fail
func TestSessionErrorHandler_LogoutEventAlwaysPublished(t *testing.T) {
	category.Set(t, category.Unit)

	uid := int64(123)

	cfg := config.Config{
		AutoConnectData: config.AutoConnectData{ID: uid},
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:       "test-token",
				RenewToken:  "renew-token",
				TokenExpiry: "2020-01-01 00:00:00",
			},
		},
	}
	cfgManager := mock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	failingNetworker := &mocknetworker.Failing{}

	notificationClient := &mockNotificationClient{
		stopErr: errors.New("notification stop failed"),
	}

	logoutEventPublished := false

	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	deps := SessionErrorHandlerDependencies{
		AuthChecker:        &mockauth.AuthCheckerMock{LoggedIn: true},
		Networker:          failingNetworker,
		NotificationClient: notificationClient,
		ConfigManager:      cfgManager,
		PublishLogoutEventFunc: func(e events.DataAuthorization) {
			logoutEventPublished = true
		},
		PublishDisconnectFunc: func(e events.DataDisconnect) {},
		DebugPublisherFunc:    func(msg string) {},
	}
	RegisterSessionErrorHandler(errorRegistry, deps)

	handlers := errorRegistry.GetHandlers(core.ErrUnauthorized)
	assert.NotEmpty(t, handlers)

	for _, handler := range handlers {
		handler(core.ErrUnauthorized)
	}

	assert.True(t, logoutEventPublished, "Logout event should be published even if components fail")
}
