package remote

import (
	"errors"
	"fmt"
	"io/fs"
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

const (
	envUseLocalConfig   = "RC_USE_LOCAL_CONFIG"
	defaultFeatureState = true
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
	Load() error
	TryPreload()
}

// remoteConfigFileOps aggregates into a common interface
// a set of file-related operations utilized for the sake of remote-config feature functionality.
type remoteConfigFileOps interface {
	IsValidExistingDir(path string) (bool, error)
	CleanupTmpFiles(targetPath, fileExt string) error
	RenameTmpFiles(targetPath, fileExt string) error
}

type fileOpsDefaultImpl struct{}

// WalkFiles iterate files by given extension and do specified action
func (fileOpsDefaultImpl) WalkFiles(targetPath, fileExt string, actionFunc func(string)) error {
	err := filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf(internal.ErrorPrefix+" accessing %s: %v\n", path, err)
			return nil // continue walking
		}
		// exclude symlinks
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), fileExt) {
			actionFunc(path)
		}
		return nil
	})
	return err
}

// RenameTmpFiles rename files by removing extra extension
func (f fileOpsDefaultImpl) RenameTmpFiles(targetPath, fileExt string) error {
	return f.WalkFiles(targetPath, fileExt, func(path string) {
		newPath := strings.TrimSuffix(path, fileExt)
		if err := os.Rename(path, newPath); err != nil {
			log.Printf(internal.ErrorPrefix+" renaming %s to %s: %s\n", path, newPath, err)
		}
	})
}

// CleanupTmpFiles remove files by specified extension
func (f fileOpsDefaultImpl) CleanupTmpFiles(targetPath, fileExt string) error {
	return f.WalkFiles(targetPath, fileExt, func(path string) {
		if err := os.Remove(path); err != nil {
			log.Printf(internal.ErrorPrefix+" removing %s: %s\n", path, err)
		}
	})
}

// IsValidExistingDir check if is valid existing directory
func (fileOpsDefaultImpl) IsValidExistingDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

type RemoteConfigEvent struct {
	MeshnetFeatureEnabled bool
}

type RemoteConfigNotifier interface {
	RemoteConfigUpdate(RemoteConfigEvent) error
}

// fileStoreOps aggregates file relevant operations under a common interface
type fileStoreOps interface {
	writeFile(name string, content []byte, mode os.FileMode) error
	readFile(name string) ([]byte, error)
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
	mu              sync.RWMutex
	notifier        events.PublishSubcriber[RemoteConfigEvent]
	fileOps         fileStoreOps
	rcFileOps       remoteConfigFileOps
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
		fileOps:         jsonFileReaderWriter{},
		rcFileOps:       &fileOpsDefaultImpl{},
	}
	rc.features.add(FeatureMain)
	rc.features.add(FeatureLibtelio)
	rc.features.add(FeatureMeshnet)
	rc.features.add(FeatureNordWhisper)
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

	// check for network timeout errors
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// check for network operation errors (includes DNS resolution failures)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// check for DNS errors that may contain network errors in their message
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		// check if the DNS error message contains network-related errors
		errMsg := dnsErr.Error()
		if strings.Contains(errMsg, "network is unreachable") ||
			strings.Contains(errMsg, "connection refused") ||
			strings.Contains(errMsg, "no route to host") {
			return true
		}
	}

	// check for specific server errors
	if errors.Is(err, core.ErrServerInternal) ||
		errors.Is(err, core.ErrTooManyRequests) {
		return true
	}

	return false
}

// TryPreload load config from disk
func (c *CdnRemoteConfig) TryPreload() {
	// try preload config from disk - but do not complain if anything wrong
	// as this happens on early run and config on disk maybe does not exist yet
	c.loadSilent()
}

// Load download from remote or load from disk
func (c *CdnRemoteConfig) Load() error {
	var err error

	reloadDone := false
	needReload := false

	useOnlyLocalConfig := internal.IsDevEnv(c.appEnvironment) && os.Getenv(envUseLocalConfig) != "" // forced load from disk?
	if useOnlyLocalConfig {
		log.Printf("%s Ignoring remote config, using only local\n", internal.InfoPrefix)
	} else {
		if needReload, err = c.download(); err != nil {
			return fmt.Errorf("downloading remote config: %w", err)
		}
	}
	// remote config files were downloaded and need to be reloaded?
	if needReload || useOnlyLocalConfig {
		c.load()
		reloadDone = true
	}

	if reloadDone {
		// notify what is current state after config reload
		c.notifier.Publish(RemoteConfigEvent{MeshnetFeatureEnabled: c.IsFeatureEnabled(FeatureMeshnet)})
		// reset event flags for some events to control emit frequency
		c.analytics.ClearEventFlags()
	}

	return nil
}

func (c *CdnRemoteConfig) download() (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newChangesDownloaded := false

	for _, f := range c.features.keys() {
		feature := c.features.get(f)
		dnld, err := feature.download(cdnFileGetter{cdn: c.cdn}, c.fileOps, jsonValidator{}, filepath.Join(c.remotePath, c.appEnvironment), c.localCachePath, c.rcFileOps)
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

func (c *CdnRemoteConfig) loadSilent() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.doLoad(false)
}

func (c *CdnRemoteConfig) load() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.doLoad(true)
}

func (c *CdnRemoteConfig) doLoad(reportErrors bool) {
	for _, f := range c.features.keys() {
		feature := c.features.get(f)
		if err := feature.load(c.localCachePath, c.fileOps, jsonValidator{}, c.rcFileOps); err != nil {
			if reportErrors {
				c.reportLoadError(feature.name, err)
				log.Printf("%s failed loading feature [%s] config from the disk: %s\n", internal.ErrorPrefix, feature.name, err)
			}
			continue
		}
		log.Printf("%s feature [%s] config loaded from: %s\n", internal.InfoPrefix, feature.name, c.localCachePath)
		c.analytics.EmitLocalUseEvent(ClientCli, feature.name, nil)
	}
}

func (c *CdnRemoteConfig) findMatchingRecord(ss []ParamValue, featureName string) (match *ParamValue) {
	for _, s := range ss {
		// find my version matching records
		ok, err := isVersionMatching(c.appVersion, s.AppVersion)
		if err != nil {
			log.Printf("%s invalid version: %s\n", internal.ErrorPrefix, err)
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
		if match.TargetRollout > 0 && match.TargetRollout < c.appRolloutGroup {
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
	return c.GetFeatureParam(FeatureLibtelio, "libtelio_config")
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
