package daemon

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

func TestServerCheck_DedicatedServersAreNotChecked(t *testing.T) {
	category.Set(t, category.Unit)

	serverBeforeUpdate := core.Server{ID: 1122, Status: core.Online, Penalty: 1.0}

	serverAfterUpdate := core.Server{ID: 1122,
		Status: core.Online,
		// Penalty is expected to be 4.0 because it's calculated by dividing Load by 10 and exponentiating it by itself.
		Penalty: 4.0,
		Load:    20}

	tests := []struct {
		name                 string
		serverGroup          config.ServerGroup
		isVPNActive          bool
		severResponse        core.Server
		expectedServerStatus core.Status
		expectedPenalty      float64
		shouldCallServersAPI bool
	}{
		{
			name:                 "VPN is active and server is not a dedicated server, server data is updated",
			serverGroup:          config.ServerGroup_STANDARD_VPN_SERVERS,
			isVPNActive:          true,
			severResponse:        serverAfterUpdate,
			expectedPenalty:      serverAfterUpdate.Penalty,
			expectedServerStatus: serverAfterUpdate.Status,
			shouldCallServersAPI: true,
		},
		{
			name:                 "VPN is active, server is a dedicated server, server data is not updated",
			serverGroup:          config.ServerGroup_DEDICATED_SERVER,
			isVPNActive:          true,
			expectedServerStatus: serverBeforeUpdate.Status,
			expectedPenalty:      serverBeforeUpdate.Penalty,
			shouldCallServersAPI: false,
		},
		{
			name:                 "VPN is not active, server is not a dedicated server, server data is not updated",
			serverGroup:          config.ServerGroup_STANDARD_VPN_SERVERS,
			isVPNActive:          false,
			expectedServerStatus: serverBeforeUpdate.Status,
			expectedPenalty:      serverBeforeUpdate.Penalty,
			shouldCallServersAPI: false,
		},
		{
			name:                 "VPN is not active, server is a dedicated server, server data is not updated",
			serverGroup:          config.ServerGroup_DEDICATED_SERVER,
			isVPNActive:          false,
			expectedServerStatus: serverBeforeUpdate.Status,
			expectedPenalty:      serverBeforeUpdate.Penalty,
			shouldCallServersAPI: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataManager := DataManager{serversData: ServersData{
				Servers: core.Servers{
					serverBeforeUpdate,
				},
			}}

			serversAPIMock := testcore.MockServersAPI{ServersList: core.Servers{test.severResponse}}
			jobFunc := JobServerCheck(&dataManager,
				&serversAPIMock,
				&testnetworker.Mock{VpnActive: test.isVPNActive},
				core.Server{ID: 1122, Groups: core.Groups{core.Group{ID: test.serverGroup}}})
			jobFunc()

			assert.Equal(t, test.expectedServerStatus, dataManager.serversData.Servers[0].Status)
			assert.Equal(t, test.expectedPenalty, dataManager.serversData.Servers[0].Penalty)
			assert.Equal(t,
				test.shouldCallServersAPI,
				serversAPIMock.ServerEndpointCalled,
				"Servers endpoint should not be called if current server is a dedicated server.")
		})
	}
}
