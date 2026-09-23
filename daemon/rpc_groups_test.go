package daemon

import (
	"context"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
	"github.com/stretchr/testify/assert"
)

func TestRPCGroups(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	tests := []struct {
		name       string
		dm         *DataManager
		cm         config.Manager
		statusCode int64
	}{
		{
			name:       "missing configuration file",
			dm:         testNewDataManager(),
			cm:         failingConfigManager{},
			statusCode: internal.CodeConfigError,
		},
		{
			name:       "no results when servers list is empty",
			dm:         testNewDataManager(),
			cm:         newMockConfigManager(),
			statusCode: internal.CodeSuccess,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpc := RPC{
				ac:                 &workingLoginChecker{},
				cm:                 test.cm,
				dm:                 test.dm,
				norduser:           &testnorduser.MockNorduserCombinedService{},
				netw:               &networker.Mock{},
				ncClient:           &mock.NotificationClientMock{},
				publisher:          &subs.Subject[string]{},
				api:                &coremock.CredentialsAPIMock{},
				remoteConfigGetter: mock.NewRemoteConfigMock(),
			}
			payload, _ := rpc.Groups(context.Background(), &pb.Empty{})

			assert.Equal(t, test.statusCode, payload.Type)
		})
	}
}

func TestRPCGroups_RegionalGroupsFiltered(t *testing.T) {
	category.Set(t, category.Unit)

	dm := testNewDataManager()
	dm.serversData.Servers = core.Servers{
		{
			Status: core.Online,
			Technologies: core.Technologies{
				{ID: core.WireguardTech, Pivot: core.Pivot{Status: core.Online}},
			},
			Groups: core.Groups{{ID: 19, Title: "Europe"}},
		},
		{
			Status: core.Online,
			Technologies: core.Technologies{
				{ID: core.WireguardTech, Pivot: core.Pivot{Status: core.Online}},
			},
			Groups: core.Groups{{ID: config.ServerGroup_P2P, Title: "P2P"}},
		},
	}

	cm := newMockConfigManager()
	cm.c.Technology = config.Technology_NORDLYNX

	rpc := RPC{
		cm: cm,
		dm: dm,
		remoteConfigGetter: &mock.RemoteConfigMock{
			FeatureToggles: map[string]bool{
				remote.FeatureDedicatedServer: false,
			},
		},
	}
	payload, _ := rpc.Groups(context.Background(), &pb.Empty{})

	assert.Equal(t, internal.CodeSuccess, payload.Type)
	assert.Equal(t, 1, len(payload.Servers))
	assert.Equal(t, "P2P", payload.Servers[0].Name)
}

func TestRPCGroups_Successful(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	tests := []struct {
		name                    string
		cm                      config.Manager
		servers                 core.Servers
		disableVirtualServers   bool
		disableDedicatedServers bool
		statusCode              int64
		expected                []*pb.ServerGroup
	}{
		{
			name:       "missing configuration file",
			cm:         failingConfigManager{},
			statusCode: internal.CodeConfigError,
		},
		{
			name:       "dedicated servers group present",
			cm:         newMockConfigManager(),
			statusCode: internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Dedicated_Server", VirtualLocation: false},
			},
		},
		{
			name:       "virtual and physical servers",
			cm:         newMockConfigManager(),
			servers:    coremock.ServersList(),
			statusCode: internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Dedicated_IP", VirtualLocation: false},
				{Name: "Double_VPN", VirtualLocation: false},
				{Name: "P2P", VirtualLocation: false},
				{Name: "Standard_VPN_Servers", VirtualLocation: false},
				{Name: "Dedicated_Server", VirtualLocation: false},
			},
		},
		{
			name:                    "virtual and physical servers, exclude dedicated servers via feature toggle",
			cm:                      newMockConfigManager(),
			servers:                 coremock.ServersList(),
			disableDedicatedServers: true,
			statusCode:              internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Dedicated_IP", VirtualLocation: false},
				{Name: "Double_VPN", VirtualLocation: false},
				{Name: "P2P", VirtualLocation: false},
				{Name: "Standard_VPN_Servers", VirtualLocation: false},
			},
		},
		{
			name:                  "return physical servers only",
			cm:                    newMockConfigManager(),
			servers:               coremock.ServersList(),
			disableVirtualServers: true,
			statusCode:            internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Dedicated_IP", VirtualLocation: false},
				{Name: "Double_VPN", VirtualLocation: false},
				{Name: "P2P", VirtualLocation: false},
				{Name: "Standard_VPN_Servers", VirtualLocation: false},
				{Name: "Dedicated_Server", VirtualLocation: false},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dm := testNewDataManager()
			dm.serversData.Servers = test.servers

			rc := mock.NewRemoteConfigMock()
			rc.AddFeatureToggle(remote.FeatureDedicatedServer, !test.disableDedicatedServers)

			if cm, ok := test.cm.(*mockConfigManager); ok {
				cm.c.AutoConnectData.Protocol = config.Protocol_UDP
				cm.c.Technology = config.Technology_NORDLYNX
				cm.c.VirtualLocation.Set(!test.disableVirtualServers)
			}

			rpc := RPC{
				ac:                 &workingLoginChecker{},
				cm:                 test.cm,
				dm:                 dm,
				norduser:           &testnorduser.MockNorduserCombinedService{},
				netw:               &networker.Mock{},
				ncClient:           &mock.NotificationClientMock{},
				publisher:          &subs.Subject[string]{},
				api:                &coremock.CredentialsAPIMock{},
				remoteConfigGetter: rc,
			}
			payload, _ := rpc.Groups(context.Background(), &pb.Empty{})

			assert.Equal(t, test.statusCode, payload.Type)
			assert.Equal(t, len(test.expected), len(payload.Servers))
			assert.ElementsMatch(t, test.expected, payload.Servers)
		})
	}
}

func TestGroups_ObfuscatedIsListedOnlyUnderNordWhisper(t *testing.T) {
	category.Set(t, category.Unit)

	allTechs := []core.ServerTechnology{
		core.OpenVPNTCP, core.OpenVPNUDP, core.WireguardTech, core.NordWhisperTech,
	}
	standard := getServer(1, "standard1", "Germany", "de", "Berlin", false,
		core.Groups{{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"}},
		allTechs)
	legacyXOR := getServer(3, "xor1", "Canada", "ca", "Toronto", false,
		core.Groups{{ID: config.ServerGroup_OBFUSCATED, Title: "Obfuscated Servers"}},
		[]core.ServerTechnology{core.OpenVPNUDPObfuscated, core.OpenVPNTCPObfuscated})

	tests := []struct {
		name     string
		servers  core.Servers
		tech     config.Technology
		proto    config.Protocol
		expected bool
	}{
		{
			name:     "listed over nordwhisper",
			servers:  core.Servers{standard, legacyXOR},
			tech:     config.Technology_NORDWHISPER,
			proto:    config.Protocol_Webtunnel,
			expected: true,
		},
		{
			name:    "not listed over nordlynx",
			servers: core.Servers{standard, legacyXOR},
			tech:    config.Technology_NORDLYNX,
			proto:   config.Protocol_UDP,
		},
		{
			name:    "not listed over openvpn tcp",
			servers: core.Servers{standard, legacyXOR},
			tech:    config.Technology_OPENVPN,
			proto:   config.Protocol_TCP,
		},
		{
			name:    "not listed over openvpn udp",
			servers: core.Servers{standard, legacyXOR},
			tech:    config.Technology_OPENVPN,
			proto:   config.Protocol_UDP,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dm := DataManager{serversData: ServersData{Servers: test.servers}}

			groups, err := dm.Groups(test.tech, test.proto, true)
			assert.NoError(t, err)

			names := make([]string, 0, len(groups))
			for _, group := range groups {
				names = append(names, group.Name)
			}

			obfuscated := internal.Title(serverpicker.ObfuscatedServersGroupTitle)
			if test.expected {
				assert.Contains(t, names, obfuscated)
				assert.Equal(t, 1, strings.Count(strings.Join(names, " "), obfuscated),
					"the group must be reported exactly once")
			} else {
				assert.NotContains(t, names, obfuscated)
			}
		})
	}
}
