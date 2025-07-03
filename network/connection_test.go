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
	}
	endpoint := DefaultEndpoint(&resolver, serverIps)

	ip, err := endpoint.ip()
	assert.NoError(t, err)
	assert.Equal(t, serverIps[0], ip)
}
