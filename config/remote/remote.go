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
	appVersion      string
	appEnvironment  string
	localCachePath  string
	remotePath      string
	cdn             core.RemoteStorage
	features        *FeatureMap
	appRolloutGroup int
	analytics       Analytics
	initOnce        sync.Once
	mu              sync.RWMutex
	notifier        events.PublishSubcriber[RemoteConfigEvent]
}

// NewCdnRemoteConfig setup RemoteStorage based remote config loaded/getter
func NewCdnRemoteConfig(buildTarget config.BuildTarget, remotePath, localPath string,
	cdn core.RemoteStorage, analytics Analytics, appRollout int) *CdnRemoteConfig {
	rc := &CdnRemoteConfig{
		appVersion:      buildTarget.Version,
		appEnvironment:  buildTarget.Environment,
		remotePath:      remotePath,
		localCachePath:  localPath,
		cdn:             cdn,
		appRolloutGroup: appRollout,
		analytics:       analytics,
		features:        NewFeatureMap(),
		notifier:        &subs.Subject[RemoteConfigEvent]{},
	}
	rc.features.add(FeatureMain)
	rc.features.add(FeatureLibtelio)
	rc.features.add(FeatureMeshnet)
	return rc
}

type jsonFileReaderWriter struct{}

func (w jsonFileReaderWriter) writeFile(name string, content []byte, mode os.FileMode) error {
	return internal.FileWrite(name, content, mode)
}
func (w jsonFileReaderWriter) readFile(name string) ([]byte, error) {
	// try to prevent overloading
	if err := internal.IsFileTooBig(name); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return internal.FileRead(name)
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
	var err error
	reloadDone := false
	c.initOnce.Do(func() {
		c.load() // on start init cache from disk
		reloadDone = true
	})
	needReload := false
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
			log.Printf("%s failed downloading feature [%s] remote config: %v\n", internal.ErrorPrefix, feature.name, err)

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
			log.Printf("%s feature [%s] remote config downloaded to: %s\n", internal.InfoPrefix, feature.name, c.localCachePath)
			c.analytics.EmitDownloadEvent(ClientCli, feature.name)
			newChangesDownloaded = true
		}
	}
	return newChangesDownloaded, nil
}

// isJsonParsingError checks if reported LoadError is related to JSON parsing
func isJsonParsingError(errKind LoadErrorKind) bool {
	return errKind == LoadErrorParsing ||
		errKind == LoadErrorParsingIncludeFile ||
		errKind == LoadErrorMainHashJsonParsing ||
		errKind == LoadErrorMainJsonValidationFailure
}

// reportLoadError emits the appropriate analytics event based on the error type
func (c *CdnRemoteConfig) reportLoadError(featureName string, err error) {
	var loadErr *LoadError
	if !errors.As(err, &loadErr) {
		// For non-LoadError types, use the default event
		c.analytics.EmitLocalUseEvent(ClientCli, featureName, err)
		return
	}

	if isJsonParsingError(loadErr.Kind) {
		// For JSON parsing errors, emit the specialized event
		c.analytics.EmitJsonParseFailureEvent(ClientCli, featureName, *loadErr)
		return
	}

	// For all other error types, emit the local use error event
	c.analytics.EmitLocalUseEvent(ClientCli, featureName, err)
}

func (c *CdnRemoteConfig) load() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, f := range c.features.keys() {
		feature := c.features.get(f)
		if err := feature.load(c.localCachePath, jsonFileReaderWriter{}, jsonValidator{}); err != nil {
			c.reportLoadError(feature.name, err)
			log.Println(internal.ErrorPrefix, "failed loading feature [", feature.name, "] config from the disk:", err)
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
		if match.TargetRollout < c.appRolloutGroup {
			c.analytics.EmitPartialRolloutEvent(ClientCli, featureName, match.TargetRollout, partialRolloutPerformedFailure)
			match = nil
		} else {
			c.analytics.EmitPartialRolloutEvent(ClientCli, featureName, match.TargetRollout, partialRolloutPerformedSuccess)
		}
	} else {
		//when there's no match (eg., due to a version value mismatch) emit the partial rollout event failure
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
