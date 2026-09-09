package serverpicker

import "github.com/NordSecurity/nordvpn-linux/config"

// IsObfuscatedTech reports whether connections over the technology are obfuscated.
func IsObfuscatedTech(tech config.Technology) bool {
	return tech == config.Technology_NORDWHISPER
}

// searchGroup returns the group to look servers up by for the requested one.
func searchGroup(requested config.ServerGroup, tech config.Technology) config.ServerGroup {
	if requested == config.ServerGroup_OBFUSCATED && IsObfuscatedTech(tech) {
		return config.ServerGroup_STANDARD_VPN_SERVERS
	}

	return requested
}
