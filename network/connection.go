package network

import "net/netip"

// EndpointResolver check if the endpoint can be used
type EndpointResolver interface {
	Resolve(endpoint netip.Addr) ([]netip.Addr, error)
}

// DefaultEndpoint returns appropriate endpoint to use.
func DefaultEndpoint(resolver EndpointResolver, serverIps []netip.Addr) Endpoint {
	return NewLocalEndpoint(serverIps)
}
