package network

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpv4OnlySystemSupport(t *testing.T) {
	serverIps := []netip.Addr{netip.MustParseAddr("1.1.1.1")}
	endpoint := NewLocalEndpoint(serverIps)

	ip, err := endpoint.ip()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[0], ip)

	ip4, err := endpoint.ip4()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[0], ip4)

	_, err = endpoint.ip6()
	assert.Error(t, err)

	network, err := endpoint.Network()
	assert.NoError(t, err)
	assert.Equal(t, netip.PrefixFrom(serverIps[0], 32), network)
	assert.False(t, endpoint.supportsIPv6)
}

func TestDualStackSystemSupport(t *testing.T) {
	serverIps := []netip.Addr{
		netip.MustParseAddr("1.1.1.1"),
		netip.MustParseAddr("2001:4860:4860::8888"),
	}
	endpoint := NewLocalEndpoint(serverIps)

	ip, err := endpoint.ip()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[0], ip)

	ip4, err := endpoint.ip4()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[0], ip4)

	ip6, err := endpoint.ip6()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[1], ip6)

	net4, err := endpoint.Network()
	assert.NoError(t, err)
	assert.Equal(t, netip.PrefixFrom(ip4, 32), net4)
}

func TestIpv6OnlySystemSupport(t *testing.T) {
	endpoint := NewIPv6Endpoint([]netip.Addr{
		netip.MustParseAddr("1.1.1.1"),
		netip.MustParseAddr("2001:4860:4860::8888"),
	})

	ip, err := endpoint.ip()
	assert.NoError(t, err)
	assert.Equal(t, netip.MustParseAddr("2001:4860:4860::8888"), ip)

	ip4, err := endpoint.ip4()
	assert.NoError(t, err)
	assert.Equal(t, netip.MustParseAddr("1.1.1.1"), ip4)

	ip6, err := endpoint.ip6()
	assert.NoError(t, err)
	assert.Equal(t, netip.MustParseAddr("2001:4860:4860::8888"), ip6)

	net6, err := endpoint.Network()
	assert.NoError(t, err)
	assert.Equal(t, netip.PrefixFrom(ip6, 128), net6)
	assert.True(t, endpoint.supportsIPv6)
}
