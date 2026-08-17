package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockalert "github.com/NordSecurity/nordvpn-linux/test/mock/alert"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatedNotifierSuppressesUntilInitialSyncCompleted(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		initialSyncCompleted bool
		urgent               bool
		wantDelivered        bool
	}{
		{
			name:                 "normal alert suppressed before initial sync",
			initialSyncCompleted: false,
			urgent:               false,
			wantDelivered:        false,
		},
		{
			name:                 "urgent alert suppressed before initial sync",
			initialSyncCompleted: false,
			urgent:               true,
			wantDelivered:        false,
		},
		{
			name:                 "normal alert delivered after initial sync",
			initialSyncCompleted: true,
			urgent:               false,
			wantDelivered:        true,
		},
		{
			name:                 "urgent alert delivered after initial sync",
			initialSyncCompleted: true,
			urgent:               true,
			wantDelivered:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockalert.NotifierMock{}
			ready := tt.initialSyncCompleted
			g := &gatedNotifier{Notifier: mock, isReady: func() bool { return ready }}

			b := g.Alert("body")
			if tt.urgent {
				b = b.Urgent()
			}
			b.Show()

			if tt.wantDelivered {
				require.Len(t, mock.Alerts, 1)
				assert.Equal(t, "body", mock.Alerts[0].Body)
			} else {
				assert.Empty(t, mock.Alerts)
			}
		})
	}
}
