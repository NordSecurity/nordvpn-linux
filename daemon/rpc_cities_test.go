package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
	"github.com/stretchr/testify/assert"
)

func TestRPCCities(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	ukServers := core.Servers{
		{
			Hostname:     "uk1.nordvpn.com",
			Status:       core.Online,
			Technologies: core.Technologies{{ID: core.WireguardTech, Pivot: core.Pivot{Status: core.Online}}},
			Locations: core.Locations{
				{Country: core.Country{Name: "United Kingdom", Code: "GB", City: core.City{Name: "London"}}},
			},
		},
	}

	tests := []struct {
		name        string
		cm          config.Manager
		servers     core.Servers
		countryName string
		statusCode  int64
		expected    []*pb.ServerGroup
	}{
		{
			name:        "missing configuration file",
			cm:          failingConfigManager{},
			countryName: "France",
			statusCode:  internal.CodeConfigError,
		},
		{
			name:        "no results when servers are not loaded",
			cm:          newMockConfigManager(),
			countryName: "France",
			statusCode:  internal.CodeSuccess,
			expected:    []*pb.ServerGroup{},
		},
		{
			name:        "no results for invalid country name",
			cm:          newMockConfigManager(),
			countryName: "invalid_country",
			statusCode:  internal.CodeSuccess,
			expected:    []*pb.ServerGroup{},
		},
		{
			name:        "results for country name",
			cm:          newMockConfigManager(),
			servers:     coremock.ServersList(),
			countryName: "alGeria",
			statusCode:  internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Algiers"},
			},
		},
		{
			name:        "results for gb country code",
			cm:          newMockConfigManager(),
			servers:     ukServers,
			countryName: "gb",
			statusCode:  internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "London"},
			},
		},
		{
			name:        "results for uk country code alias",
			cm:          newMockConfigManager(),
			servers:     ukServers,
			countryName: "uk",
			statusCode:  internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "London"},
			},
		},
		{
			name:        "results for uppercase uk country code alias",
			cm:          newMockConfigManager(),
			servers:     ukServers,
			countryName: "UK",
			statusCode:  internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "London"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dm := testNewDataManager()
			dm.serversData.Servers = test.servers

			if cm, ok := test.cm.(*mockConfigManager); ok {
				cm.c.VPNProtocol = config.VPNProtocol_VPN_PROTOCOL_NORDLYNX
			}

			rpc := RPC{
				ac:        &workingLoginChecker{},
				cm:        test.cm,
				dm:        dm,
				norduser:  &testnorduser.MockNorduserCombinedService{},
				netw:      &networker.Mock{},
				ncClient:  &mock.NotificationClientMock{},
				publisher: &subs.Subject[string]{},
				api:       &coremock.CredentialsAPIMock{},
			}

			payload, _ := rpc.Cities(context.Background(), &pb.CitiesRequest{Country: test.countryName})

			assert.Equal(t, test.statusCode, payload.Type)
			assert.Equal(t, len(test.expected), len(payload.Servers))
			assert.Equal(t, test.expected, payload.Servers)
		})
	}
}
