package daemon

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
)

type mockVersionGetter struct {
	Version semver.Version
	Error   error
}

func (m *mockVersionGetter) GetMinFeatureVersion(featureKey string) (*semver.Version, error) {
	return &m.Version, m.Error
}
func (m *mockVersionGetter) MinimalVersion() (*semver.Version, error) { return nil, nil }

type mockCredentialsAPI struct {
	Resp  core.ServicesResponse
	Error error
}

func (m *mockCredentialsAPI) Services(string) (core.ServicesResponse, error) {
	return m.Resp, m.Error
}
func (m *mockCredentialsAPI) NotificationCredentials(string, string) (r core.NotificationCredentialsResponse, e error) {
	return r, e
}
func (m *mockCredentialsAPI) ServiceCredentials(string) (r *core.CredentialsResponse, e error) {
	return r, e
}
func (m *mockCredentialsAPI) TokenRenew(string) (r *core.TokenRenewResponse, e error)   { return r, e }
func (m *mockCredentialsAPI) DeleteToken(string) (e error)                              { return e }
func (m *mockCredentialsAPI) CurrentUser(string) (r *core.CurrentUserResponse, e error) { return r, e }

type mockSCConfigManager struct {
	featureConfigs map[config.Feature]config.FeatureConfig
	userID         int64
}

func (m *mockSCConfigManager) SaveWith(f config.SaveFunc) error {
	m.featureConfigs = f(config.Config{Features: map[config.Feature]config.FeatureConfig{}}).Features
	return nil
}
func (m *mockSCConfigManager) Load(c *config.Config) error {
	c.Features = m.featureConfigs
	c.AutoConnectData.ID = m.userID
	return nil
}
func (*mockSCConfigManager) Reset() error { return nil }

func getBasicAPISupportChecker(t *testing.T) (*APISupportChecker, *mockSCConfigManager) {
	conf := &mockSCConfigManager{map[config.Feature]config.FeatureConfig{}, 0}
	c, err := NewAPISupportChecker(
		conf,
		"1.0.0",
		&mockVersionGetter{Version: *semver.New("2.0.0")},
		&mockCredentialsAPI{},
		time.Second*0,
	)
	assert.NoError(t, err)
	return c, conf
}

func TestNewAPISupportChecker_DevVersion(t *testing.T) {
	c, err := NewAPISupportChecker(
		&mockSCConfigManager{map[config.Feature]config.FeatureConfig{}, 0},
		"3.14.2+b2a71f00 - dev (b2a71f00)",
		&mockVersionGetter{},
		&mockCredentialsAPI{},
		time.Second*0,
	)
	assert.NoError(t, err)
	assert.Equal(t, int(c.appVersion.Major), 3)
	assert.Equal(t, int(c.appVersion.Minor), 14)
	assert.Equal(t, int(c.appVersion.Patch), 2)
}

func TestNewAPISupportChecker_InvalidVersion(t *testing.T) {
	_, err := NewAPISupportChecker(
		&mockSCConfigManager{map[config.Feature]config.FeatureConfig{}, 0},
		"3.14.2.1",
		&mockVersionGetter{},
		&mockCredentialsAPI{},
		time.Second*0,
	)
	assert.Error(t, err)
}

func TestIsSupported_UnknownFeature(t *testing.T) {
	category.Set(t, category.Unit)

	checker, conf := getBasicAPISupportChecker(t)
	isSupported, err := checker.IsSupported(config.Feature_UNKNOWN_FEATURE)
	assert.NoError(t, err)
	assert.False(t, isSupported)
	_, ok := conf.featureConfigs[config.Feature_MESHNET]
	assert.False(t, ok)
}

func TestIsSupported_Variations(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name               string
		isServiceEnabled   bool
		isServiceError     bool
		isVersionSupported bool
		isVersionError     bool
		wasSupported       bool
		result             bool
	}{
		{
			name:               "not supported at all",
			isServiceEnabled:   false,
			isVersionSupported: false,
			result:             false,
		},
		{
			name:               "only version supported",
			isServiceEnabled:   false,
			isVersionSupported: true,
			result:             false,
		},
		{
			name:               "only service enabled",
			isServiceEnabled:   true,
			isVersionSupported: false,
			// result: true because there are currently no services which would follow
			// this logic.
			// TODO: Change this to false once other feature than meshnet is used
			result: true,
		},
		{
			name:               "fully supported",
			isServiceEnabled:   true,
			isVersionSupported: true,
			result:             true,
		},
		{
			name:           "API down",
			isServiceError: true,
			isVersionError: true,
			wasSupported:   false,
			result:         false,
		},
		{
			name:           "API down but was supported",
			isServiceError: true,
			isVersionError: true,
			wasSupported:   true,
			result:         true,
		},
		{
			name:               "was supported 1",
			isServiceError:     true,
			isVersionSupported: false,
			wasSupported:       true,
			// result: true because there are currently no services which would follow
			// this logic.
			// TODO: Change this to false once other feature than meshnet is used
			result: true,
		},
		{
			name:               "was supported 2",
			isServiceError:     true,
			isVersionSupported: true,
			wasSupported:       true,
			result:             true,
		},
		{
			name:             "was supported 3",
			isServiceEnabled: false,
			isVersionError:   true,
			wasSupported:     true,
			result:           false,
		},
		{
			name:             "was supported 4",
			isServiceEnabled: true,
			isVersionError:   true,
			wasSupported:     true,
			result:           true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checker, conf := getBasicAPISupportChecker(t)
			if test.isServiceEnabled {
				checker.credentialsAPI = &mockCredentialsAPI{
					Resp: core.ServicesResponse{core.ServiceData{
						Service:   core.Service{ID: meshnetFeatureID},
						ExpiresAt: "2100-01-01 01:01:01",
					}},
				}
			}
			if test.isServiceError {
				checker.credentialsAPI = &mockCredentialsAPI{
					Error: errors.New("credentials API is unavailable"),
				}
			}
			if test.isVersionSupported {
				checker.versionGetter = &mockVersionGetter{Version: *semver.New("1.0.0")}
			}
			if test.isVersionError {
				checker.versionGetter = &mockVersionGetter{Error: errors.New("version API is unavailable")}
			}
			if test.wasSupported {
				conf = &mockSCConfigManager{featureConfigs: map[config.Feature]config.FeatureConfig{config.Feature_MESHNET: {IsSupported: true}}}
				checker.configManager = conf
			}

			isSupported, err := checker.IsSupported(config.Feature_MESHNET)
			assert.NoError(t, err)
			assert.Equal(t, test.result, isSupported)
			featureConf, ok := conf.featureConfigs[config.Feature_MESHNET]
			assert.True(t, ok)
			assert.Equal(t, test.result, featureConf.IsSupported)
		})
	}
}

func TestIsSupported_FeatureDisabledNotification(t *testing.T) {
	category.Set(t, category.Unit)

	checker, _ := getBasicAPISupportChecker(t)

	notified := false
	checker.GetFeatureDisabledSubs()[config.Feature_NAT_TRAVERSAL].Subscribe(func(a any) error {
		notified = true
		return nil
	})

	checker.credentialsAPI = &mockCredentialsAPI{}
	checker.versionGetter = &mockVersionGetter{Version: *semver.New("1.0.0")}
	checker.IsSupported(config.Feature_NAT_TRAVERSAL)
	assert.False(t, notified)

	checker.versionGetter = &mockVersionGetter{Version: *semver.New("2.0.0")}
	checker.IsSupported(config.Feature_NAT_TRAVERSAL)
	assert.True(t, notified)
}

func TestIsSupported_ServiceCaching(t *testing.T) {
	category.Set(t, category.Unit)

	checker, cm := getBasicAPISupportChecker(t)
	checker.updatePeriod = time.Hour * 24
	checker.credentialsAPI = &mockCredentialsAPI{
		Resp: core.ServicesResponse{core.ServiceData{
			Service:   core.Service{ID: meshnetFeatureID},
			ExpiresAt: "2100-01-01 01:01:01",
		}},
	}
	checker.versionGetter = &mockVersionGetter{Version: *semver.New("1.0.0")}
	isSupported, err := checker.IsSupported(config.Feature_MESHNET)
	assert.NoError(t, err)
	assert.True(t, isSupported)

	checker.credentialsAPI = &mockCredentialsAPI{}
	isSupported, err = checker.IsSupported(config.Feature_MESHNET)
	assert.NoError(t, err)
	assert.True(t, isSupported)

	meshnetConfig := cm.featureConfigs[config.Feature_MESHNET]
	meshnetConfig.LastUpdate = time.Now().Add(-time.Hour * 48)
	cm.featureConfigs[config.Feature_MESHNET] = meshnetConfig
	isSupported, err = checker.IsSupported(config.Feature_MESHNET)
	assert.NoError(t, err)
	assert.False(t, isSupported)
}

func TestIsSupported_ServiceExpired(t *testing.T) {
	category.Set(t, category.Unit)

	checker, _ := getBasicAPISupportChecker(t)
	checker.credentialsAPI = &mockCredentialsAPI{
		Resp: core.ServicesResponse{core.ServiceData{
			Service:   core.Service{ID: meshnetFeatureID},
			ExpiresAt: "2022-07-01 00:00:00",
		}},
	}
	checker.versionGetter = &mockVersionGetter{Version: *semver.New("1.0.0")}
	isSupported, err := checker.IsSupported(config.Feature_MESHNET)
	assert.NoError(t, err)
	assert.False(t, isSupported)
}
