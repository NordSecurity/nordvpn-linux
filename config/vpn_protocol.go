package config

import (
	"encoding/json"

	"github.com/NordSecurity/nordvpn-linux/features"
)

// Technology returns technology used for the selected vpn protocol
func (p VPNProtocol) Technology() Technology {
	//exhaustive:ignore
	switch p {
	case VPNProtocol_VPN_PROTOCOL_NORDLYNX:
		return TechnologyNordLynx
	case VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP, VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP:
		return TechnologyOpenVPN
	case VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
		return TechnologyNordWhisper
	}
	return TechnologyUnknown
}

// Transport returns transport used for the selected vpn protocol
func (p VPNProtocol) Transport() Protocol {
	//exhaustive:ignore
	switch p {
	case VPNProtocol_VPN_PROTOCOL_NORDLYNX:
		return Protocol_UNKNOWN_PROTOCOL
	case VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP:
		return Protocol_TCP
	case VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP:
		return Protocol_UDP
	case VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
		return Protocol_Webtunnel
	}
	return Protocol_UNKNOWN_PROTOCOL
}

// DisplayName converts vpn protocol to string representation
func (p VPNProtocol) DisplayName() string {
	//exhaustive:ignore
	switch p {
	case VPNProtocol_VPN_PROTOCOL_NORDLYNX:
		return "NordLynx"
	case VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP:
		return "OpenVPN (TCP)"
	case VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP:
		return "OpenVPN (UDP)"
	case VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
		return "NordWhisper (WebTunnel)"
	}
	return VPNProtocol_VPN_PROTOCOL_UNSPECIFIED.String()
}

// IsNordLynx reports whether p uses the NordLynx technology.
func (p VPNProtocol) IsNordLynx() bool {
	return p.Technology() == TechnologyNordLynx
}

// IsOpenVPN reports whether p uses the OpenVPN technology (TCP or UDP).
func (p VPNProtocol) IsOpenVPN() bool {
	return p.Technology() == TechnologyOpenVPN
}

// IsNordWhisper reports whether p uses the NordWhisper technology.
func (p VPNProtocol) IsNordWhisper() bool {
	return p.Technology() == TechnologyNordWhisper
}

// vpnProtocolFromLegacy maps the technology and protocol stored by versions before vpn_protocol.
func vpnProtocolFromLegacy(tech Technology, proto Protocol) VPNProtocol {
	//exhaustive:ignore
	switch tech {
	case TechnologyOpenVPN:
		if proto == Protocol_TCP {
			return VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP
		}
		return VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP
	case TechnologyNordWhisper:
		return VPNProtocol_VPN_PROTOCOL_NORDWHISPER
	}
	return VPNProtocol_VPN_PROTOCOL_NORDLYNX
}

// migrateVPNProtocol sets an unset VPN protocol from the legacy config keys.
func migrateVPNProtocol(c *Config, data []byte) {
	if c.VPNProtocol != VPNProtocol_VPN_PROTOCOL_UNSPECIFIED {
		return
	}
	var legacy struct {
		Technology      Technology `json:"technology"`
		AutoConnectData struct {
			Protocol Protocol `json:"protocol"`
		} `json:"auto_connect_data"`
	}
	_ = json.Unmarshal(data, &legacy)
	c.VPNProtocol = vpnProtocolFromLegacy(legacy.Technology, legacy.AutoConnectData.Protocol)
}

// AvailableVPNProtocols returns VPN protocols supported by this build, in display order.
func AvailableVPNProtocols() []VPNProtocol {
	protocols := []VPNProtocol{
		VPNProtocol_VPN_PROTOCOL_NORDLYNX,
		VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
	}

	if features.NordWhisperEnabled {
		protocols = append(protocols, VPNProtocol_VPN_PROTOCOL_NORDWHISPER)
	}

	return protocols
}
