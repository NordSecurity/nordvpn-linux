package network

import (
	"errors"
	"net/netip"
)

// Endpoint is responsible for picking the correct IP
// to use when connecting to the server.
type Endpoint struct {
	ips []netip.Addr
}

func (e Endpoint) ip4() (netip.Addr, error) {
	for _, ip := range e.ips {
		if ip.Is4() {
			return ip, nil
		}
	}
	return netip.Addr{}, errors.New("no IPv4 addresses")
}

func (e Endpoint) ip() (netip.Addr, error) {
	return e.ip4()
}

// Network returns a parsed CIDR for IP to used to connect.
func (e Endpoint) Network() (netip.Prefix, error) {
	ip, err := e.ip()
	if err != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(ip, ip.BitLen()), nil
}

func NewIPv4Endpoint(ip netip.Addr) Endpoint {
	return Endpoint{ips: []netip.Addr{ip}}
}

func NewLocalEndpoint(ips []netip.Addr) Endpoint {
	return Endpoint{
		ips: ips,
	}
}
