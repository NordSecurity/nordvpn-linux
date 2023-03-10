package network

import (
	"errors"
	"net/netip"
)

// Endpoint is responsible for picking the correct IP
// to use when connecting to the server. Sometimes,
// even if the server supports IPv6, it cannot be used
// to connect to it, due to limitations on the client.
type Endpoint struct {
	ips          []netip.Addr
	supportsIPv6 bool
}

func (e Endpoint) ip6() (netip.Addr, error) {
	for _, ip := range e.ips {
		if ip.Is6() {
			return ip, nil
		}
	}
	return netip.Addr{}, errors.New("no IPv6 addresses")
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
	if e.supportsIPv6 {
		return e.ip6()
	}
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

func NewIPv6Endpoint(ips []netip.Addr) Endpoint {
	return Endpoint{
		ips:          ips,
		supportsIPv6: true,
	}
}

func NewLocalEndpoint(ips []netip.Addr) Endpoint {
	return Endpoint{
		ips: ips,
	}
}
