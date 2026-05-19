package tray

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/uievent"
)

// ItemValueFromServerGroup maps a server group to the corresponding UI event item value.
func ItemValueFromServerGroup(group config.ServerGroup) pb.UIEvent_ItemValue {
	return uievent.ItemValueFromServerGroup(group)
}

// ItemValueFromRecentConnection determines the UI event item value from a RecentConnection.
// Specialty groups (Double VPN, P2P, etc.) take priority over connection type as they
// represent the user's primary intent. Falls back to CITY/COUNTRY for standard VPN.
func ItemValueFromRecentConnection(conn *RecentConnection) pb.UIEvent_ItemValue {
	if conn == nil {
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}

	if itemValue := ItemValueFromServerGroup(conn.Group); itemValue != pb.UIEvent_ITEM_VALUE_UNSPECIFIED {
		return itemValue
	}

	//exhaustive:ignore
	switch conn.ConnectionType {
	case config.ServerSelectionRule_CITY:
		return pb.UIEvent_CITY
	case config.ServerSelectionRule_COUNTRY:
		return pb.UIEvent_COUNTRY
	default:
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
}

// newUIEventContext creates a UIEventContext for tray actions.
func newUIEventContext(
	itemName pb.UIEvent_ItemName,
	itemValue pb.UIEvent_ItemValue,
) *uievent.UIEventContext {
	return &uievent.UIEventContext{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      itemName,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     itemValue,
	}
}

// attachUIEventMetadata attaches UI event metadata to a context for gRPC calls.
func attachUIEventMetadata(
	ctx context.Context,
	itemName pb.UIEvent_ItemName,
	itemValue pb.UIEvent_ItemValue,
) context.Context {
	uiCtx := newUIEventContext(itemName, itemValue)
	return uievent.AttachToOutgoingContext(ctx, uiCtx)
}
