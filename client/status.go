package client

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// SpecialtyGroupLabel returns the display name of the connected specialty group, or "" for none.
func SpecialtyGroupLabel(status *pb.StatusResponse) string {
	if status.GetState() != pb.ConnectionState_CONNECTED || status.GetIsMeshPeer() {
		return ""
	}

	group := status.GetParameters().GetGroup()
	// TODO: Move to switch below after LVPN-10704
	if status.GetObfuscated() {
		group = config.ServerGroup_OBFUSCATED
	}

	//exhaustive:ignore [only specialty groups have a label]
	switch group {
	case config.ServerGroup_DOUBLE_VPN,
		config.ServerGroup_ONION_OVER_VPN,
		config.ServerGroup_DEDICATED_IP,
		config.ServerGroup_OBFUSCATED:
		return config.GroupDisplayName(group)
	}
	return ""
}
