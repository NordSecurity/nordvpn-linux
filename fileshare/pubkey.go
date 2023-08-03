package fileshare

import (
	"encoding/base64"
	"log"

	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

// PubkeyProvider for libdrop
type PubkeyProvider struct {
	meshClient meshpb.MeshnetClient
}

// NewPubkeyProvider must be used to create PubkeyProvider
func NewPubkeyProvider(meshClient meshpb.MeshnetClient) *PubkeyProvider {
	return &PubkeyProvider{meshClient}
}

// PubkeyFunc is called by libdrop on incoming requests to verify their validity
func (c *PubkeyProvider) PubkeyFunc(peerIP string) []byte {
	peers, err := getPeers(c.meshClient)
	if err != nil {
		log.Print(err)
	}

	for _, peer := range peers {
		if peer.Ip == peerIP {
			pubkeyBytes, err := base64.StdEncoding.DecodeString(peer.Pubkey)
			if err != nil || len(pubkeyBytes) != 32 { // libdrop gives exactly 32 bytes buffer to write pubkey
				log.Printf("invalid pubkey %s: %v", peer.Pubkey, err)
				return make([]byte, 32)
			}
			return pubkeyBytes
		}
	}

	log.Printf("couldn't find pubkey for ip %s", peerIP)
	return make([]byte, 32)
}
