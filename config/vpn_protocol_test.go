package config

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestAvailableVPNProtocols_ReturnsBuildProtocolsInDisplayOrder(t *testing.T) {
	category.Set(t, category.Unit)

	expected := []VPNProtocol{
		VPNProtocol_VPN_PROTOCOL_NORDLYNX,
		VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
	}
	if features.NordWhisperEnabled {
		expected = append(expected, VPNProtocol_VPN_PROTOCOL_NORDWHISPER)
	}

	assert.Equal(t, expected, AvailableVPNProtocols())
}

func TestAvailableVPNProtocols_NeverContainsUnspecified(t *testing.T) {
	category.Set(t, category.Unit)

	assert.NotContains(t, AvailableVPNProtocols(), VPNProtocol_VPN_PROTOCOL_UNSPECIFIED)
}

func TestVPNProtocol_MapsToTechnologyTransportAndDisplayName(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		protocol    VPNProtocol
		technology  Technology
		transport   Protocol
		displayName string
	}{
		{VPNProtocol_VPN_PROTOCOL_NORDLYNX, Technology_NORDLYNX, Protocol_UNKNOWN_PROTOCOL, "NordLynx"},
		{VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP, Technology_OPENVPN, Protocol_UDP, "OpenVPN (UDP)"},
		{VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP, Technology_OPENVPN, Protocol_TCP, "OpenVPN (TCP)"},
		{VPNProtocol_VPN_PROTOCOL_NORDWHISPER, Technology_NORDWHISPER, Protocol_Webtunnel, "NordWhisper (WebTunnel)"},
		{VPNProtocol_VPN_PROTOCOL_UNSPECIFIED, Technology_UNKNOWN_TECHNOLOGY, Protocol_UNKNOWN_PROTOCOL, "VPN_PROTOCOL_UNSPECIFIED"},
		{VPNProtocol(99), Technology_UNKNOWN_TECHNOLOGY, Protocol_UNKNOWN_PROTOCOL, "VPN_PROTOCOL_UNSPECIFIED"},
	}

	for _, test := range tests {
		assert.Equal(t, test.technology, test.protocol.Technology(), "Technology() of %v", test.protocol)
		assert.Equal(t, test.transport, test.protocol.Transport(), "Transport() of %v", test.protocol)
		assert.Equal(t, test.displayName, test.protocol.DisplayName(), "DisplayName() of %v", test.protocol)
	}
}

func TestVPNProtocol_TechnologyPredicates(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		protocol    VPNProtocol
		nordLynx    bool
		openVPN     bool
		nordWhisper bool
	}{
		{VPNProtocol_VPN_PROTOCOL_NORDLYNX, true, false, false},
		{VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP, false, true, false},
		{VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP, false, true, false},
		{VPNProtocol_VPN_PROTOCOL_NORDWHISPER, false, false, true},
		{VPNProtocol_VPN_PROTOCOL_UNSPECIFIED, false, false, false},
	}

	for _, test := range tests {
		assert.Equal(t, test.nordLynx, test.protocol.IsNordLynx(), "IsNordLynx() of %v", test.protocol)
		assert.Equal(t, test.openVPN, test.protocol.IsOpenVPN(), "IsOpenVPN() of %v", test.protocol)
		assert.Equal(t, test.nordWhisper, test.protocol.IsNordWhisper(), "IsNordWhisper() of %v", test.protocol)
	}
}

func TestVPNProtocolFromLegacy(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		tech     Technology
		proto    Protocol
		expected VPNProtocol
	}{
		{Technology_NORDLYNX, Protocol_UDP, VPNProtocol_VPN_PROTOCOL_NORDLYNX},
		{Technology_NORDLYNX, Protocol_TCP, VPNProtocol_VPN_PROTOCOL_NORDLYNX},
		{Technology_OPENVPN, Protocol_TCP, VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP},
		{Technology_OPENVPN, Protocol_UDP, VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP},
		{Technology_OPENVPN, Protocol_UNKNOWN_PROTOCOL, VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP},
		{Technology_NORDWHISPER, Protocol_Webtunnel, VPNProtocol_VPN_PROTOCOL_NORDWHISPER},
		{Technology_UNKNOWN_TECHNOLOGY, Protocol_UDP, VPNProtocol_VPN_PROTOCOL_NORDLYNX},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, vpnProtocolFromLegacy(test.tech, test.proto), "%v + %v", test.tech, test.proto)
	}
}
