package client

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestENSConnectionLimitReached(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		appID       core.AppID
		expectedMsg string
	}{
		{
			name:        "tray limit reached message is constructed properly",
			appID:       core.TrayAppID,
			expectedMsg: "Wait a while before trying again. Retrying now can make the waiting period longer. If the issue persists, check our help guide for other possible causes.",
		},
		{
			name:        "CLI limit reached message is constructed properly",
			appID:       core.CLIAppID,
			expectedMsg: "Too many connection attempts. Wait a while before trying again. Retrying now can make the waiting period longer. If the issue persists, check our help guide for other possible causes: https://support.nordvpn.com/hc/en-us/articles/47181405478417-I-get-the-Session-Limit-Reached-error-on-NordVPN?utm_medium=app&utm_source=nordvpn-linux-cli&utm_campaign=ens_error-session_limit&nm=app&ns=nordvpn-linux-cli&nc=ens_error-session_limit",
		},
		{
			name:        "unknown AppID returns fallback message",
			appID:       core.AppID("unknown"),
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMsg := ENSConnectionLimitReached(tt.appID)
			assert.Equal(t, tt.expectedMsg, actualMsg)
		})
	}
}
