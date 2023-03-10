package network

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

type resolver struct {
	CustomError error
}

func (r *resolver) Resolve(ip netip.Addr) ([]netip.Addr, error) {
	return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, r.CustomError
}

func TestDefaultEndpoint(t *testing.T) {
	resolver := resolver{}
	serverIps := []netip.Addr{
		netip.MustParseAddr("1.1.1.1"),
		netip.MustParseAddr("2606:4700:4700::1111"),
	}
	endpoint := DefaultEndpoint(&resolver, serverIps)

	ip, err := endpoint.ip()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[1], ip)

	ip4, err := endpoint.ip4()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[0], ip4)

	ip6, err := endpoint.ip6()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[1], ip6)
	assert.True(t, endpoint.supportsIPv6)
}
