package nft

import (
	"fmt"
	"net/netip"
	"os/exec"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/stretchr/testify/assert"
)

func runNftCommand(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("nft", args...).Output()

	assert.NoError(t, err)
	return string(out)
}

func withNftCommandOutput(t *testing.T, args []string, fn func(out string)) {
	t.Helper()
	out := runNftCommand(t, args...)
	fn(out)
}

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

	assert.NoError(t, n.Configure(cfg))

	args := []string{"list", "chain", "inet", "nordvpn", "mesh_input"}

	withNftCommandOutput(t, args, func(out string) {
		assert.Contains(t, out, fmt.Sprintf("tcp dport 49111 ip saddr @%s accept", fileshareAllowedPeersSet))
		assert.Contains(t, out, "tcp dport 49111 drop")
	})

	// block the fileshare
	cfg.MeshnetInfo.BlockFileshare = true
	assert.NoError(t, n.Configure(cfg))

	withNftCommandOutput(t, args, func(out string) {
		assert.NotContains(t, out, fmt.Sprintf("tcp dport 49111 ip saddr @%s accept", fileshareAllowedPeersSet))
		assert.Contains(t, out, "tcp dport 49111 drop", fileshareAllowedPeersSet)
	})

	// allow fileshare
	cfg.MeshnetInfo.BlockFileshare = false
	assert.NoError(t, n.Configure(cfg))

	withNftCommandOutput(t, args, func(out string) {
		assert.Contains(t, out, fmt.Sprintf("tcp dport 49111 ip saddr @%s accept", fileshareAllowedPeersSet))
		assert.Contains(t, out, "tcp dport 49111 drop", fileshareAllowedPeersSet)
	})
}
