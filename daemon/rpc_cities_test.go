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

func TestRPCCities(t *testing.T) {
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
			name:       "return no results when no servers are loaded",
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
			payload, _ := rpc.Cities(context.Background(), &pb.CitiesRequest{})

			assert.Equal(t, test.statusCode, payload.Type)
		})
	}
}

func TestRPCCities_Successful(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	dm := testNewDataManager()
	dm.serversData.Servers = serversList()

	cm := newMockConfigManager()
	cm.c.AutoConnectData.Protocol = config.Protocol_UDP

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

	request := &pb.CitiesRequest{}
	request.Country = "LT"

	payload, _ := rpc.Cities(context.Background(), request)
	assert.Equal(t, internal.CodeSuccess, payload.GetType())
	assert.Equal(t, 1, len(payload.Servers))

	assert.Equal(t, &pb.ServerGroup{Name: "Vilnius", VirtualLocation: true}, payload.Servers[0])
}
