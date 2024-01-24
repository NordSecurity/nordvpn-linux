package nc

import mqtt "github.com/eclipse/paho.mqtt.golang"

type MockClientBuilder struct {
	Client mqtt.Client
}

func (m *MockClientBuilder) Build(*mqtt.ClientOptions) mqtt.Client {
	return m.Client
}
