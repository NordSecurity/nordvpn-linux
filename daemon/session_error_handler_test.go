package daemon

import (
	"errors"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
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
			name:           "handles ErrAccessTokenRevoked",
			triggeredError: session.ErrAccessTokenRevoked,
			expectHandled:  true,
		},
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

				// Verify logout was attempted
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

	// Trigger renewal which should fail and call error handler
	err := store.Renew()

	// HandleError returns nil when handlers are found, so we expect no error
	assert.NoError(t, err)
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

	// The Renew method internally validates the token first, and since it's expired,
	// it will proceed with renewal which will fail with unauthorized
	err := store.Renew()
	assert.NoError(t, err) // HandleError returns nil when handlers are found

	// Verify the flow
	assert.Contains(t, flowSteps, "renewal_attempted", "Expected renewal to be attempted")
	assert.Contains(t, flowSteps, "logout_event", "Expected logout event to be published")
	assert.Contains(t, flowSteps, "debug_message", "Expected debug message to be logged")

	// Verify the order
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

	// Test with properly wrapped error using fmt.Errorf with %w
	wrappedErr := fmt.Errorf("wrapped: %w", core.ErrUnauthorized)
	handlers := registry.GetHandlers(wrappedErr)

	// Should handle wrapped errors that properly use errors.Is
	assert.NotEmpty(t, handlers, "Should handle properly wrapped errors")

	// Execute the handler to verify it's called
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

	// Test with incorrectly wrapped error (string concatenation instead of %w)
	incorrectlyWrappedErr := errors.New("wrapped: " + core.ErrUnauthorized.Error())
	handlers := registry.GetHandlers(incorrectlyWrappedErr)

	// Should NOT handle errors that don't properly wrap with %w
	assert.Empty(t, handlers, "Should not handle incorrectly wrapped errors")
	assert.False(t, handlerCalled, "Handler should not be called for incorrectly wrapped errors")
}
