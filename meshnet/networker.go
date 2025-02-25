package meshnet

import (
	"context"
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
	// PermitFileshare creates a rules enabling fileshare port for all available peers and sets fileshare as permitted
	PermitFileshare() error
	// ForbidFileshare removes a rules enabling fileshare port for all available peers and sets fileshare as forbidden
	ForbidFileshare() error
	StatusMap() (map[string]string, error)
	LastServerName() string
	Start(
		context.Context,
		vpn.Credentials,
		vpn.ServerData,
		config.Allowlist,
		config.DNS,
		bool, // enableLocalTraffic
	) error
	Stop() error
	GetConnectionParameters() (vpn.ServerData, bool)
}

// UniqueAddress a member of mesh network.
type UniqueAddress struct {
	// UID is a base64 encoded unique string
	UID     string
	Address netip.Addr
}
