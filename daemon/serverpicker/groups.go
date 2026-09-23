package serverpicker

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

// EffectiveGroups returns the groups the server offers to a user of the given technology.
func EffectiveGroups(s core.Server, tech config.Technology) core.Groups {
	groups := withoutRetiredGroups(s.Groups)
	return withObfuscatedAlias(groups, s, tech)
}

// withoutRetiredGroups drops the OBFUSCATED tag the API puts on the retired OpenVPN XOR servers.
// Those servers are not connectable anymore, so the tag carries no meaning for the app.
func withoutRetiredGroups(groups core.Groups) core.Groups {
	kept := make(core.Groups, 0, len(groups)+1)
	for _, g := range groups {
		if g.ID != config.ServerGroup_OBFUSCATED {
			kept = append(kept, g)
		}
	}
	return kept
}
