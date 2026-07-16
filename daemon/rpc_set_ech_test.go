package daemon

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

// echRemoteGetterMock is a controllable remote.ConfigGetter for the ECH remote gate.
type echRemoteGetterMock struct {
	param string
	err   error
}

func (echRemoteGetterMock) GetTelioConfig() (string, error) { return "", nil }
func (echRemoteGetterMock) IsFeatureEnabled(string) bool    { return false }
func (m echRemoteGetterMock) GetFeatureParam(_, _ string) (string, error) {
	return m.param, m.err
}

func TestSetECH(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		tech          config.Technology
		remoteParam   string // remote enable_ech value; "" with no err means default-on
		remoteErr     error
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
			tech:          config.Technology_NORDWHISPER,
			remoteParam:   "true",
			currentECH:    false,
			requested:     true,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: true,
		},
		{
			name:          "disable when enabled",
			tech:          config.Technology_NORDWHISPER,
			remoteParam:   "true",
			currentECH:    true,
			requested:     false,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: false,
		},
		{
			name:         "no-op when already enabled",
			tech:         config.Technology_NORDWHISPER,
			remoteParam:  "true",
			currentECH:   true,
			requested:    true,
			expectedType: internal.CodeNothingToDo,
			expectSaved:  false,
		},
		{
			name:          "enable while vpn active reports active",
			tech:          config.Technology_NORDWHISPER,
			remoteParam:   "true",
			currentECH:    false,
			requested:     true,
			vpnActive:     true,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: true,
		},
		{
			name:          "remote missing defaults to enabled",
			tech:          config.Technology_NORDWHISPER,
			remoteParam:   "",
			currentECH:    false,
			requested:     true,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: true,
		},
		{
			name:          "remote error defaults to enabled",
			tech:          config.Technology_NORDWHISPER,
			remoteErr:     errors.New("boom"),
			currentECH:    false,
			requested:     true,
			expectedType:  internal.CodeSuccess,
			expectSaved:   true,
			expectedValue: true,
		},
		{
			name:         "wrong technology openvpn",
			tech:         config.Technology_OPENVPN,
			remoteParam:  "true",
			currentECH:   false,
			requested:    true,
			expectedType: internal.CodeECHTechUnsupported,
			expectSaved:  false,
		},
		{
			name:         "wrong technology nordlynx",
			tech:         config.Technology_NORDLYNX,
			remoteParam:  "true",
			currentECH:   false,
			requested:    true,
			expectedType: internal.CodeECHTechUnsupported,
			expectSaved:  false,
		},
		{
			name:         "globally disabled by remote config",
			tech:         config.Technology_NORDWHISPER,
			remoteParam:  "false",
			currentECH:   false,
			requested:    true,
			expectedType: internal.CodeECHGloballyDisabled,
			expectSaved:  false,
		},
		{
			name:         "globally disabled blocks even matching value",
			tech:         config.Technology_NORDWHISPER,
			remoteParam:  "false",
			currentECH:   true,
			requested:    true, // matches current, but guard runs before no-op
			expectedType: internal.CodeECHGloballyDisabled,
			expectSaved:  false,
		},
		{
			name:         "load error",
			tech:         config.Technology_NORDWHISPER,
			remoteParam:  "true",
			currentECH:   false,
			requested:    true,
			loadErr:      true,
			expectedType: internal.CodeConfigError,
			expectSaved:  false,
		},
		{
			name:         "save error",
			tech:         config.Technology_NORDWHISPER,
			remoteParam:  "true",
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
			cm.Cfg.Technology = test.tech
			cm.Cfg.ECH.Set(test.currentECH)
			if test.loadErr {
				cm.LoadErr = assert.AnError
			}
			if test.saveErr {
				cm.SaveErr = assert.AnError
			}

			netw := networker.Mock{VpnActive: test.vpnActive}

			r := RPC{
				cm:                 cm,
				netw:               &netw,
				remoteConfigGetter: echRemoteGetterMock{param: test.remoteParam, err: test.remoteErr},
			}

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
				// When not saved, the stored value must be unchanged.
				assert.Equal(t, test.currentECH, cm.Cfg.ECH.Get())
			}
		})
	}
}
