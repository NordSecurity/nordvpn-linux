package daemon

import (
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

	newSubnets := make(config.Subnets)

	for subnet := range allowlist.Subnets {
		newSubnets[subnet] = true
	}

	for _, network := range localNetworks {
		if _, ok := newSubnets[network]; !ok {
			newSubnets[network] = true
		}
	}

	newAllowlist := config.Allowlist{
		Ports:   allowlist.Ports,
		Subnets: newSubnets,
	}

	return newAllowlist
}
