package nft

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/google/nftables"
	"github.com/stretchr/testify/assert"
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
				Allowlist:       config.Allowlist{config.Ports{}, []string{}},
				KillSwitch:      false,
				MeshnetInfo:     nil,
			},
		},
		{
			name: "only killswitch",
			fwConfig: firewall.Config{
				TunnelInterface: "",
				Allowlist:       config.Allowlist{config.Ports{}, []string{}},
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

			n.Configure(tt.fwConfig)
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
