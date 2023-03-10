// Package mesh provides data types and interfaces for implementing peer to peer
// communication.
package meshnet

import (
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// Mesh defines a set of operations that any type that wants to act as a mesh must implement.
type Mesh interface {
	// Enable creates a tunnel interface with a given IP.
	Enable(netip.Addr, string) error
	// Disable removes a tunnel interface
	Disable() error
	// IsActive returns false when the tunnel is gone.
	IsActive() bool
	// Refresh peer list
	// Has to be called at least once after Enable
	Refresh(mesh.MachineMap) error
	// Tun retrieves a tunnel used for the meshnet
	Tun() tunnel.T
	// StatusMap retrieves the current status map for the related
	// meshnet peers
	StatusMap() (map[string]string, error)
}

// KeyGenerator for use in meshnet.
type KeyGenerator interface {
	// Private returns base64 encoded private key
	Private() string
	// Public expects base64 encoded private key and returns base64 encoded public key
	Public(string) string
}
