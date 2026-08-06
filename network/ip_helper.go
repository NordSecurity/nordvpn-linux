package network

import (
	"net/netip"
)

func FilterInvalidIPs(addresses []string) []string {
	result := []string{}
	for _, address := range addresses {
		_, err := netip.ParseAddr(address)
		if err != nil {
			continue
		}
		result = append(result, address)
	}
	return result
}
