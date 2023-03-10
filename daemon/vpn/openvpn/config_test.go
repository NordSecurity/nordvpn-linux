package openvpn

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigIdentifier(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		protocol   config.Protocol
		obfuscated bool
		expected   openvpnID
	}{
		{
			protocol:   config.Protocol_UDP,
			obfuscated: true,
			expected:   techXORUDP,
		},
		{
			protocol:   config.Protocol_UDP,
			obfuscated: false,
			expected:   techUDP,
		},
		{
			protocol:   config.Protocol_TCP,
			obfuscated: true, expected: techXORTCP,
		},
		{
			protocol:   config.Protocol_TCP,
			obfuscated: false,
			expected:   techTCP,
		},
	}

	for _, test := range tests {
		t.Run(string(test.expected), func(t *testing.T) {
			got, err := getConfigIdentifier(test.protocol, test.obfuscated)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGenerateConfigXML(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		identifier openvpnID
		expected   string
	}{
		{
			identifier: techXORUDP,
			expected:   "<?xml version=\"1.0\"?>\n<?xml-stylesheet type=\"xml/xsl\"?>\n<config>\n  <ips>\n    <ip address=\"1.1.1.1\" />\n  </ips>\n  <technology identifier=\"openvpn_xor_udp\"/>\n</config>\n",
		},
		{
			identifier: techUDP,
			expected:   "<?xml version=\"1.0\"?>\n<?xml-stylesheet type=\"xml/xsl\"?>\n<config>\n  <ips>\n    <ip address=\"1.1.1.1\" />\n  </ips>\n  <technology identifier=\"openvpn_udp\"/>\n</config>\n",
		},
		{
			identifier: techXORTCP,
			expected:   "<?xml version=\"1.0\"?>\n<?xml-stylesheet type=\"xml/xsl\"?>\n<config>\n  <ips>\n    <ip address=\"1.1.1.1\" />\n  </ips>\n  <technology identifier=\"openvpn_xor_tcp\"/>\n</config>\n",
		},
		{
			identifier: techTCP,
			expected:   "<?xml version=\"1.0\"?>\n<?xml-stylesheet type=\"xml/xsl\"?>\n<config>\n  <ips>\n    <ip address=\"1.1.1.1\" />\n  </ips>\n  <technology identifier=\"openvpn_tcp\"/>\n</config>\n",
		},
	}

	for _, test := range tests {
		t.Run(string(test.identifier), func(t *testing.T) {
			got, err := generateConfigXML(netip.MustParseAddr("1.1.1.1"), test.identifier)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, string(got))
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name       string
		ip         netip.Addr
		identifier openvpnID
		template   string
		config     string
		err        bool
	}{
		{
			name:       "1.0",
			ip:         netip.MustParseAddr("1.1.1.1"),
			identifier: techUDP,
			template:   configV1Template,
			config:     configV1,
		},
		{
			name:       "XOR 1.0",
			ip:         netip.MustParseAddr("5.5.5.5"),
			identifier: techXORTCP,
			template:   configXORV1Template,
			config:     configXORV1,
		},
		{
			name:       "invalid template",
			ip:         netip.MustParseAddr("5.5.5.5"),
			identifier: techXORTCP,
			template:   "invalid",
			config:     "",
			err:        true,
		},
		{
			name:       "importingTemplate",
			ip:         netip.MustParseAddr("5.5.5.5"),
			identifier: techXORTCP,
			template: `<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform" version="1.0">
	<xsl:include href="nonexistant.xsl"/>
	<xsl:apply-templates/>
</xsl:stylesheet>`,
			config: "",
			err:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := generateConfig(tt.ip, tt.identifier, []byte(tt.template))
			assert.Equal(t, tt.err, err != nil)
			assert.Equal(t, tt.config, string(out))
		})
	}
}
