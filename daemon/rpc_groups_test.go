package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
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
				ac:        &workingLoginChecker{},
				cm:        test.cm,
				dm:        test.dm,
				norduser:  &testnorduser.MockNorduserCombinedService{},
				netw:      &networker.Mock{},
				ncClient:  mockNC{},
				publisher: &subs.Subject[string]{},
				api:       mockApi{},
			}
			payload, _ := rpc.Groups(context.Background(), &pb.Empty{})

			assert.Equal(t, test.statusCode, payload.Type)
		})
	}
}

func TestRPCGroups_Successful(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	dm := testNewDataManager()
	dm.serversData.Servers = serversList()

	cm := newMockConfigManager()
	cm.c.AutoConnectData.Protocol = config.Protocol_TCP
	cm.c.Technology = config.Technology_NORDLYNX

	rpc := RPC{
		ac:        &workingLoginChecker{},
		cm:        cm,
		dm:        dm,
		norduser:  &testnorduser.MockNorduserCombinedService{},
		netw:      &networker.Mock{},
		ncClient:  mockNC{},
		publisher: &subs.Subject[string]{},
		api:       mockApi{},
	}

	payload, _ := rpc.Groups(context.Background(), &pb.Empty{})
	assert.Equal(t, internal.CodeSuccess, payload.GetType())
	assert.Equal(t, 2, len(payload.Servers))

	assert.Equal(t, &pb.ServerGroup{Name: "Double_VPN", VirtualLocation: false}, payload.Servers[0])
	assert.Equal(t, &pb.ServerGroup{Name: "P2P", VirtualLocation: false}, payload.Servers[1])
}
