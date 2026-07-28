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

func TestReportUIEvent_PublishesValidEvents(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    *pb.UIEvent
		expected events.UiItemsAction
	}{
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
			name: "CLI login",
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
			name: "Tray open app",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_OPEN_APP,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: events.UiItemsAction{
				FormReference: "tray",
				ItemName:      "open_app",
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

			resp, err := r.ReportUIEvent(context.Background(), tt.input)

			require.NoError(t, err)
			assert.Equal(t, internal.CodeSuccess, resp.Type)
			require.Len(t, captured, 1)
			assert.Equal(t, tt.expected, captured[0])
		})
	}
}

func TestReportUIEvent_DropsIncompleteEvents(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name  string
		input *pb.UIEvent
	}{
		{
			name: "missing form reference",
			input: &pb.UIEvent{
				ItemName: pb.UIEvent_OPEN_APP,
				ItemType: pb.UIEvent_CLICK,
			},
		},
		{
			name: "missing item name",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemType:      pb.UIEvent_CLICK,
			},
		},
		{
			name: "missing item type",
			input: &pb.UIEvent{
				FormReference: pb.UIEvent_TRAY,
				ItemName:      pb.UIEvent_OPEN_APP,
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

			resp, err := r.ReportUIEvent(context.Background(), tt.input)

			require.NoError(t, err)
			assert.Equal(t, internal.CodeSuccess, resp.Type)
			assert.Empty(t, captured, "incomplete event must not be published")
		})
	}
}
