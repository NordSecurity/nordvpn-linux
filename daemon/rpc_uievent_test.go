package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportUIEvent_NilInput(t *testing.T) {
	category.Set(t, category.Unit)

	r := testRPC()
	resp, err := r.ReportUIEvent(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, internal.CodeSuccess, resp.Type)
}

func TestReportUIEvent_NilInput_NoPublish(t *testing.T) {
	category.Set(t, category.Unit)

	r := testRPC()
	var captured []events.UiItemsAction
	r.events.Service.UiItemsClick.Subscribe(
		func(a events.UiItemsAction) error {
			captured = append(captured, a)
			return nil
		},
	)

	resp, err := r.ReportUIEvent(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, internal.CodeSuccess, resp.Type)
	assert.Empty(t, captured, "nil input must not publish any event")
}

func TestReportUIEvent_PublishesAllInventoryEvents(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    *pb.UIEvent
		expected events.UiItemsAction
	}{
		// GUI events
		{
			name: "GUI connect by country",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_COUNTRY,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "country",
			},
		},
		{
			name: "GUI connect by city",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_CITY,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "city",
			},
		},
		{
			name: "GUI connect dedicated IP",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_DIP,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "dip",
			},
		},
		{
			name: "GUI connect P2P",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_P2P,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "p2p",
			},
		},
		{
			name: "GUI connect recents by city",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_CONNECT_RECENTS,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_CITY,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "connect_recents",
				ItemType:      "click",
				ItemValue:     "city",
			},
		},
		{
			name: "GUI reconnect",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CONNECTION_INFO,
				ItemName:      pb.UIEvent_RECONNECT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "connection_info",
				ItemName:      "reconnect",
				ItemType:      "click",
			},
		},
		{
			name: "GUI disconnect",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_DISCONNECT,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "disconnect",
			},
		},
		{
			name: "GUI pause 5 minutes",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_5_MIN,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "5_minutes",
			},
		},
		{
			name: "GUI pause 1 hour",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_HOME_SCREEN,
				ItemName:      pb.UIEvent_PAUSE,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_PAUSE_1_HOUR,
			},
			expected: events.UiItemsAction{
				FormReference: "home_screen",
				ItemName:      "pause",
				ItemType:      "click",
				ItemValue:     "1_hour",
			},
		},
		{
			name: "GUI change settings",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CONNECTION_INFO,
				ItemName:      pb.UIEvent_CHANGE_SETTINGS,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "connection_info",
				ItemName:      "change_vpn_settings",
				ItemType:      "click",
			},
		},
		{
			name: "GUI get help",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CONNECTION_INFO,
				ItemName:      pb.UIEvent_GET_HELP,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "connection_info",
				ItemName:      "help",
				ItemType:      "click",
			},
		},
		{
			name: "GUI login",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_GUI,
				ItemName:      pb.UIEvent_LOGIN,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "gui",
				ItemName:      "login",
				ItemType:      "click",
			},
		},
		{
			name: "GUI logout",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_GUI,
				ItemName:      pb.UIEvent_LOGOUT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "gui",
				ItemName:      "logout",
				ItemType:      "click",
			},
		},
		// CLI events
		{
			name: "CLI connect without server group",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "connect",
				ItemType:      "click",
			},
		},
		{
			name: "CLI connect P2P group",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_P2P,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "p2p",
			},
		},
		{
			name: "CLI disconnect",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_DISCONNECT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "disconnect",
				ItemType:      "click",
			},
		},
		{
			name: "CLI login OAuth2",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_LOGIN,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "login",
				ItemType:      "click",
			},
		},
		{
			name: "CLI login with token",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_LOGIN_TOKEN,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "login_token",
				ItemType:      "click",
			},
		},
		{
			name: "CLI logout",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_LOGOUT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "logout",
				ItemType:      "click",
			},
		},
		{
			name: "CLI rate connection",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_RATE_CONNECTION,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "rate_connection",
				ItemType:      "click",
			},
		},
		{
			name: "CLI meshnet invite send",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_MESHNET_INVITE_SEND,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "meshnet_invite_send",
				ItemType:      "click",
			},
		},
		{
			name: "CLI connect dedicated server",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_DEDICATED_SERVER,
			},
			expected: events.UiItemsAction{
				FormReference: "cli",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "dedicated_server",
			},
		},
		// Tray events
		{
			name: "Tray connect by country",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_COUNTRY,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "connect",
				ItemType:      "click",
				ItemValue:     "country",
			},
		},
		{
			name: "Tray connect recents by city",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_CONNECT_RECENTS,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_CITY,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "connect_recents",
				ItemType:      "click",
				ItemValue:     "city",
			},
		},
		{
			name: "Tray login",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_LOGIN,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "login",
				ItemType:      "click",
			},
		},
		{
			name: "Tray logout",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_LOGOUT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "logout",
				ItemType:      "click",
			},
		},
		{
			name: "Tray disconnect",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_DISCONNECT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "disconnect",
				ItemType:      "click",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := testRPC()
			var captured []events.UiItemsAction
			r.events.Service.UiItemsClick.Subscribe(
				func(a events.UiItemsAction) error {
					captured = append(captured, a)
					return nil
				},
			)

			resp, err := r.ReportUIEvent(
				context.Background(), tt.input,
			)

			require.NoError(t, err)
			assert.Equal(t, internal.CodeSuccess, resp.Type)
			require.Len(t, captured, 1)
			assert.Equal(t, tt.expected, captured[0])
		})
	}
}
