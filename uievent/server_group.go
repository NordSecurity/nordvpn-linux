package uievent

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// ItemValueFromServerGroup maps a server group to the
// corresponding UIEvent item value.
func ItemValueFromServerGroup(
	group config.ServerGroup,
) pb.UIEvent_ItemValue {
	//exhaustive:ignore
	switch group {
	case config.ServerGroup_DEDICATED_IP:
		return pb.UIEvent_DIP
	case config.ServerGroup_OBFUSCATED:
		return pb.UIEvent_OBFUSCATED
	case config.ServerGroup_ONION_OVER_VPN:
		return pb.UIEvent_ONION_OVER_VPN
	case config.ServerGroup_DOUBLE_VPN:
		return pb.UIEvent_DOUBLE_VPN
	case config.ServerGroup_P2P:
		return pb.UIEvent_P2P
	case config.ServerGroup_DEDICATED_SERVER:
		return pb.UIEvent_DEDICATED_SERVER
	default:
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
}

// ItemValueFromServerGroupString resolves a group name string
// to the corresponding UIEvent item value via config.GroupMap.
func ItemValueFromServerGroupString(
	group string,
) pb.UIEvent_ItemValue {
	serverGroup, ok := config.GroupMap[group]
	if !ok {
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
	return ItemValueFromServerGroup(serverGroup)
}
