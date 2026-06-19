package daemon

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/sharedctx"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	core_test "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
)

func mockTimeout(tries int) time.Duration {
	return time.Millisecond
}

type failingLoginChecker struct{}

func (failingLoginChecker) IsLoggedIn() (bool, error)   { return false, nil }
func (failingLoginChecker) IsMFAEnabled() (bool, error) { return false, nil }
func (failingLoginChecker) IsVPNExpired() (bool, error) {
	return true, errors.New("IsVPNExpired error")
}

func (failingLoginChecker) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	return nil, fmt.Errorf("Not implemented")
}
func (failingLoginChecker) GetDedicatedServerService() (auth.DedicatedServerService, error) {
	return auth.DedicatedServerService{}, fmt.Errorf("Not implemented")
}

func updateAutoconnectData(c *mockConfigManager, data config.AutoConnectData) {
	c.c.AutoConnect = true
	c.c.AutoConnectData.ServerTag = data.ServerTag
	c.c.AutoConnectData.Country = data.Country
	c.c.AutoConnectData.City = data.City
	c.c.AutoConnectData.Group = data.Group
}

func TestStartAutoConnect(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		expectError bool
		setup       func(*RPC)
	}{
		{
			name:        "not logged-in",
			setup:       func(rpc *RPC) { rpc.ac = failingLoginChecker{} },
			expectError: false,
		},
		{
			name:        "config load fail",
			setup:       func(rpc *RPC) { rpc.cm = failingConfigManager{} },
			expectError: false,
		},
		{
			name:        "failing servers API",
			setup:       func(rpc *RPC) { rpc.serversAPI = core_test.NewMockFailingServersAPI(errors.New("500")) },
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpc := testRPC()
			if test.setup != nil {
				test.setup(rpc)
			}
			err := rpc.StartAutoConnect(mockTimeout)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDoAutoConnect_BailsOutWhenAutoConnectDisabled(t *testing.T) {
	category.Set(t, category.Unit)

	rpc := testRPC()
	cm := newMockConfigManager()
	cm.c.AutoConnect = false
	rpc.cm = cm

	err := rpc.doAutoConnect()
	assert.ErrorIs(t, err, errAutoConnectDisabled)
}

func TestDoAutoconnectHandlesServerAvailabilityIssues(t *testing.T) {
	category.Set(t, category.Unit)

	rpc := testRPC()
	rpc.serversAPI = core_test.NewMockFailingServersAPI(errors.New("500"))
	mockConfigManager := newMockConfigManager()
	updateAutoconnectData(mockConfigManager, config.AutoConnectData{Country: "DE"})
	rpc.cm = mockConfigManager

	rpc.dm.SetServersData(time.Now(), []core.Server{}, "")

	err := rpc.doAutoConnect()
	assert.ErrorIs(t, err, errServersUnavailable, "doAutoconnect has ignored server availability errors")

	rpc.dm.SetServersData(time.Now(), core_test.ServersList(), "")
	err = rpc.doAutoConnect()
	assert.Nil(t, err, "unexpected error returned by doAutoconnect")
}

func TestDoAutoConnect(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name  string
		setup func(*RPC)
	}{
		{
			name: "connects to obfuscated group",
			setup: func(rpc *RPC) {
				rpc.serversAPI = core_test.NewMockServersAPI()
				mockConfigManager := newMockConfigManager()

				// For obfuscated the server group from API is Obfuscated_servers
				updateAutoconnectData(mockConfigManager, config.AutoConnectData{Group: config.ServerGroup_OBFUSCATED, ServerTag: "obfuscated_servers"})
				mockConfigManager.c.AutoConnectData.Obfuscate = true
				mockConfigManager.c.Technology = config.Technology_OPENVPN

				rpc.cm = mockConfigManager
			},
		},
		{
			name: "connects to country code",
			setup: func(rpc *RPC) {
				rpc.serversAPI = core_test.NewMockServersAPI()
				mockConfigManager := newMockConfigManager()

				updateAutoconnectData(mockConfigManager, config.AutoConnectData{Country: "DE"})

				rpc.cm = mockConfigManager
			},
		},
		{
			name: "connects to country + city",
			setup: func(rpc *RPC) {
				rpc.serversAPI = core_test.NewMockServersAPI()
				mockConfigManager := newMockConfigManager()

				updateAutoconnectData(mockConfigManager, config.AutoConnectData{Country: "DE", City: "Berlin"})

				rpc.cm = mockConfigManager
			},
		},
		{
			name: "connects to country + city + group",
			setup: func(rpc *RPC) {
				rpc.serversAPI = core_test.NewMockServersAPI()
				mockConfigManager := newMockConfigManager()

				updateAutoconnectData(mockConfigManager, config.AutoConnectData{Country: "DE", City: "Berlin", Group: config.ServerGroup_P2P, ServerTag: "p2p"})

				rpc.cm = mockConfigManager
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpc := testRPC()
			if test.setup != nil {
				test.setup(rpc)
			}
			err := rpc.doAutoConnect()
			assert.NoError(t, err)
		})
	}
}

func TestDoAutoConnect_DedicatedServerFallback_ServiceExpired(t *testing.T) {
	category.Set(t, category.Unit)

	mockConfigManager := newMockConfigManager()
	updateAutoconnectData(mockConfigManager, config.AutoConnectData{
		Group:     config.ServerGroup_DEDICATED_SERVER,
		ServerTag: "Dedicated Server",
	})

	rpc := testRPC()
	rpc.cm = mockConfigManager
	rpc.doAutoConnect()

	// Verify that settings are not modified if service is available
	assert.Equal(t, config.ServerGroup_DEDICATED_SERVER, mockConfigManager.c.AutoConnectData.Group,
		"Unexpected autoconnect group target change after doAutoConnect. Group should remain set to %s.", config.ServerGroup_DEDICATED_SERVER.String())
	assert.Equal(t, "Dedicated Server", mockConfigManager.c.AutoConnectData.ServerTag,
		"Unexpected autoconnect target ServerTag change after doAutoConnect. ServerTag should remain set to DedicatedServer.")

	authMock := rpc.ac.(*workingLoginChecker)
	authMock.dedicatedServerErr = errors.New("failed to fetch ds service")

	rpc.doAutoConnect()
	// Verify that setting are not modified if service fetch failed
	assert.Equal(t, config.ServerGroup_DEDICATED_SERVER, mockConfigManager.c.AutoConnectData.Group,
		"Unexpected autoconnect group target change after doAutoConnect. Group should remain set to %s.", config.ServerGroup_DEDICATED_SERVER.String())
	assert.Equal(t, "Dedicated Server", mockConfigManager.c.AutoConnectData.ServerTag,
		"Unexpected autoconnect target ServerTag change after doAutoConnect. ServerTag should remain set to DedicatedServer.")

	authMock.dedicatedServerErr = nil
	authMock.isDedicatedServersExpired = true

	rpc.doAutoConnect()

	// Verify that settings are  modified if service is not available
	assert.NotEqual(t, config.ServerGroup_DEDICATED_SERVER, mockConfigManager.c.AutoConnectData.Group,
		"Group should be unset when dedicated servers service is not available.")
	assert.Equal(t, "", mockConfigManager.c.AutoConnectData.ServerTag,
		"ServerTag should be unset when dedicated servers service is not available.")
}

func TestDoAutoConnect_DedicatedServerFallback_FeatureDisabledInRemoteConfig(t *testing.T) {
	category.Set(t, category.Unit)

	mockConfigManager := newMockConfigManager()
	updateAutoconnectData(mockConfigManager, config.AutoConnectData{
		Group:     config.ServerGroup_DEDICATED_SERVER,
		ServerTag: "Dedicated Server",
	})

	remoteConfigMock := mock.NewRemoteConfigMock()
	remoteConfigMock.AddFeatureToggle(remote.FeatureDedicatedServer, false)

	rpc := testRPC()
	rpc.cm = mockConfigManager
	rpc.remoteConfigGetter = remoteConfigMock

	rpc.doAutoConnect()
	assert.NotEqual(t, config.ServerGroup_DEDICATED_SERVER, mockConfigManager.c.AutoConnectData.Group,
		"Group should be unset when dedicated servers feature is disabled in remote config.")
	assert.Equal(t, "", mockConfigManager.c.AutoConnectData.ServerTag,
		"ServerTag should be unset when dedicated servers feature is disabled in remote config.")
}

type meshRenewChecker struct{}

func (meshRenewChecker) CheckAndRegisterMeshnet() bool { return true }
func (meshRenewChecker) ForceRegisterMeshnet() error   { return nil }

type invitationsAPI struct{}

func (invitationsAPI) Invite(string, uuid.UUID, string, bool, bool, bool, bool) error { return nil }
func (invitationsAPI) Sent(string, uuid.UUID) (mesh.Invitations, error) {
	return mesh.Invitations{}, nil
}

func (invitationsAPI) Received(string, uuid.UUID) (mesh.Invitations, error) {
	return mesh.Invitations{}, nil
}

func (invitationsAPI) Accept(string, uuid.UUID, uuid.UUID, bool, bool, bool, bool) error { return nil }
func (invitationsAPI) Revoke(string, uuid.UUID, uuid.UUID) error                         { return nil }
func (invitationsAPI) Reject(string, uuid.UUID, uuid.UUID) error                         { return nil }

type meshNetworker struct{}

func (meshNetworker) Start(
	context.Context,
	vpn.Credentials,
	vpn.ServerData,
	config.Allowlist,
	config.DNS,
	bool,
	events.DisconnectCallback,
) error {
	return nil
}

func (*meshNetworker) Stop() error                                       { return nil }
func (*meshNetworker) SetMesh(mesh.MachineMap, netip.Addr, string) error { return nil }
func (*meshNetworker) UnSetMesh() error                                  { return nil }

func (n *meshNetworker) PermitFileshare() error {
	return nil
}

func (n *meshNetworker) ForbidFileshare() error {
	return nil
}

func (*meshNetworker) Refresh(mesh.MachineMap) error { return nil }
func (*meshNetworker) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}
func (*meshNetworker) LastServerName() string { return "" }
func (*meshNetworker) GetConnectionParameters() (vpn.ServerData, bool) {
	return vpn.ServerData{}, false
}

func TestStartAutoMeshnet(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		cfg         config.Manager
		authChecker auth.Checker
		serversAPI  core.ServersAPI
		expectError bool
	}{
		{
			name:        "not logged-in",
			authChecker: &failingLoginChecker{},
			expectError: true,
		},
		{
			name:        "config load fail",
			cfg:         &failingConfigManager{},
			expectError: true,
		},
		{
			name:        "failing servers API",
			serversAPI:  core_test.NewMockFailingServersAPI(errors.New("500")),
			expectError: false,
		},
		{
			name: "meshnet not enabled",
			cfg: func() config.Manager {
				cm := newMockConfigManager()
				cm.c.Mesh = false
				return cm
			}(),
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpc := testRPC()
			if test.cfg != nil {
				rpc.cm = test.cfg
			}
			if test.authChecker != nil {
				rpc.ac = test.authChecker
			}
			if test.serversAPI != nil {
				rpc.serversAPI = test.serversAPI
			}

			registry := &mock.RegistryMock{}
			meshService := meshnet.NewServer(
				rpc.ac,
				rpc.cm,
				&meshRenewChecker{},
				&invitationsAPI{},
				&meshNetworker{},
				registry,
				registry,
				&mock.DNSGetter{},
				&subs.Subject[error]{},
				&daemonevents.Events{},
				&testnorduser.MockNorduserClient{},
				sharedctx.New(),
			)

			err := rpc.StartAutoMeshnet(meshService, mockTimeout)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
