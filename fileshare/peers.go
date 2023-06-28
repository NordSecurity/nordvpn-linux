package fileshare

import (
	"context"
	"fmt"

	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"golang.org/x/exp/slices"
)

// GetSelfPeer from meshnet client
func GetSelfPeer(meshClient meshpb.MeshnetClient) (*meshpb.Peer, error) {
	peerList, err := getAllPeers(meshClient)
	if err != nil {
		return nil, err
	}
	return peerList.Self, nil
}

func getPeers(meshClient meshpb.MeshnetClient) ([]*meshpb.Peer, error) {
	peerList, err := getAllPeers(meshClient)
	if err != nil {
		return nil, err
	}
	peers := peerList.External
	peers = append(peers, peerList.Local...)
	return peers, nil
}

func getPeerByIP(meshClient meshpb.MeshnetClient, peerIP string) (*meshpb.Peer, error) {
	peers, err := getPeers(meshClient)
	if err != nil {
		return nil, err
	}
	peerIndex := slices.IndexFunc(peers, func(peer *meshpb.Peer) bool {
		return peer.Ip == peerIP
	})
	if peerIndex == -1 {
		return nil, fmt.Errorf("peer %s not found", peerIP)
	}
	return peers[peerIndex], nil
}

func getAllPeers(meshClient meshpb.MeshnetClient) (*meshpb.PeerList, error) {
	resp, err := meshClient.GetPeers(context.Background(), &meshpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}
	switch resp := resp.Response.(type) {
	case *meshpb.GetPeersResponse_Peers:
		return resp.Peers, nil
	case *meshpb.GetPeersResponse_ServiceErrorCode:
		return nil, fmt.Errorf("GetPeers failed, service error: %s", meshpb.ServiceErrorCode_name[int32(resp.ServiceErrorCode)])
	case *meshpb.GetPeersResponse_MeshnetErrorCode:
		return nil, fmt.Errorf("GetPeers failed, meshnet error: %s", meshpb.ServiceErrorCode_name[int32(resp.MeshnetErrorCode)])
	default:
		return nil, fmt.Errorf("GetPeers failed, unknown error")
	}
}
