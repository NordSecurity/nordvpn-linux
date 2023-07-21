package daemon

import "github.com/NordSecurity/nordvpn-linux/config"

// addLANPermissions adds or removes LANs to the whitelist
func addLANPermissions(whitelist config.Allowlist) config.Allowlist {
	localNetworks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16"}

	for _, network := range localNetworks {
		if _, ok := whitelist.Subnets[network]; !ok {
			whitelist.Subnets[network] = true
		}
	}

	return whitelist
}
