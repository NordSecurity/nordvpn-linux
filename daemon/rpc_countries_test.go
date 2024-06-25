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

func TestRPCCountries(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	tests := []struct {
		name                  string
		cm                    config.Manager
		servers               core.Servers
		disableVirtualServers bool
		statusCode            int64
		expected              []*pb.ServerGroup
	}{
		{
			name:       "missing configuration file",
			cm:         failingConfigManager{},
			statusCode: internal.CodeConfigError,
		},
		{
			name:       "return no results when no servers are loaded",
			cm:         newMockConfigManager(),
			statusCode: internal.CodeSuccess,
			expected:   []*pb.ServerGroup{},
		},
		{
			name:       "return virtual and physical servers",
			cm:         newMockConfigManager(),
			servers:    serversList(),
			statusCode: internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "France", VirtualLocation: false},
				{Name: "Lithuania", VirtualLocation: true},
			},
		},
		{
			name:                  "return physical servers only",
			cm:                    newMockConfigManager(),
			servers:               serversList(),
			disableVirtualServers: true,
			statusCode:            internal.CodeSuccess,
			expected: []*pb.ServerGroup{
				{Name: "France", VirtualLocation: false},
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
			payload, _ := rpc.Countries(context.Background(), &pb.Empty{})

			assert.Equal(t, test.statusCode, payload.Type)
			assert.Equal(t, len(test.expected), len(payload.Servers))
			assert.Equal(t, test.expected, payload.Servers)
		})
	}
}
