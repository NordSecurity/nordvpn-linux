package vpn

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceIPv6(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		serverIP    string
		interfaceID [8]byte
		interfaceIP string
		hasError    bool
	}{
		{
			serverIP:    "2a00:7c80:0:eb::11",
			interfaceID: [8]byte{0x0, 0x0, 0x0, 0x11, 0x0, 0x5, 0x0, 0x2},
			interfaceIP: "2a00:7c80:0:eb:0:11:5:2",
		},
		{
			serverIP:    "2a02:5740:1:9::11",
			interfaceID: [8]byte{0x0, 0x0, 0x0, 0x11, 0x0, 0x5, 0x0, 0x2},
			interfaceIP: "2a02:5740:1:9:0:11:5:2",
		},
		{
			serverIP:    "1.1.1.1",
			interfaceID: [8]byte{0x0, 0x0, 0x0, 0x11, 0x0, 0x5, 0x0, 0x2},
			interfaceIP: "invalid IP",
			hasError:    true,
		},
		{
			serverIP:    "::1",
			interfaceID: [8]byte{0x0, 0x0, 0x0, 0x11, 0x0, 0x5, 0x0, 0x2},
			interfaceIP: "::11:5:2",
		},
		{
			serverIP:    "::1",
			interfaceID: [8]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2},
			interfaceIP: "::2",
		},
	}

	for _, test := range tests {
		t.Run(test.serverIP, func(t *testing.T) {
			serverIP := netip.MustParseAddr(test.serverIP)
			ip, err := InterfaceIPv6(serverIP, test.interfaceID)
			assert.Equal(t, test.hasError, err != nil)
			assert.Equal(t, test.interfaceIP, ip.String())
			assert.NotEqual(t, test.interfaceIP, serverIP.String())
		})
	}

	ip, err := InterfaceIPv6(netip.Addr{}, [8]byte{0x0, 0x0, 0x0, 0x11, 0x0, 0x5, 0x0, 0x2})
	assert.Error(t, err)
	assert.Equal(t, netip.Addr{}, ip)
}
