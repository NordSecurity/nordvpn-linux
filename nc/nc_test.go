package nc

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	cfgmock "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	"github.com/NordSecurity/nordvpn-linux/test/mock/core"
	ncmock "github.com/NordSecurity/nordvpn-linux/test/mock/nc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

type mockMqttClient struct {
	mqtt.Client
	stopped        bool
	connectSuccess bool
}

type mockMqttToken struct {
	mqtt.Token
	success bool
	err     error
}

func (m *mockMqttToken) WaitTimeout(time.Duration) bool {
	return m.success
}

func (m *mockMqttToken) Error() error {
	return m.err
}

func (m *mockMqttClient) Connect() mqtt.Token {
	m.stopped = false
	return &mockMqttToken{success: m.connectSuccess}
}

func (m *mockMqttClient) Unsubscribe(topics ...string) mqtt.Token {
	return &mockMqttToken{success: true}
}

func (m *mockMqttClient) Disconnect(uint) { m.stopped = true }

func getConfigManagerMock(t *testing.T) cfgmock.ConfigManagerMock {
	t.Helper()

	cfg := config.Config{}
	cfg.TokensData = make(map[int64]config.TokenData)
	cfgManager := cfgmock.ConfigManagerMock{
		Cfg: cfg,
	}

	return cfgManager
}

func TestRaceConditions(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := getConfigManagerMock(t)
	credsFetcher := NewCredsFetcher(&core.CredentialsAPIMock{}, &cfgManager, &ncmock.MockTime{})
	mockClient := mockMqttClient{}
	mockClientBuilder := &ncmock.MockClientBuilder{Client: &mockClient}

	client := NewClient(mockClientBuilder,
		&subs.Subject[string]{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		credsFetcher)
	client.client = &mockClient

	ctx, cancel := context.WithCancel(context.Background())
	client.cancelConnecting = cancel
	waitgroup := sync.WaitGroup{}
	waitgroup.Add(2)
	go func() {
		client.connectWithBackoff(ctx, func(int) time.Duration { return time.Nanosecond })
		waitgroup.Done()
	}()
	go func() {
		// Introducing randomness to so that various scenarios of start/stop timing would be tested
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
		_ = client.Stop() // Not checking for error because stop might happen before connect
		waitgroup.Done()
	}()
	waitgroup.Wait()

	assert.True(t, mockClient.stopped)
}

func TestConnectionRetrying(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := getConfigManagerMock(t)
	credsFetcher := NewCredsFetcher(&core.CredentialsAPIMock{}, &cfgManager, &ncmock.MockTime{})
	mockClient := mockMqttClient{}
	mockClientBuilder := &ncmock.MockClientBuilder{Client: &mockClient}

	client := NewClient(
		mockClientBuilder,
		&subs.Subject[string]{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		credsFetcher)
	client.client = &mockClient

	client.connectWithBackoff(context.Background(), func(r int) time.Duration {
		if r == 3 {
			mockClient.connectSuccess = true
		}
		return time.Nanosecond
	})
	assert.False(t, mockClient.stopped)
}

func TestStopsTryingToConnectWhenStopped(t *testing.T) {
	category.Set(t, category.Unit)

	cfgManager := getConfigManagerMock(t)
	credsFetcher := NewCredsFetcher(&core.CredentialsAPIMock{}, &cfgManager, &ncmock.MockTime{})
	mockClient := mockMqttClient{}
	mockClientBuilder := &ncmock.MockClientBuilder{Client: &mockClient}

	client := NewClient(
		mockClientBuilder,
		&subs.Subject[string]{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		credsFetcher)
	client.client = &mockClient

	ctx, cancel := context.WithCancel(context.Background())
	client.cancelConnecting = cancel
	waitgroup := sync.WaitGroup{}
	waitgroup.Add(2)
	go func() {
		client.connectWithBackoff(ctx, func(int) time.Duration { return time.Second })
		waitgroup.Done()
	}()
	go func() {
		err := client.Stop()
		assert.NoError(t, err)
		waitgroup.Done()
	}()
	start := time.Now()
	waitgroup.Wait()

	assert.True(t, mockClient.stopped)
	assert.Less(t, time.Since(start), time.Second)
}
