package nc

import mqtt "github.com/eclipse/paho.mqtt.golang"

type MockClientBuilder struct {
	Client                mqtt.Client
	ConnectionLostHandler mqtt.ConnectionLostHandler
}

func (m *MockClientBuilder) Build(opts *mqtt.ClientOptions) mqtt.Client {
	m.ConnectionLostHandler = opts.OnConnectionLost
	return m.Client
}
