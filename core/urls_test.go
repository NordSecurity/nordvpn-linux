package core

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestConnectionLimitReachedGuideURL(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		appID       AppID
		expectedURL string
	}{
		{
			name:        "tray limit reached guide URL is constructed properly",
			appID:       TrayAppID,
			expectedURL: "https://support.nordvpn.com/hc/en-us/articles/47181405478417-I-get-the-Session-Limit-Reached-error-on-NordVPN?utm_medium=app&utm_source=nordvpn-linux-tray&utm_campaign=ens_error-session_limit&nm=app&ns=nordvpn-linux-tray&nc=ens_error-session_limit",
		},

		{
			name:        "CLI limit reached guide URL is constructed properly",
			appID:       CLIAppID,
			expectedURL: "https://support.nordvpn.com/hc/en-us/articles/47181405478417-I-get-the-Session-Limit-Reached-error-on-NordVPN?utm_medium=app&utm_source=nordvpn-linux-cli&utm_campaign=ens_error-session_limit&nm=app&ns=nordvpn-linux-cli&nc=ens_error-session_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualURL := ConnectionLimitReachedGuideURL(tt.appID)
			assert.Equal(t, tt.expectedURL, actualURL)
		})
	}
}
