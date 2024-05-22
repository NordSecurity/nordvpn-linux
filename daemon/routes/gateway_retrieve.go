package routes

import (
	"errors"
	"net"
	"net/netip"
)

var (
	// ErrNotFound defines that gateway is not found for a given address
	ErrNotFound = errors.New("gateway not found")
)

// GatewayRetriever is responsible for retrieving gateways for the given networks in current
// system.
type GatewayRetriever interface {
	// Retrieve a gateway to a given prefix while ignoring the given routing table.
	//
	// `ignoreTable` is used in order to not receive actual gateway due to the following
	// reasons:
	// 1. In case VPN connection is active, retrieved gateway will be default route to VPN
	//    tunnel interface, which is useless for allowlisting functionality.
	// 2. Assuming main routing table and default gateway is an incorrect way to determine
	//    gateway before VPN in environments with multiple physical interfaces.
	//    Conditional route adding for non-private IPs is not viable solution because IP rule
	//    setup blocks any traffic for physical network interfaces. `192.168.0.0/16` is
	//    considered a private IP range and is usually routed through a physical network
	//    interface.
	//
	// Default gateway can be retrieved with such `prefix` values:
	// * IPv4: `netip.Prefix{}` or `0.0.0.0/0`
	// * IPv6: `::/0`
	Retrieve(prefix netip.Prefix, ignoreTable uint) (netip.Addr, net.Interface, error)
}
