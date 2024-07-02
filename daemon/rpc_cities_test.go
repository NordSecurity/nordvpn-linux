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
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
	"github.com/stretchr/testify/assert"
)

func TestRPCCities(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	tests := []struct {
		name                  string
		cm                    config.Manager
		servers               core.Servers
		disableVirtualServers bool
		countryName           string
		statusCode            int64
		expected              []*pb.ServerGroup
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
			name:        "results for country name with virtual servers only",
			cm:          newMockConfigManager(),
			servers:     serversList(),
			countryName: "liThuania",
			statusCode:  internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Kaunas", VirtualLocation: true},
				{Name: "Vilnius", VirtualLocation: true},
			},
		},
		{
			name:        "results for country code with virtual servers only",
			cm:          newMockConfigManager(),
			servers:     serversList(),
			countryName: "lT",
			statusCode:  internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Kaunas", VirtualLocation: true},
				{Name: "Vilnius", VirtualLocation: true},
			},
		},
		{
			name:                  "no servers for country with virtual servers only and virtual location is off",
			cm:                    newMockConfigManager(),
			servers:               serversList(),
			countryName:           "lT",
			disableVirtualServers: true,
			statusCode:            internal.CodeSuccess,
			expected:              []*pb.ServerGroup{},
		},
		{
			name:        "results for country code with physical servers only",
			cm:          newMockConfigManager(),
			servers:     serversList(),
			countryName: "fR",
			statusCode:  internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "Paris", VirtualLocation: false},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dm := testNewDataManager()
			dm.serversData.Servers = test.servers

			if cm, ok := test.cm.(*mockConfigManager); ok {
				cm.c.AutoConnectData.Protocol = config.Protocol_UDP
				cm.c.Technology = config.Technology_NORDLYNX
				cm.c.VirtualLocation.Set(!test.disableVirtualServers)
			}

			rpc := RPC{
				ac:        &workingLoginChecker{},
				cm:        test.cm,
				dm:        dm,
				norduser:  &testnorduser.MockNorduserCombinedService{},
				netw:      &networker.Mock{},
				ncClient:  mockNC{},
				publisher: &subs.Subject[string]{},
				api:       mockApi{},
			}

			payload, _ := rpc.Cities(context.Background(), &pb.CitiesRequest{Country: test.countryName})

			assert.Equal(t, test.statusCode, payload.Type)
			assert.Equal(t, len(test.expected), len(payload.Servers))
			assert.Equal(t, test.expected, payload.Servers)
		})
	}
}
