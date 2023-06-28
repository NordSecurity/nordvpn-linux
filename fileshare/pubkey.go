package fileshare

import (
	"encoding/base64"
	"log"

	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

// PubkeyProvider for libdrop
type PubkeyProvider struct {
	meshClient  meshpb.MeshnetClient
	pubkeyCache map[string][]byte // ip to pubkey
}

// NewPubkeyProvider must be used to create PubkeyProvider
func NewPubkeyProvider(meshClient meshpb.MeshnetClient) *PubkeyProvider {
	return &PubkeyProvider{meshClient, map[string][]byte{}}
}

// PubkeyFunc is called by libdrop on incoming requests to verify their validity
func (c *PubkeyProvider) PubkeyFunc(peerIP string) []byte {
	pubkeyInternal, ok := c.pubkeyCache[peerIP]
	if !ok {
		c.updateCache()
		pubkeyInternal, ok = c.pubkeyCache[peerIP]
		if !ok {
			log.Printf("can't provide pubkey for ip %s", peerIP)
			return nil
		}
	}

	return pubkeyInternal
}

func (c *PubkeyProvider) updateCache() {
	peers, err := getPeers(c.meshClient)
	if err != nil {
		log.Print(err)
	}

	c.pubkeyCache = map[string][]byte{}
	for _, peer := range peers {
		pubkeyBytes, err := base64.StdEncoding.DecodeString(peer.Pubkey)
		if err != nil || len(pubkeyBytes) != 32 { // libdrop gives exactly 32 bytes buffer to write pubkey
			log.Printf("invalid pubkey %s: %v", peer.Pubkey, err)
			continue
		}
		c.pubkeyCache[peer.Ip] = pubkeyBytes
	}
}
