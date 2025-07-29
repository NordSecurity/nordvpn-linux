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

	network, err := endpoint.Network()
	assert.NoError(t, err)
	assert.Equal(t, netip.PrefixFrom(serverIps[0], 32), network)
}
