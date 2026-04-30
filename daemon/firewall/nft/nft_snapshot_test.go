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
	tests := []struct {
		name   string
		config *helpers.FirewallConfigBuilder
	}{
		{
			name:   "kill_switch_only",
			config: helpers.NewFWConfig().KillSwitch(),
		},
		{
			name:   "vpn_only",
			config: helpers.NewFWConfig().TunnelInterface(ifName),
		},
		{
			name:   "vpn_and_kill_switch",
			config: helpers.NewFWConfig().TunnelInterface(ifName).KillSwitch(),
		},
		{
			name:   "tcp_port_allowlisted",
			config: helpers.NewFWConfig().TunnelInterface(ifName).AllowlistTCPPort(1337),
		},
		{
			name:   "udp_port_allowlisted",
			config: helpers.NewFWConfig().TunnelInterface(ifName).AllowlistUDPPort(8080),
		},
		{
			name:   "subnet_allowlisted",
			config: helpers.NewFWConfig().TunnelInterface(ifName).AllowlistSubnet("10.0.0.0/24"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runSnapshotTest(t, tt.config)
		})
	}
}

func TestMeshnetRuleset(t *testing.T) {
	tests := []struct {
		name   string
		config *helpers.FirewallConfigBuilder
	}{
		{
			name:   "without_peers",
			config: helpers.NewFWConfig().Meshnet(ifName),
		},
		{
			name: "peer_with_lan_access",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:              netip.MustParseAddr(peerIP),
					DoIAllowLocalNetwork: true,
					DoIAllowInbound:      true,
				}),
		},
		{
			name: "host_allows_routing",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:         netip.MustParseAddr(peerIP),
					DoIAllowRouting: true,
				}),
		},
		{
			name: "host_allows_inbound_but_no_routing",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:         netip.MustParseAddr(peerIP),
					DoIAllowInbound: true,
					DoIAllowRouting: false,
				}),
		},
		{
			name: "with_fileshare",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:           netip.MustParseAddr(peerIP),
					DoIAllowFileshare: true,
				}),
		},
		{
			name: "with_blocked_fileshare",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				BlockFileshare().
				MeshPeer(mesh.MachinePeer{
					Address:           netip.MustParseAddr(peerIP),
					DoIAllowFileshare: true,
				}),
		},
		{
			name: "peer_with_full_permissions",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:              netip.MustParseAddr(peerIP),
					DoIAllowFileshare:    true,
					DoIAllowInbound:      true,
					DoIAllowRouting:      true,
					DoIAllowLocalNetwork: true,
				}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runSnapshotTest(t, tt.config)
		})
	}
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
