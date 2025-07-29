package clientid

import (
	"context"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockevents "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"google.golang.org/grpc/metadata"
	"gotest.tools/v3/assert"
)

func getUiItemsActionEvent(t *testing.T, itemName string, formReference string) events.UiItemsAction {
	t.Helper()

	return events.UiItemsAction{
		ItemName:      itemName,
		ItemType:      "button",
		FormReference: formReference,
	}
}

func TestNotifyAboutClickEvent(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		clientID        string
		fullMethod      string
		shouldNotify    bool
		expectedMessage events.UiItemsAction
	}{
		{
			name:            "connect request from cli",
			clientID:        strconv.Itoa(int(pb.ClientID_CLI.Number())),
			fullMethod:      pb.Daemon_Connect_FullMethodName,
			shouldNotify:    true,
			expectedMessage: getUiItemsActionEvent(t, connectItemName, cliClientString),
		},
		{
			name:            "connect request from gui",
			clientID:        strconv.Itoa(int(pb.ClientID_GUI.Number())),
			fullMethod:      pb.Daemon_Connect_FullMethodName,
			shouldNotify:    true,
			expectedMessage: getUiItemsActionEvent(t, connectItemName, guiClientString),
		},
		{
			name:            "connect request from tray",
			clientID:        strconv.Itoa(int(pb.ClientID_TRAY.Number())),
			fullMethod:      pb.Daemon_Connect_FullMethodName,
			shouldNotify:    true,
			expectedMessage: getUiItemsActionEvent(t, connectItemName, trayClientString),
		},
		{
			name:         "unrelated request from cli",
			clientID:     strconv.Itoa(int(pb.ClientID_CLI.Number())),
			fullMethod:   "unknown",
			shouldNotify: false,
		},
		{
			name:         "unrelated request from gui",
			clientID:     strconv.Itoa(int(pb.ClientID_GUI.Number())),
			fullMethod:   "unknown",
			shouldNotify: false,
		},
		{
			name:         "unrealated request from tray",
			clientID:     strconv.Itoa(int(pb.ClientID_TRAY.Number())),
			fullMethod:   "unknown",
			shouldNotify: false,
		},
		{
			name:         "connect request from unknown client",
			clientID:     "105",
			fullMethod:   pb.Daemon_Connect_FullMethodName,
			shouldNotify: false,
		},
		{
			name:         "connect request from invalid client id",
			clientID:     "aaa",
			fullMethod:   pb.Daemon_Connect_FullMethodName,
			shouldNotify: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := mockevents.MockPublisher[events.UiItemsAction]{}
			middleware := NewClientIDMiddleware(&events)

			md := metadata.Pairs(clientIDMetadataKey, test.clientID)
			ctx := metadata.NewIncomingContext(context.Background(), md)

			middleware.notifyAboutClickEvent(ctx, test.fullMethod)
			if test.shouldNotify {
				message, remaingMessages, eventFound := events.PopEvent()
				assert.Equal(t, true, eventFound, "Middleware did not emit an event.")
				assert.Equal(t, 0, remaingMessages, "Only one event should be emitted by middleware.")
				assert.Equal(t, test.expectedMessage, message, "Unexpected events emitted by the middleware.")
			} else {
				_, _, eventFound := events.PopEvent()
				assert.Equal(t, false, eventFound, "Middleware should not emitt an event.")
			}
		})
	}
}
