package nc

import (
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MockClientBuilder struct {
	mu                    sync.Mutex
	Client                mqtt.Client
	ConnectionLostHandler mqtt.ConnectionLostHandler
}

func (m *MockClientBuilder) CallConnectionLost(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConnectionLostHandler(m.Client, err)
}

func (m *MockClientBuilder) Build(opts *mqtt.ClientOptions) mqtt.Client {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConnectionLostHandler = opts.OnConnectionLost
	return m.Client
}
