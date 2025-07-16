package remote

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type RemoteStorage interface {
	GetRemoteFile(name string) ([]byte, error)
}

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
	envUseLocalConfig = "USE_LOCAL_CONFIG"
)

type CdnRemoteConfig struct {
	appVersion     string
	appEnvironment string
	localCachePath string
	remotePath     string
	cdn            RemoteStorage
	features       FeatureMap
	mu             sync.Mutex
}

// NewCdnRemoteConfig setup RemoteStorage based remote config loaded/getter
func NewCdnRemoteConfig(buildTarget config.BuildTarget, remotePath, localPath string, cdn RemoteStorage) *CdnRemoteConfig {
	rc := &CdnRemoteConfig{
		appVersion:     buildTarget.Version,
		appEnvironment: buildTarget.Environment,
		remotePath:     remotePath,
		localCachePath: localPath,
		cdn:            cdn,
		features:       make(FeatureMap),
	}
	rc.features.Add(FeatureMain)
	rc.features.Add(FeatureLibtelio)
	rc.features.Add(FeatureMeshnet)
	return rc
}

type jsonFileReaderWriter struct{}

func (w jsonFileReaderWriter) writeFile(name string, content []byte, mode os.FileMode) error {
	return internal.FileWrite(name, content, mode)
}
func (w jsonFileReaderWriter) readFile(name string) ([]byte, error) {
	return internal.FileRead(name)
}

type jsonValidator struct{}

func (v jsonValidator) validate(content []byte) error {
	return validateJsonString(content)
}

// LoadConfig download from remote or load from disk
func (c *CdnRemoteConfig) LoadConfig() error {
	useOnlyLocalConfig := internal.IsDevEnv(c.appEnvironment) && os.Getenv(envUseLocalConfig) != "" // forced load from disk?
	if !useOnlyLocalConfig {
		for _, f := range c.features {
			dnld, err := f.download(c.cdn, jsonFileReaderWriter{}, jsonValidator{}, filepath.Join(c.remotePath, c.appEnvironment), c.localCachePath)
			if err != nil {
				log.Println(internal.ErrorPrefix, "failed downloading feature [", f.name, "] remote config:", err)
				continue
			}
			if dnld {
				// only if remote config was really downloaded
				log.Println(internal.InfoPrefix, "Feature [", f.name, "] remote config downloaded to:", c.localCachePath)
			}
		}
	}

	// lock only when loading from local files
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, f := range c.features {
		if err := f.load(c.localCachePath, jsonFileReaderWriter{}, jsonValidator{}); err != nil {
			log.Println(internal.ErrorPrefix, "failed loading feature [", f.name, "] config from the disk:", err)
			continue
		}
		log.Println(internal.InfoPrefix, "Feature [", f.name, "] config loaded from:", c.localCachePath)
	}

	return nil
}

func findMatchingRecord(ss []ParamValue, ver string) *ParamValue {
	matches := []ParamValue{}
	for _, s := range ss {
		// find all my version matching records
		ok, err := isVersionMatching(ver, s.AppVersion)
		if err != nil {
			log.Println(internal.ErrorPrefix, "invalid version:", err)
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

func (c *CdnRemoteConfig) GetTelioConfig() (string, error) {
	return c.GetFeatureParam(FeatureLibtelio, FeatureLibtelio)
}

// TODO/FIXME: add `rollout` support
func (c *CdnRemoteConfig) IsFeatureEnabled(featureName string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// find by name, expect param name to be the same as feature name and expect boolean type
	f, found := c.features[featureName]
	if !found {
		return false
	}
	p, found := f.params[featureName]
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

// TODO/FIXME: add `rollout` support
func (c *CdnRemoteConfig) GetFeatureParam(featureName, paramName string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, found := c.features[featureName]
	if !found {
		return "", fmt.Errorf("feature [%s] not found", featureName)
	}
	p, found := f.params[paramName]
	if !found {
		return "", fmt.Errorf("feature [%s] param [%s] not found", featureName, paramName)
	}
	if item := findMatchingRecord(p.Settings, c.appVersion); item != nil {
		switch p.Type {
		case "bool", "boolean":
			val, err := item.AsBool()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%t", val), nil
		case "string", "object":
			val, err := item.AsString()
			if err != nil {
				return "", err
			}
			return val, nil
		case "integer", "int", "number":
			val, err := item.AsInt()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", val), nil
		case "array":
			val, err := item.AsStringArray()
			if err != nil {
				return "", err
			}
			return strings.Join(val, ", "), nil
		case "file":
			return item.incValue, nil
		}
	}
	return "", fmt.Errorf("feature [%s] param [%s] value not found", featureName, paramName)
}
