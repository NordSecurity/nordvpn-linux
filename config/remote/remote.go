package remote

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type RemoteStorage interface {
	GetRemoteFile(name string) ([]byte, error)
}

// RemoteConfigGetter get values from remote config
type RemoteConfigGetter interface {
	GetTelioConfig() (string, error)
	IsFeatureEnabled(featureName string) bool
	GetFeatureParam(featureName, paramName string) //TODO/FIXME: return type?
	LoadConfig() error
}

var (
	envUseLocalConfig = "USE_LOCAL_CONFIG"
)

type CdnRemoteConfig struct {
	appVersion     string
	appEnvironment string
	localCachePath string
	remotePath     string
	cdn            RemoteStorage
	Features       FeatureMap
	mu             sync.Mutex
}

// NewCdnRemoteConfig setup RemoteStorage based remote config loaded/getter
func NewCdnRemoteConfig(ver, env, remotePath, localPath string, cdn RemoteStorage) *CdnRemoteConfig {
	rc := &CdnRemoteConfig{
		appVersion:     ver,
		appEnvironment: env,
		remotePath:     remotePath,
		localCachePath: localPath,
		cdn:            cdn,
		Features:       make(FeatureMap),
	}
	rc.Features.Add(featureMain)
	rc.Features.Add("nordwhisper") // TODO/FIXME: debug/remove
	rc.Features.Add(featureLibtelio)
	return rc
}

// LoadConfig download from remote or load from disk
func (c *CdnRemoteConfig) LoadConfig() error {
	useOnlyLocalConfig := c.appEnvironment == "dev" && os.Getenv(envUseLocalConfig) != "" // forced load from disk?
	if !useOnlyLocalConfig {
		log.Println(internal.DebugPrefix, "Downloading remote config to:", c.localCachePath)
		for _, f := range c.Features {
			if err := f.download(c.cdn, filepath.Join(c.remotePath, c.appEnvironment), c.localCachePath); err != nil {
				log.Println(internal.ErrorPrefix, "failed downloading config for [", f.Name, "]:", err)
				continue
			}
			log.Println(internal.DebugPrefix, "Feature [", f.Name, "] config downloaded.")
		}
	}

	// local only when loading from local files
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println(internal.DebugPrefix, "Loading config from:", c.localCachePath)

	for _, f := range c.Features {
		if err := f.load(c.localCachePath); err != nil {
			log.Println(internal.ErrorPrefix, "failed loading config from disk for [", f.Name, "]:", err)
			continue
		}
		log.Println(internal.DebugPrefix, "Feature [", f.Name, "] config loaded.")
	}

	return nil
}

func findMatchingRecord(ss []ParamValue, ver string) *ParamValue {
	matches := []ParamValue{}
	for _, s := range ss {
		// find all my version matching records
		ok, err := isVersionMatching(ver, s.AppVersion)
		if err != nil {
			//TODO/FIXME: abort? or ignore?
			log.Println("invalid version:", err)
			continue
		}
		if ok {
			matches = append(matches, s)
		}
	}
	if len(matches) > 0 {
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Weight > matches[j].Weight
		})
		return &matches[0]
	}
	return nil
}

// TODO/FIXME: add `rollout` support
func (c *CdnRemoteConfig) GetTelioConfig() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// have app version; find the right telio config;
	// if multiple matching records found, sort by weight and use highest weight
	f, found := c.Features[featureLibtelio]
	if !found {
		return "", fmt.Errorf("libtelio feature config not found")
	}
	p, found := f.Params[featureLibtelio]
	if !found {
		return "", fmt.Errorf("libtelio config not found")
	}
	if item := findMatchingRecord(p.Settings, c.appVersion); item != nil {
		return item.IncValue, nil
	}
	return "", fmt.Errorf("telio config for current app version not found")
}

func (c *CdnRemoteConfig) IsFeatureEnabled(featureName string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// find by name, expect param name to be the same as feature name and expect boolean type
	f, found := c.Features[featureName]
	if !found {
		return false
	}
	p, found := f.Params[featureName]
	if !found {
		return false
	}
	switch p.Type {
	case "bool", "boolean":
		if item := findMatchingRecord(p.Settings, c.appVersion); item != nil {
			val, _ := item.AsBool()
			return val
		}
	}
	return false
}

// TODO/FIXME: return type?
func (c *CdnRemoteConfig) GetFeatureParam(featureName, paramName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, found := c.Features[featureName]
	if !found {
		return
	}
	p, found := f.Params[paramName]
	if !found {
		return
	}
	switch p.Type {
	case "bool", "boolean":
		if item := findMatchingRecord(p.Settings, c.appVersion); item != nil {
			item.AsBool()
			return
		}
	}
}
