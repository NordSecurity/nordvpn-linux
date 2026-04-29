package nft

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/golden"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/stretchr/testify/require"
)

const (
	peerIP = "100.113.144.142"
	ifName = "nordlynx"
)

func TestVPNRuleset(t *testing.T) {
	t.Run("kill_switch_only", func(t *testing.T) {
		runSnapshotTest(t, helpers.NewFWConfig().KillSwitch())
	})

	t.Run("vpn_only", func(t *testing.T) {
		runSnapshotTest(t, helpers.NewFWConfig().TunnelInterface(ifName))
	})

	t.Run("vpn_and_kill_switch", func(t *testing.T) {
		runSnapshotTest(t, helpers.NewFWConfig().TunnelInterface(ifName).KillSwitch())
	})

	t.Run("tcp_port_allowlisted", func(t *testing.T) {
		runSnapshotTest(t, helpers.NewFWConfig().TunnelInterface(ifName).AllowlistTCPPort(1337))
	})

	t.Run("udp_port_allowlisted", func(t *testing.T) {
		runSnapshotTest(t, helpers.NewFWConfig().TunnelInterface(ifName).AllowlistUDPPort(8080))
	})

	t.Run("subnet_allowlisted", func(t *testing.T) {
		runSnapshotTest(t, helpers.NewFWConfig().TunnelInterface(ifName).AllowlistSubnet("10.0.0.0/24"))
	})
}

func TestMeshnetRuleset(t *testing.T) {
	t.Run("without_peers", func(t *testing.T) {
		runSnapshotTest(t, helpers.NewFWConfig().Meshnet(ifName))
	})

	t.Run("peer_with_lan_access", func(t *testing.T) {
		runSnapshotTest(t,
			helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:              netip.MustParseAddr(peerIP),
					DoIAllowLocalNetwork: true,
					DoIAllowInbound:      true,
				}),
		)
	})

	t.Run("host_allows_routing", func(t *testing.T) {
		runSnapshotTest(t,
			helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:         netip.MustParseAddr(peerIP),
					DoIAllowRouting: true,
				}),
		)
	})

	t.Run("host_allows_inbound_but_no_routing", func(t *testing.T) {
		runSnapshotTest(t,
			helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:         netip.MustParseAddr(peerIP),
					DoIAllowInbound: true,
					DoIAllowRouting: false,
				}),
		)
	})

	t.Run("with_fileshare", func(t *testing.T) {
		runSnapshotTest(t,
			helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:           netip.MustParseAddr(peerIP),
					DoIAllowFileshare: true,
				}),
		)
	})

	t.Run("with_blocked_fileshare", func(t *testing.T) {
		runSnapshotTest(t,
			helpers.NewFWConfig().
				Meshnet(ifName).
				BlockFileshare().
				MeshPeer(mesh.MachinePeer{
					Address:           netip.MustParseAddr(peerIP),
					DoIAllowFileshare: true,
				}),
		)
	})

	t.Run("peer_with_full_permissions", func(t *testing.T) {
		runSnapshotTest(t,
			helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:              netip.MustParseAddr(peerIP),
					DoIAllowFileshare:    true,
					DoIAllowInbound:      true,
					DoIAllowRouting:      true,
					DoIAllowLocalNetwork: true,
				}),
		)
	})
}

func runSnapshotTest(t *testing.T, b *helpers.FirewallConfigBuilder) {
	t.Helper()
	category.Set(t, category.Root)
	ns := helpers.OpenNewNamespace(t)
	defer helpers.CleanNamespace(t, ns)

	n := GetTestNft()
	require.NoError(t, n.Configure(b.Build()))

	helpers.WithNftCommandOutput(t, helpers.ListTable(tableName), func(out string) {
		golden.AssertMatchesGolden(t, out)
	})
}
