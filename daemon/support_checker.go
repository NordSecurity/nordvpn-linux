package daemon

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/coreos/go-semver/semver"
)

// SupportChecker is used to check whether specified feature can be enabled
type SupportChecker interface {
	IsSupported(feature config.Feature) (bool, error)
}

// MockSupportChecker supports all features
type MockSupportChecker struct {
}

// NewMockSupportChecker creates MockSupportChecker
func NewMockSupportChecker() *MockSupportChecker {
	return &MockSupportChecker{}
}

func (*MockSupportChecker) IsSupported(feature config.Feature) (bool, error) {
	return true, nil
}

// APISupportChecker checks feature support using credentialsAPI and versionGetter
type APISupportChecker struct {
	configManager       config.Manager
	appVersion          semver.Version
	versionGetter       remote.SupportedVersionGetter
	credentialsAPI      core.CredentialsAPI
	featureDisabledSubs map[config.Feature]*subs.Subject[any]
	updatePeriod        time.Duration
	services            core.ServicesResponse
	mutex               sync.Mutex
}

// NewAPISupportChecker creates API based support checker
func NewAPISupportChecker(
	cm config.Manager,
	version string,
	versionGetter remote.SupportedVersionGetter,
	credsAPI core.CredentialsAPI,
	updatePeriod time.Duration,
) (*APISupportChecker, error) {
	appVersion := version
	// if version development
	if strings.Contains(appVersion, "+") {
		appVersion = strings.Split(appVersion, "+")[0]
	}
	appVersionSemver, err := semver.NewVersion(appVersion)
	if err != nil {
		return nil, fmt.Errorf("parsing app version: %w", err)
	}

	featureDisabledSubs := map[config.Feature]*subs.Subject[any]{}
	for feature := range config.Feature_name {
		featureDisabledSubs[config.Feature(feature)] = &subs.Subject[any]{}
	}

	return &APISupportChecker{
		configManager:       cm,
		appVersion:          *appVersionSemver,
		versionGetter:       versionGetter,
		credentialsAPI:      credsAPI,
		featureDisabledSubs: featureDisabledSubs,
		updatePeriod:        updatePeriod,
	}, nil
}

// GetFeatureDisabledSubs returns Subjects to get notifications when a previously available
// feature become unavailable
func (sc *APISupportChecker) GetFeatureDisabledSubs() map[config.Feature]*subs.Subject[any] {
	return sc.featureDisabledSubs
}

// IsSupported checks whether requested feature can be used
// Thread safe
func (sc *APISupportChecker) IsSupported(feature config.Feature) (bool, error) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	var cfg config.Config
	if err := sc.configManager.Load(&cfg); err != nil {
		return false, fmt.Errorf("loading config: %w", err)
	}

	featureConfig := cfg.Features[feature]
	if time.Since(featureConfig.LastUpdate) < sc.updatePeriod {
		return featureConfig.IsSupported, nil
	}
	isFeatureEnabled, err := sc.isFeatureEnabledForUser(feature, cfg)
	if err != nil {
		log.Printf("error checking if feature is enabled for user: %s", err)
		isFeatureEnabled = featureConfig.IsSupported
	}
	isVersionCompatible, err := sc.isFeatureVersionCompatible(feature)
	if err != nil {
		log.Printf("error checking if feature is version compatible: %s", err)
		isVersionCompatible = featureConfig.IsSupported
	}

	isSupported := isFeatureEnabled && isVersionCompatible
	if !isSupported && featureConfig.IsSupported {
		sc.featureDisabledSubs[feature].Publish(nil)
	}
	featureConfig.IsSupported = isSupported
	featureConfig.LastUpdate = time.Now()

	err = sc.configManager.SaveWith(func(c config.Config) config.Config {
		c.Features[feature] = featureConfig
		return c
	})
	if err != nil {
		log.Printf("error saving config: %s", err)
	}

	return featureConfig.IsSupported, nil
}

const (
	meshnetFeatureID = 19
)

func (sc *APISupportChecker) isFeatureEnabledForUser(feature config.Feature, cfg config.Config) (bool, error) {
	var featureID int64
	switch feature {
	// Fileshare does not have its' own service but it depends on meshnet
	case config.Feature_MESHNET, config.Feature_FILESHARE:
		featureID = meshnetFeatureID
	case config.Feature_NAT_TRAVERSAL, config.Feature_TELIO_ANALYTICS:
		// These features cannot be enabled per user
		return true, nil
	case config.Feature_UNKNOWN_FEATURE:
		fallthrough
	default:
		return false, fmt.Errorf("unknown feature %s", feature)
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	services, err := sc.credentialsAPI.Services(token)
	if err != nil {
		return false, fmt.Errorf("requesting services: %w", err)
	}

	sc.services = services

	for _, service := range sc.services {
		if service.Service.ID == featureID {
			expiry, err := time.Parse(internal.ServerDateFormat, service.ExpiresAt)
			if err != nil || expiry.Before(time.Now()) {
				return false, nil
			}
			return true, nil
		}
	}
	return false, nil
}

func (sc *APISupportChecker) isFeatureVersionCompatible(feature config.Feature) (bool, error) {
	var versionKey string
	switch feature {
	case config.Feature_MESHNET:
		// Enabled for life
		return true, nil
	case config.Feature_FILESHARE:
		versionKey = remote.RcFileSharingMinVerKey
	case config.Feature_NAT_TRAVERSAL:
		versionKey = remote.RcNatTraversalMinVerKey
	case config.Feature_TELIO_ANALYTICS:
		versionKey = remote.RcTelioAnalyticsMinVerKey
	case config.Feature_UNKNOWN_FEATURE:
		fallthrough
	default:
		return false, fmt.Errorf(
			"can't retrieve version from remote config for unknown feature %s", feature,
		)
	}

	minVersion, err := sc.versionGetter.GetMinFeatureVersion(versionKey)
	if err != nil {
		return false, fmt.Errorf("could not get min version for feature %s, %w", feature, err)
	}

	return !sc.appVersion.LessThan(*minVersion), nil
}
