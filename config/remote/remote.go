package remote

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// ConfigGetter get values from remote config
type ConfigGetter interface {
	GetTelioConfig() (string, error)
	FeatureConfig
}

type FeatureConfig interface {
	IsFeatureEnabled(featureName string) bool
	GetFeatureParam(featureName, paramName string) (string, error)
}

type ConfigLoader interface {
	LoadConfig() error
}

const (
	envUseLocalConfig   = "RC_USE_LOCAL_CONFIG"
	defaultFeatureState = true
)

type RemoteConfigEvent struct {
	MeshnetFeatureEnabled bool
}

type RemoteConfigNotifier interface {
	RemoteConfigUpdate(RemoteConfigEvent) error
}

type CdnRemoteConfig struct {
	appVersion     string
	appEnvironment string
	localCachePath string
	remotePath     string
	cdn            core.RemoteStorage
	features       *FeatureMap
	rolloutGroup   int
	analytics      Analytics
	initOnce       sync.Once
	mu             sync.RWMutex
	notifier       events.PublishSubcriber[RemoteConfigEvent]
}

// NewCdnRemoteConfig setup RemoteStorage based remote config loaded/getter
func NewCdnRemoteConfig(buildTarget config.BuildTarget, remotePath, localPath string,
	cdn core.RemoteStorage, analytics Analytics, appRollout int) *CdnRemoteConfig {
	rc := &CdnRemoteConfig{
		appVersion:     buildTarget.Version,
		appEnvironment: buildTarget.Environment,
		remotePath:     remotePath,
		localCachePath: localPath,
		cdn:            cdn,
		rolloutGroup:   appRollout,
		analytics:      analytics,
		features:       NewFeatureMap(),
		notifier:       &subs.Subject[RemoteConfigEvent]{},
	}
	rc.features.add(FeatureMain)
	rc.features.add(FeatureLibtelio)
	rc.features.add(FeatureMeshnet)
	return rc
}

type jsonFileReaderWriter struct{}

func (w jsonFileReaderWriter) writeFile(name string, content []byte, mode os.FileMode) error {
	return internal.SecureFileWrite(name, content, mode)
}
func (w jsonFileReaderWriter) readFile(name string) ([]byte, error) {
	return internal.SecureFileRead(name)
}

type noopWriter struct{}

func (w noopWriter) writeFile(name string, content []byte, mode os.FileMode) error {
	return nil
}

type jsonValidator struct{}

func (v jsonValidator) validate(content []byte) error {
	return validateJsonString(content)
}

type cdnFileGetter struct {
	cdn core.RemoteStorage
}

func (cfg cdnFileGetter) readFile(fname string) ([]byte, error) {
	return cfg.cdn.GetRemoteFile(fname)
}

// isNetworkRetryable returns true if the error is due to a transient network issue
func isNetworkRetryable(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	var opErr *net.OpError
	return errors.As(err, &opErr) ||
		errors.Is(err, core.ErrServerInternal) ||
		errors.Is(err, core.ErrTooManyRequests)
}

// LoadConfig download from remote or load from disk
func (c *CdnRemoteConfig) LoadConfig() error {
	var initErr error
	var reloadDone bool
	c.initOnce.Do(func() {
		validDir, err := internal.IsValidExistingDir(c.localCachePath)
		if err != nil {
			initErr = fmt.Errorf("accessing config path on init: %w", err)
		} else if validDir {
			c.load() // on start init cache from disk
			reloadDone = true
		}
	})
	if initErr != nil {
		return initErr
	}
	var err error
	var needReload bool
	useOnlyLocalConfig := internal.IsDevEnv(c.appEnvironment) && os.Getenv(envUseLocalConfig) != "" // forced load from disk?
	if !useOnlyLocalConfig {
		if needReload, err = c.download(); err != nil {
			return fmt.Errorf("downloading remote config: %w", err)
		}
	}

	// remote config files were downloaded and need to be reloaded?
	if needReload {
		c.load()
		reloadDone = true
	}

	if reloadDone {
		// notify what is current state after config reload
		c.notifier.Publish(RemoteConfigEvent{MeshnetFeatureEnabled: c.IsFeatureEnabled(FeatureMeshnet)})
	}

	return nil
}

func (c *CdnRemoteConfig) download() (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newChangesDownloaded := false

	for _, f := range c.features.keys() {
		feature := c.features.get(f)
		dnld, err := feature.download(cdnFileGetter{cdn: c.cdn}, jsonFileReaderWriter{}, jsonValidator{}, filepath.Join(c.remotePath, c.appEnvironment), c.localCachePath)
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed downloading feature [", feature, "] remote config:", err)

			var downloadErr *DownloadError
			if errors.As(err, &downloadErr) {
				c.analytics.EmitDownloadFailureEvent(ClientCli, feature.name, *downloadErr)
			}
			if isNetworkRetryable(err) {
				return false, err
			}
			continue
		}
		if dnld {
			// only if remote config was really downloaded
			log.Println(internal.InfoPrefix, "feature [", feature, "] remote config downloaded to:", c.localCachePath)
			c.analytics.EmitDownloadEvent(ClientCli, feature.name)
			newChangesDownloaded = true
		}
	}
	return newChangesDownloaded, nil
}

func (c *CdnRemoteConfig) load() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, f := range c.features.keys() {
		feature := c.features.get(f)
		if err := feature.load(c.localCachePath, jsonFileReaderWriter{}, jsonValidator{}); err != nil {
			var loadErr *LoadError
			if errors.As(err, &loadErr) {
				// only specific load errors are used for JSON related failures
				if loadErr.Kind == LoadErrorParsing || loadErr.Kind == LoadErrorParsingIncludeFile || loadErr.Kind == LoadErrorMainHashJsonParsing || loadErr.Kind == LoadErrorMainJsonValidationFailure {
					c.analytics.EmitJsonParseFailureEvent(ClientCli, feature.name, *loadErr)
				}
			}
			log.Println(internal.ErrorPrefix, "failed loading feature [", feature.name, "] config from the disk:", err)
			c.analytics.EmitLocalUseEvent(ClientCli, feature.name, err)
			continue
		}
		log.Println(internal.InfoPrefix, "feature [", feature.name, "] config loaded from:", c.localCachePath)
		c.analytics.EmitLocalUseEvent(ClientCli, feature.name, nil)
	}
}

func (c *CdnRemoteConfig) findMatchingRecord(ss []ParamValue, featureName string) (match *ParamValue) {
	for _, s := range ss {
		// find my version matching records
		ok, err := isVersionMatching(c.appVersion, s.AppVersion)
		if err != nil {
			log.Println(internal.ErrorPrefix, "invalid version:", err)
			continue
		}
		if ok {
			// find matching item with highest weight
			if match == nil {
				match = &s
			} else {
				if s.Weight > match.Weight {
					match = &s
				}
			}
		}
	}
	// as a last step, check if app's rollout group matches feature's rollout value
	// (do not try to use other match with lesser weight)
	if match != nil {
		if match.Rollout > c.rolloutGroup {
			c.analytics.EmitPartialRolloutEvent(ClientCli, featureName, match.Rollout, partialRolloutPerformedFailure)
			match = nil
		} else {
			c.analytics.EmitPartialRolloutEvent(ClientCli, featureName, match.Rollout, partialRolloutPerformedSuccess)
		}
	} else {
		//when there's no match (eg., due to a version value mismatch) emit the partial rollout event failure
		//TODO: change the structure of this event to hold error anyway?
		c.analytics.EmitPartialRolloutEvent(ClientCli, featureName, 0, partialRolloutPerformedFailure)
	}
	return match
}

func (c *CdnRemoteConfig) GetTelioConfig() (string, error) {
	return c.GetFeatureParam(FeatureLibtelio, FeatureLibtelio)
}

func (c *CdnRemoteConfig) IsFeatureEnabled(featureName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// find by name, expect param name to be the same as feature name and expect boolean type
	f := c.features.get(featureName)
	if f == nil {
		return defaultFeatureState
	}
	p, found := f.params[featureName]
	if !found {
		return defaultFeatureState
	}
	switch p.Type {
	case fieldTypeBool:
		if item := c.findMatchingRecord(p.Settings, featureName); item != nil {
			val, _ := item.AsBool()
			return val
		}
	}
	return defaultFeatureState
}

func (c *CdnRemoteConfig) GetFeatureParam(featureName, paramName string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	f := c.features.get(featureName)
	if f == nil {
		return "", fmt.Errorf("feature [%s] not found", featureName)
	}
	p, found := f.params[paramName]
	if !found {
		return "", fmt.Errorf("feature [%s] param [%s] not found", featureName, paramName)
	}
	if item := c.findMatchingRecord(p.Settings, featureName); item != nil {
		switch p.Type {
		case fieldTypeBool:
			val, err := item.AsBool()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%t", val), nil
		case fieldTypeString, fieldTypeObject:
			val, err := item.AsString()
			if err != nil {
				return "", err
			}
			return val, nil
		case fieldTypeInt, fieldTypeNumber:
			val, err := item.AsInt()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", val), nil
		case fieldTypeArray:
			val, err := item.AsStringArray()
			if err != nil {
				return "", err
			}
			return strings.Join(val, ", "), nil
		case fieldTypeFile:
			return item.incValue, nil
		}
	}
	return "", fmt.Errorf("feature [%s] param [%s] value not found", featureName, paramName)
}

func (c *CdnRemoteConfig) Subscribe(to RemoteConfigNotifier) {
	c.notifier.Subscribe(to.RemoteConfigUpdate)
}
