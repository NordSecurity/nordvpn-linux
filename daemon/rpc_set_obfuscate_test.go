package daemon

import (
	"context"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	mockN "github.com/NordSecurity/nordvpn-linux/test/mock/networker"

	"github.com/stretchr/testify/assert"
)

type mockObfuscateConfigManager struct {
	c config.Config
}

func (*mockObfuscateConfigManager) SaveWith(f config.SaveFunc) error {
	return nil
}

func (m *mockObfuscateConfigManager) Load(c *config.Config) error {
	c.AutoConnect = m.c.AutoConnect
	c.AutoConnectData = m.c.AutoConnectData
	return nil
}

func (*mockObfuscateConfigManager) Reset(bool, bool) error {
	return nil
}

func TestSetObfuscate(t *testing.T) {
	mockConfigManager := mockObfuscateConfigManager{}

	mockPublisherSubscriber := events.MockPublisherSubscriber[bool]{}
	mockEvents := events.Events{Settings: &events.SettingsEvents{Obfuscate: &mockPublisherSubscriber}}

	r := RPC{
		cm:     &mockConfigManager,
		events: &mockEvents,
		netw:   &mockN.Mock{VpnActive: true, MeshActive: true},
	}

	successPayload := pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{strconv.FormatBool(true)},
	}
	nothingToDoPayload := pb.Payload{Type: internal.CodeNothingToDo}

	tests := []struct {
		testName           string
		current            bool
		requested          bool
		autoconnectEnabled bool
		autoconnectServer  string
		payload            *pb.Payload
		eventPublished     bool
	}{
		{
			testName:       "turned on while autoconnect is off",
			current:        false,
			requested:      true,
			payload:        &successPayload,
			eventPublished: true,
		},
		{
			testName:           "turned on while autoconnect targets a specific server",
			current:            false,
			requested:          true,
			autoconnectEnabled: true,
			autoconnectServer:  "lt15",
			payload:            &successPayload,
			eventPublished:     true,
		},
		{
			testName:       "turned off",
			current:        true,
			requested:      false,
			payload:        &successPayload,
			eventPublished: true,
		},
		{
			testName:       "nothing to do when already set",
			current:        true,
			requested:      true,
			payload:        &nothingToDoPayload,
			eventPublished: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConfigManager.c.AutoConnect = test.autoconnectEnabled
			mockConfigManager.c.AutoConnectData.ServerTag = test.autoconnectServer
			mockConfigManager.c.AutoConnectData.Obfuscate = test.current

			resp, err := r.SetObfuscate(context.Background(), &pb.SetGenericRequest{Enabled: test.requested})

			assert.NoError(t, err)
			assert.Equal(t, test.payload, resp)
			assert.Equal(t, test.eventPublished, mockPublisherSubscriber.EventPublished)
			mockPublisherSubscriber.EventPublished = false
		})
	}
}
