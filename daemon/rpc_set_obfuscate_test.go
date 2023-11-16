package daemon

import (
	"context"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
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

func (*mockObfuscateConfigManager) Reset() error {
	return nil
}

func TestSetObfuscate(t *testing.T) {
	mockConfigManager := mockObfuscateConfigManager{c: config.Config{AutoConnect: false}}

	mockPublisherSubscriber := mockPublisherSubcriber{}
	mockEvents := Events{Settings: &SettingsEvents{Obfuscate: &mockPublisherSubscriber}}

	obfuscatedTechnologies := core.Technologies{
		core.Technology{
			ID:    core.OpenVPNTCPObfuscated,
			Pivot: core.Pivot{Status: core.Online},
		},
		core.Technology{
			ID:    core.OpenVPNUDPObfuscated,
			Pivot: core.Pivot{Status: core.Online},
		},
	}
	servers := core.Servers{
		core.Server{Hostname: "lt16.nordvpn.com", Technologies: obfuscatedTechnologies, Status: core.Online},
		core.Server{Hostname: "lt15.nordvpn.com", Status: core.Online}}
	dm := DataManager{serversData: ServersData{Servers: servers}}

	r := RPC{
		cm:     &mockConfigManager,
		events: &mockEvents,
		netw:   &mockN.Mock{VpnActive: true, MeshActive: true},
		dm:     &dm,
	}

	successPayload := pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{strconv.FormatBool(true)},
	}

	serverNotObfuscatedPayload := pb.Payload{
		Type: internal.CodeAutoConnectServerNotObfuscated,
	}

	serverObfuscatedPayload := pb.Payload{
		Type: internal.CodeAutoConnectServerObfuscated,
	}

	tests := []struct {
		testName           string
		obfuscate          bool
		server             string
		autoconnectEnabled bool
		payload            *pb.Payload
		eventPublished     bool
	}{
		{
			testName:           "obfuscate off autoconnect is off",
			obfuscate:          true,
			server:             "",
			autoconnectEnabled: false,
			payload:            &successPayload,
			eventPublished:     true,
		},
		{
			testName:           "obfuscate on autoconnect is on server tag is not a server name",
			obfuscate:          true,
			server:             "lt",
			autoconnectEnabled: true,
			payload:            &successPayload,
			eventPublished:     true,
		},
		{
			testName:           "obfuscate on autoconnect is on to obfuscated server",
			obfuscate:          true,
			server:             "lt16",
			autoconnectEnabled: true,
			payload:            &successPayload,
			eventPublished:     true,
		},
		{
			testName:           "obfuscate on autoconnect is on to unknown server",
			obfuscate:          true,
			server:             "lt17",
			autoconnectEnabled: true,
			payload:            &successPayload,
			eventPublished:     true,
		},
		{
			testName:           "obfuscate on autoconnect is on to non obfuscated server",
			obfuscate:          true,
			server:             "lt15",
			autoconnectEnabled: true,
			payload:            &serverNotObfuscatedPayload,
			eventPublished:     false,
		},
		{
			testName:           "obfuscate off autoconnect is on to obfuscated server",
			obfuscate:          false,
			server:             "lt16",
			autoconnectEnabled: true,
			payload:            &serverObfuscatedPayload,
			eventPublished:     false,
		},
		{
			testName:           "obfuscate off autoconnect is on to unknown server",
			obfuscate:          false,
			server:             "lt17",
			autoconnectEnabled: true,
			payload:            &successPayload,
			eventPublished:     true,
		},
		{
			testName:           "obfuscate off autoconnect is on server tag is not a server name",
			obfuscate:          false,
			server:             "lt",
			autoconnectEnabled: true,
			payload:            &successPayload,
			eventPublished:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConfigManager.c.AutoConnect = test.autoconnectEnabled
			mockConfigManager.c.AutoConnectData.ServerTag = test.server
			mockConfigManager.c.AutoConnectData.Obfuscate = !test.obfuscate

			req := pb.SetGenericRequest{Enabled: test.obfuscate}
			resp, err := r.SetObfuscate(context.Background(), &req)

			assert.NoError(t, err)
			assert.Equal(t, test.payload, resp)
			assert.Equal(t, test.eventPublished, mockPublisherSubscriber.eventPublished)
			mockPublisherSubscriber.eventPublished = false
		})
	}
}
