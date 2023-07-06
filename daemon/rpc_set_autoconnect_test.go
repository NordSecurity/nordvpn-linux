package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/stretchr/testify/assert"
)

type mockAutoconnectAuthChecker struct {
}

func (mockAutoconnectAuthChecker) IsLoggedIn() bool { return true }

type mockAutoconnectConfigManager struct {
	c config.Config
}

func (*mockAutoconnectConfigManager) SaveWith(f config.SaveFunc) error {
	return nil
}

func (m *mockAutoconnectConfigManager) Load(c *config.Config) error {
	c.AutoConnect = m.c.AutoConnect
	c.AutoConnectData = m.c.AutoConnectData
	return nil
}

func (*mockAutoconnectConfigManager) Reset() error {
	return nil
}

func TestAutoconnectObfuscateInteraction(t *testing.T) {
	mockAuthChecker := mockAutoconnectAuthChecker{}

	mockConfigManager := mockAutoconnectConfigManager{}

	mockPublisherSubscriber := mockPublisherSubcriber{}
	mockEvents := Events{Settings: &SettingsEvents{Autoconnect: &mockPublisherSubscriber}}

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

	r := RPC{cm: &mockConfigManager, ac: mockAuthChecker, events: &mockEvents, dm: &dm}

	request := pb.SetAutoconnectRequest{AutoConnect: true}

	tests := []struct {
		testName         string
		server           string
		obfuscateEnabled bool
		returnCode       int64
		eventPublished   bool
	}{
		{"obfuscate is off", "", false, internal.CodeSuccess, true},
		{"obfuscate is on unknown server", "lt", true, internal.CodeSuccess, true},
		{"obfuscate is on server is obfuscated", "lt16", true, internal.CodeSuccess, true},
		{"obfuscate is on server is not obfuscated", "lt15", true, internal.CodeAutoConnectServerNotObfuscated, false},
		{"obfuscate is off server is obfuscated", "lt16", false, internal.CodeAutoConnectServerObfuscated, false},
		{"obfuscate is off server is not obfuscated", "lt15", false, internal.CodeSuccess, true},
	}

	for _, test := range tests {
		request.ServerTag = test.server
		mockConfigManager.c.AutoConnectData.Obfuscate = test.obfuscateEnabled

		t.Run(test.testName, func(t *testing.T) {
			resp, err := r.SetAutoConnect(context.Background(), &request)

			assert.NoError(t, err)
			assert.Equal(t, &pb.Payload{
				Type: test.returnCode,
			}, resp)
			assert.Equal(t, test.eventPublished, mockPublisherSubscriber.eventPublished)
			mockPublisherSubscriber.eventPublished = false
		})
	}
}
