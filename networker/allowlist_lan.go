package networker

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// addLANPermissions creates a new Allowlist. Subnets map is copied and updated with LANs, Port maps
// remain unchanged.
func addLANPermissions(allowlist config.Allowlist) config.Allowlist {
	newSubnets := append([]string{}, allowlist.Subnets...)
	for _, network := range internal.LocalNetworks {
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
