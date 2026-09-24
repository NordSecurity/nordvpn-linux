package serverpicker

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

// ObfuscatedServersGroupTitle is the title of the synthesized "Obfuscated Servers" group. No server
// carries that group anymore, so the title is not taken from the API.
const ObfuscatedServersGroupTitle = "Obfuscated Servers"

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

// withObfuscatedAlias adds the Obfuscated group for obfuscated technologies.
func withObfuscatedAlias(groups core.Groups, s core.Server, tech config.Technology) core.Groups {
	if !IsObfuscatedTech(tech) {
		return groups
	}

	isStandard := slices.ContainsFunc(groups, core.ByGroup(config.ServerGroup_STANDARD_VPN_SERVERS))
	if !isStandard || !core.IsConnectableVia(core.NordWhisperTech)(s) {
		return groups
	}

	return append(groups, core.Group{
		ID:    config.ServerGroup_OBFUSCATED,
		Title: ObfuscatedServersGroupTitle,
	})
}
