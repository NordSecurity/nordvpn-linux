package meshnet

import (
	"context"
	"fmt"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/mock"
	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"
)

type meshRenewChecker struct{}

func (meshRenewChecker) IsLoggedIn() bool { return true }

type registrationChecker struct{}

func (registrationChecker) IsRegistered() bool { return true }

type workingNetworker struct{}

func (workingNetworker) Start(
	vpn.Credentials,
	vpn.ServerData,
	config.Whitelist,
	config.DNS,
) error {
	return nil
}

func (workingNetworker) Stop() error                                       { return nil }
func (workingNetworker) SetMesh(mesh.MachineMap, netip.Addr, string) error { return nil }
func (workingNetworker) UnSetMesh() error                                  { return nil }
func (workingNetworker) AllowIncoming(UniqueAddress) error                 { return nil }
func (workingNetworker) BlockIncoming(UniqueAddress) error                 { return nil }
func (workingNetworker) ResetRouting(mesh.MachinePeers) error              { return nil }
func (workingNetworker) BlockRouting(UniqueAddress) error                  { return nil }
func (workingNetworker) Refresh(mesh.MachineMap) error                     { return nil }
func (workingNetworker) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}

type memory struct {
	cfg *config.Config
}

func newMemory() *memory {
	return &memory{}
}

func (m *memory) SaveWith(fn config.SaveFunc) error {
	if m.cfg == nil {
		m.cfg = &config.Config{}
	}
	cfg := fn(*m.cfg)
	*m.cfg = *&cfg
	return nil
}

func (m *memory) Load(c *config.Config) error {
	if m.cfg == nil {
		m.cfg = &config.Config{}
	}
	if m.cfg.MeshDevice == nil {
		m.cfg.MeshDevice = &mesh.Machine{}
	}
	*c = *m.cfg
	return nil
}

func (m *memory) Reset() error {
	*m = *newMemory()
	return nil
}

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

type limitedInvitationsAPI struct {
	invitationsAPI
}

func (limitedInvitationsAPI) Invite(string, uuid.UUID, string, bool, bool, bool, bool) error {
	return core.ErrTooManyRequests
}

type maximumInvitationsAPI struct {
	invitationsAPI
}

func (maximumInvitationsAPI) Invite(string, uuid.UUID, string, bool, bool, bool, bool) error {
	return core.ErrMaximumDeviceCount
}

type acceptInvitationsAPI struct {
	invitationsAPI
}

func (acceptInvitationsAPI) Accept(string, uuid.UUID, uuid.UUID, bool, bool, bool, bool) error {
	return core.ErrMaximumDeviceCount
}

func (acceptInvitationsAPI) Received(string, uuid.UUID) (mesh.Invitations, error) {
	return mesh.Invitations{
		mesh.Invitation{Email: "inviter@nordvpn.com"},
	}, nil
}

type registryAPI struct{}

func (registryAPI) Register(string, mesh.Machine) (*mesh.Machine, error) {
	return &mesh.Machine{}, nil
}

func (registryAPI) Update(string, uuid.UUID, []netip.AddrPort) error                     { return nil }
func (registryAPI) Configure(string, uuid.UUID, uuid.UUID, bool, bool, bool, bool) error { return nil }
func (registryAPI) Unregister(string, uuid.UUID) error                                   { return nil }
func (registryAPI) List(string, uuid.UUID) (mesh.MachinePeers, error) {
	return mesh.MachinePeers{}, nil
}
func (registryAPI) Local(string) (mesh.Machines, error)       { return mesh.Machines{}, nil }
func (registryAPI) Unpair(string, uuid.UUID, uuid.UUID) error { return nil }

func (registryAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	return &mesh.MachineMap{}, nil
}
func (registryAPI) NotifyNewTransfer(
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

type failingFileshare struct{ Fileshare }

func (failingFileshare) Enable(uint32, uint32) error  { return fmt.Errorf("error") }
func (failingFileshare) Disable(uint32, uint32) error { return fmt.Errorf("error") }

func TestServer_EnableMeshnet(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		netw      Networker
		ac        auth.Checker
		inv       mesh.Inviter
		rc        Checker
		reg       mesh.Registry
		cm        config.Manager
		dns       dns.Getter
		fileshare Fileshare
		success   bool
	}{
		{
			name:      "everything works",
			netw:      workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       registryAPI{},
			cm:        newMemory(),
			dns:       dnsGetter{},
			fileshare: mock.Fileshare{},
			success:   true,
		},
		{
			name:      "fileshare fails",
			netw:      workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       registryAPI{},
			cm:        newMemory(),
			dns:       dnsGetter{},
			fileshare: failingFileshare{},
			success:   true, // Fileshare shouldn't impact meshnet enabling
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Check server creation
			mserver := NewServer(
				test.ac,
				test.cm,
				test.rc,
				test.inv,
				test.netw,
				test.reg,
				test.dns,
				&subs.Subject[string]{},
				&subs.Subject[[]string]{},
				&subs.Subject[bool]{},
				test.fileshare,
			)
			assert.NotEqual(t, nil, mserver)
			assert.Equal(t, test.cm, mserver.cm)
			assert.Equal(t, test.netw, mserver.netw)

			//Check server configuration
			var cfg config.Config
			err := mserver.cm.Load(&cfg)
			assert.NoError(t, err)
			assert.False(t, cfg.Mesh)

			//Enable Mesh
			peerCtx := peer.NewContext(context.Background(), &peer.Peer{AuthInfo: internal.UcredAuth{}})
			resp, err := mserver.EnableMeshnet(peerCtx, &pb.Empty{})
			assert.NoError(t, err)
			_, ok := resp.GetResponse().(*pb.MeshnetResponse_Empty)
			assert.Equal(t, test.success, ok)

			//Check new server configuration
			err = mserver.cm.Load(&cfg)
			assert.NoError(t, err)
			assert.Equal(t, test.success, cfg.Mesh)
		})
	}
}

func TestServer_DisableMeshnet(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		netw      Networker
		ac        auth.Checker
		inv       mesh.Inviter
		rc        Checker
		reg       mesh.Registry
		cm        config.Manager
		dns       dns.Getter
		fileshare Fileshare
	}{
		{
			name:      "everything works",
			netw:      workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       registryAPI{},
			cm:        newMemory(),
			dns:       dnsGetter{},
			fileshare: mock.Fileshare{},
		},
		{
			name:      "fileshare fails",
			netw:      workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       registryAPI{},
			cm:        newMemory(),
			dns:       dnsGetter{},
			fileshare: failingFileshare{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Check server creation
			mserver := NewServer(
				test.ac,
				test.cm,
				test.rc,
				test.inv,
				test.netw,
				test.reg,
				test.dns,
				&subs.Subject[string]{},
				&subs.Subject[[]string]{},
				&subs.Subject[bool]{},
				test.fileshare,
			)
			assert.NotEqual(t, nil, mserver)
			assert.Equal(t, test.cm, mserver.cm)
			assert.Equal(t, test.netw, mserver.netw)

			//Set server configuration
			var cfg config.Config
			mserver.cm.SaveWith(func(c config.Config) config.Config { c.Mesh = true; return c })

			//Disable Mesh
			resp, err := mserver.DisableMeshnet(context.Background(), &pb.Empty{})
			assert.NoError(t, err)
			_, ok := resp.GetResponse().(*pb.MeshnetResponse_Empty)
			assert.Equal(t, true, ok)

			//Check new server configuration
			err = mserver.cm.Load(&cfg)
			assert.NoError(t, err)
			assert.Equal(t, false, cfg.Mesh)
		})
	}
}

func TestServer_Invite(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		inv      mesh.Inviter
		expected pb.InviteResponseErrorCode
	}{
		{
			name:     "invitation limit",
			inv:      limitedInvitationsAPI{},
			expected: pb.InviteResponseErrorCode_LIMIT_REACHED,
		},
		{
			name:     "invited device count",
			inv:      maximumInvitationsAPI{},
			expected: pb.InviteResponseErrorCode_PEER_COUNT,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(
				meshRenewChecker{},
				newMemory(),
				registrationChecker{},
				test.inv,
				workingNetworker{},
				registryAPI{},
				dnsGetter{},
				&subs.Subject[string]{},
				&subs.Subject[[]string]{},
				&subs.Subject[bool]{},
				mock.Fileshare{},
			)
			server.EnableMeshnet(context.Background(), &pb.Empty{})
			resp, err := server.Invite(context.Background(), &pb.InviteRequest{})
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t,
				test.expected,
				resp.Response.(*pb.InviteResponse_InviteResponseErrorCode).InviteResponseErrorCode,
			)
		})
	}
}

func TestServer_AcceptInvite(t *testing.T) {
	category.Set(t, category.Unit)

	server := NewServer(
		meshRenewChecker{},
		newMemory(),
		registrationChecker{},
		acceptInvitationsAPI{},
		workingNetworker{},
		registryAPI{},
		dnsGetter{},
		&subs.Subject[string]{},
		&subs.Subject[[]string]{},
		&subs.Subject[bool]{},
		mock.Fileshare{},
	)
	server.EnableMeshnet(context.Background(), &pb.Empty{})
	resp, err := server.AcceptInvite(context.Background(), &pb.InviteRequest{
		Email: "inviter@nordvpn.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t,
		pb.RespondToInviteErrorCode_DEVICE_COUNT,
		resp.Response.(*pb.RespondToInviteResponse_RespondToInviteErrorCode).RespondToInviteErrorCode,
	)
}
