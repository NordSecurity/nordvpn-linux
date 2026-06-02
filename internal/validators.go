package internal

import "net/netip"

func IsAddressValidAsDNSServer(address string) bool {
	parsedAddress, err := netip.ParseAddr(address)
	// Is4 is true only for genuine 4-byte IPv4 addresses; it rejects IPv6 and
	// IPv4-mapped IPv6 forms such as "::ffff:192.168.1.1" (those are Is4In6).
	return err == nil && parsedAddress.Is4()
}
