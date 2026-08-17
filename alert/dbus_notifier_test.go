package alert

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	mocknotify "github.com/NordSecurity/nordvpn-linux/test/mock/notify"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDbusNotifier() *DbusNotifier {
	return &DbusNotifier{
		notifier: &mocknotify.UpstreamNotifierMock{},
		actions:  make(map[AlertID]map[string]func()),
		isActive: true,
	}
}

func TestDbusNotifierDispatchAction(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		actionKey  string
		wantCalled bool
	}{
		{
			name:       "known action key invokes callback",
			actionKey:  "open",
			wantCalled: true,
		},
		{
			name:       "unknown action key is a no-op",
			actionKey:  "unknown-key",
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newTestDbusNotifier()

			called := false
			n.Alert("body").Action("open", "Open", func() { called = true }).Show()
			n.dispatchAction(&notify.ActionInvokedSignal{ID: 0, ActionKey: tt.actionKey})

			assert.Equal(t, tt.wantCalled, called)
		})
	}
}

func TestDbusNotifierSingleActionIsConsumedAtMostOnce(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		firstStep  func(n *DbusNotifier, id AlertID)
		wantCalled bool
	}{
		{
			name: "dispatching twice invokes callback only once",
			firstStep: func(n *DbusNotifier, id AlertID) {
				n.dispatchAction(&notify.ActionInvokedSignal{ID: id, ActionKey: "open"})
			},
			wantCalled: true,
		},
		{
			name: "forgetting before dispatch discards the callback",
			firstStep: func(n *DbusNotifier, id AlertID) {
				n.forget(&notify.NotificationClosedSignal{ID: id})
			},
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newTestDbusNotifier()

			callCount := 0
			n.Alert("body").Action("open", "Open", func() { callCount++ }).Show()

			const id AlertID = 0
			tt.firstStep(n, id)
			n.dispatchAction(&notify.ActionInvokedSignal{ID: id, ActionKey: "open"})

			wantCount := 0
			if tt.wantCalled {
				wantCount = 1
			}
			assert.Equal(t, wantCount, callCount)
		})
	}
}

func TestDbusNotifierDispatchActionMutualExclusion(t *testing.T) {
	category.Set(t, category.Unit)

	n := newTestDbusNotifier()

	var accepted, declined bool
	n.Alert("body").
		Action("accept", "Accept", func() { accepted = true }).
		Action("decline", "Decline", func() { declined = true }).
		Show()

	n.dispatchAction(&notify.ActionInvokedSignal{ID: 0, ActionKey: "accept"})
	n.dispatchAction(&notify.ActionInvokedSignal{ID: 0, ActionKey: "decline"})

	assert.True(t, accepted)
	assert.False(t, declined)
}

func TestDbusNotifierDisableRespectsUrgency(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		urgent        bool
		wantDelivered bool
	}{
		{
			name:          "normal alert suppressed when disabled",
			urgent:        false,
			wantDelivered: false,
		},
		{
			name:          "urgent alert bypasses disable",
			urgent:        true,
			wantDelivered: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &mocknotify.UpstreamNotifierMock{}
			n := &DbusNotifier{
				notifier: upstream,
				actions:  make(map[AlertID]map[string]func()),
				isActive: true,
			}

			n.Mute()

			b := n.Alert("body")
			if tt.urgent {
				b = b.Urgent()
			}
			b.Show()

			assert.Equal(t, tt.wantDelivered, len(upstream.Sent) == 1)
		})
	}
}

func TestDbusNotifierTransientHint(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		transient bool
	}{
		{name: "transient option sets the hint", transient: true},
		{name: "without the option no hint is set", transient: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &mocknotify.UpstreamNotifierMock{}
			n := &DbusNotifier{
				notifier:  upstream,
				actions:   make(map[AlertID]map[string]func()),
				isActive:  true,
				transient: tt.transient,
			}

			n.Alert("body").Show()

			require.Len(t, upstream.Sent, 1)
			hint, ok := upstream.Sent[0].Hints["transient"]
			if tt.transient {
				assert.True(t, ok)
				assert.Equal(t, dbus.MakeVariant(1), hint)
			} else {
				assert.False(t, ok)
			}
		})
	}
}
