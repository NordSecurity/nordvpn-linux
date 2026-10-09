package serverpicker

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"golang.org/x/exp/slices"
)

// MatchesUserSettings reports whether the server can be connected to with the
// technology and protocol from cfg.
func MatchesUserSettings(s core.Server, cfg config.Config) bool {
	return core.IsConnectableWithProtocol(cfg.VPNProtocol)(s)
}

// selectFilterForLocalServers - it will return a filter function that is compatible only with local cached server
// This is because most of the checks are based on server.Key, which is only computed for cached list
func selectFilterForLocalServers(tag string, group config.ServerGroup) core.Predicate {
	if tag != "" && group != config.ServerGroup_UNDEFINED {
		return func(s core.Server) bool {
			return slices.ContainsFunc(s.Groups, core.ByGroup(group)) && slices.Contains(s.Keys, tag)
		}
	}

	if group != config.ServerGroup_UNDEFINED {
		return func(s core.Server) bool {
			return slices.ContainsFunc(s.Groups, core.ByGroup(group))
		}
	}

	if tag != "" {
		return func(s core.Server) bool {
			return slices.Contains(s.Keys, tag)
		}
	}

	return func(s core.Server) bool {
		return slices.ContainsFunc(s.Groups, core.ByGroup(config.ServerGroup_STANDARD_VPN_SERVERS))
	}
}

// TechToServerTech maps the user connection settings to the corresponding core
// server technology. It is exported because it is also used directly by the
// daemon (e.g. when filtering recent connections).
func TechToServerTech(p config.VPNProtocol) core.ServerTechnology {
	//exhaustive:ignore
	switch p {
	case config.VPNProtocol_VPN_PROTOCOL_NORDLYNX:
		return core.WireguardTech
	case config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP:
		return core.OpenVPNTCP
	case config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP:
		return core.OpenVPNUDP
	case config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
		return core.NordWhisperTech
	}
	return core.Unknown
}
