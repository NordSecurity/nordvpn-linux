package remote

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

type ensStubAnalytics struct{}

func (ensStubAnalytics) EmitDownloadEvent(string, string)                       {}
func (ensStubAnalytics) EmitDownloadFailureEvent(string, string, DownloadError) {}
func (ensStubAnalytics) EmitLocalUseEvent(string, string, error)                {}
func (ensStubAnalytics) EmitJsonParseFailureEvent(string, string, LoadError)    {}
func (ensStubAnalytics) EmitPartialRolloutEvent(string, string, int, bool)      {}
func (ensStubAnalytics) ClearEventFlags()                                       {}

func newENSTestRemoteConfig() *CdnRemoteConfig {
	return NewCdnRemoteConfig(
		config.BuildTarget{Version: "1.2.3", Environment: "test"},
		"/remote/path",
		"/local/path",
		&MockRemoteStorage{},
		ensStubAnalytics{},
		100,
	)
}

func setENSRecord(rc *CdnRemoteConfig, value bool) {
	rc.features.get(FeatureENS).params = map[string]*Param{
		FeatureENS: {
			Type: fieldTypeBool,
			Settings: []ParamValue{
				{AppVersion: "^1.0.0", Value: value, Weight: 100, TargetRollout: 0},
			},
		},
	}
}

func TestCdnRemoteConfig_IsFeatureEnabled_ENSDefaultsEnabledWhenAbsent(t *testing.T) {
	category.Set(t, category.Unit)

	rc := newENSTestRemoteConfig()
	assert.True(t, rc.IsFeatureEnabled(FeatureENS))
}

func TestCdnRemoteConfig_IsFeatureEnabled_ENSRespectsRemoteConfigValue(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		rcValue  bool
		expected bool
	}{
		{name: "remote config disables ENS", rcValue: false, expected: false},
		{name: "remote config enables ENS", rcValue: true, expected: true},
	}

	for _, tt := range tests {
		rc := newENSTestRemoteConfig()
		setENSRecord(rc, tt.rcValue)

		assert.Equal(t, tt.expected, rc.IsFeatureEnabled(FeatureENS), tt.name)
	}
}
