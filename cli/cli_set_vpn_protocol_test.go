package cli

import (
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestVPNProtocolToArg_ReturnsLowercaseNameWithoutPrefix(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		protocol config.VPNProtocol
		expected string
	}{
		{config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, "nordlynx"},
		{config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP, "openvpn_udp"},
		{config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP, "openvpn_tcp"},
		{config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER, "nordwhisper"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, vpnProtocolToArg(test.protocol), "protocol %v", test.protocol)
	}
}

func TestVPNProtocolFromArg_RoundTripsAvailableProtocols(t *testing.T) {
	category.Set(t, category.Unit)

	for _, protocol := range config.AvailableVPNProtocols() {
		parsed, ok := vpnProtocolFromArg(vpnProtocolToArg(protocol))

		assert.True(t, ok, "protocol %v", protocol)
		assert.Equal(t, protocol, parsed, "protocol %v", protocol)
	}
}

func TestVPNProtocolFromArg_IsCaseInsensitive(t *testing.T) {
	category.Set(t, category.Unit)

	for _, arg := range []string{"OPENVPN_TCP", "OpenVPN_TCP", "openvpn_TCP"} {
		parsed, ok := vpnProtocolFromArg(arg)

		assert.True(t, ok, "arg %q", arg)
		assert.Equal(t, config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP, parsed, "arg %q", arg)
	}
}

func TestVPNProtocolFromArg_RejectsInvalidArgs(t *testing.T) {
	category.Set(t, category.Unit)

	invalid := []string{
		"",
		"foo",
		"unspecified",
		"UNSPECIFIED",
		"udp",
		"tcp",
		"openvpn",
		"vpn_protocol_nordlynx",
		" nordlynx",
	}

	for _, arg := range invalid {
		parsed, ok := vpnProtocolFromArg(arg)

		assert.False(t, ok, "arg %q", arg)
		assert.Equal(t, config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED, parsed, "arg %q", arg)
	}
}

func TestBuildVPNProtocolDescription_ListsAvailableProtocolsAndExample(t *testing.T) {
	category.Set(t, category.Unit)

	description := buildVPNProtocolDescription()
	args := vpnProtocolArgs()

	assert.Contains(t, description, "Supported values for <protocol>: "+strings.Join(args, ", "))
	assert.Contains(t, description, "'nordvpn set protocol "+args[0]+"'")
	assert.NotContains(t, description, "%s", "all format verbs must be filled")
}
