package fileshare

import (
	"testing"

	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/stretchr/testify/assert"
)

func getTestMeshClient() meshpb.MeshnetClient {
	meshClient := mockMeshClient{}
	meshClient.externalPeers = []*meshpb.Peer{
		{
			Ip: "1.1.1.1",
		},
	}
	meshClient.localPeers = []*meshpb.Peer{
		{
			Ip: "2.2.2.2",
		},
	}
	meshClient.selfPeer = &meshpb.Peer{
		Ip: "3.3.3.3",
	}
	return &meshClient
}

func TestGetSelfPeer(t *testing.T) {
	peer, err := GetSelfPeer(getTestMeshClient())
	assert.Nil(t, err)
	assert.Equal(t, "3.3.3.3", peer.Ip)
}

func TestGetPeers(t *testing.T) {
	peers, err := getPeers(getTestMeshClient())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(peers))
}

func TestGetPeerByIP(t *testing.T) {
	peer, err := getPeerByIP(getTestMeshClient(), "1.1.1.1")
	assert.Nil(t, err)
	assert.NotNil(t, peer)
}

func TestGetPeerByIP_NotFound(t *testing.T) {
	_, err := getPeerByIP(getTestMeshClient(), "3.3.3.3")
	assert.NotNil(t, err)
}
