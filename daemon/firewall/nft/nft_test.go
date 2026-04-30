package nft

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/google/nftables"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetTestNft() *nft {
	return NewNft(0xe1f1).(*nft)
}

func TestConfigure(t *testing.T) {
	category.Set(t, category.Root)
	tests := []struct {
		name     string
		fwConfig firewall.Config
	}{
		{
			name: "only vpn interface",
			fwConfig: firewall.Config{
				TunnelInterface: "dummynlx",
				Allowlist:       config.Allowlist{Ports: config.Ports{}, Subnets: []string{}},
				KillSwitch:      false,
				MeshnetInfo:     nil,
			},
		},
		{
			name: "only killswitch",
			fwConfig: firewall.Config{
				TunnelInterface: "",
				Allowlist:       config.Allowlist{Ports: config.Ports{}, Subnets: []string{}},
				KillSwitch:      true,
				MeshnetInfo:     nil,
			},
		},
		// this should eventually contain all of the cases and be moved from python to here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple framework for testing the Configure func and corectness of set nft rules
			n := GetTestNft()
			ns := helpers.OpenNewNamespace(t)
			defer helpers.CleanNamespace(t, ns)

			require.NoError(t, n.Configure(tt.fwConfig))
			// Currently just checking if the table was created
			// When rules are finalized, we can start comparing hard coded expected strings to
			// whatever output we get after calling Configure()

			// Output can be checked via help of
			// exec.Command("nft", "list", "ruleset")
			table, err := n.conn.ListTableOfFamily(tableName, nftables.TableFamilyINet)
			if err != nil {
				t.Fatalf("unable to list default configured table %v", err)
			}
			assert.NotNil(t, table)
		})
	}
}

func TestAllowlistPortsMarkedBeforeAcceptedByOtherRule(t *testing.T) {
	category.Set(t, category.Root)
	ns := helpers.OpenNewNamespace(t)
	defer helpers.CleanNamespace(t, ns)

	tunnelIface := "nordlynx"
	n := GetTestNft()
	require.NoError(t, n.Configure(firewall.Config{
		TunnelInterface: tunnelIface,
		KillSwitch:      true,
		Allowlist: config.Allowlist{
			Ports: config.Ports{
				TCP: config.PortSet{55: true},
				UDP: config.PortSet{55: true},
			},
		},
		MeshnetInfo: &firewall.MeshInfo{
			MeshInterface: tunnelIface,
			MeshnetMap:    mesh.MachineMap{},
		},
	}))

	tcpPortRule := fmt.Sprintf("tcp sport @%s meta mark set 0x%08x accept", tcpAllowlistSetName, n.fwmark)
	udpPortRule := fmt.Sprintf("udp sport @%s meta mark set 0x%08x accept", udpAllowlistSetName, n.fwmark)
	vpnAccept := fmt.Sprintf(`oifname "%s" accept`, tunnelIface)
	meshAccept := fmt.Sprintf(`oifname "%s" ip daddr %s accept`, tunnelIface, internal.MeshSubnet)

	helpers.WithNftCommandOutput(t, helpers.ListChain(outputChainName), func(out string) {
		for _, portRule := range []string{tcpPortRule, udpPortRule} {
			helpers.AssertRulesOrder(t, out, portRule, vpnAccept)
			helpers.AssertRulesOrder(t, out, portRule, meshAccept)
		}
	})
}

func TestLanDNSDropBeforeAllowlistPorts(t *testing.T) {
	category.Set(t, category.Root)
	ns := helpers.OpenNewNamespace(t)
	defer helpers.CleanNamespace(t, ns)

	n := GetTestNft()
	require.NoError(t, n.Configure(firewall.Config{
		TunnelInterface: "nordlynx",
		KillSwitch:      true,
		Allowlist: config.Allowlist{
			Ports: config.Ports{
				TCP: config.PortSet{55: true},
				UDP: config.PortSet{55: true},
			},
		},
	}))

	tcpDNSDrop := fmt.Sprintf("ip daddr @%s tcp dport %d drop", lanPrivateIpsSetName, defaultDNSPort)
	udpDNSDrop := fmt.Sprintf("ip daddr @%s udp dport %d drop", lanPrivateIpsSetName, defaultDNSPort)
	tcpPortRule := fmt.Sprintf("tcp sport @%s meta mark set 0x%08x accept", tcpAllowlistSetName, n.fwmark)
	udpPortRule := fmt.Sprintf("udp sport @%s meta mark set 0x%08x accept", udpAllowlistSetName, n.fwmark)

	helpers.WithNftCommandOutput(t, helpers.ListChain(outputChainName), func(out string) {
		helpers.AssertRulesOrder(t, out, tcpDNSDrop, tcpPortRule)
		helpers.AssertRulesOrder(t, out, tcpDNSDrop, udpPortRule)
		helpers.AssertRulesOrder(t, out, udpDNSDrop, tcpPortRule)
		helpers.AssertRulesOrder(t, out, udpDNSDrop, udpPortRule)
	})
}
