package meshnet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
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
	"github.com/stretchr/testify/require"
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

type workingNetworker struct{}

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

func (n *workingNetworker) PermitFileshare() error {
	return nil
}

func (n *workingNetworker) ForbidFileshare() error {
	return nil
}

func (*workingNetworker) Refresh(mesh.MachineMap) error { return nil }
func (*workingNetworker) StatusMap() (map[string]string, error) {
	return map[string]string{}, nil
}
func (*workingNetworker) LastServerName() string { return "" }
func (*workingNetworker) GetConnectionParameters() (vpn.ServerData, bool) {
	return vpn.ServerData{}, false
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

func newMockedServer(
	t *testing.T,
	enableMesh bool,
) *Server {
	t.Helper()

	registryApi := mock.RegistryMock{}
	configManager := mock.NewMockConfigManager()
	configManager.Cfg.Technology = config.Technology_NORDLYNX

	server := NewServer(
		meshRenewChecker{},
		configManager,
		registrationChecker{},
		acceptInvitationsAPI{},
		&workingNetworker{},
		&registryApi,
		&registryApi,
		&mock.DNSGetter{},
		&subs.Subject[error]{},
		&daemonevents.Events{
			Settings: &daemonevents.SettingsEvents{
				Meshnet: &daemonevents.MockPublisherSubscriber[bool]{},
			},
			User: &daemonevents.LoginEvents{
				Logout: &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{},
			},
			Service: &daemonevents.ServiceEvents{
				UiItemsClick: &daemonevents.MockPublisherSubscriber[events.UiItemsAction]{},
				Connect:      &daemonevents.MockPublisherSubscriber[events.DataConnect]{},
			},
		},
		testnorduser.NewMockNorduserClient(nil),
		sharedctx.New(),
	)

	if enableMesh {
		server.EnableMeshnet(context.Background(), &pb.Empty{})
	}
	return server
}

func TestServer_EnableMeshnet(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name                string
		startFileshareError error
		success             bool
	}{
		{
			name:                "everything works",
			startFileshareError: nil,
			success:             true,
		},
		{
			name:                "fileshare fails",
			startFileshareError: fmt.Errorf("failed to disable fileshare"),
			success:             true, // Fileshare shouldn't impact meshnet enabling
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Check server creation
			mserver := newMockedServer(t, false)
			mserver.norduser = testnorduser.NewMockNorduserClient(test.startFileshareError)
			assert.NotEqual(t, nil, mserver)

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
		startFileshareError error
	}{
		{
			name:                "everything works",
			startFileshareError: nil,
		},
		{
			name:                "fileshare fails",
			startFileshareError: fmt.Errorf("failed to start fileshare"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Check server creation
			mserver := newMockedServer(t, true)
			mserver.norduser = testnorduser.NewMockNorduserClient(test.startFileshareError)

			assert.NotEqual(t, nil, mserver)

			// Disable Mesh
			resp, err := mserver.DisableMeshnet(context.Background(), &pb.Empty{})
			assert.NoError(t, err)
			_, ok := resp.GetResponse().(*pb.MeshnetResponse_Empty)
			assert.Equal(t, true, ok)

			// Check new server configuration
			var cfg config.Config
			err = mserver.cm.Load(&cfg)
			assert.NoError(t, err)
			assert.Equal(t, false, cfg.Mesh)
		})
	}
}

func TestServer_DisableMeshnetViaRemoteConfig(t *testing.T) {
	category.Set(t, category.Unit)
	t.Run("basic test", func(t *testing.T) {
		// Check server creation
		mserver := newMockedServer(t, true)
		assert.NotEqual(t, nil, mserver)

		// Disable Mesh via Remote config update
		err := mserver.RemoteConfigUpdate(remote.RemoteConfigEvent{MeshnetFeatureEnabled: false})
		assert.NoError(t, err)

		// Check new server configuration
		var cfg config.Config
		err = mserver.cm.Load(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, false, cfg.Mesh)
	})
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
			server := newMockedServer(t, true)
			server.invitationAPI = test.inv
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
	server := newMockedServer(t, true)
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
		registryApi := mock.RegistryMock{}
		registryApi.Peers = test.peers
		server := newMockedServer(t, true)
		server.mapper = &registryApi

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
			server := newMockedServer(t, true)
			registryApi := mock.RegistryMock{}
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
			server.mapper = &registryApi
			resp, err := server.Connect(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_AcceptIncoming(t *testing.T) {
	peerValidUuid := exampleUUID3
	peerIncomingAlreadyAllowedUuid := exampleUUID1

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyAllowedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := examplePublicKey1
	peerIncomingAlreadyAllowedPublicKey := examplePublicKey2

	tests := []struct {
		name             string
		peerUuid         string
		expectedResponse *pb.AllowIncomingResponse
	}{
		{
			name:             "allow valid peer",
			peerUuid:         peerValidUuid,
			expectedResponse: &pb.AllowIncomingResponse{Response: &pb.AllowIncomingResponse_Empty{}},
		},
		{
			name:     "peer traffic routing already allowed",
			peerUuid: peerIncomingAlreadyAllowedUuid,
			expectedResponse: &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_AllowIncomingErrorCode{
					AllowIncomingErrorCode: pb.AllowIncomingErrorCode_INCOMING_ALREADY_ALLOWED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newMockedServer(t, true)
			registryApi := mock.RegistryMock{}
			registryApi.Peers = []mesh.MachinePeer{
				{
					ID:              uuid.MustParse(peerValidUuid),
					DoIAllowInbound: false,
					Address:         peerValidAddress,
					PublicKey:       peerValidPublicKey,
				},
				{
					ID:              uuid.MustParse(peerIncomingAlreadyAllowedUuid),
					DoIAllowInbound: true,
					Address:         peerIncomingAlreadyAllowedAddress,
					PublicKey:       peerIncomingAlreadyAllowedPublicKey,
				},
			}
			server.mapper = &registryApi

			resp, err := server.AllowIncoming(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_DenyIncoming(t *testing.T) {
	peerValidUuid := exampleUUID3
	peerIncomingAlreadyDeniedUuid := exampleUUID1

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := examplePublicKey1
	peerIncomingAlreadyDeniedPublicKey := examplePublicKey2

	tests := []struct {
		name             string
		peerUuid         string
		expectedResponse *pb.DenyIncomingResponse
	}{
		{
			name:             "deny valid peer",
			peerUuid:         peerValidUuid,
			expectedResponse: &pb.DenyIncomingResponse{Response: &pb.DenyIncomingResponse_Empty{}},
		},
		{
			name:     "peer traffic routing already denied",
			peerUuid: peerIncomingAlreadyDeniedUuid,
			expectedResponse: &pb.DenyIncomingResponse{
				Response: &pb.DenyIncomingResponse_DenyIncomingErrorCode{
					DenyIncomingErrorCode: pb.DenyIncomingErrorCode_INCOMING_ALREADY_DENIED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.DenyIncomingResponse{
				Response: &pb.DenyIncomingResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newMockedServer(t, true)
			registryApi := mock.RegistryMock{}
			registryApi.Peers = []mesh.MachinePeer{
				{
					ID:              uuid.MustParse(peerValidUuid),
					DoIAllowInbound: true,
					Address:         peerValidAddress,
					PublicKey:       peerValidPublicKey,
				},
				{
					ID:              uuid.MustParse(peerIncomingAlreadyDeniedUuid),
					DoIAllowInbound: false,
					Address:         peerIncomingAlreadyDeniedAddress,
					PublicKey:       peerIncomingAlreadyDeniedPublicKey,
				},
			}
			server.mapper = &registryApi
			resp, err := server.DenyIncoming(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_AllowFileshare(t *testing.T) {
	peerValidUuid := exampleUUID3
	peerIncomingAlreadyDeniedUuid := exampleUUID1

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := examplePublicKey2
	peerIncomingAlreadyDeniedPublicKey := examplePublicKey1

	tests := []struct {
		name             string
		peerUuid         string
		expectedResponse *pb.AllowFileshareResponse
	}{
		{
			name:             "allow valid peer",
			peerUuid:         peerValidUuid,
			expectedResponse: &pb.AllowFileshareResponse{Response: &pb.AllowFileshareResponse_Empty{}},
		},
		{
			name:     "fileshare already denied",
			peerUuid: peerIncomingAlreadyDeniedUuid,
			expectedResponse: &pb.AllowFileshareResponse{
				Response: &pb.AllowFileshareResponse_AllowSendErrorCode{
					AllowSendErrorCode: pb.AllowFileshareErrorCode_SEND_ALREADY_ALLOWED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.AllowFileshareResponse{
				Response: &pb.AllowFileshareResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newMockedServer(t, true)
			registryApi := mock.RegistryMock{}
			registryApi.Peers = []mesh.MachinePeer{
				{
					ID:                uuid.MustParse(peerValidUuid),
					DoIAllowFileshare: false,
					Address:           peerValidAddress,
					PublicKey:         peerValidPublicKey,
				},
				{
					ID:                uuid.MustParse(peerIncomingAlreadyDeniedUuid),
					DoIAllowFileshare: true,
					Address:           peerIncomingAlreadyDeniedAddress,
					PublicKey:         peerIncomingAlreadyDeniedPublicKey,
				},
			}
			server.mapper = &registryApi
			resp, err := server.AllowFileshare(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_DenyFileshare(t *testing.T) {
	peerValidUuid := exampleUUID3
	peerIncomingAlreadyDeniedUuid := exampleUUID1

	peerValidAddress := netip.MustParseAddr("220.16.61.136")
	peerIncomingAlreadyDeniedAddress := netip.MustParseAddr("87.169.173.253")

	peerValidPublicKey := examplePublicKey1
	peerIncomingAlreadyDeniedPublicKey := examplePublicKey2

	tests := []struct {
		name             string
		peerUuid         string
		expectedResponse *pb.DenyFileshareResponse
	}{
		{
			name:             "deny valid peer",
			peerUuid:         peerValidUuid,
			expectedResponse: &pb.DenyFileshareResponse{Response: &pb.DenyFileshareResponse_Empty{}},
		},
		{
			name:     "fileshare already denied",
			peerUuid: peerIncomingAlreadyDeniedUuid,
			expectedResponse: &pb.DenyFileshareResponse{
				Response: &pb.DenyFileshareResponse_DenySendErrorCode{
					DenySendErrorCode: pb.DenyFileshareErrorCode_SEND_ALREADY_DENIED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.DenyFileshareResponse{
				Response: &pb.DenyFileshareResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newMockedServer(t, true)
			registryApi := mock.RegistryMock{}
			registryApi.Peers = []mesh.MachinePeer{
				{
					ID:                uuid.MustParse(peerValidUuid),
					DoIAllowFileshare: true,
					Address:           peerValidAddress,
					PublicKey:         peerValidPublicKey,
				},
				{
					ID:                uuid.MustParse(peerIncomingAlreadyDeniedUuid),
					DoIAllowFileshare: false,
					Address:           peerIncomingAlreadyDeniedAddress,
					PublicKey:         peerIncomingAlreadyDeniedPublicKey,
				},
			}
			server.mapper = &registryApi

			resp, err := server.DenyFileshare(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_EnableAutomaticFileshare(t *testing.T) {
	peerValidUuid := exampleUUID3
	peerAlreadyEnabledUuid := exampleUUID1

	tests := []struct {
		name             string
		peerUuid         string
		configureErr     error
		expectedResponse *pb.EnableAutomaticFileshareResponse
	}{
		{
			name:             "enable for valid peer",
			peerUuid:         peerValidUuid,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{Response: &pb.EnableAutomaticFileshareResponse_Empty{}},
		},
		{
			name:     "automatic fileshare already enabled",
			peerUuid: peerAlreadyEnabledUuid,
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_EnableAutomaticFileshareErrorCode{
					EnableAutomaticFileshareErrorCode: pb.EnableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_ENABLED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
		{
			name:         "failed to configure peer",
			peerUuid:     peerValidUuid,
			configureErr: fmt.Errorf("generic error"),
			expectedResponse: &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
			reg := &mock.RegistryMock{}
			reg.ConfigureErr = test.configureErr
			reg.Peers = peers
			server := newMockedServer(t, true)
			server.reg = reg
			server.mapper = reg
			resp, err := server.EnableAutomaticFileshare(context.Background(), &pb.UpdatePeerRequest{Identifier: test.peerUuid})

			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}

func TestServer_DisableAutomaticFileshare(t *testing.T) {
	peerValidUuid := exampleUUID3
	peerAleradyDisabledUuid := exampleUUID1

	tests := []struct {
		name             string
		peerUuid         string
		configureErr     error
		expectedResponse *pb.DisableAutomaticFileshareResponse
	}{
		{
			name:             "enable for valid peer",
			peerUuid:         peerValidUuid,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{Response: &pb.DisableAutomaticFileshareResponse_Empty{}},
		},
		{
			name:     "automatic fileshare already enabled",
			peerUuid: peerAleradyDisabledUuid,
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_DisableAutomaticFileshareErrorCode{
					DisableAutomaticFileshareErrorCode: pb.DisableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_DISABLED,
				},
			},
		},
		{
			name:     "unknown peer",
			peerUuid: "invalid",
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
		{
			name:         "failed to configure peer",
			peerUuid:     peerValidUuid,
			configureErr: fmt.Errorf("generic error"),
			expectedResponse: &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			},
		},
	}

	for _, test := range tests {
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

		t.Run(test.name, func(t *testing.T) {
			reg := &mock.RegistryMock{}
			reg.ConfigureErr = test.configureErr
			reg.Peers = peers
			server := newMockedServer(t, true)
			server.reg = reg
			server.mapper = reg
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
		mapErr           error
		configureErr     error
		reservedDNSNames mock.RegisteredDomainsList
		expectedResponse *pb.ChangeNicknameResponse
	}{
		{
			name: "set successful using peer ID",
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
			name: "set successful case insensitive using nickname",
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
			name: "reset successful",
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
			name: "fails settings the same nickname",
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
			name: "peer not found",
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:      exampleUUID2,
			newNickname: peerNickname1,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
		{
			name: "fails to get peers list",
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:      exampleUUID1,
			newNickname: peerNickname1,
			mapErr:      fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			},
		},
		{
			name: "fails to get peers list with ErrUnauthorized",
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:      exampleUUID1,
			newNickname: peerNickname1,
			mapErr:      core.ErrUnauthorized,

			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN),
				},
			},
		},
		{
			name: "fails to register because of DNS conflict",
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
			name: "successful change nickname for same value but different caps",
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
			peerId:      exampleUUID1,
			newNickname: peerNickname1,

			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
				},
			},
		},
		{
			name: "set nickname API fails",
			peersList: []mesh.MachinePeer{
				{
					ID: uuid.MustParse(exampleUUID1),
				},
			},
			peerId:       exampleUUID1,
			newNickname:  peerNickname1,
			configureErr: fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			},
		},
		{
			name: "API returns error that nickname contains invalid chars",
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
			registryApi.MapErr = test.mapErr
			registryApi.ConfigureErr = test.configureErr

			server := newMockedServer(t, true)
			server.reg = &registryApi
			server.mapper = &registryApi
			server.nameservers = &mock.DNSGetter{RegisteredDomains: test.reservedDNSNames}

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
		mapErr           error
		configureErr     error
		isNotLoggedIn    bool
		updateErr        error
		machine          mesh.Machine
		reservedDNSNames mock.RegisteredDomainsList
		expectedResponse *pb.ChangeNicknameResponse
	}{
		{
			name:             "set nickname successfully",
			newNickname:      machineNickname,
			machine:          mesh.Machine{SupportsRouting: true},
			expectedResponse: changedSuccessfully,
		},
		{
			name:             "clear nickname",
			machine:          mesh.Machine{SupportsRouting: true, Nickname: strings.ToUpper(machineNickname)},
			expectedResponse: changedSuccessfully,
		},
		{
			name:    "clear nickname, when is already empty",
			machine: mesh.Machine{SupportsRouting: true},
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_NICKNAME_ALREADY_EMPTY,
				},
			},
		},
		{
			name:          "fails because not logged in",
			isNotLoggedIn: true,
			newNickname:   machineNickname,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN),
				},
			},
		},
		{
			name:        "set same nickname",
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
			newNickname:      strings.ToUpper(machineNickname),
			machine:          mesh.Machine{SupportsRouting: true, Nickname: machineNickname},
			expectedResponse: changedSuccessfully,
		},
		{
			name:        "update API fails with ErrUnauthorized",
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   core.ErrUnauthorized,
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN),
				},
			},
		},
		{
			name:        "update API fails",
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			},
		},
		{
			name:             "fails to register because of DNS conflict",
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
			newNickname: machineNickname,
			machine:     mesh.Machine{SupportsRouting: true},
			updateErr:   fmt.Errorf("error"),
			expectedResponse: &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			},
		},
		{
			name:        "API returns error rate limit reach",
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
			registryApi.MapErr = test.mapErr
			registryApi.ConfigureErr = test.configureErr
			registryApi.UpdateErr = test.updateErr
			registryApi.CurrentMachine = test.machine

			configManager := mock.NewMockConfigManager()
			configManager.Cfg.MeshDevice = &test.machine
			configManager.Cfg.Mesh = true

			ac := meshRenewChecker{}
			ac.IsNotLoggedIn = test.isNotLoggedIn

			server := newMockedServer(t, true)
			server.ac = ac
			server.reg = &registryApi
			server.mapper = &registryApi
			server.cm = configManager
			server.nameservers = &mock.DNSGetter{RegisteredDomains: test.reservedDNSNames}

			request := pb.ChangeMachineNicknameRequest{
				Nickname: test.newNickname,
			}

			resp, err := server.ChangeMachineNickname(context.Background(), &request)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedResponse, resp)

			if !assert.ObjectsAreEqual(resp, changedSuccessfully) {
				// if the API is able to change the nickname, but save fails => local info and server info are out-of-sync
				// the out-of-sync state will remain until current machine receives a NC notification or after mesh restart
				assert.Equal(t, test.machine.Nickname, registryApi.CurrentMachine.Nickname)
			} else {
				assert.Equal(t, test.newNickname, registryApi.CurrentMachine.Nickname)
				assert.True(t, configManager.Cfg.MeshDevice.SupportsRouting)
				assert.True(t, registryApi.CurrentMachine.SupportsRouting)
			}
		})
	}
}

func TestServer_fetchCfg(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name          string
		isNotLoggedIn bool
		cm            config.Manager
		mc            Checker
		err           *pb.Error
	}{
		{
			name: "success",
		},
		{
			name:          "not logged in",
			isNotLoggedIn: true,
			err:           generalServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN),
		},
		{
			name: "config load failed",
			cm: func() config.Manager {
				cm := mock.NewMockConfigManager()
				cm.LoadErr = errors.New("some err")
				return cm
			}(),
			err: generalServiceError(pb.ServiceErrorCode_CONFIG_FAILURE),
		},
		{
			name: "meshnet not enabled",
			cm: func() config.Manager {
				cm := mock.NewMockConfigManager()
				cm.Cfg.Mesh = false
				return cm
			}(),
			err: generalMeshError(pb.MeshnetErrorCode_NOT_ENABLED),
		},
		{
			name: "auth checker failed",
			mc:   registrationChecker{registrationErr: errors.New("some err")},
			err:  generalMeshError(pb.MeshnetErrorCode_NOT_REGISTERED),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newMockedServer(t, true)
			if tt.mc != nil {
				s.mc = tt.mc
			}
			if tt.cm != nil {
				s.cm = tt.cm
			}
			ac := meshRenewChecker{}
			ac.IsNotLoggedIn = tt.isNotLoggedIn
			s.ac = ac

			// Make sure it fetches the same config as cm would
			var expectedCfg config.Config
			s.cm.Load(&expectedCfg)

			cfg, err := s.fetchCfg()

			// Ignore meshnet settings as they are likely to be changed by reg checker
			cfg.Mesh = false
			cfg.MeshDevice = nil
			cfg.MeshPrivateKey = ""
			cfg.Technology = config.Technology_UNKNOWN_TECHNOLOGY
			expectedCfg.Mesh = false
			expectedCfg.MeshDevice = nil
			expectedCfg.MeshPrivateKey = ""
			expectedCfg.Technology = config.Technology_UNKNOWN_TECHNOLOGY

			assert.EqualValues(t, tt.err, err)
			assert.Equal(t, expectedCfg, cfg)
		})
	}
}

func TestServer_fetchPeers(t *testing.T) {
	category.Set(t, category.Unit)

	for _, tt := range []struct {
		name   string
		err    *pb.Error
		cm     config.Manager
		mapErr error
	}{
		{
			name: "success",
		},
		{
			name:   "invalid token",
			mapErr: core.ErrUnauthorized,
			err:    generalServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN),
		},
		{
			name:   "config save error on logout",
			err:    generalServiceError(pb.ServiceErrorCode_CONFIG_FAILURE),
			mapErr: core.ErrUnauthorized,
			cm: func() config.Manager {
				cm := mock.NewMockConfigManager()
				cm.Cfg.Mesh = true
				cm.SaveErr = errors.New("some err")
				return cm
			}(),
		},
		{
			name:   "self removed",
			err:    generalMeshError(pb.MeshnetErrorCode_NOT_ENABLED),
			mapErr: core.ErrConflict,
			cm: func() config.Manager {
				cm := mock.NewMockConfigManager()
				cm.Cfg.Mesh = false
				return cm
			}(),
		},
		{
			name:   "list failure",
			err:    generalServiceError(pb.ServiceErrorCode_API_FAILURE),
			mapErr: core.ErrConflict,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			peers := []mesh.MachinePeer{
				{
					ID:                uuid.MustParse(exampleUUID1),
					DoIAllowFileshare: false,
					AlwaysAcceptFiles: true,
				},
				{
					ID:                uuid.MustParse(exampleUUID1),
					DoIAllowFileshare: true,
				},
			}

			reg := &mock.RegistryMock{}
			reg.MapErr = tt.mapErr
			reg.Peers = peers
			s := newMockedServer(t, true)
			s.mapper = reg
			if tt.cm != nil {
				s.cm = tt.cm
			}
			token, self, peers, err := s.fetchPeers()

			// Make sure it fetches the same config as cm would
			var cfg config.Config
			require.NoError(t, s.cm.Load(&cfg))
			assert.Equal(t, cfg.TokensData[cfg.AutoConnectData.ID].Token, token)
			assert.EqualValues(t, *cfg.MeshDevice, self)
			if tt.mapErr == nil {
				mmap, _ := s.mapper.Map(token, self.ID, true)
				assert.EqualValues(t, mmap.Peers, peers)
			}
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestServer_fetchPeer(t *testing.T) {
	category.Set(t, category.Unit)

	peers := []mesh.MachinePeer{
		{
			ID:                uuid.MustParse(exampleUUID1),
			DoIAllowFileshare: false,
			AlwaysAcceptFiles: true,
		},
		{
			ID:                uuid.MustParse(exampleUUID2),
			DoIAllowFileshare: true,
		},
	}

	for _, tt := range []struct {
		name     string
		peerUUID string
		err      *pb.UpdatePeerError
		mapErr   error
	}{
		{
			name:     "success 1",
			peerUUID: exampleUUID1,
		},
		{
			name:     "success 2",
			peerUUID: exampleUUID2,
		},
		{
			name:     "peer not found",
			peerUUID: exampleUUID3,
			err:      updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND),
		},
		{
			name:     "list failure",
			peerUUID: exampleUUID1,
			mapErr:   core.ErrConflict,
			err:      updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reg := &mock.RegistryMock{}
			reg.MapErr = tt.mapErr
			reg.Peers = peers
			s := newMockedServer(t, true)
			s.mapper = reg
			token, self, peer, err := s.fetchPeer(tt.peerUUID)

			// Make sure it fetches the same config as cm would
			var cfg config.Config
			require.NoError(t, s.cm.Load(&cfg))
			assert.Equal(t, cfg.TokensData[cfg.AutoConnectData.ID].Token, token)
			assert.EqualValues(t, *cfg.MeshDevice, self)
			assert.EqualValues(t, tt.err, err)

			if tt.err == nil {
				assert.Equal(t, uuid.MustParse(tt.peerUUID), peer.ID)
			}
		})
	}
}
