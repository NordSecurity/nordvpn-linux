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

	dataManager := DataManager{}
	serversDataBeforeUpdate := dataManager.serversData
	serversAPIMock := testcore.ServersAPIMock{}
	jobFunc := JobServerCheck(&dataManager,
		&serversAPIMock,
		&testnetworker.Mock{VpnActive: true},
		core.Server{Groups: core.Groups{core.Group{ID: config.ServerGroup_DEDICATED_SERVERS}}})
	jobFunc()

	assert.Equal(t, serversDataBeforeUpdate,
		dataManager.serversData,
		"Servers data should not be updated if the current server is a dedicated server.")
	assert.Equal(t,
		serversAPIMock.ServerEndpointCalled,
		false,
		"Servers endpoint should not be called if current server is a dedicated server.")
}
