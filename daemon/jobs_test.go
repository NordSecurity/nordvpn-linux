package daemon

import (
	"errors"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func mockTimeout(tries int) time.Duration {
	return time.Duration(time.Millisecond)
}

type failingLoginChecker struct{}

func (failingLoginChecker) IsLoggedIn() bool { return false }
func (failingLoginChecker) IsVPNExpired() (bool, error) {
	return true, errors.New("IsVPNExpired error")
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
				NewEvents(
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[events.DataDNS]{},
					&subs.Subject[bool]{},
					&subs.Subject[config.Protocol]{},
					&subs.Subject[events.DataAllowlist]{},
					&subs.Subject[config.Technology]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[any]{},
					&subs.Subject[events.DataConnect]{},
					&subs.Subject[events.DataDisconnect]{},
					&subs.Subject[any]{},
					&subs.Subject[core.ServicesResponse]{},
					&subs.Subject[events.ServerRating]{},
					&subs.Subject[int]{},
				),
				func(config.Technology) (vpn.VPN, error) {
					return &workingVPN{}, nil
				},
				newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
				netw,
				&subs.Subject[string]{},
				mockNameservers([]string{"1.1.1.1"}),
				nil,
				&mockAnalytics{},
				service.NoopFileshare{},
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

func (meshRenewChecker) IsRegistered() bool { return true }

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

type registryAPI struct {
	localPeers   mesh.Machines
	machinePeers mesh.MachinePeers
	listErr      error
	configureErr error
}

func (registryAPI) Register(string, mesh.Machine) (*mesh.Machine, error) {
	return &mesh.Machine{}, nil
}

func (*registryAPI) Update(string, uuid.UUID, []netip.AddrPort) error { return nil }

func (r *registryAPI) Configure(string, uuid.UUID, uuid.UUID, bool, bool, bool, bool, bool) error {
	return r.configureErr
}

func (*registryAPI) Unregister(string, uuid.UUID) error { return nil }

func (r *registryAPI) List(string, uuid.UUID) (mesh.MachinePeers, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.machinePeers, nil
}

func (r *registryAPI) Local(string) (mesh.Machines, error) {
	return r.localPeers, nil
}

func (r *registryAPI) Unpair(string, uuid.UUID, uuid.UUID) error { return nil }

func (r *registryAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	return &mesh.MachineMap{}, nil
}
func (r *registryAPI) NotifyNewTransfer(
	token string,
	self uuid.UUID,
	peer uuid.UUID,
	fileName string,
	fileCount int,
) error {
	return nil
}

type dnsGetter struct{}

func (dnsGetter) Get(bool, bool) []string { return nil }

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

func (n *meshNetworker) AllowIncoming(address meshnet.UniqueAddress) error {
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

func (*meshNetworker) ResetRouting(mesh.MachinePeers) error     { return nil }
func (*meshNetworker) BlockRouting(meshnet.UniqueAddress) error { return nil }
func (*meshNetworker) Refresh(mesh.MachineMap) error            { return nil }
func (*meshNetworker) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}

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
				NewEvents(
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[events.DataDNS]{},
					&subs.Subject[bool]{},
					&subs.Subject[config.Protocol]{},
					&subs.Subject[events.DataAllowlist]{},
					&subs.Subject[config.Technology]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[any]{},
					&subs.Subject[events.DataConnect]{},
					&subs.Subject[events.DataDisconnect]{},
					&subs.Subject[any]{},
					&subs.Subject[core.ServicesResponse]{},
					&subs.Subject[events.ServerRating]{},
					&subs.Subject[int]{},
				),
				func(config.Technology) (vpn.VPN, error) {
					return &workingVPN{}, nil
				},
				newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
				test.netw,
				&subs.Subject[string]{},
				mockNameservers([]string{"1.1.1.1"}),
				nil,
				&mockAnalytics{},
				service.NoopFileshare{},
				&RegistryMock{},
			)

			meshService := meshnet.NewServer(
				test.authChecker,
				test.cfg,
				&meshRenewChecker{},
				&invitationsAPI{},
				&meshNetworker{},
				&registryAPI{},
				&dnsGetter{},
				&subs.Subject[error]{},
				nil,
				&subs.Subject[bool]{},
				&subs.Subject[events.DataConnect]{},
				service.NoopFileshare{},
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
