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
			name:   "kill switch only",
			config: helpers.NewFWConfig().KillSwitch(),
		},
		{
			name:   "vpn only",
			config: helpers.NewFWConfig().TunnelInterface(ifName),
		},
		{
			name:   "vpn and kill switch",
			config: helpers.NewFWConfig().TunnelInterface(ifName).KillSwitch(),
		},
		{
			name:   "tcp port allowlisted",
			config: helpers.NewFWConfig().TunnelInterface(ifName).AllowlistTCPPort(1337),
		},
		{
			name:   "udp port allowlisted",
			config: helpers.NewFWConfig().TunnelInterface(ifName).AllowlistUDPPort(8080),
		},
		{
			name: "DNS port allowlisted for both protocols",
			config: helpers.NewFWConfig().
				TunnelInterface(ifName).
				AllowlistUDPPort(53).
				AllowlistTCPPort(53),
		},
		{
			name:   "subnet allowlisted",
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
			name:   "without peers",
			config: helpers.NewFWConfig().Meshnet(ifName),
		},
		{
			name: "peer with lan_access",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:              netip.MustParseAddr(peerIP),
					DoIAllowLocalNetwork: true,
					DoIAllowInbound:      true,
				}),
		},
		{
			name: "host allows routing",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:         netip.MustParseAddr(peerIP),
					DoIAllowRouting: true,
				}),
		},
		{
			name: "host allows inbound but_no_routing",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:         netip.MustParseAddr(peerIP),
					DoIAllowInbound: true,
					DoIAllowRouting: false,
				}),
		},
		{
			name: "with fileshare",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				MeshPeer(mesh.MachinePeer{
					Address:           netip.MustParseAddr(peerIP),
					DoIAllowFileshare: true,
				}),
		},
		{
			name: "with blocked fileshare",
			config: helpers.NewFWConfig().
				Meshnet(ifName).
				BlockFileshare().
				MeshPeer(mesh.MachinePeer{
					Address:           netip.MustParseAddr(peerIP),
					DoIAllowFileshare: true,
				}),
		},
		{
			name: "peer with full permissions",
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

	n := NewNft(0xe1f1)
	require.NoError(t, n.Configure(b.Build()))

	helpers.WithNftCommandOutput(t, helpers.ListTable(tableName), func(out string) {
		golden.AssertMatchesGolden(t, out)
	})
}
