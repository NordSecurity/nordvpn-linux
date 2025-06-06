package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type mockAutoconnectAuthChecker struct{}

func (mockAutoconnectAuthChecker) IsLoggedIn() bool            { return true }
func (mockAutoconnectAuthChecker) IsMFAEnabled() (bool, error) { return false, nil }
func (mockAutoconnectAuthChecker) IsVPNExpired() (bool, error) { return false, nil }
func (mockAutoconnectAuthChecker) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	return []auth.DedicatedIPService{}, nil
}

func TestAutoconnect(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		testName             string
		server               string
		config               config.Config
		isDedicatedIPExpired bool
		returnCode           int64
		eventPublished       bool
		expectedError        error
	}{
		{
			testName:       "autoconnect works for OpenVPN, obfuscate = off",
			server:         "",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "autoconnect works for OpenVPN, obfuscate = on",
			server:         "",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: true, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "autoconnect works for NordLynx",
			server:         "",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for country code using Nordlynx",
			server:         "de",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "autoconnect works for country code using OpenVPN and obfuscate = off",
			server:         "de",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "autoconnect works for country code using OpenVPN and obfuscate = on",
			server:         "de",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: true, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for country name using Nordlynx",
			server:         "germany",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for city name using Nordlynx",
			server:         "berlin",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for country and city name using Nordlynx",
			server:         "germany berlin",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for group name using Nordlynx",
			server:         "double_vpn",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for server name using Nordlynx",
			server:         "fr1",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for server name using OpenVPN, obfuscate = off",
			server:         "fr1",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "works for server name using OpenVPN, obfuscate = on",
			server:         "lt17",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: true, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeSuccess,
			eventPublished: true,
		},
		{
			testName:       "fails for invalid name server name using Nordlynx",
			server:         "invalid_name",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			eventPublished: false,
			expectedError:  internal.ErrTagDoesNotExist,
		},
		{
			testName:       "fails when connecting to obfuscated OpenVPN server using OpenVPN and obfuscate = off",
			server:         "lt17",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeAutoConnectServerObfuscated,
			eventPublished: false,
		},
		{
			testName:       "fails when connecting to regular OpenVPN server using OpenVPN and obfuscate = on",
			server:         "lt15",
			config:         config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: true, Protocol: config.Protocol_TCP}, Technology: config.Technology_OPENVPN},
			returnCode:     internal.CodeAutoConnectServerNotObfuscated,
			eventPublished: false,
		},
		{
			testName:             "fails to connect dedicated IP when subscription expired",
			server:               "dedicated_ip",
			config:               config.Config{AutoConnectData: config.AutoConnectData{Obfuscate: false, Protocol: config.Protocol_UDP}, Technology: config.Technology_NORDLYNX},
			isDedicatedIPExpired: true,
			returnCode:           internal.CodeDedicatedIPRenewError,
			eventPublished:       false,
		},
	}

	for _, test := range tests {
		mockAuthChecker := mockAutoconnectAuthChecker{}
		mockConfigManager := newMockConfigManager()
		mockPublisherSubscriber := events.MockPublisherSubscriber[bool]{}
		mockEvents := events.Events{Settings: &events.SettingsEvents{Autoconnect: &mockPublisherSubscriber}}
		dm := DataManager{serversData: ServersData{Servers: serversList()}}
		r := RPC{cm: mockConfigManager, ac: mockAuthChecker, events: &mockEvents, dm: &dm, serversAPI: &mockServersAPI{}}
		request := pb.SetAutoconnectRequest{Enabled: true}

		request.ServerTag = test.server
		mockConfigManager.c = test.config

		t.Run(test.testName, func(t *testing.T) {
			resp, err := r.SetAutoConnect(context.Background(), &request)

			assert.Equal(t, test.expectedError, err)
			if err == nil {
				assert.Equal(t, test.returnCode, resp.Type)
			}
			assert.Equal(t, test.eventPublished, mockPublisherSubscriber.EventPublished)
		})
	}
}

func TestAutoconnect_SavesCorrectAutoconnectData(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		testName    string
		serverGroup string
		tag         string
		expected    config.AutoConnectData
	}{
		{
			testName:    "for standard",
			serverGroup: "standard_vpn_servers",
			expected:    config.AutoConnectData{Group: config.ServerGroup_STANDARD_VPN_SERVERS},
		},
		{
			testName:    "for p2p",
			serverGroup: "p2p",
			expected:    config.AutoConnectData{Group: config.ServerGroup_P2P},
		},
		{
			testName:    "for obfuscated servers",
			serverGroup: "obfuscated_servers",
			expected:    config.AutoConnectData{Group: config.ServerGroup_OBFUSCATED},
		},
		{
			testName:    "for double_vpn",
			serverGroup: "double_vpn",
			expected:    config.AutoConnectData{Group: config.ServerGroup_DOUBLE_VPN},
		},
		{
			testName:    "for onion_over_vpn",
			serverGroup: "onion_over_vpn",
			expected:    config.AutoConnectData{Group: config.ServerGroup_ONION_OVER_VPN},
		},
		{
			testName:    "group name is in tag",
			serverGroup: "",
			tag:         "p2p",
			expected:    config.AutoConnectData{Group: config.ServerGroup_P2P, ServerTag: "p2p"},
		},
		{
			testName: "for country name",
			tag:      "germany",
			expected: config.AutoConnectData{Group: config.ServerGroup_UNDEFINED, Country: "DE", ServerTag: "germany"},
		},
		{
			testName: "for country code and city name",
			tag:      "de berlin",
			expected: config.AutoConnectData{Group: config.ServerGroup_UNDEFINED, Country: "DE", City: "Berlin", ServerTag: "de berlin"},
		},
		{
			testName:    "for country code, city name and group",
			tag:         "de berlin",
			serverGroup: "p2p",
			expected:    config.AutoConnectData{Group: config.ServerGroup_P2P, Country: "DE", City: "Berlin", ServerTag: "de berlin"},
		},
	}

	for _, test := range tests {
		mockAuthChecker := mockAutoconnectAuthChecker{}
		mockConfigManager := newMockConfigManager()
		mockPublisherSubscriber := events.MockPublisherSubscriber[bool]{}
		mockEvents := events.Events{
			Settings: &events.SettingsEvents{
				Autoconnect: &mockPublisherSubscriber,
			},
		}
		dm := DataManager{
			serversData: ServersData{
				Servers: serversList(),
			},
			countryData: CountryData{
				Countries: countriesList(),
			},
		}
		r := RPC{
			cm:         mockConfigManager,
			ac:         mockAuthChecker,
			events:     &mockEvents,
			dm:         &dm,
			serversAPI: &mockServersAPI{},
		}
		request := pb.SetAutoconnectRequest{
			Enabled:     true,
			ServerTag:   test.tag,
			ServerGroup: test.serverGroup,
		}

		t.Run(test.testName, func(t *testing.T) {
			assert.Equal(t, config.ServerGroup_UNDEFINED, mockConfigManager.c.AutoConnectData.Group)
			resp, err := r.SetAutoConnect(context.Background(), &request)

			assert.Nil(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, internal.CodeSuccess, resp.Type)

			assert.Equal(t, test.expected.ServerTag, mockConfigManager.c.AutoConnectData.ServerTag)
			assert.Equal(t, test.expected.Country, mockConfigManager.c.AutoConnectData.Country)
			assert.Equal(t, test.expected.City, mockConfigManager.c.AutoConnectData.City)
			assert.Equal(t, test.expected.Group, mockConfigManager.c.AutoConnectData.Group)
		})
	}
}
