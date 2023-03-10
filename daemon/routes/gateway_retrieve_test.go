package routes

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestGrepDefaultGatewayIPFromOutput(t *testing.T) {
	category.Set(t, category.Unit)

	longOutput := `default via 192.168.0.2 dev wlan0 proto dhcp metric 600
192.168.0.0/16 dev nordlynx proto kernel scope link src 192.168.0.1 linkdown`
	tests := []struct {
		output  string
		gateway []byte
		devices []string
	}{
		{
			output:  "default via 192.168.0.1 dev wlan0 proto dhcp metric 600",
			gateway: []byte("192.168.0.1"),
			devices: []string{"wlan0"},
		},
		{
			output:  "192.168.0.0/16 via 192.168.0.1 dev wlan0 proto dhcp metric 600",
			gateway: nil,
			devices: nil,
		},
		{output: longOutput,
			gateway: []byte("192.168.0.2"),
			devices: []string{"wlan0"},
		},
		{
			output:  "default via 192.168.0.1 dev docker0 proto dhcp metric 600",
			gateway: nil,
			devices: []string{"wlan0"},
		},
		{
			output:  "default via fe80::1 dev eth0 proto ra metric 1024 expires 296sec hoplimit 64 pref medium",
			gateway: []byte("fe80::1"),
			devices: []string{"eth0"},
		},
	}
	for _, test := range tests {
		t.Run(test.output, func(t *testing.T) {
			assert.Equal(t, test.gateway,
				grepDefaultGatewayIPFromOutput(
					[]byte(test.output), test.devices),
			)
		})
	}
}

func TestIPGatewayRetriever_Default(t *testing.T) {
	category.Set(t, category.Route)
	// Just assume that default gateway exists on system
	retriever := IPGatewayRetriever{}
	gateway, _, err := retriever.Default(false)
	assert.NoError(t, err)
	assert.NotNil(t, gateway, gateway)
}

func TestStrContainsAny(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name     string
		str      string
		parts    []string
		contains bool
	}{
		{name: "contains_single", str: "hello, world", parts: []string{"hello"}, contains: true},
		{name: "contains_all", str: "hello, world", parts: []string{"hello", "world", "o, w"}, contains: true},
		{name: "contains_none", str: "hello, world", parts: []string{"goodbye"}, contains: false},
		{name: "contains_single", str: "hello, world", parts: []string{"hello", "goodbye"}, contains: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contains := strContainsAny(tt.str, tt.parts)
			assert.Equal(t, tt.contains, contains)
		})
	}
}
