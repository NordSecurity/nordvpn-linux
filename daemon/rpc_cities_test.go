package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	mapset "github.com/deckarep/golang-set"
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
				fileshare: service.NoopFileshare{},
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

	cityNames := map[bool]map[config.Protocol]map[string]mapset.Set{
		false: {
			config.Protocol_UDP: {
				"lt": mapset.NewSet("Vilnius"),
			},
			config.Protocol_TCP: {
				"de": mapset.NewSet("Berlin"),
			},
		},
	}
	dm.SetAppData(nil, cityNames, nil)

	cm := newMockConfigManager()
	cm.c.AutoConnectData.Protocol = config.Protocol_UDP

	rpc := RPC{
		ac:        &workingLoginChecker{},
		cm:        cm,
		dm:        dm,
		fileshare: service.NoopFileshare{},
		netw:      &networker.Mock{},
		ncClient:  mockNC{},
		publisher: &subs.Subject[string]{},
		api:       mockApi{},
	}

	request := &pb.CitiesRequest{}
	request.Obfuscate = false
	request.Country = "LT"

	payload, _ := rpc.Cities(context.Background(), request)
	assert.Equal(t, internal.CodeSuccess, payload.GetType())
	assert.Equal(t, []string{"Vilnius"}, payload.GetData())
}
