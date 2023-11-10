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

func TestRPCCountries(t *testing.T) {
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
			payload, _ := rpc.Countries(context.Background(), &pb.CountriesRequest{})

			assert.Equal(t, test.statusCode, payload.Type)
		})
	}
}

func TestRPCCountries_Successful(t *testing.T) {
	category.Set(t, category.Unit)
	defer testsCleanup()

	dm := testNewDataManager()

	countryNames := map[bool]map[config.Protocol]mapset.Set{
		false: {
			config.Protocol_UDP: mapset.NewSet("LT"),
			config.Protocol_TCP: mapset.NewSet("DE"),
		},
	}
	dm.SetAppData(countryNames, nil, nil)

	cm := newMockConfigManager()
	cm.c.AutoConnectData.Protocol = config.Protocol_UDP

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

	request := &pb.CountriesRequest{}
	request.Obfuscate = false

	payload, _ := rpc.Countries(context.Background(), request)
	assert.Equal(t, internal.CodeSuccess, payload.GetType())
	assert.Equal(t, []string{"LT"}, payload.GetData())
}
