package vpn

import (
	"errors"
	"net/netip"
)

// InterfaceIPv6 is made from server IP and static interface id and is different per protocol.
func InterfaceIPv6(serverIP netip.Addr, interfaceID [8]byte) (netip.Addr, error) {
	if !serverIP.Is6() {
		return netip.Addr{}, errors.New("not an IPv6 address")
	}

	var interfaceIP [16]byte
	serverIPBytes := serverIP.As16()
	copy(interfaceIP[:8], serverIPBytes[:])
	copy(interfaceIP[8:], interfaceID[:])
	return netip.AddrFrom16(interfaceIP), nil
}
