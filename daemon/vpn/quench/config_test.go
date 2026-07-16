//go:build quench

package quench

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

// fakeRemoteGetter is a controllable remote.ConfigGetter for the ECH gate matrix.
type fakeRemoteGetter struct {
	param string
	err   error
}

func (f fakeRemoteGetter) GetTelioConfig() (string, error) { return "", nil }
func (f fakeRemoteGetter) IsFeatureEnabled(string) bool    { return false }
func (f fakeRemoteGetter) GetFeatureParam(_, _ string) (string, error) {
	return f.param, f.err
}

func TestNordWhisperConfig_GetConfig_ECHGate(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		remoteParam string
		remoteErr   error
		userSet     bool // whether the user ECH field is explicitly set
		userValue   bool
		loadErr     bool
		expectedECH bool
	}{
		{name: "remote on, user on", remoteParam: "true", userSet: true, userValue: true, expectedECH: true},
		{name: "remote off gates user on", remoteParam: "false", userSet: true, userValue: true, expectedECH: false},
		{name: "user off gates remote on", remoteParam: "true", userSet: true, userValue: false, expectedECH: false},
		{name: "remote off, user off", remoteParam: "false", userSet: true, userValue: false, expectedECH: false},
		{name: "remote error defaults on, user on", remoteErr: errors.New("boom"), userSet: true, userValue: true, expectedECH: true},
		{name: "remote malformed defaults on, user on", remoteParam: "not-a-bool", userSet: true, userValue: true, expectedECH: true},
		{name: "remote on, user unset defaults on", remoteParam: "true", userSet: false, expectedECH: true},
		{name: "remote on, user load error defaults on", remoteParam: "true", loadErr: true, expectedECH: true},
		{name: "remote off, user unset", remoteParam: "false", userSet: false, expectedECH: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := mock.NewMockConfigManager()
			if test.userSet {
				cm.Cfg.ECH.Set(test.userValue)
			}
			if test.loadErr {
				cm.LoadErr = errors.New("load failed")
			}

			qc := NewNordWhisperConfig(cm, fakeRemoteGetter{param: test.remoteParam, err: test.remoteErr})

			features, err := qc.GetConfig()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedECH, features.EnableECH)
		})
	}
}
