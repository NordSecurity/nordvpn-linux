package fileshare

import (
	"context"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"

	"google.golang.org/grpc"
)

// MockFileshare is a mock implementation of fileshare. It is used when libdrop
// is not available and should be used only for development purposes.
type MockFileshare struct{}

// Enable is a stub
func (MockFileshare) Enable(netip.Addr) error { return nil }

// Disable is a stub
func (MockFileshare) Disable() error { return nil }

// Send is a stub
func (MockFileshare) Send(netip.Addr, []string) (string, error) { return "", nil }

// Accept is a stub
func (MockFileshare) Accept(string, string, string) error { return nil }

// Reject is a stub
func (MockFileshare) Cancel(string) error { return nil }

// CancelFile is a stub
func (MockFileshare) CancelFile(transferID string, fileID string) error { return nil }

// MockStorage is a mock implementation of fileshare MockStorage. It is used when no persistence is desired.
type MockStorage struct{}

// Load is a stub
func (MockStorage) Load() (map[string]*pb.Transfer, error) {
	return map[string]*pb.Transfer{}, nil
}

// Save is a stub
func (MockStorage) Save(map[string]*pb.Transfer) error { return nil }

type mockMeshClient struct {
	meshpb.MeshnetClient
	isEnabled      bool
	localPeers     []*meshpb.Peer
	externalPeers  []*meshpb.Peer
	selfPeer       *meshpb.Peer
	getPeersCalled bool
}

// IsEnabled mock implementation
func (m *mockMeshClient) IsEnabled(ctx context.Context, in *meshpb.Empty, opts ...grpc.CallOption) (*meshpb.ServiceBoolResponse, error) {
	return &meshpb.ServiceBoolResponse{Response: &meshpb.ServiceBoolResponse_Value{Value: m.isEnabled}}, nil
}

// GetPeers mock implementation
func (m *mockMeshClient) GetPeers(ctx context.Context, in *meshpb.Empty, opts ...grpc.CallOption) (*meshpb.GetPeersResponse, error) {
	response := &meshpb.GetPeersResponse{
		Response: &meshpb.GetPeersResponse_Peers{
			Peers: &meshpb.PeerList{
				Local:    m.localPeers,
				External: m.externalPeers,
				Self:     m.selfPeer,
			},
		},
	}
	m.getPeersCalled = true
	return response, nil
}

// NotifyNewTransfer mock implementation
func (m *mockMeshClient) NotifyNewTransfer(ctx context.Context, in *meshpb.NewTransferNotification, opts ...grpc.CallOption) (*meshpb.NotifyNewTransferResponse, error) {
	return &meshpb.NotifyNewTransferResponse{
		Response: &meshpb.NotifyNewTransferResponse_Empty{},
	}, nil
}
