package meshnet

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/sharedctx"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"
)

const (
	examplePublicKey1 = "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t="
	examplePublicKey2 = "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp="
	exampleUUID1      = "cb5a8446-e404-11ed-b5ea-0242ac120002"
	exampleUUID2      = "c4a11926-e404-11ed-b5ea-0242ac120002"
	exampleUUID3      = "a7e4e7d6-e404-11ed-b5ea-0242ac120002"
)

type meshRenewChecker struct {
	IsNotLoggedIn bool // by default is logged in
}

func (m meshRenewChecker) IsLoggedIn() bool {
	return !m.IsNotLoggedIn
}
func (m meshRenewChecker) IsMFAEnabled() (bool, error) {
	return false, nil
}
func (meshRenewChecker) IsVPNExpired() (bool, error) { return false, nil }
func (meshRenewChecker) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	return nil, fmt.Errorf("Not implemented")
}

type registrationChecker struct {
	registrationErr error
}

func (r registrationChecker) IsRegistrationInfoCorrect() bool { return r.registrationErr == nil }
func (r registrationChecker) Register() error                 { return r.registrationErr }

type allowedIncoming struct {
	address    UniqueAddress
	lanAllowed bool
}

type workingNetworker struct {
	allowedIncoming  []allowedIncoming
	blockedIncoming  []UniqueAddress
	allowedFileshare []UniqueAddress
	blockedFileshare []UniqueAddress
	resetPeers       []string
}

func (workingNetworker) Start(
	context.Context,
	vpn.Credentials,
	vpn.ServerData,
	config.Allowlist,
	config.DNS,
	bool,
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

func (n *workingNetworker) AllowIncoming(address UniqueAddress, lanAllowed bool) error {
	n.allowedIncoming = append(n.allowedIncoming, allowedIncoming{
		address:    address,
		lanAllowed: lanAllowed,
	})

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

func (n *workingNetworker) ResetRouting(changedPeer mesh.MachinePeer, peer mesh.MachinePeers) error {
	n.resetPeers = append(n.resetPeers, changedPeer.PublicKey)

	return nil
}

func (*workingNetworker) BlockRouting(UniqueAddress) error { return nil }
func (*workingNetworker) Refresh(mesh.MachineMap) error    { return nil }
func (*workingNetworker) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}
func (*workingNetworker) LastServerName() string { return "" }

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

func newMockedServer(
	t *testing.T,
	listErr error,
	saveConfigErr error,
	configureErr error,
	isMeshOn bool,
	peers []mesh.MachinePeer,
) *Server {
	t.Helper()

	registryApi := mock.RegistryMock{}
	registryApi.Peers = peers
	registryApi.ListErr = listErr
	registryApi.ConfigureErr = configureErr

	configManager := &mock.ConfigManager{}
	configManager.SaveErr = saveConfigErr

	server := NewServer(
		meshRenewChecker{},
		configManager,
		registrationChecker{},
		acceptInvitationsAPI{},
		&workingNetworker{},
		&registryApi,
		&mock.DNSGetter{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
		testnorduser.NewMockNorduserClient(nil),
		sharedctx.New(),
	)

	if isMeshOn {
		server.EnableMeshnet(context.Background(), &pb.Empty{})
	}

	return server
}

func TestServer_EnableMeshnet(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name                string
		netw                Networker
		ac                  auth.Checker
		inv                 mesh.Inviter
		rc                  Checker
		reg                 mesh.Registry
		cm                  config.Manager
		dns                 dns.Getter
		startFileshareError error
		success             bool
	}{
		{
			name:                "everything works",
			netw:                &workingNetworker{},
			ac:                  meshRenewChecker{},
			inv:                 invitationsAPI{},
			rc:                  registrationChecker{},
			reg:                 &mock.RegistryMock{},
			cm:                  &mock.ConfigManager{},
			dns:                 &mock.DNSGetter{},
			startFileshareError: nil,
			success:             true,
		},
		{
			name:                "fileshare fails",
			netw:                &workingNetworker{},
			ac:                  meshRenewChecker{},
			inv:                 invitationsAPI{},
			rc:                  registrationChecker{},
			reg:                 &mock.RegistryMock{},
			cm:                  &mock.ConfigManager{},
			dns:                 &mock.DNSGetter{},
			startFileshareError: fmt.Errorf("failed to disable fileshare"),
			success:             true, // Fileshare shouldn't impact meshnet enabling
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
				&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
				testnorduser.NewMockNorduserClient(test.startFileshareError),
				sharedctx.New(),
			)
			assert.NotEqual(t, nil, mserver)
			assert.Equal(t, test.cm, mserver.cm)
			assert.Equal(t, test.netw, mserver.netw)

			// Check server configuration
			var cfg config.Config
			err := mserver.cm.Load(&cfg)
			assert.NoError(t, err)
			assert.False(t, cfg.Mesh)

			// Enable Mesh
			peerCtx := peer.NewContext(context.Background(), &peer.Peer{AuthInfo: internal.UcredAuth{}})
			resp, err := mserver.EnableMeshnet(peerCtx, &pb.Empty{})
			assert.NoError(t, err)
			_, ok := resp.GetResponse().(*pb.MeshnetResponse_Empty)
			assert.Equal(t, test.success, ok)

			// Check new server configuration
			err = mserver.cm.Load(&cfg)
			assert.NoError(t, err)
			assert.Equal(t, test.success, cfg.Mesh)
		})
	}
}

func TestServer_DisableMeshnet(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name                string
		netw                Networker
		ac                  auth.Checker
		inv                 mesh.Inviter
		rc                  Checker
		reg                 mesh.Registry
		cm                  config.Manager
		dns                 dns.Getter
		startFileshareError error
	}{
		{
			name:                "everything works",
			netw:                &workingNetworker{},
			ac:                  meshRenewChecker{},
			inv:                 invitationsAPI{},
			rc:                  registrationChecker{},
			reg:                 &mock.RegistryMock{},
			cm:                  &mock.ConfigManager{},
			dns:                 &mock.DNSGetter{},
			startFileshareError: nil,
		},
		{
			name:                "fileshare fails",
			netw:                &workingNetworker{},
			ac:                  meshRenewChecker{},
			inv:                 invitationsAPI{},
			rc:                  registrationChecker{},
			reg:                 &mock.RegistryMock{},
			cm:                  &mock.ConfigManager{},
			dns:                 &mock.DNSGetter{},
			startFileshareError: fmt.Errorf("failed to start fileshare"),
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
				&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
				testnorduser.NewMockNorduserClient(test.startFileshareError),
				sharedctx.New(),
			)
			assert.NotEqual(t, nil, mserver)
			assert.Equal(t, test.cm, mserver.cm)
			assert.Equal(t, test.netw, mserver.netw)

			// Set server configuration
			var cfg config.Config
			mserver.cm.SaveWith(func(c config.Config) config.Config { c.Mesh = true; return c })

			// Disable Mesh
			resp, err := mserver.DisableMeshnet(context.Background(), &pb.Empty{})
			assert.NoError(t, err)
			_, ok := resp.GetResponse().(*pb.MeshnetResponse_Empty)
			assert.Equal(t, true, ok)

			// Check new server configuration
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
				&mock.ConfigManager{},
				registrationChecker{},
				test.inv,
				&workingNetworker{},
				&mock.RegistryMock{},
				&mock.DNSGetter{},
				&subs.Subject[error]{},
				&subs.Subject[[]string]{},
				&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}},
					Service: &daemonevents.ServiceEvents{UiItemsClick: &daemonevents.MockPublisherSubscriber[events.UiItemsAction]{}}},
				testnorduser.NewMockNorduserClient(nil),
				sharedctx.New(),
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
		&mock.ConfigManager{},
		registrationChecker{},
		acceptInvitationsAPI{},
		&workingNetworker{},
		&mock.RegistryMock{},
		&mock.DNSGetter{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
		testnorduser.NewMockNorduserClient(nil),
		sharedctx.New(),
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
	registryApi := mock.RegistryMock{}

	server := NewServer(
		meshRenewChecker{},
		&mock.ConfigManager{},
		registrationChecker{},
		acceptInvitationsAPI{},
		&workingNetworker{},
		&registryApi,
		&mock.DNSGetter{},
		&subs.Subject[error]{},
		&subs.Subject[[]string]{},
		&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
		testnorduser.NewMockNorduserClient(nil),
		sharedctx.New(),
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
		registryApi.Peers = test.peers

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
	peerValidUuid := exampleUUID3
	peerNoIpUuid := exampleUUID2
	peerNoRoutingUuid := exampleUUID1

	getServer := func() *Server {
		registryApi := mock.RegistryMock{}
		configManager := &mock.ConfigManager{}
		configManager.Cfg = &config.Config{Technology: config.Technology_NORDLYNX}

		registryApi.Peers = []mesh.MachinePeer{
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
			&mock.DNSGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}},
				Service: &daemonevents.ServiceEvents{Connect: &daemonevents.MockPublisherSubscriber[events.DataConnect]{}}},
			testnorduser.NewMockNorduserClient(nil),
			sharedctx.New(),
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
	peerValidUuid := exampleUUID3
	peerNoIpUuid := exampleUUID2
	peerIncomingAlreadyAllowedUuid := exampleUUID1
	peerNoRoutingUuid := "8c9f1e11-4b67-4ba4-a2df-4308757f2d59"
	peerNoLANUuid := "06d3c1ba-997c-4b2c-9d61-0d718becdd89"
	peerLANAndRoutingUuid := "7505abad-527f-442f-b17a-820451ff8e8a"

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyAllowedAddress := netip.MustParseAddr("87.169.173.253")
	peerNoRoutingAddress := netip.MustParseAddr("54.1.218.8")
	peerNoLANAddress := netip.MustParseAddr("18.203.48.39")
	peerLANAndRoutingAddress := netip.MustParseAddr("249.205.110.178")

	peerValidPublicKey := examplePublicKey1
	peerIncomingAlreadyAllowedPublicKey := examplePublicKey2
	peerNoRoutingPublicKey := "ubQBAfx1VXCI2yXqx5oqmcoc5wpBuRxvXRfXXC8qeR="
	peerNoLANAddressPublicKey := "OwJTUXZmqOvXtiC8viXIlezSGe5uEZjTkhVWPyNSnA="
	peerLANAndRoutingPublicKey := "SNoKCfCdi6OKHGI1dRM8QCLwuUMZ5Q2oltlYsLG1kA="

	getServer := func() (*Server, *workingNetworker) {
		registryApi := mock.RegistryMock{}
		registryApi.Peers = []mesh.MachinePeer{
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
			{
				ID:                   uuid.MustParse(peerNoRoutingUuid),
				DoIAllowInbound:      false,
				DoIAllowLocalNetwork: true,
				DoIAllowRouting:      false,
				Address:              peerNoRoutingAddress,
				PublicKey:            peerNoRoutingPublicKey,
			},
			{
				ID:                   uuid.MustParse(peerNoRoutingUuid),
				DoIAllowInbound:      false,
				DoIAllowLocalNetwork: true,
				DoIAllowRouting:      false,
				Address:              peerNoRoutingAddress,
				PublicKey:            peerNoRoutingPublicKey,
			},
			{
				ID:                   uuid.MustParse(peerNoLANUuid),
				DoIAllowInbound:      false,
				DoIAllowLocalNetwork: false,
				DoIAllowRouting:      true,
				Address:              peerNoLANAddress,
				PublicKey:            peerNoLANAddressPublicKey,
			},
			{
				ID:                   uuid.MustParse(peerLANAndRoutingUuid),
				DoIAllowInbound:      false,
				DoIAllowLocalNetwork: true,
				DoIAllowRouting:      true,
				Address:              peerLANAndRoutingAddress,
				PublicKey:            peerLANAndRoutingPublicKey,
			},
		}

		networker := workingNetworker{}
		networker.allowedIncoming = []allowedIncoming{}

		server := NewServer(
			meshRenewChecker{},
			&mock.ConfigManager{},
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			&mock.DNSGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
			testnorduser.NewMockNorduserClient(nil),
			sharedctx.New(),
		)
		server.EnableMeshnet(context.Background(), &pb.Empty{})
		return server, &networker
	}

	tests := []struct {
		name               string
		peerUuid           string
		expectedResponse   *pb.AllowIncomingResponse
		expectedAllowedIPs []allowedIncoming
	}{
		{
			name:               "allow valid peer",
			peerUuid:           peerValidUuid,
			expectedResponse:   &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
			expectedAllowedIPs: []allowedIncoming{{address: UniqueAddress{UID: peerValidPublicKey, Address: peerValidAddress}, lanAllowed: false}},
		},
		{
			name:               "allow peer with no ip",
			peerUuid:           peerNoIpUuid,
			expectedResponse:   &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
			expectedAllowedIPs: []allowedIncoming{},
		},
		{
			name:     "peer traffic routing already allowed",
			peerUuid: peerIncomingAlreadyAllowedUuid,
			expectedResponse: &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_AllowIncomingErrorCode{
					AllowIncomingErrorCode: pb.AllowIncomingErrorCode_INCOMING_ALREADY_ALLOWED,
				},
			},
			expectedAllowedIPs: []allowedIncoming{},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
			expectedAllowedIPs: []allowedIncoming{},
		},
		{
			name:               "allow peer no routing",
			peerUuid:           peerNoRoutingUuid,
			expectedResponse:   &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
			expectedAllowedIPs: []allowedIncoming{{address: UniqueAddress{UID: peerNoRoutingPublicKey, Address: peerNoRoutingAddress}, lanAllowed: false}},
		},
		{
			name:               "allow peer no lan",
			peerUuid:           peerNoLANUuid,
			expectedResponse:   &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
			expectedAllowedIPs: []allowedIncoming{{address: UniqueAddress{UID: peerNoLANAddressPublicKey, Address: peerNoLANAddress}, lanAllowed: false}},
		},
		{
			name:               "allow peer routing and lan",
			peerUuid:           peerLANAndRoutingUuid,
			expectedResponse:   &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
			expectedAllowedIPs: []allowedIncoming{{address: UniqueAddress{UID: peerLANAndRoutingPublicKey, Address: peerLANAndRoutingAddress}, lanAllowed: true}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, networker := getServer()
			resp, err := server.AllowIncoming(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedAllowedIPs, networker.allowedIncoming, "Invalid addresses were allowed.")
		})
	}
}

func TestServer_DenyIncoming(t *testing.T) {
	peerValidUuid := exampleUUID3
	peerNoIpUuid := exampleUUID2
	peerIncomingAlreadyDeniedUuid := exampleUUID1

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := examplePublicKey1
	peerIncomingAlreadyDeniedPublicKey := examplePublicKey2

	getServer := func() (*Server, *workingNetworker) {
		registryApi := mock.RegistryMock{}
		registryApi.Peers = []mesh.MachinePeer{
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
			&mock.ConfigManager{},
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			&mock.DNSGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
			testnorduser.NewMockNorduserClient(nil),
			sharedctx.New(),
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
	peerValidUuid := exampleUUID3
	peerNoIpUuid := exampleUUID2
	peerIncomingAlreadyDeniedUuid := exampleUUID1

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := examplePublicKey2
	peerIncomingAlreadyDeniedPublicKey := examplePublicKey1

	getServer := func() (*Server, *workingNetworker) {
		registryApi := mock.RegistryMock{}
		registryApi.Peers = []mesh.MachinePeer{
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
			&mock.ConfigManager{},
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			&mock.DNSGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
			testnorduser.NewMockNorduserClient(nil),
			sharedctx.New(),
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
	peerValidUuid := exampleUUID3
	peerNoIpUuid := exampleUUID2
	peerIncomingAlreadyDeniedUuid := exampleUUID1

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := examplePublicKey1
	peerIncomingAlreadyDeniedPublicKey := examplePublicKey2

	getServer := func() (*Server, *workingNetworker) {
		registryApi := mock.RegistryMock{}
		registryApi.Peers = []mesh.MachinePeer{
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
			&mock.ConfigManager{},
			registrationChecker{},
			acceptInvitationsAPI{},
			&networker,
			&registryApi,
			&mock.DNSGetter{},
			&subs.Subject[error]{},
			&subs.Subject[[]string]{},
			&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
			testnorduser.NewMockNorduserClient(nil),
			sharedctx.New(),
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
	peerValidUuid := exampleUUID3
	peerAlreadyEnabledUuid := exampleUUID1

	peers := []mesh.MachinePeer{
		{
			ID:                uuid.MustParse(peerValidUuid),
			DoIAllowFileshare: false,
		},
		{
			ID:                uuid.MustParse(peerAlreadyEnabledUuid),
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
			peerUuid: peerAlreadyEnabledUuid,
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
			server := newMockedServer(t,
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
	peerValidUuid := exampleUUID3
	peerAleradyDisabledUuid := exampleUUID1

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
			server := newMockedServer(t,
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

func TestServer_Peer_Nickname(t *testing.T) {
	category.Set(t, category.Unit)

	peerNickname1 := "nickname1"
	changedSuccessfully := &pb.ChangeNicknameResponse{
		Response: &pb.ChangeNicknameResponse_Empty{},
	}

	tests := []struct {
		name             string
		peersList        mesh.MachinePeers
		peerId           string
		newNickname      string
		isMeshOn         bool
		listErr          error
		saveConfigErr    error
		loadConfigErr    error
		configureErr     error
		isNotLoggedIn    bool
		expectedErr      error
		registrationErr  error
		reservedDNSNames mock.RegisteredDomainsList
		expectedResponse *pb.ChangeNicknameResponse
	}{
		{
			name:     "set successful using peer ID",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:           exampleUUID1,
			newNickname:      peerNickname1,
			expectedResponse: changedSuccessfully,
		},
		{
			name:     "set successful case insensitive using nickname",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID:       uuid.MustParse(exampleUUID1),
					Nickname: peerNickname1,
				},
			},
			peerId:           peerNickname1,
			newNickname:      strings.ToUpper(peerNickname1),
			expectedResponse: changedSuccessfully,
		},
		{
			name:     "reset successful",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID:       uuid.MustParse(exampleUUID1),
					Nickname: peerNickname1,
				},
			},
			peerId:           exampleUUID1,
			expectedResponse: changedSuccessfully,
		},
		{
			name:     "fails settings the same nickname",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID:       uuid.MustParse(exampleUUID1),
					Nickname: peerNickname1,
				},
			},
			peerId:      exampleUUID1,
			newNickname: peerNickname1,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_SAME_NICKNAME,
				},
			},
		},
		{
			name:     "fails because is not logged in",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:        exampleUUID1,
			newNickname:   peerNickname1,
			isNotLoggedIn: true,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			},
		},
		{
			name:     "fails to load config",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:        exampleUUID1,
			newNickname:   peerNickname1,
			loadConfigErr: fmt.Errorf("load failed"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			},
		},
		{
			name:     "meshnet not running",
			isMeshOn: false,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:      exampleUUID1,
			newNickname: peerNickname1,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			},
		},
		{
			name:     "fails at registration info incorrect",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:          exampleUUID1,
			newNickname:     peerNickname1,
			registrationErr: fmt.Errorf("failed"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			},
		},
		{
			name:     "peer not found",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:      exampleUUID2,
			newNickname: peerNickname1,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
		},
		{
			name:     "fails to get peers list",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:      exampleUUID1,
			newNickname: peerNickname1,
			listErr:     fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
		{
			name:     "fails to get peers list with ErrUnauthorized",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:      exampleUUID1,
			newNickname: peerNickname1,
			listErr:     core.ErrUnauthorized,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			},
		},
		{
			name:     "fails to register because of DNS conflict",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:           exampleUUID1,
			newNickname:      "peer1",
			reservedDNSNames: mock.RegisteredDomainsList{"peer1": []net.IP{net.IPv4bcast}},
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_DOMAIN_NAME_EXISTS,
				},
			},
		},
		{
			name:     "successful change nickname for same value but different caps",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID:       uuid.MustParse(exampleUUID1),
					Nickname: "peer1",
				},
			},
			peerId:           exampleUUID1,
			newNickname:      "PEER1",
			reservedDNSNames: mock.RegisteredDomainsList{"peer1": []net.IP{net.IPv4bcast}},
			expectedResponse: changedSuccessfully,
		},
		{
			name:        "machine has no peers",
			isMeshOn:    true,
			peerId:      exampleUUID1,
			newNickname: peerNickname1,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			},
		},
		{
			name:     "set nickname API fails",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:       exampleUUID1,
			newNickname:  peerNickname1,
			configureErr: fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
		{
			name:     "API returns error that nickname contains invalid chars",
			isMeshOn: true,
			peersList: []mesh.MachinePeer{
				{
					ID:       uuid.MustParse(exampleUUID1),
					Nickname: "peer1",
				},
			},
			peerId:       exampleUUID1,
			newNickname:  peerNickname1,
			configureErr: core.ErrContainsInvalidChars,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_INVALID_CHARS,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registryApi := mock.RegistryMock{}
			registryApi.Peers = test.peersList
			registryApi.ListErr = test.listErr
			registryApi.ConfigureErr = test.configureErr

			configManager := &mock.ConfigManager{}
			configManager.SaveErr = test.saveConfigErr
			configManager.LoadErr = test.loadConfigErr

			ac := meshRenewChecker{}
			ac.IsNotLoggedIn = test.isNotLoggedIn

			checker := registrationChecker{}

			server := NewServer(
				&ac,
				configManager,
				&checker,
				acceptInvitationsAPI{},
				&workingNetworker{},
				&registryApi,
				&mock.DNSGetter{RegisteredDomains: test.reservedDNSNames},
				&subs.Subject[error]{},
				&subs.Subject[[]string]{},
				&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
				testnorduser.NewMockNorduserClient(nil),
				sharedctx.New(),
			)

			if test.isMeshOn {
				server.EnableMeshnet(context.Background(), &pb.Empty{})
			}

			checker.registrationErr = test.registrationErr

			request := pb.ChangePeerNicknameRequest{
				Identifier: test.peerId,
				Nickname:   test.newNickname,
			}

			resp, err := server.ChangePeerNickname(context.Background(), &request)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)

			if p := registryApi.GetPeerWithIdentifier(test.peerId); p != nil {
				if !assert.ObjectsAreEqual(resp, changedSuccessfully) {
					assert.Equal(t, test.peersList[0].Nickname, p.Nickname)
				} else {
					assert.Equal(t, test.newNickname, p.Nickname)
				}
			}
		})
	}
}

func TestServer_Current_Machine_Nickname(t *testing.T) {
	category.Set(t, category.Unit)

	machineNickname := "nickname1"
	changedSuccessfully := &pb.ChangeNicknameResponse{
		Response: &pb.ChangeNicknameResponse_Empty{},
	}

	tests := []struct {
		name             string
		newNickname      string
		isMeshOn         bool
		listErr          error
		saveConfigErr    error
		loadConfigErr    error
		configureErr     error
		isNotLoggedIn    bool
		expectedErr      error
		registrationErr  error
		updateErr        error
		machine          mesh.Machine
		reservedDNSNames mock.RegisteredDomainsList
		expectedResponse *pb.ChangeNicknameResponse
	}{
		{
			name:             "set nickname successfully",
			isMeshOn:         true,
			newNickname:      machineNickname,
			machine:          mesh.Machine{SupportsRouting: true},
			expectedResponse: changedSuccessfully,
		},
		{
			name:             "clear nickname",
			isMeshOn:         true,
			machine:          mesh.Machine{SupportsRouting: true, Nickname: strings.ToUpper(machineNickname)},
			expectedResponse: changedSuccessfully,
		},
		{
			name:     "clear nickname, when is already empty",
			isMeshOn: true,
			machine:  mesh.Machine{SupportsRouting: true},
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_NICKNAME_ALREADY_EMPTY,
				},
			},
		},
		{
			name:          "fails because not logged in",
			isMeshOn:      true,
			isNotLoggedIn: true,
			newNickname:   machineNickname,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			},
		},
		{
			name:          "load config files fails",
			isMeshOn:      true,
			newNickname:   machineNickname,
			loadConfigErr: fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			},
		},
		{
			name:        "set nickname successfully",
			isMeshOn:    false,
			newNickname: machineNickname,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			},
		},
		{
			name:            "meshnet registration is not correct",
			isMeshOn:        true,
			newNickname:     machineNickname,
			registrationErr: fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			},
		},
		{
			name:        "set same nickname",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true, Nickname: machineNickname},
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_SAME_NICKNAME,
				},
			},
		},
		{
			name:             "set same nickname, but different caps",
			isMeshOn:         true,
			newNickname:      strings.ToUpper(machineNickname),
			machine:          mesh.Machine{SupportsRouting: true, Nickname: machineNickname},
			expectedResponse: changedSuccessfully,
		},
		{
			name:        "update API fails with ErrUnauthorized",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrUnauthorized,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			},
		},
		{
			name:        "update API fails",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
		{
			name:          "config save fails",
			isMeshOn:      true,
			newNickname:   machineNickname,
			machine:       mesh.Machine{SupportsRouting: true},
			saveConfigErr: fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			},
		},
		{
			name:             "fails to register because of DNS conflict",
			isMeshOn:         true,
			newNickname:      "peer1",
			machine:          mesh.Machine{SupportsRouting: true},
			reservedDNSNames: mock.RegisteredDomainsList{"peer1": []net.IP{net.IPv4bcast}},
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_DOMAIN_NAME_EXISTS,
				},
			},
		},
		{
			name:        "generic API error",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			},
		},
		{
			name:        "API returns error rate limit reach",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrRateLimitReach,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_RATE_LIMIT_REACH,
				},
			},
		},
		{
			name:        "API returns error nickname too long",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrNicknameTooLong,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_NICKNAME_TOO_LONG,
				},
			},
		},
		{
			name:        "API returns error duplicate nickname",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrDuplicateNickname,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_DUPLICATE_NICKNAME,
				},
			},
		},
		{
			name:        "API returns forbidden word",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrContainsForbiddenWord,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_CONTAINS_FORBIDDEN_WORD,
				},
			},
		},
		{
			name:        "API returns invalid suffix or prefix",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrInvalidPrefixOrSuffix,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_SUFFIX_OR_PREFIX_ARE_INVALID,
				},
			},
		},
		{
			name:        "API error double hyphens",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrNicknameWithDoubleHyphens,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_NICKNAME_HAS_DOUBLE_HYPHENS,
				},
			},
		},
		{
			name:        "API error nickname contains invalid chars",
			isMeshOn:    true,
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrContainsInvalidChars,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_INVALID_CHARS,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registryApi := mock.RegistryMock{}
			registryApi.ListErr = test.listErr
			registryApi.ConfigureErr = test.configureErr
			registryApi.UpdateErr = test.updateErr
			registryApi.CurrentMachine = test.machine

			configManager := mock.NewMockConfigManager()
			configManager.LoadErr = test.loadConfigErr
			configManager.Cfg.MeshDevice = &test.machine

			ac := meshRenewChecker{}
			ac.IsNotLoggedIn = test.isNotLoggedIn

			checker := registrationChecker{}

			server := NewServer(
				&ac,
				configManager,
				&checker,
				acceptInvitationsAPI{},
				&workingNetworker{},
				&registryApi,
				&mock.DNSGetter{RegisteredDomains: test.reservedDNSNames},
				&subs.Subject[error]{},
				&subs.Subject[[]string]{},
				&daemonevents.Events{Settings: &daemonevents.SettingsEvents{Meshnet: &daemonevents.MockPublisherSubscriber[bool]{}}},
				testnorduser.NewMockNorduserClient(nil),
				sharedctx.New(),
			)

			if test.isMeshOn {
				server.EnableMeshnet(context.Background(), &pb.Empty{})
			}

			checker.registrationErr = test.registrationErr

			request := pb.ChangeMachineNicknameRequest{
				Nickname: test.newNickname,
			}

			configManager.SaveErr = test.saveConfigErr

			resp, err := server.ChangeMachineNickname(context.Background(), &request)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)

			if !assert.ObjectsAreEqual(resp, changedSuccessfully) {
				if test.saveConfigErr == nil {
					// if the API is able to change the nickname, but save fails => local info and server info are out-of-sync
					// the out-of-sync state will remain until current machine receives a NC notification or after mesh restart
					assert.Equal(t, test.machine.Nickname, registryApi.CurrentMachine.Nickname)
				}
			} else {
				assert.Equal(t, test.newNickname, registryApi.CurrentMachine.Nickname)
				assert.True(t, configManager.Cfg.MeshDevice.SupportsRouting)
				assert.True(t, registryApi.CurrentMachine.SupportsRouting)
			}
		})
	}
}
