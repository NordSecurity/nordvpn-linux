package daemon

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestHandleNotificationType(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name             string
		notificationType NotificationType
		args             []string
		expected         string
	}{
		{
			name:             "connected",
			notificationType: internal.NotificationConnected,
			args:             []string{},
			expected:         "You are connected to %!s(MISSING) (%!s(MISSING))!",
		},
		{
			name:             "reconnected",
			notificationType: internal.NotificationReconnected,
			args: []string{
				"en0",
				"virtual_en0",
			},
			expected: "You have been reconnected to en0 (virtual_en0)",
		},
		{
			name:             "disconnected",
			notificationType: internal.NotificationDisconnected,
			args:             []string{},
			expected:         "You are disconnected from NordVPN.",
		},
		{
			name:             "notificationType unknown",
			notificationType: 65,
			args:             []string{},
			expected:         "Unknown type (65)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := handleNotificationType(test.notificationType, test.args)
			assert.Equal(t, got, test.expected)
		})
	}
}
