package uievent

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestToMooseStrings(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		ctx      *UIEventContext
		expected events.UiItemsAction
	}{
		{
			name: "nil context returns empty",
			ctx:  nil,
			expected: events.UiItemsAction{
				FormReference: "",
				ItemName:      "",
				ItemType:      "",
				ItemValue:     "",
			},
		},
		{
			name: "all fields set",
			ctx: &UIEventContext{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToMooseStrings(tt.ctx))
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
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, itemValueToString(tt.input))
		})
	}
}
