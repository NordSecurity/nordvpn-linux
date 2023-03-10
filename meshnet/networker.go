package meshnet

import (
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
)

// Networker defines functions responsible for network configuration
type Networker interface {
	// SetMesh sets the meshnet configuration up
	SetMesh(
		mesh.MachineMap,
		netip.Addr,
		string,
	) error
	Refresh(mesh.MachineMap) error // Remove
	// UnSetMesh unsets the meshnet configuration
	UnSetMesh() error
	// AllowIncoming creates an allowing fw rule for the given
	// address
	AllowIncoming(UniqueAddress) error
	// BlockIncoming creates a blocking fw rule for the given
	// address
	BlockIncoming(UniqueAddress) error
	// ResetRouting is used when there are routing setting changes,
	// except when routing is denied - then BlockRouting must be used
	ResetRouting(mesh.MachinePeers) error
	StatusMap() (map[string]string, error)
	Start(
		vpn.Credentials,
		vpn.ServerData,
		config.Whitelist,
		config.DNS,
	) error
	Stop() error
}

// UniqueAddress a member of mesh network.
type UniqueAddress struct {
	// UID is a base64 encoded unique string
	UID     string
	Address netip.Addr
}
