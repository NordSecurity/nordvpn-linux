package utils

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/netip"
	"strings"

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

func GetPeers(meshClient meshpb.MeshnetClient) ([]*meshpb.Peer, error) {
	peerList, err := getAllPeers(meshClient)
	if err != nil {
		return nil, err
	}
	peers := peerList.External
	peers = append(peers, peerList.Local...)
	return peers, nil
}

func GetPeerByIP(meshClient meshpb.MeshnetClient, peerIP string) (*meshpb.Peer, error) {
	peers, err := GetPeers(meshClient)
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
	case *meshpb.GetPeersResponse_Error:
		return nil, fmt.Errorf("GetPeers failed, error: %s", resp.Error.String())
	default:
		return nil, fmt.Errorf("GetPeers failed, unknown error")
	}
}

// XXX: Can I use nickname?
func PeerName(p *meshpb.Peer) string {
	name := p.Hostname
	if p.Nickname != "" {
		name = p.Nickname
	}
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\x00' {
			return '_'
		}
		return r
	}, name)
}

// XXX: probably shouldn't be here
func StableIno(parts ...string) uint64 {
	h := fnv.New64a()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	ino := h.Sum64()
	// 0 means "auto-assign" and ^uint64(0) is reserved - can't use
	if ino == 0 || ino == ^uint64(0) {
		ino = 1
	}
	return ino
}

func PeerAddr(p *meshpb.Peer) (netip.Addr, error) {
	return netip.ParseAddr(p.Ip)
}
