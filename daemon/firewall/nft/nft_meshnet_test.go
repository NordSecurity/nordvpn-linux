package nft

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileshareBlockAllow(t *testing.T) {
	category.Set(t, category.Root)
	ns := helpers.OpenNewNamespace(t)
	defer helpers.CleanNamespace(t, ns)

	n := NewNft(0xe1f1)
	cfg := firewall.Config{
		MeshnetInfo: &firewall.MeshInfo{
			MeshInterface: "nordlynx",
			MeshnetMap: mesh.MachineMap{
				Peers: mesh.MachinePeers{
					mesh.MachinePeer{
						DoIAllowFileshare: true,
						Address:           netip.MustParseAddr("100.77.197.112"),
					},
				},
			},
		},
	}

	require.NoError(t, n.Configure(cfg))

	listMeshInput := helpers.ListChain(meshInputChainName)

	helpers.WithNftCommandOutput(t, listMeshInput, func(out string) {
		assert.Contains(t, out, fmt.Sprintf("tcp dport 49111 ip saddr @%s accept", fileshareAllowedPeersSet))
		assert.Contains(t, out, "tcp dport 49111 drop")
	})

	// block the fileshare
	cfg.BlockFileshare = true
	assert.NoError(t, n.Configure(cfg))

	helpers.WithNftCommandOutput(t, listMeshInput, func(out string) {
		assert.NotContains(t, out, fmt.Sprintf("tcp dport 49111 ip saddr @%s accept", fileshareAllowedPeersSet))
		assert.Contains(t, out, "tcp dport 49111 drop")
	})

	// allow fileshare
	cfg.BlockFileshare = false
	assert.NoError(t, n.Configure(cfg))

	helpers.WithNftCommandOutput(t, listMeshInput, func(out string) {
		assert.Contains(t, out, fmt.Sprintf("tcp dport 49111 ip saddr @%s accept", fileshareAllowedPeersSet))
		assert.Contains(t, out, "tcp dport 49111 drop")
	})
}
