package fileshare

import (
	"encoding/base64"
	"testing"

	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/stretchr/testify/assert"
)

func TestPubkeyFunc(t *testing.T) {
	meshClient := mockMeshClient{}
	pubkey := "PWOUuWuCLIj6xCbPOCBaMA2ez29g8aTMuYkCQx9kfj4="
	meshClient.externalPeers = []*meshpb.Peer{
		{
			Ip:     "1.2.3.4",
			Pubkey: pubkey,
		},
	}
	pubkeyProvider := NewPubkeyProvider(&meshClient)
	pubkeyActual := pubkeyProvider.PubkeyFunc("1.2.3.4")
	assert.Equal(t, pubkey, base64.StdEncoding.EncodeToString(pubkeyActual))
	assert.True(t, meshClient.getPeersCalled)
}
