package network

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestStringsToIPs(t *testing.T) {
	category.Set(t, category.Unit)

	expected := []netip.Addr{
		netip.MustParseAddr("127.0.0.1"),
		netip.MustParseAddr("1.1.1.1"),
		netip.MustParseAddr("fe80::1"),
		netip.MustParseAddr("2a00:7c80:0:eb:0:11:8:1000"),
	}
	result := StringsToIPs([]string{"127.0.0.1", "1.1.1.1", "127", "any", "fe80::1", "2a00:7c80:0:eb:0:11:8:1000", "ff:", ""})
	assert.Equal(t, expected, result)
	assert.Equal(t, []netip.Addr{}, StringsToIPs([]string{"invalid"}))
	assert.Equal(t, []netip.Addr{}, StringsToIPs(nil))
}

func TestNetworkToRouteString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name   string
		subnet netip.Prefix
	}{
		{
			name:   "127.0.0.1/16",
			subnet: netip.MustParsePrefix("127.0.0.1/16"),
		},
		{
			name:   "1.1.1.1",
			subnet: netip.MustParsePrefix("1.1.1.1/32"),
		},
		{
			name:   "fe80::1/16",
			subnet: netip.MustParsePrefix("fe80::1/16"),
		},
		{
			name:   "2a00:7c80:0:eb:0:11:8:1000",
			subnet: netip.MustParsePrefix("2a00:7c80:0:eb:0:11:8:1000/128"),
		},
		{
			name:   "2a00:7c80:0:eb:0:11:8:1000/32",
			subnet: netip.MustParsePrefix("2a00:7c80:0:eb:0:11:8:1000/32"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.name, ToRouteString(test.subnet))
		})
	}
}
