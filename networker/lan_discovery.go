package networker

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// addLANDiscoverySubnets returns a copy of the given Allowlist with local
// network and mDNS subnets appended (if not already present) for LAN service
// discovery. Ports are shared with the original; Subnets are copied.
func addLANDiscoverySubnets(allowlist config.Allowlist) config.Allowlist {
	newSubnets := append([]string{}, allowlist.Subnets...)
	for _, network := range slices.Concat(internal.LocalNetworks, []string{internal.MDNSSubnet}) {
		if !slices.Contains(newSubnets, network) {
			newSubnets = append(newSubnets, network)
		}
	}

	newAllowlist := config.Allowlist{
		Ports:   allowlist.Ports,
		Subnets: newSubnets,
	}

	return newAllowlist
}
