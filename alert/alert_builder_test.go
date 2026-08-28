package alert

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertBuilderOnShown(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		setOnShown      bool
		callShow        bool
		sendReturns     bool
		wantSendCalled  bool
		wantShownCalled bool
	}{
		{
			name:            "show invokes OnShown after sending the alert",
			setOnShown:      true,
			callShow:        true,
			sendReturns:     true,
			wantSendCalled:  true,
			wantShownCalled: true,
		},
		{
			name:            "show does not panic when OnShown is unset",
			setOnShown:      false,
			callShow:        true,
			sendReturns:     true,
			wantSendCalled:  true,
			wantShownCalled: false,
		},
		{
			name:            "OnShown is not invoked before Show is called",
			setOnShown:      true,
			callShow:        false,
			sendReturns:     true,
			wantSendCalled:  false,
			wantShownCalled: false,
		},
		{
			name:            "OnShown is not invoked when send reports the alert was not shown",
			setOnShown:      true,
			callShow:        true,
			sendReturns:     false,
			wantSendCalled:  true,
			wantShownCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sendCalled bool
			var sentBody string
			var shownCalled bool
			var bodyAtShownTime string

			send := func(a Alert) bool {
				sendCalled = true
				sentBody = a.Body
				return tt.sendReturns
			}

			b := NewAlertBuilder(send, "body")
			if tt.setOnShown {
				b = b.OnShown(func() {
					shownCalled = true
					bodyAtShownTime = sentBody
				})
			}
			require.NotNil(t, b)

			if tt.callShow {
				assert.NotPanics(t, func() { b.Show() })
			}

			assert.Equal(t, tt.wantSendCalled, sendCalled)
			assert.Equal(t, tt.wantShownCalled, shownCalled)
			if tt.wantShownCalled {
				assert.Equal(t, "body", bodyAtShownTime)
			}
		})
	}
}
