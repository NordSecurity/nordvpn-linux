package uievent

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// To execute tests:
// go test -v -run TestEventInventory ./uievent/... -count=1 2>&1

// TestEventInventory_AllEventsReachMoose verifies that every UI event across all
// clients (GUI, CLI, Tray) reaches Moose analytics with the correct field values.
//
// This test serves as a regression safety net for UI event refactoring: run it
// before and after changes to confirm no events are lost or altered.
func TestEventInventory_AllEventsReachMoose(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    UIEventContext
		expected events.UiItemsAction
	}{
		// GUI events
		{
			name: "GUI connect by country",
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			input: UIEventContext{
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
			publisher := &mockPublisher{}
			middleware := NewMiddleware(publisher)

			clientCtx := AttachToOutgoingContext(
				context.Background(), &tt.input,
			)
			outgoingMD, _ := metadata.FromOutgoingContext(clientCtx)

			_, _ = middleware.UnaryMiddleware(
				metadata.NewIncomingContext(
					context.Background(), outgoingMD,
				),
				nil,
				&grpc.UnaryServerInfo{},
			)

			require.Len(t, publisher.published, 1)
			assert.Equal(t, tt.expected, publisher.published[0])
		})
	}
}
