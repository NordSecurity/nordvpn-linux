package nc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	cfgmock "github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/core"
	ncmock "github.com/NordSecurity/nordvpn-linux/test/mock/nc"
	"github.com/stretchr/testify/assert"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttp "github.com/eclipse/paho.mqtt.golang/packets"
)

type mockMqttClient struct {
	mqtt.Client
	// connecting indicates if client is in connecting or disconnecting state
	connecting   bool
	connectToken mockMqttToken
}

func (m *mockMqttClient) Connect() mqtt.Token {
	m.connecting = true
	return &m.connectToken
}

func (m *mockMqttClient) Unsubscribe(topics ...string) mqtt.Token {
	return &mockMqttToken{timesOut: true}
}

func (m *mockMqttClient) Disconnect(uint) { m.connecting = false }

type mockMqttToken struct {
	mqtt.Token
	timesOut bool
	err      error
}

func (m *mockMqttToken) WaitTimeout(time.Duration) bool {
	return !m.timesOut
}

func (m *mockMqttToken) Error() error {
	return m.err
}

func connectionStateToString(t *testing.T, state connectionState) string {
	t.Helper()

	switch state {
	case needsAuthorization:
		return "needsAuthorization"
	case connecting:
		return "connecting"
	case connectedSuccessfully:
		return "connectedSuccessfully"
	}

	return "unknown"
}

func TestStartStopNotificationClient(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := config.Config{}
	cfg.TokensData = make(map[int64]config.TokenData)
	cfgManager := cfgmock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	tests := []struct {
		name                    string
		initialState            connectionState
		expectedConnectionState connectionState
		credentialsFetchError   error
		connectionTimeout       bool
		connectionTokenErr      error
		expectedClientState     bool
	}{
		{
			name:                    "unauthorized client connects successfully",
			initialState:            needsAuthorization,
			expectedConnectionState: connectedSuccessfully,
			expectedClientState:     true,
			credentialsFetchError:   nil,
			connectionTimeout:       false,
			connectionTokenErr:      nil,
		},
		{
			name:                    "authorized client connects successfully",
			initialState:            connecting,
			expectedConnectionState: connectedSuccessfully,
			expectedClientState:     true,
			credentialsFetchError:   nil,
			connectionTimeout:       false,
			connectionTokenErr:      nil,
		},
		{
			name:         "unauthorized client times out when attempting to connect",
			initialState: needsAuthorization,
			// in case of timeout, MQTT client should be manually disconnected to clean the state
			expectedClientState:     false,
			expectedConnectionState: connecting,
			credentialsFetchError:   nil,
			connectionTimeout:       true,
			connectionTokenErr:      nil,
		},
		{
			name:                    "authorized client times out when attempting to connect",
			initialState:            connecting,
			expectedConnectionState: connecting,
			expectedClientState:     false,
			credentialsFetchError:   nil,
			connectionTimeout:       true,
			connectionTokenErr:      nil,
		},
		{
			name:                    "authorized client loses authorization",
			initialState:            connecting,
			expectedConnectionState: needsAuthorization,
			expectedClientState:     true,
			credentialsFetchError:   nil,
			connectionTimeout:       false,
			connectionTokenErr:      mqttp.ErrorRefusedNotAuthorised,
		},
		{
			name:                    "authorized client fails to connect",
			initialState:            needsAuthorization,
			expectedConnectionState: needsAuthorization,
			expectedClientState:     false,
			credentialsFetchError:   fmt.Errorf("failed to fetch credentials"),
			connectionTimeout:       false,
			connectionTokenErr:      nil,
		},
	}

	for _, test := range tests {
		connectionToken := mockMqttToken{
			timesOut: test.connectionTimeout,
			err:      test.connectionTokenErr,
		}
		mockMqttClient := mockMqttClient{
			connectToken: connectionToken,
		}
		clientBuilderMock := ncmock.MockClientBuilder{
			Client: &mockMqttClient,
		}

		credsFetcher := NewCredsFetcher(&core.CredentialsAPIMock{
			NotificationCredentialsError: test.credentialsFetchError,
		}, cfgManager)
		notificationClient := NewClient(&clientBuilderMock,
			&subs.Subject[string]{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			credsFetcher)

		mgmtChan := make(chan interface{})
		go func() {
			for {
				<-mgmtChan
			}
		}()

		t.Run(test.name, func(t *testing.T) {
			_, newConnectionState := notificationClient.tryConnect(&mockMqttClient,
				nil,
				test.initialState,
				mgmtChan,
				context.Background())

			assert.Equal(t,
				test.expectedConnectionState,
				newConnectionState,
				"Invalid connection status after trying to connect, expected '%s', got '%s'.",
				connectionStateToString(t, test.expectedConnectionState),
				connectionStateToString(t, newConnectionState))
			assert.Equal(t, test.expectedClientState, mockMqttClient.connecting,
				"MQTT client left in invalid state after calling tryConnect.")
		})
	}
}
