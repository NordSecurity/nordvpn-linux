package uievent

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestProtoToMooseStrings(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    *pb.UIEvent
		expected events.UiItemsAction
	}{
		{
			name:  "nil event returns empty",
			input: nil,
			expected: events.UiItemsAction{
				FormReference: "",
				ItemName:      "",
				ItemType:      "",
				ItemValue:     "",
			},
		},
		{
			name: "all fields set",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_COUNTRY,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "country",
			},
		},
		{
			name: "pause event for 5 minutes",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_5_MIN,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "5_minutes",
			},
		},
		{
			name: "pause event for 15 minutes",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_15_MIN,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "15_minutes",
			},
		},
		{
			name: "pause event for 30 minutes",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_30_MIN,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "30_minutes",
			},
		},
		{
			name: "pause event for 1 hour",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_1_HOUR,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "1_hour",
			},
		},
		{
			name: "pause event for 24 hours",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_24_HOURS,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "24_hours",
			},
		},
		{
			name: "change settings event",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CONNECTION_INFO,
				ItemName:      pb.UIEvent_CHANGE_SETTINGS,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "connection_info",
				ItemName:      "change_vpn_settings",
				ItemType:      "click",
				ItemValue:     "",
			},
		},
		{
			name: "get help event",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CONNECTION_INFO,
				ItemName:      pb.UIEvent_GET_HELP,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "connection_info",
				ItemName:      "help",
				ItemType:      "click",
				ItemValue:     "",
			},
		},
		{
			name: "pause event disconnect",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_DISCONNECT,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "disconnect",
			},
		},
		{
			name: "dedicated server connect from tray",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_DEDICATED_SERVER,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "dedicated_server",
			},
		},
		{
			name: "open gui app from the tray",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_OPEN_APP,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "open_app",
				ItemType:      "click",
				ItemValue:     "",
			},
		},
		{
			name: "open gui app download page from the tray",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_DOWNLOAD_APP,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "download_app",
				ItemType:      "click",
				ItemValue:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ProtoToMooseStrings(tt.input))
		})
	}
}

func TestFormReferenceToString(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		input    pb.UIEvent_FormReference
		expected string
	}{
		{pb.UIEvent_FORM_REFERENCE_UNSPECIFIED, ""},
		{pb.UIEvent_CLI, "cli"},
		{pb.UIEvent_TRAY, "tray"},
		{pb.UIEvent_HOME_SCREEN, "home_screen"},
		{pb.UIEvent_GUI, "gui"},
		{pb.UIEvent_CONNECTION_INFO, "connection_info"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, formReferenceToString(tt.input))
		})
	}
}

func TestItemNameToString(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		input    pb.UIEvent_ItemName
		expected string
	}{
		{pb.UIEvent_ITEM_NAME_UNSPECIFIED, ""},
		{pb.UIEvent_CONNECT, "connect"},
		{pb.UIEvent_CONNECT_RECENTS, "connect_recents"},
		{pb.UIEvent_DISCONNECT, "disconnect"},
		{pb.UIEvent_LOGIN, "login"},
		{pb.UIEvent_LOGIN_TOKEN, "login_token"},
		{pb.UIEvent_LOGOUT, "logout"},
		{pb.UIEvent_RATE_CONNECTION, "rate_connection"},
		{pb.UIEvent_MESHNET_INVITE_SEND, "meshnet_invite_send"},
		{pb.UIEvent_PAUSE, "pause"},
		{pb.UIEvent_RECONNECT, "reconnect"},
		{pb.UIEvent_CHANGE_SETTINGS, "change_vpn_settings"},
		{pb.UIEvent_GET_HELP, "help"},
		{pb.UIEvent_OPEN_APP, "open_app"},
		{pb.UIEvent_DOWNLOAD_APP, "download_app"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, itemNameToString(tt.input))
		})
	}
}

func TestItemTypeToString(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		input    pb.UIEvent_ItemType
		expected string
	}{
		{pb.UIEvent_ITEM_TYPE_UNSPECIFIED, ""},
		{pb.UIEvent_CLICK, "click"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, itemTypeToString(tt.input))
		})
	}
}

func TestItemValueToString(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		input    pb.UIEvent_ItemValue
		expected string
	}{
		{pb.UIEvent_ITEM_VALUE_UNSPECIFIED, ""},
		{pb.UIEvent_COUNTRY, "country"},
		{pb.UIEvent_CITY, "city"},
		{pb.UIEvent_DIP, "dip"},
		{pb.UIEvent_MESHNET, "meshnet"},
		{pb.UIEvent_OBFUSCATED, "obfuscated"},
		{pb.UIEvent_ONION_OVER_VPN, "onion_over_vpn"},
		{pb.UIEvent_DOUBLE_VPN, "double_vpn"},
		{pb.UIEvent_P2P, "p2p"},
		{pb.UIEvent_PAUSE_5_MIN, "5_minutes"},
		{pb.UIEvent_PAUSE_15_MIN, "15_minutes"},
		{pb.UIEvent_PAUSE_30_MIN, "30_minutes"},
		{pb.UIEvent_PAUSE_1_HOUR, "1_hour"},
		{pb.UIEvent_PAUSE_24_HOURS, "24_hours"},
		{pb.UIEvent_PAUSE_DISCONNECT, "disconnect"},
		{pb.UIEvent_DEDICATED_SERVER, "dedicated_server"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, itemValueToString(tt.input))
		})
	}
}
