package networker

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
)

// addLANPermissions creates a new Allowlist. Subnets map is copied and updated with LANs, Port maps
// remain unchanged.
func addLANPermissions(allowlist config.Allowlist) config.Allowlist {
	localNetworks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16"}

	var newSubnets []string
	copy(newSubnets, allowlist.Subnets)
	for _, network := range localNetworks {
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
