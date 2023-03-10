package network

import (
	"net/netip"
)

func StringsToIPs(addresses []string) []netip.Addr {
	ips := []netip.Addr{}
	for _, address := range addresses {
		ip, err := netip.ParseAddr(address)
		if err != nil {
			continue
		}
		ips = append(ips, ip)
	}
	return ips
}

func ToRouteString(network netip.Prefix) string {
	ip := network.Addr()
	bits := network.Bits()
	if (ip.Is4() && bits == 32) || (ip.Is6() && bits == 128) {
		return network.Addr().String()
	}
	return network.String()
}
