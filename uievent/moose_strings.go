package uievent

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// ToMooseStrings converts a UIEventContext to Moose analytics string values.
func ToMooseStrings(ctx *UIEventContext) events.UiItemsAction {
	if ctx == nil {
		return events.UiItemsAction{}
	}

	return events.UiItemsAction{
		FormReference: formReferenceToString(ctx.FormReference),
		ItemName:      itemNameToString(ctx.ItemName),
		ItemType:      itemTypeToString(ctx.ItemType),
		ItemValue:     itemValueToString(ctx.ItemValue),
	}
}

// formReferenceToString converts FormReference enum to Moose string value
func formReferenceToString(ref pb.UIEvent_FormReference) string {
	switch ref {
	case pb.UIEvent_FORM_REFERENCE_UNSPECIFIED:
		return ""
	case pb.UIEvent_CLI:
		return "cli"
	case pb.UIEvent_TRAY:
		return "tray"
	case pb.UIEvent_HOME_SCREEN:
		return "home_screen"
	}
	return ""
}

// itemNameToString converts ItemName enum to Moose string value
func itemNameToString(name pb.UIEvent_ItemName) string {
	switch name {
	case pb.UIEvent_ITEM_NAME_UNSPECIFIED:
		return ""
	case pb.UIEvent_CONNECT:
		return "connect"
	case pb.UIEvent_CONNECT_RECENTS:
		return "connect_recents"
	case pb.UIEvent_DISCONNECT:
		return "disconnect"
	case pb.UIEvent_LOGIN:
		return "login"
	case pb.UIEvent_LOGIN_TOKEN:
		return "login_token"
	case pb.UIEvent_LOGOUT:
		return "logout"
	case pb.UIEvent_RATE_CONNECTION:
		return "rate_connection"
	case pb.UIEvent_MESHNET_INVITE_SEND:
		return "meshnet_invite_send"
	}
	return ""
}

// itemTypeToString converts ItemType enum to Moose string value
func itemTypeToString(itemType pb.UIEvent_ItemType) string {
	switch itemType {
	case pb.UIEvent_ITEM_TYPE_UNSPECIFIED:
		return ""
	case pb.UIEvent_CLICK:
		return "click"
	}
	return ""
}

// itemValueToString converts ItemValue enum to Moose string value
func itemValueToString(value pb.UIEvent_ItemValue) string {
	switch value {
	case pb.UIEvent_ITEM_VALUE_UNSPECIFIED:
		return ""
	case pb.UIEvent_COUNTRY:
		return "country"
	case pb.UIEvent_CITY:
		return "city"
	case pb.UIEvent_DIP:
		return "dip"
	case pb.UIEvent_MESHNET:
		return "meshnet"
	case pb.UIEvent_OBFUSCATED:
		return "obfuscated"
	case pb.UIEvent_ONION_OVER_VPN:
		return "onion_over_vpn"
	case pb.UIEvent_DOUBLE_VPN:
		return "double_vpn"
	case pb.UIEvent_P2P:
		return "p2p"
	}
	return ""
}
