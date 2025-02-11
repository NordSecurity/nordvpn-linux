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
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/sharedctx"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
)

func mockTimeout(tries int) time.Duration {
	return time.Millisecond
}

type failingLoginChecker struct{}

func (failingLoginChecker) IsLoggedIn() bool            { return false }
func (failingLoginChecker) IsMFAEnabled() (bool, error) { return false, nil }
func (failingLoginChecker) IsVPNExpired() (bool, error) {
	return true, errors.New("IsVPNExpired error")
}

func (failingLoginChecker) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	return nil, fmt.Errorf("Not implemented")
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
			expectError: true,
		},
		{
			name:        "failing servers API",
			setup:       func(rpc *RPC) { rpc.serversAPI = &mockFailingServersAPI{} },
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

type meshRenewChecker struct{}

func (meshRenewChecker) IsRegistrationInfoCorrect() bool   { return true }
func (meshRenewChecker) Register() error                   { return nil }
func (meshRenewChecker) GetMeshPrivateKey() (string, bool) { return "", true }
func (meshRenewChecker) ClearMeshPrivateKey()              {}

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
			serversAPI:  &mockFailingServersAPI{},
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
