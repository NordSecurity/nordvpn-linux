package daemon

import (
	"context"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

func TestSetECH(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		currentECH    bool
		requested     bool
		vpnActive     bool
		loadErr       bool
		saveErr       bool
		expectedType  int64
		expectSaved   bool
		expectedValue bool // persisted ECH value when saved
	}{
		{
			name:          "enable when disabled",
			currentECH:    false,
			requested:     true,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: true,
		},
		{
			name:          "disable when enabled",
			currentECH:    true,
			requested:     false,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: false,
		},
		{
			name:         "no-op when already enabled",
			currentECH:   true,
			requested:    true,
			expectedType: internal.CodeNothingToDo,
			expectSaved:  false,
		},
		{
			name:         "no-op when already disabled",
			currentECH:   false,
			requested:    false,
			expectedType: internal.CodeNothingToDo,
			expectSaved:  false,
		},
		{
			name:          "enable while vpn active reports active",
			currentECH:    false,
			requested:     true,
			vpnActive:     true,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: true,
		},
		{
			name:         "load error",
			currentECH:   false,
			requested:    true,
			loadErr:      true,
			expectedType: internal.CodeConfigError,
			expectSaved:  false,
		},
		{
			name:         "save error",
			currentECH:   false,
			requested:    true,
			saveErr:      true,
			expectedType: internal.CodeConfigError,
			expectSaved:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := mock.NewMockConfigManager()
			cm.Cfg.ECH.Set(test.currentECH)
			if test.loadErr {
				cm.LoadErr = assert.AnError
			}
			if test.saveErr {
				cm.SaveErr = assert.AnError
			}

			netw := networker.Mock{VpnActive: test.vpnActive}

			r := RPC{cm: cm, netw: &netw}

			resp, err := r.SetECH(context.Background(), &pb.SetGenericRequest{Enabled: test.requested})
			assert.NoError(t, err)
			assert.Equal(t, test.expectedType, resp.Type)

			if test.expectedType == internal.CodeSuccess {
				assert.Equal(t, []string{strconv.FormatBool(test.vpnActive)}, resp.Data)
			}

			if test.expectSaved {
				assert.True(t, cm.Saved)
				assert.Equal(t, test.expectedValue, cm.Cfg.ECH.Get())
			} else if !test.loadErr {
				// On no-op the value must be unchanged; save must not have persisted a change.
				assert.Equal(t, test.currentECH, cm.Cfg.ECH.Get())
			}
		})
	}
}
