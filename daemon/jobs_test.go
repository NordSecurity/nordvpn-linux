package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
)

func mockTimeout(tries int) time.Duration {
	return time.Millisecond
}

type failingLoginChecker struct{}

func (failingLoginChecker) IsLoggedIn() bool { return false }
func (failingLoginChecker) IsVPNExpired() (bool, error) {
	return true, errors.New("IsVPNExpired error")
}
func (failingLoginChecker) IsDedicatedIPExpired() (bool, error) {
	return true, errors.New("IsDedicatedIPExipred error")
}
func (failingLoginChecker) ServiceData(serviceID int64) (*config.ServiceData, error) {
	return nil, fmt.Errorf("Not implemented")
}

func TestStartAutoConnect(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		cfg         config.Manager
		authChecker auth.Checker
		serversAPI  core.ServersAPI
		netw        networker.Networker
		expectError bool
	}{
		{
			name:        "not logged-in",
			cfg:         newMockConfigManager(),
			authChecker: &failingLoginChecker{},
			serversAPI:  &mockServersAPI{},
			netw:        &testnetworker.Mock{},
			expectError: false,
		},
		{
			name:        "config load fail",
			cfg:         &failingConfigManager{},
			authChecker: &workingLoginChecker{},
			serversAPI:  &mockServersAPI{},
			netw:        testnetworker.Failing{},
			expectError: true,
		},
		{
			name:        "failing servers API",
			cfg:         newMockConfigManager(),
			authChecker: &workingLoginChecker{},
			serversAPI:  &mockFailingServersAPI{},
			netw:        &testnetworker.Mock{},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := test.cfg
			dm := testNewDataManager()
			api := core.NewDefaultAPI(
				"1.0.0",
				"",
				http.DefaultClient,
				response.NoopValidator{},
			)

			netw := &testnetworker.Mock{}

			rpc := NewRPC(
				internal.Development,
				test.authChecker,
				cm,
				dm,
				api,
				test.serversAPI,
				&validCredentialsAPI{},
				testNewCDNAPI(),
				testNewRepoAPI(),
				&mockAuthenticationAPI{},
				"1.0.0",
				&workingFirewall{},
				daemonevents.NewEventsEmpty(),
				func(config.Technology) (vpn.VPN, error) {
					return &mock.WorkingVPN{}, nil
				},
				newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
				netw,
				&subs.Subject[string]{},
				&mock.DNSGetter{Names: []string{"1.1.1.1"}},
				nil,
				&mockAnalytics{},
				&testnorduser.MockNorduserCombinedService{},
				&RegistryMock{},
			)

			err := rpc.StartAutoConnect(mockTimeout)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type meshRenewChecker struct{}

func (meshRenewChecker) IsRegistrationInfoCorrect() bool { return true }
func (meshRenewChecker) Register() error                 { return nil }

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

type meshNetworker struct {
	allowedIncoming  []meshnet.UniqueAddress
	blockedIncoming  []meshnet.UniqueAddress
	allowedFileshare []meshnet.UniqueAddress
	blockedFileshare []meshnet.UniqueAddress
}

func (meshNetworker) Start(
	vpn.Credentials,
	vpn.ServerData,
	config.Allowlist,
	config.DNS,
	bool,
) error {
	return nil
}

func (*meshNetworker) Stop() error                                       { return nil }
func (*meshNetworker) SetMesh(mesh.MachineMap, netip.Addr, string) error { return nil }
func (*meshNetworker) UnSetMesh() error                                  { return nil }

func (n *meshNetworker) AllowFileshare(address meshnet.UniqueAddress) error {
	n.allowedFileshare = append(n.allowedFileshare, address)
	return nil
}

func (n *meshNetworker) AllowIncoming(address meshnet.UniqueAddress, lanAllowed bool) error {
	n.allowedIncoming = append(n.allowedIncoming, address)
	return nil
}

func (n *meshNetworker) BlockIncoming(address meshnet.UniqueAddress) error {
	n.blockedIncoming = append(n.blockedIncoming, address)
	return nil
}

func (n *meshNetworker) BlockFileshare(address meshnet.UniqueAddress) error {
	n.blockedFileshare = append(n.blockedFileshare, address)
	return nil
}

func (*meshNetworker) ResetRouting(mesh.MachinePeer, mesh.MachinePeers) error { return nil }
func (*meshNetworker) BlockRouting(meshnet.UniqueAddress) error               { return nil }
func (*meshNetworker) Refresh(mesh.MachineMap) error                          { return nil }
func (*meshNetworker) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}
func (*meshNetworker) LastServerName() string { return "" }

func TestStartAutoMeshnet(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		cfg         config.Manager
		authChecker auth.Checker
		serversAPI  core.ServersAPI
		netw        networker.Networker
		expectError bool
	}{
		{
			name:        "not logged-in",
			cfg:         newMockConfigManager(),
			authChecker: &failingLoginChecker{},
			serversAPI:  &mockServersAPI{},
			netw:        &testnetworker.Mock{},
			expectError: true,
		},
		{
			name:        "config load fail",
			cfg:         &failingConfigManager{},
			authChecker: &workingLoginChecker{},
			serversAPI:  &mockServersAPI{},
			netw:        &testnetworker.Mock{},
			expectError: true,
		},
		{
			name:        "failing servers API",
			cfg:         newMockConfigManager(),
			authChecker: &workingLoginChecker{},
			serversAPI:  &mockFailingServersAPI{},
			netw:        &testnetworker.Mock{},
			expectError: false,
		},
		{
			name: "meshnet not enabled",
			cfg: func() config.Manager {
				cm := newMockConfigManager()
				cm.c.Mesh = false
				return cm
			}(),
			authChecker: &workingLoginChecker{},
			serversAPI:  &mockFailingServersAPI{},
			netw:        &testnetworker.Mock{},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			api := core.NewDefaultAPI(
				"1.0.0",
				"",
				http.DefaultClient,
				response.NoopValidator{},
			)

			rpc := NewRPC(
				internal.Development,
				test.authChecker,
				test.cfg,
				testNewDataManager(),
				api,
				test.serversAPI,
				&validCredentialsAPI{},
				testNewCDNAPI(),
				testNewRepoAPI(),
				&mockAuthenticationAPI{},
				"1.0.0",
				&workingFirewall{},
				daemonevents.NewEventsEmpty(),
				func(config.Technology) (vpn.VPN, error) {
					return &mock.WorkingVPN{}, nil
				},
				newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
				test.netw,
				&subs.Subject[string]{},
				&mock.DNSGetter{Names: []string{"1.1.1.1"}},
				nil,
				&mockAnalytics{},
				&testnorduser.MockNorduserCombinedService{},
				&RegistryMock{},
			)

			meshService := meshnet.NewServer(
				test.authChecker,
				test.cfg,
				&meshRenewChecker{},
				&invitationsAPI{},
				&meshNetworker{},
				&mock.RegistryMock{},
				&mock.DNSGetter{},
				&subs.Subject[error]{},
				nil,
				&daemonevents.Events{},
				&testnorduser.MockNorduserClient{},
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
