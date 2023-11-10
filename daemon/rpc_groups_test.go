package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/fileshare/daemon"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	mapset "github.com/deckarep/golang-set"
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
			name:       "app data is empty",
			dm:         testNewDataManager(),
			cm:         newMockConfigManager(),
			statusCode: internal.CodeEmptyPayloadError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpc := RPC{
				ac:        &workingLoginChecker{},
				cm:        test.cm,
				dm:        test.dm,
				fileshare: daemon.NoopFileshare{},
				netw:      &networker.Mock{},
				ncClient:  mockNC{},
				publisher: &subs.Subject[string]{},
				api:       mockApi{},
			}
			payload, _ := rpc.Groups(context.Background(), &pb.GroupsRequest{})

			assert.Equal(t, test.statusCode, payload.Type)
		})
	}
}

func TestRPCGroups_Successful(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	dm := testNewDataManager()

	groupNames := map[bool]map[config.Protocol]mapset.Set{
		false: {
			config.Protocol_UDP: mapset.NewSet("false_Protocol_UDP"),
			config.Protocol_TCP: mapset.NewSet("false_Protocol_TCP"),
		},
	}
	dm.SetAppData(nil, nil, groupNames)

	cm := newMockConfigManager()
	cm.c.AutoConnectData.Protocol = config.Protocol_TCP

	rpc := RPC{
		ac:        &workingLoginChecker{},
		cm:        cm,
		dm:        dm,
		fileshare: daemon.NoopFileshare{},
		netw:      &networker.Mock{},
		ncClient:  mockNC{},
		publisher: &subs.Subject[string]{},
		api:       mockApi{},
	}

	groupsRequest := &pb.GroupsRequest{}
	groupsRequest.Obfuscate = false

	payload, _ := rpc.Groups(context.Background(), groupsRequest)
	assert.Equal(t, internal.CodeSuccess, payload.GetType())
	assert.Equal(t, []string{"false_Protocol_TCP"}, payload.GetData())
}
