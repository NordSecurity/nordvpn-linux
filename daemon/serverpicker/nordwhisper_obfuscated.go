package serverpicker

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

// ObfuscatedServersGroupTitle is the title of the synthesized "Obfuscated" group. No server
// carries that group anymore, so the title is not taken from the API.
const ObfuscatedServersGroupTitle = "Obfuscated"

// IsObfuscatedProtocol reports whether connections over the protocol are obfuscated.
func IsObfuscatedProtocol(protocol config.VPNProtocol) bool {
	return protocol == config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER
}

// searchGroup returns the group to look servers up by for the requested one.
func searchGroup(requested config.ServerGroup, protocol config.VPNProtocol) config.ServerGroup {
	if requested == config.ServerGroup_NW_OBFUSCATED && IsObfuscatedProtocol(protocol) {
		return config.ServerGroup_STANDARD_VPN_SERVERS
	}

	return requested
}

// withObfuscatedAlias adds the Obfuscated group for obfuscated protocol.
func withObfuscatedAlias(
	groups core.Groups,
	s core.Server,
	protocol config.VPNProtocol,
) core.Groups {
	if !IsObfuscatedProtocol(protocol) {
		return groups
	}

	isStandard := slices.ContainsFunc(groups, core.ByGroup(config.ServerGroup_STANDARD_VPN_SERVERS))
	if !isStandard || !core.IsConnectableVia(core.NordWhisperTech)(s) {
		return groups
	}

	return append(groups, core.Group{
		ID:    config.ServerGroup_NW_OBFUSCATED,
		Title: ObfuscatedServersGroupTitle,
	})
}
