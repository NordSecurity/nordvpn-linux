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

type workingNetworker struct {
	allowedIncoming  []UniqueAddress
	blockedIncoming  []UniqueAddress
	allowedFileshare []UniqueAddress
	blockedFileshare []UniqueAddress
}

func (workingNetworker) Start(
	vpn.Credentials,
	vpn.ServerData,
	config.Whitelist,
	config.DNS,
) error {
	return nil
}

func (*workingNetworker) Stop() error                                       { return nil }
func (*workingNetworker) SetMesh(mesh.MachineMap, netip.Addr, string) error { return nil }
func (*workingNetworker) UnSetMesh() error                                  { return nil }

func (n *workingNetworker) AllowFileshare(address UniqueAddress) error {
	n.allowedFileshare = append(n.allowedFileshare, address)
	return nil
}

func (n *workingNetworker) AllowIncoming(address UniqueAddress) error {
	n.allowedIncoming = append(n.allowedIncoming, address)
	return nil
}

func (n *workingNetworker) BlockIncoming(address UniqueAddress) error {
	n.blockedIncoming = append(n.blockedIncoming, address)
	return nil
}

func (n *workingNetworker) BlockFileshare(address UniqueAddress) error {
	n.blockedFileshare = append(n.blockedFileshare, address)
	return nil
}

func (*workingNetworker) ResetRouting(mesh.MachinePeers) error { return nil }
func (*workingNetworker) BlockRouting(UniqueAddress) error     { return nil }
func (*workingNetworker) Refresh(mesh.MachineMap) error        { return nil }
func (*workingNetworker) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}

type memory struct {
	cfg     *config.Config
	saveErr error
}

func newMemory() *memory {
	return &memory{}
}

func (m *memory) SaveWith(fn config.SaveFunc) error {
	if m.saveErr != nil {
		return m.saveErr
	}

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

type failingFileshare struct{ Fileshare }

func (failingFileshare) Enable(uint32, uint32) error  { return fmt.Errorf("error") }
func (failingFileshare) Disable(uint32, uint32) error { return fmt.Errorf("error") }

func newMockedServer(
	t *testing.T,
	listErr error,
	saveConfigErr error,
	configureErr error,
	isMeshOn bool,
	peers []mesh.MachinePeer) (*Server, *workingNetworker) {
	t.Helper()

	registryApi := registryAPI{}
	registryApi.machinePeers = peers
	registryApi.listErr = listErr
	registryApi.configureErr = configureErr

	networker := workingNetworker{}
	networker.allowedFileshare = []UniqueAddress{}

	configManager := newMemory()
	configManager.saveErr = saveConfigErr

	server := NewServer(
		meshRenewChecker{},
		configManager,
		registrationChecker{},
		acceptInvitationsAPI{},
		&networker,
		&registryApi,
		dnsGetter{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		&subs.Subject[bool]{},
		mock.Fileshare{},
	)

	if isMeshOn {
		server.EnableMeshnet(context.Background(), &pb.Empty{})
	}

	return server, &networker
}

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
			netw:      &workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       &registryAPI{},
			cm:        newMemory(),
			dns:       dnsGetter{},
			fileshare: mock.Fileshare{},
			success:   true,
		},
		{
			name:      "fileshare fails",
			netw:      &workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       &registryAPI{},
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
				&subs.Subject[error]{},
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
			netw:      &workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       &registryAPI{},
			cm:        newMemory(),
			dns:       dnsGetter{},
			fileshare: mock.Fileshare{},
		},
		{
			name:      "fileshare fails",
			netw:      &workingNetworker{},
			ac:        meshRenewChecker{},
			inv:       invitationsAPI{},
			rc:        registrationChecker{},
			reg:       &registryAPI{},
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
				&subs.Subject[error]{},
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
				&workingNetworker{},
				&registryAPI{},
				dnsGetter{},
				&subs.Subject[error]{},
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
		&workingNetworker{},
		&registryAPI{},
		dnsGetter{},
		&subs.Subject[error]{},
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

func TestServer_GetPeersIPHandling(t *testing.T) {
	registryApi := registryAPI{}

	server := NewServer(
		meshRenewChecker{},
		newMemory(),
		registrationChecker{},
		acceptInvitationsAPI{},
		&workingNetworker{},
		&registryApi,
		dnsGetter{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		&subs.Subject[bool]{},
		mock.Fileshare{},
	)
	server.EnableMeshnet(context.Background(), &pb.Empty{})

	localPeerIP := "172.17.0.1"
	externalPeerIP := "192.17.30.5"

	localPeer := mesh.MachinePeer{
		IsLocal:   true,
		Hostname:  "test0-everest.nord",
		PublicKey: "sfB1pvE4RavTwF6oAQlNbJpVp2RqEmEcB6YyoD4tYWG=",
		Address:   netip.MustParseAddr(localPeerIP),
	}
	externalPeer := mesh.MachinePeer{
		IsLocal:   false,
		Hostname:  "test1-everest.nord",
		PublicKey: "sfB1pvE4RavTwF6oAQlNbJpVp2RqEmEcB6YyoD4tYWG=",
		Address:   netip.MustParseAddr(externalPeerIP),
	}

	localPeerNoIP := localPeer
	localPeerNoIP.Address = netip.Addr{}

	externalPeerNoIP := externalPeer
	externalPeerNoIP.Address = netip.Addr{}

	tests := []struct {
		name                   string
		peers                  mesh.MachinePeers
		expectedLocalPeerIP    string
		expectedExternalPeerIP string
	}{
		{
			name:                   "both peers have assigned IP",
			peers:                  mesh.MachinePeers{localPeer, externalPeer},
			expectedLocalPeerIP:    localPeerIP,
			expectedExternalPeerIP: externalPeerIP,
		},
		{
			name:                   "both peers do not have assigned IP",
			peers:                  mesh.MachinePeers{localPeerNoIP, externalPeerNoIP},
			expectedLocalPeerIP:    "",
			expectedExternalPeerIP: "",
		},
	}

	for _, test := range tests {
		registryApi.machinePeers = test.peers

		resp, _ := server.GetPeers(context.Background(), &pb.Empty{})

		t.Run(test.name, func(t *testing.T) {
			assert.IsType(t, &pb.GetPeersResponse_Peers{}, resp.Response)
			assert.Equal(t, 1, len(resp.GetPeers().Local))
			assert.Equal(t, test.expectedLocalPeerIP, resp.GetPeers().Local[0].GetIp())
			assert.Equal(t, 1, len(resp.GetPeers().External))
			assert.Equal(t, test.expectedExternalPeerIP, resp.GetPeers().External[0].GetIp())
		})
	}
}

func TestServer_Connect(t *testing.T) {
	peerValidUuid := "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
	peerNoIpUuid := "c4a11926-e404-11ed-b5ea-0242ac120002"
	peerNoRoutingUuid := "cb5a8446-e404-11ed-b5ea-0242ac120002"

	getServer := func() *Server {
		registryApi := registryAPI{}
		configManager := newMemory()
		configManager.cfg = &config.Config{Technology: config.Technology_NORDLYNX}

		registryApi.machinePeers = []mesh.MachinePeer{
			{
				ID:                   uuid.MustParse(peerValidUuid),
				DoesPeerAllowRouting: true,
				Address:              netip.MustParseAddr("220.16.61.136"),
			},
			{
				ID:                   uuid.MustParse(peerNoIpUuid),
				DoesPeerAllowRouting: true,
				Address:              netip.Addr{},
			},
			{
				ID:                   uuid.MustParse(peerNoRoutingUuid),
				DoesPeerAllowRouting: false,
				Address:              netip.MustParseAddr("87.169.173.253"),
			},
		}

		server := NewServer(
			meshRenewChecker{},
			configManager,
			registrationChecker{},
			acceptInvitationsAPI{},
			&workingNetworker{},
			&registryApi,
			dnsGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&subs.Subject[bool]{},
			mock.Fileshare{},
		)
		server.EnableMeshnet(context.Background(), &pb.Empty{})
		return server
	}

	tests := []struct {
		name             string
		peerUuid         string
		expectedResponse *pb.ConnectResponse
	}{
		{
			name:             "connect to valid peer",
			peerUuid:         peerValidUuid,
			expectedResponse: &pb.ConnectResponse{Response: &pb.ConnectResponse_Empty{}},
		},
		{
			name:     "connect to peer with no ip",
			peerUuid: peerNoIpUuid,
			expectedResponse: &pb.ConnectResponse{Response: &pb.ConnectResponse_ConnectErrorCode{
				ConnectErrorCode: pb.ConnectErrorCode_PEER_NO_IP,
			}},
		},
		{
			name:     "peer forbids traffic routing",
			peerUuid: peerNoRoutingUuid,
			expectedResponse: &pb.ConnectResponse{
				Response: &pb.ConnectResponse_ConnectErrorCode{
					ConnectErrorCode: pb.ConnectErrorCode_PEER_DOES_NOT_ALLOW_ROUTING,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := getServer()
			resp, err := server.Connect(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_AcceptIncoming(t *testing.T) {
	peerValidUuid := "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
	peerNoIpUuid := "c4a11926-e404-11ed-b5ea-0242ac120002"
	peerIncomingAlreadyAllowedUuid := "cb5a8446-e404-11ed-b5ea-0242ac120002"

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyAllowedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t="
	peerIncomingAlreadyAllowedPublicKey := "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp="

	getServer := func() (*Server, *workingNetworker) {
		registryApi := registryAPI{}
		registryApi.machinePeers = []mesh.MachinePeer{
			{
				ID:              uuid.MustParse(peerValidUuid),
				DoIAllowInbound: false,
				Address:         peerValidAddress,
				PublicKey:       peerValidPublicKey,
			},
			{
				ID:              uuid.MustParse(peerNoIpUuid),
				DoIAllowInbound: false,
				Address:         netip.Addr{},
			},
			{
				ID:              uuid.MustParse(peerIncomingAlreadyAllowedUuid),
				DoIAllowInbound: true,
				Address:         peerIncomingAlreadyAllowedAddress,
				PublicKey:       peerIncomingAlreadyAllowedPublicKey,
			},
		}

		networker := workingNetworker{}
		networker.allowedIncoming = []UniqueAddress{}

		server := NewServer(
			meshRenewChecker{},
			newMemory(),
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			dnsGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&subs.Subject[bool]{},
			mock.Fileshare{},
		)
		server.EnableMeshnet(context.Background(), &pb.Empty{})
		return server, &networker
	}

	tests := []struct {
		name               string
		peerUuid           string
		expectedResponse   *pb.AllowIncomingResponse
		expectedAllowedIPs []UniqueAddress
	}{
		{
			name:               "allow valid peer",
			peerUuid:           peerValidUuid,
			expectedResponse:   &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
			expectedAllowedIPs: []UniqueAddress{{UID: peerValidPublicKey, Address: peerValidAddress}},
		},
		{
			name:               "allow peer with no ip",
			peerUuid:           peerNoIpUuid,
			expectedResponse:   &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
			expectedAllowedIPs: []UniqueAddress{},
		},
		{
			name:     "peer traffic routing already allowed",
			peerUuid: peerIncomingAlreadyAllowedUuid,
			expectedResponse: &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_AllowIncomingErrorCode{
					AllowIncomingErrorCode: pb.AllowIncomingErrorCode_INCOMING_ALREADY_ALLOWED,
				},
			},
			expectedAllowedIPs: []UniqueAddress{},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
			expectedAllowedIPs: []UniqueAddress{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, networker := getServer()
			resp, err := server.AllowIncoming(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedAllowedIPs, networker.allowedIncoming)
		})
	}
}

func TestServer_DenyIncoming(t *testing.T) {
	peerValidUuid := "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
	peerNoIpUuid := "c4a11926-e404-11ed-b5ea-0242ac120002"
	peerIncomingAlreadyDeniedUuid := "cb5a8446-e404-11ed-b5ea-0242ac120002"

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t="
	peerIncomingAlreadyDeniedPublicKey := "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp="

	getServer := func() (*Server, *workingNetworker) {
		registryApi := registryAPI{}
		registryApi.machinePeers = []mesh.MachinePeer{
			{
				ID:              uuid.MustParse(peerValidUuid),
				DoIAllowInbound: true,
				Address:         peerValidAddress,
				PublicKey:       peerValidPublicKey,
			},
			{
				ID:              uuid.MustParse(peerNoIpUuid),
				DoIAllowInbound: true,
				Address:         netip.Addr{},
			},
			{
				ID:              uuid.MustParse(peerIncomingAlreadyDeniedUuid),
				DoIAllowInbound: false,
				Address:         peerIncomingAlreadyDeniedAddress,
				PublicKey:       peerIncomingAlreadyDeniedPublicKey,
			},
		}

		networker := workingNetworker{}
		networker.blockedIncoming = []UniqueAddress{}

		server := NewServer(
			meshRenewChecker{},
			newMemory(),
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			dnsGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&subs.Subject[bool]{},
			mock.Fileshare{},
		)
		server.EnableMeshnet(context.Background(), &pb.Empty{})
		return server, &networker
	}

	tests := []struct {
		name               string
		peerUuid           string
		expectedResponse   *pb.DenyIncomingResponse
		expectedBlockedIPs []UniqueAddress
	}{
		{
			name:               "deny valid peer",
			peerUuid:           peerValidUuid,
			expectedResponse:   &pb.DenyIncomingResponse{Response: &pb.DenyIncomingResponse_Empty{}},
			expectedBlockedIPs: []UniqueAddress{{UID: peerValidPublicKey, Address: peerValidAddress}},
		},
		{
			name:               "connect to peer with no ip",
			peerUuid:           peerNoIpUuid,
			expectedResponse:   &pb.DenyIncomingResponse{Response: &pb.DenyIncomingResponse_Empty{}},
			expectedBlockedIPs: []UniqueAddress{},
		},
		{
			name:     "peer traffic routing already denied",
			peerUuid: peerIncomingAlreadyDeniedUuid,
			expectedResponse: &pb.DenyIncomingResponse{
				Response: &pb.DenyIncomingResponse_DenyIncomingErrorCode{
					DenyIncomingErrorCode: pb.DenyIncomingErrorCode_INCOMING_ALREADY_DENIED,
				},
			},
			expectedBlockedIPs: []UniqueAddress{},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.DenyIncomingResponse{
				Response: &pb.DenyIncomingResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
			expectedBlockedIPs: []UniqueAddress{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, networker := getServer()
			resp, err := server.DenyIncoming(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedBlockedIPs, networker.blockedIncoming)
		})
	}
}

func TestServer_AllowFileshare(t *testing.T) {
	peerValidUuid := "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
	peerNoIpUuid := "c4a11926-e404-11ed-b5ea-0242ac120002"
	peerIncomingAlreadyDeniedUuid := "cb5a8446-e404-11ed-b5ea-0242ac120002"

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t="
	peerIncomingAlreadyDeniedPublicKey := "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp="

	getServer := func() (*Server, *workingNetworker) {
		registryApi := registryAPI{}
		registryApi.machinePeers = []mesh.MachinePeer{
			{
				ID:                uuid.MustParse(peerValidUuid),
				DoIAllowFileshare: false,
				Address:           peerValidAddress,
				PublicKey:         peerValidPublicKey,
			},
			{
				ID:                uuid.MustParse(peerNoIpUuid),
				DoIAllowFileshare: false,
				Address:           netip.Addr{},
			},
			{
				ID:                uuid.MustParse(peerIncomingAlreadyDeniedUuid),
				DoIAllowFileshare: true,
				Address:           peerIncomingAlreadyDeniedAddress,
				PublicKey:         peerIncomingAlreadyDeniedPublicKey,
			},
		}

		networker := workingNetworker{}
		networker.allowedFileshare = []UniqueAddress{}

		server := NewServer(
			meshRenewChecker{},
			newMemory(),
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			dnsGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&subs.Subject[bool]{},
			mock.Fileshare{},
		)
		server.EnableMeshnet(context.Background(), &pb.Empty{})
		return server, &networker
	}

	tests := []struct {
		name               string
		peerUuid           string
		expectedResponse   *pb.AllowFileshareResponse
		expectedAllowedIPs []UniqueAddress
	}{
		{
			name:               "allow valid peer",
			peerUuid:           peerValidUuid,
			expectedResponse:   &pb.AllowFileshareResponse{Response: &pb.AllowFileshareResponse_Empty{}},
			expectedAllowedIPs: []UniqueAddress{{UID: peerValidPublicKey, Address: peerValidAddress}},
		},
		{
			name:               "allow fileshare to peer with no ip",
			peerUuid:           peerNoIpUuid,
			expectedResponse:   &pb.AllowFileshareResponse{Response: &pb.AllowFileshareResponse_Empty{}},
			expectedAllowedIPs: []UniqueAddress{},
		},
		{
			name:     "fileshare already denied",
			peerUuid: peerIncomingAlreadyDeniedUuid,
			expectedResponse: &pb.AllowFileshareResponse{
				Response: &pb.AllowFileshareResponse_AllowSendErrorCode{
					AllowSendErrorCode: pb.AllowFileshareErrorCode_SEND_ALREADY_ALLOWED,
				},
			},
			expectedAllowedIPs: []UniqueAddress{},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.AllowFileshareResponse{
				Response: &pb.AllowFileshareResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
			expectedAllowedIPs: []UniqueAddress{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, networker := getServer()
			resp, err := server.AllowFileshare(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedAllowedIPs, networker.allowedFileshare)
		})
	}
}

func TestServer_DenyFileshare(t *testing.T) {
	peerValidUuid := "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
	peerNoIpUuid := "c4a11926-e404-11ed-b5ea-0242ac120002"
	peerIncomingAlreadyDeniedUuid := "cb5a8446-e404-11ed-b5ea-0242ac120002"

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t="
	peerIncomingAlreadyDeniedPublicKey := "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp="

	getServer := func() (*Server, *workingNetworker) {
		registryApi := registryAPI{}
		registryApi.machinePeers = []mesh.MachinePeer{
			{
				ID:                uuid.MustParse(peerValidUuid),
				DoIAllowFileshare: true,
				Address:           peerValidAddress,
				PublicKey:         peerValidPublicKey,
			},
			{
				ID:                uuid.MustParse(peerNoIpUuid),
				DoIAllowFileshare: true,
				Address:           netip.Addr{},
			},
			{
				ID:                uuid.MustParse(peerIncomingAlreadyDeniedUuid),
				DoIAllowFileshare: false,
				Address:           peerIncomingAlreadyDeniedAddress,
				PublicKey:         peerIncomingAlreadyDeniedPublicKey,
			},
		}

		networker := workingNetworker{}
		networker.blockedFileshare = []UniqueAddress{}

		server := NewServer(
			meshRenewChecker{},
			newMemory(),
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			dnsGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&subs.Subject[bool]{},
			mock.Fileshare{},
		)
		server.EnableMeshnet(context.Background(), &pb.Empty{})
		return server, &networker
	}

	tests := []struct {
		name               string
		peerUuid           string
		expectedResponse   *pb.DenyFileshareResponse
		expectedAllowedIPs []UniqueAddress
	}{
		{
			name:               "deny valid peer",
			peerUuid:           peerValidUuid,
			expectedResponse:   &pb.DenyFileshareResponse{Response: &pb.DenyFileshareResponse_Empty{}},
			expectedAllowedIPs: []UniqueAddress{{UID: peerValidPublicKey, Address: peerValidAddress}},
		},
		{
			name:               "allow fileshare to peer with no ip",
			peerUuid:           peerNoIpUuid,
			expectedResponse:   &pb.DenyFileshareResponse{Response: &pb.DenyFileshareResponse_Empty{}},
			expectedAllowedIPs: []UniqueAddress{},
		},
		{
			name:     "fileshare already denied",
			peerUuid: peerIncomingAlreadyDeniedUuid,
			expectedResponse: &pb.DenyFileshareResponse{
				Response: &pb.DenyFileshareResponse_DenySendErrorCode{
					DenySendErrorCode: pb.DenyFileshareErrorCode_SEND_ALREADY_DENIED,
				},
			},
			expectedAllowedIPs: []UniqueAddress{},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.DenyFileshareResponse{
				Response: &pb.DenyFileshareResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
			expectedAllowedIPs: []UniqueAddress{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, networker := getServer()
			resp, err := server.DenyFileshare(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedAllowedIPs, networker.blockedFileshare)
		})
	}
}

func TestServer_EnableAutomaticFileshare(t *testing.T) {
	peerValidUuid := "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
	peerAleradyEnabledUuid := "cb5a8446-e404-11ed-b5ea-0242ac120002"

	peers := []mesh.MachinePeer{
		{
			ID:                uuid.MustParse(peerValidUuid),
			DoIAllowFileshare: false,
		},
		{
			ID:                uuid.MustParse(peerAleradyEnabledUuid),
			DoIAllowFileshare: true,
			AlwaysAcceptFiles: true,
		},
	}

	tests := []struct {
		name             string
		peerUuid         string
		listErr          error
		configureErr     error
		saveConfigErr    error
		isMeshOn         bool
		expectedResponse *pb.EnableAutomaticFileshareResponse
	}{
		{
			name:             "enable for valid peer",
			peerUuid:         peerValidUuid,
			isMeshOn:         true,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{Response: &pb.EnableAutomaticFileshareResponse_Empty{}},
		},
		{
			name:     "automatic fileshare already enabled",
			peerUuid: peerAleradyEnabledUuid,
			isMeshOn: true,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_EnableAutomaticFileshareErrorCode{
					EnableAutomaticFileshareErrorCode: pb.EnableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_ENABLED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			isMeshOn: true,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
		},
		{
			name:     "not authorized to list peers",
			peerUuid: peerValidUuid,
			listErr:  core.ErrUnauthorized,
			isMeshOn: true,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			},
		},
		{
			name:     "meshnet is not on",
			peerUuid: peerValidUuid,
			listErr:  fmt.Errorf("generic error"),
			isMeshOn: false,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			},
		},
		{
			name:     "unknown error",
			peerUuid: peerValidUuid,
			listErr:  fmt.Errorf("generic error"),
			isMeshOn: true,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
		{
			name:          "failed to save config",
			peerUuid:      peerValidUuid,
			listErr:       core.ErrUnauthorized,
			saveConfigErr: fmt.Errorf("generic error"),
			isMeshOn:      true,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			},
		},
		{
			name:         "failed to configure peer",
			peerUuid:     peerValidUuid,
			configureErr: fmt.Errorf("generic error"),
			isMeshOn:     true,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, _ := newMockedServer(t,
				test.listErr,
				test.saveConfigErr,
				test.configureErr,
				test.isMeshOn,
				peers)
			resp, err := server.EnableAutomaticFileshare(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_DisableAutomaticFileshare(t *testing.T) {
	peerValidUuid := "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
	peerAleradyDisabledUuid := "cb5a8446-e404-11ed-b5ea-0242ac120002"

	peers := []mesh.MachinePeer{
		{
			ID:                uuid.MustParse(peerValidUuid),
			DoIAllowFileshare: false,
			AlwaysAcceptFiles: true,
		},
		{
			ID:                uuid.MustParse(peerAleradyDisabledUuid),
			DoIAllowFileshare: true,
		},
	}

	tests := []struct {
		name             string
		peerUuid         string
		isMeshOn         bool
		listErr          error
		saveConfigErr    error
		configureErr     error
		expectedResponse *pb.DisableAutomaticFileshareResponse
	}{
		{
			name:             "enable for valid peer",
			peerUuid:         peerValidUuid,
			isMeshOn:         true,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{Response: &pb.DisableAutomaticFileshareResponse_Empty{}},
		},
		{
			name:     "automatic fileshare already enabled",
			peerUuid: peerAleradyDisabledUuid,
			isMeshOn: true,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_DisableAutomaticFileshareErrorCode{
					DisableAutomaticFileshareErrorCode: pb.DisableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_DISABLED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			isMeshOn: true,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
		},
		{
			name:     "not authorized to list peers",
			peerUuid: peerValidUuid,
			listErr:  core.ErrUnauthorized,
			isMeshOn: true,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			},
		},
		{
			name:     "meshnet is not on",
			peerUuid: peerValidUuid,
			listErr:  fmt.Errorf("generic error"),
			isMeshOn: false,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			},
		},
		{
			name:     "unknown error",
			peerUuid: peerValidUuid,
			listErr:  fmt.Errorf("generic error"),
			isMeshOn: true,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
		{
			name:          "failed to save config",
			peerUuid:      peerAleradyDisabledUuid,
			listErr:       core.ErrUnauthorized,
			saveConfigErr: fmt.Errorf("generic error"),
			isMeshOn:      true,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			},
		},
		{
			name:         "failed to configure peer",
			peerUuid:     peerValidUuid,
			configureErr: fmt.Errorf("generic error"),
			isMeshOn:     true,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, _ := newMockedServer(t,
				test.listErr,
				test.saveConfigErr,
				test.configureErr,
				test.isMeshOn,
				peers)
			resp, err := server.DisableAutomaticFileshare(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}
