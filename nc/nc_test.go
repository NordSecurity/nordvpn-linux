package nc

import (
	"context"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	cfgmock "github.com/NordSecurity/nordvpn-linux/test/mock"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	ncmock "github.com/NordSecurity/nordvpn-linux/test/mock/nc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttp "github.com/eclipse/paho.mqtt.golang/packets"
)

type mockResolver struct {
	ips []netip.Addr
	err error
}

func (m *mockResolver) Resolve(domain string) ([]netip.Addr, error) {
	return m.ips, m.err
}

func newMockResolver(ips ...string) *mockResolver {
	addrs := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, netip.MustParseAddr(ip))
	}
	return &mockResolver{ips: addrs}
}

type mockMqttMessage struct {
	payload []byte
}

func (m *mockMqttMessage) Duplicate() bool   { return false }
func (m *mockMqttMessage) Qos() byte         { return 0 }
func (m *mockMqttMessage) Retained() bool    { return false }
func (m *mockMqttMessage) Topic() string     { return "" }
func (m *mockMqttMessage) MessageID() uint16 { return 0 }
func (m *mockMqttMessage) Payload() []byte   { return m.payload }
func (m *mockMqttMessage) Ack()              {}

type mockMqttClient struct {
	mqtt.Client
	// connecting indicates if client is in connecting or disconnecting state
	connecting     bool
	connectToken   mockMqttToken
	subscribeToken mockMqttToken
	clientID       string
}

func (m *mockMqttClient) Connect() mqtt.Token {
	m.connecting = true
	return &m.connectToken
}

func (m *mockMqttClient) Unsubscribe(topics ...string) mqtt.Token {
	return &mockMqttToken{timesOut: true}
}

func (m *mockMqttClient) Disconnect(uint) { m.connecting = false }

func (m *mockMqttClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return &m.subscribeToken
}

func (m *mockMqttClient) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.NewOptionsReader(&mqtt.ClientOptions{
		ClientID: m.clientID,
	})
}

func (m *mockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &mockMqttToken{}
}

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

func (m *mockMqttToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
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

	mgmtChan := make(chan interface{})
	go func() {
		for {
			<-mgmtChan
		}
	}()

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

		credsFetcher := NewCredsFetcher(&coremock.CredentialsAPIMock{
			NotificationCredentialsError: test.credentialsFetchError,
		}, cfgManager)
		notificationClient := NewClient(&clientBuilderMock,
			&subs.Subject[string]{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&subs.Subject[any]{},
			credsFetcher,
			0,
			newMockResolver("127.0.0.1"),
		)

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

func TestConnectionCancellation(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := config.Config{}
	cfg.TokensData = make(map[int64]config.TokenData)
	cfgManager := cfgmock.NewMockConfigManager()
	cfgManager.Cfg = &cfg

	tests := []struct {
		name                string
		connectionErr       error
		fetchCredentialsErr error
		tokenTimeout        time.Duration // how long client will wait for connection to be established
		delayBeforeCancel   time.Duration
	}{
		{
			name: "connection success",
		},
		{
			name:          "connection failure",
			connectionErr: fmt.Errorf("failed to connect"),
		},
		{
			name:          "connection auth failure",
			connectionErr: mqttp.ErrorRefusedNotAuthorised,
		},
		{
			name:                "fetch credentials failure",
			fetchCredentialsErr: fmt.Errorf("failed to fetch credentials"),
		},
		{
			name:         "cancel while waiting for connection",
			tokenTimeout: 10 * time.Second,
		},
		{
			name:              "delay before cancel",
			delayBeforeCancel: 10 * time.Millisecond,
		},
	}

	for _, test := range tests {
		connectionToken := mockMqttToken{
			timesOut: false,
			err:      test.connectionErr,
		}
		mockMqttClient := mockMqttClient{
			connectToken: connectionToken,
		}
		clientBuilderMock := ncmock.MockClientBuilder{
			Client: &mockMqttClient,
		}

		credsFetcher := NewCredsFetcher(&coremock.CredentialsAPIMock{
			NotificationCredentialsError: test.fetchCredentialsErr,
		}, cfgManager)

		notificationClient := Client{
			clientBuilder:     &clientBuilderMock,
			subjectInfo:       &subs.Subject[string]{},
			subjectErr:        &subs.Subject[error]{},
			subjectPeerUpdate: &subs.Subject[[]string]{},
			credsFetcher:      credsFetcher,
			retryDelayFunc:    func(i int) time.Duration { return test.tokenTimeout },
			resolver:          newMockResolver("127.0.0.1"),
		}

		t.Run(test.name, func(t *testing.T) {
			ctx, cancelFunc := context.WithCancel(context.Background())
			connectedChan := make(chan interface{})
			go func() {
				notificationClient.connect(&mockMqttClient, false, ctx, make(chan<- interface{}), make(chan<- mqtt.Client))
				connectedChan <- true
			}()

			time.Sleep(test.delayBeforeCancel)
			cancelFunc()

			select {
			case <-time.After(1 * time.Second):
				assert.FailNow(t, "Time out when waiting for connect to finish.")
			case <-connectedChan:
			}
		})
	}
}

func serverURLs(opts *mqtt.ClientOptions) []string {
	urls := make([]string, 0, len(opts.Servers))
	for _, u := range opts.Servers {
		urls = append(urls, u.String())
	}
	return urls
}

func TestCreateClientOptions(t *testing.T) {
	category.Set(t, category.Unit)

	creds := config.NCData{
		Endpoint: "ssl://mqtt.example.com:8883",
		Username: "user",
		Password: "pass",
	}

	tests := []struct {
		name           string
		resolver       *mockResolver
		fwmark         uint32
		expectedBroker []string
		expectTLSName  string
		expectDialer   bool
	}{
		{
			name:           "single IPv4 resolved",
			resolver:       newMockResolver("1.2.3.4"),
			expectedBroker: []string{"ssl://1.2.3.4:8883"},
			expectTLSName:  "mqtt.example.com",
		},
		{
			name:           "multiple IPv4 resolved",
			resolver:       newMockResolver("1.2.3.4", "5.6.7.8"),
			expectedBroker: []string{"ssl://1.2.3.4:8883", "ssl://5.6.7.8:8883"},
			expectTLSName:  "mqtt.example.com",
		},
		{
			name:           "IPv6 only falls back to original endpoint",
			resolver:       newMockResolver("::1", "fe80::1"),
			expectedBroker: []string{"ssl://mqtt.example.com:8883"},
			expectTLSName:  "mqtt.example.com",
		},
		{
			name:           "mixed IPv4 and IPv6 uses only IPv4",
			resolver:       newMockResolver("::1", "1.2.3.4", "fe80::1"),
			expectedBroker: []string{"ssl://1.2.3.4:8883"},
			expectTLSName:  "mqtt.example.com",
		},
		{
			name: "resolution error falls back to original endpoint",
			resolver: &mockResolver{
				err: fmt.Errorf("dns failure"),
			},
			expectedBroker: []string{"ssl://mqtt.example.com:8883"},
			expectTLSName:  "mqtt.example.com",
		},
		{
			name:           "fwmark sets custom dialer",
			resolver:       newMockResolver("1.2.3.4"),
			fwmark:         0xe1f1,
			expectedBroker: []string{"ssl://1.2.3.4:8883"},
			expectTLSName:  "mqtt.example.com",
			expectDialer:   true,
		},
		{
			name:           "zero fwmark uses default dialer",
			resolver:       newMockResolver("1.2.3.4"),
			fwmark:         0,
			expectedBroker: []string{"ssl://1.2.3.4:8883"},
			expectTLSName:  "mqtt.example.com",
			expectDialer:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				subjectInfo:       &subs.Subject[string]{},
				subjectErr:        &subs.Subject[error]{},
				subjectPeerUpdate: &subs.Subject[[]string]{},
				fwmark:            tt.fwmark,
				resolver:          tt.resolver,
			}

			mgmtChan := make(chan interface{}, 10)
			opts, err := client.createClientOptions(
				creds, mgmtChan, context.Background(),
			)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedBroker, serverURLs(opts))

			if tt.expectTLSName != "" {
				require.NotNil(t, opts.TLSConfig)
				assert.Equal(t, tt.expectTLSName, opts.TLSConfig.ServerName)
			}

			if tt.expectDialer {
				assert.NotNil(t, opts.Dialer.Control,
					"dialer should have Control function when fwmark is set")
			}
		})
	}
}

func TestCreateClientOptionsInvalidEndpoint(t *testing.T) {
	category.Set(t, category.Unit)

	client := &Client{
		subjectInfo:       &subs.Subject[string]{},
		subjectErr:        &subs.Subject[error]{},
		subjectPeerUpdate: &subs.Subject[[]string]{},
		resolver:          newMockResolver("1.2.3.4"),
	}

	creds := config.NCData{
		Endpoint: "://invalid",
	}

	mgmtChan := make(chan interface{}, 10)
	_, err := client.createClientOptions(
		creds, mgmtChan, context.Background(),
	)
	assert.Error(t, err)
}

func TestCreateClientOptionsEndpointWithoutPort(t *testing.T) {
	category.Set(t, category.Unit)

	client := &Client{
		subjectInfo:       &subs.Subject[string]{},
		subjectErr:        &subs.Subject[error]{},
		subjectPeerUpdate: &subs.Subject[[]string]{},
		resolver:          newMockResolver("10.0.0.1"),
	}

	creds := config.NCData{
		Endpoint: "ssl://mqtt.example.com",
	}

	mgmtChan := make(chan interface{}, 10)
	opts, err := client.createClientOptions(
		creds, mgmtChan, context.Background(),
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"ssl://10.0.0.1"}, serverURLs(opts))
}
