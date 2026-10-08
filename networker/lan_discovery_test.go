package networker

import (
	"slices"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestAddLANDiscoverySubnets(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		inputSubnets    []string
		expectedSubnets []string
	}{
		{
			name:            "empty allowlist gets LAN and mDNS subnets",
			inputSubnets:    nil,
			expectedSubnets: append(internal.LocalNetworks, internal.MDNSSubnet),
		},
		{
			name:            "existing non-private subnets are preserved",
			inputSubnets:    []string{"1.1.1.1/32"},
			expectedSubnets: append([]string{"1.1.1.1/32"}, append(internal.LocalNetworks, internal.MDNSSubnet)...),
		},
		{
			name:            "duplicate LAN subnets are not added twice",
			inputSubnets:    []string{"10.0.0.0/8"},
			expectedSubnets: append(internal.LocalNetworks, internal.MDNSSubnet),
		},
		{
			name:            "duplicate mDNS subnet is not added twice",
			inputSubnets:    []string{internal.MDNSSubnet},
			expectedSubnets: append(internal.LocalNetworks, internal.MDNSSubnet),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputCopy := slices.Clone(test.inputSubnets)
			allowlist := config.NewAllowlist([]int64{80}, []int64{443}, test.inputSubnets)
			result := addLANDiscoverySubnets(allowlist)

			for _, expected := range test.expectedSubnets {
				assert.True(t, slices.Contains(result.Subnets, expected))
			}

			seen := map[string]bool{}
			for _, s := range result.Subnets {
				assert.False(t, seen[s])
				seen[s] = true
			}

			assert.Equal(t, allowlist.Ports, result.Ports)
			assert.Equal(t, inputCopy, test.inputSubnets)
		})
	}
}
