package internal

import "net"

func IsDNSAddressValid(address string) bool {
	parsedAddress := net.ParseIP(address)
	return parsedAddress != nil && parsedAddress.To4() != nil
}
