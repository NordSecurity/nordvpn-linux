package config

import (
	"encoding/json"

	"github.com/NordSecurity/nordvpn-linux/features"
)

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
		return "NordWhisper"
	}
	return VPNProtocol_VPN_PROTOCOL_UNSPECIFIED.String()
}

// TechNameToUpperCamelCase returns the technology name of the vpn protocol as an UpperCamelCase string
func TechNameToUpperCamelCase(p VPNProtocol) string {
	//exhaustive:ignore
	switch p {
	case VPNProtocol_VPN_PROTOCOL_NORDLYNX:
		return "NordLynx"
	case VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP, VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP:
		return "OpenVPN"
	case VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
		return "NordWhisper"
	}
	return ""
}

// IsNordLynx reports whether p uses the NordLynx protocol.
func (p VPNProtocol) IsNordLynx() bool {
	return p == VPNProtocol_VPN_PROTOCOL_NORDLYNX
}

// IsOpenVPN reports whether p uses the OpenVPN protocol family.
func (p VPNProtocol) IsOpenVPN() bool {
	return p == VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP || p == VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP
}

// IsNordWhisper reports whether p uses the NordWhisper protocol.
func (p VPNProtocol) IsNordWhisper() bool {
	return p == VPNProtocol_VPN_PROTOCOL_NORDWHISPER
}

// technology and protocol values stored by versions before vpn_protocol
const (
	legacyTechnologyOpenVPN     = 1
	legacyTechnologyNordWhisper = 3
	legacyProtocolTCP           = 2
)

// vpnProtocolFromLegacy maps the technology and protocol stored by versions before vpn_protocol.
func vpnProtocolFromLegacy(tech int32, proto int32) VPNProtocol {
	switch tech {
	case legacyTechnologyOpenVPN:
		if proto == legacyProtocolTCP {
			return VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP
		}
		return VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP
	case legacyTechnologyNordWhisper:
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
		Technology      int32 `json:"technology"`
		AutoConnectData struct {
			Protocol int32 `json:"protocol"`
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
