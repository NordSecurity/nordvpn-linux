package config

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestAvailableVPNProtocols_NeverContainsUnspecified(t *testing.T) {
	category.Set(t, category.Unit)

	assert.NotContains(t, AvailableVPNProtocols(), VPNProtocol_VPN_PROTOCOL_UNSPECIFIED)
}

func TestVPNProtocol_MapsToTechnologyTransportAndDisplayName(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		protocol    VPNProtocol
		technology  Technology
		transport   Transport
		displayName string
	}{
		{VPNProtocol_VPN_PROTOCOL_NORDLYNX, TechnologyNordLynx, TransportUnknown, "NordLynx"},
		{VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP, TechnologyOpenVPN, TransportUDP, "OpenVPN (UDP)"},
		{VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP, TechnologyOpenVPN, TransportTCP, "OpenVPN (TCP)"},
		{VPNProtocol_VPN_PROTOCOL_NORDWHISPER, TechnologyNordWhisper, TransportWebTunnel, "NordWhisper"},
		{VPNProtocol_VPN_PROTOCOL_UNSPECIFIED, TechnologyUnknown, TransportUnknown, "VPN_PROTOCOL_UNSPECIFIED"},
		{VPNProtocol(99), TechnologyUnknown, TransportUnknown, "VPN_PROTOCOL_UNSPECIFIED"},
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
