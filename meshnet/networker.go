package meshnet

import (
	"net/netip"

	teliogo "github.com/NordSecurity/libtelio-go/v5"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	_ "github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio/symbols" // required for linking process
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
	AllowIncoming(address UniqueAddress, lanAllowed bool) error
	// BlockIncoming creates a blocking fw rule for the given
	// address
	BlockIncoming(UniqueAddress) error
	// AllowFileshare creates a rule enabling fileshare port for the given address
	AllowFileshare(UniqueAddress) error
	// BlockFileshare removes a rule enabling fileshare port for the given address if it exists
	BlockFileshare(UniqueAddress) error
	// ResetRouting is used when there are routing setting changes,
	// except when routing is denied - then BlockRouting must be used. changedPeer is the peer whose routing settings
	// changed, peers is the map of all the machine peers(including the changed peer).
	ResetRouting(changedPeer mesh.MachinePeer, peers mesh.MachinePeers) error
	StatusMap() (map[string]teliogo.NodeState, error)
	LastServerName() string
	Start(
		vpn.Credentials,
		vpn.ServerData,
		config.Allowlist,
		config.DNS,
		bool, // enableLocalTraffic
	) error
	Stop() error
}

// UniqueAddress a member of mesh network.
type UniqueAddress struct {
	// UID is a base64 encoded unique string
	UID     string
	Address netip.Addr
}
