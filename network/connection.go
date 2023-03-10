package network

import "net/netip"

// EndpointResolver check if the endpoint can be used
type EndpointResolver interface {
	Resolve(endpoint netip.Addr) ([]netip.Addr, error)
}

// DefaultEndpoint returns appropriate endpoint to use.
func DefaultEndpoint(resolver EndpointResolver, serverIps []netip.Addr) Endpoint {
	for _, ip := range serverIps {
		if ip.Is6() {
			_, err := resolver.Resolve(ip)
			if err == nil {
				return NewIPv6Endpoint(serverIps)
			}
		}
	}
	return NewLocalEndpoint(serverIps)
}
