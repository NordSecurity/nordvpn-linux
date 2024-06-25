package libtelio

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	// telioLocalConfigName environment variable name
	telioLocalConfigName = "TELIO_LOCAL_CFG"
)

// TelioConfigFetcher is abstract interface to cover local and remote config
type TelioConfigFetcher interface {
	IsAvailable() bool
	Fetch(appVer string) (string, error)
}

type TelioConfig struct {
	fetchers []TelioConfigFetcher
}

func NewTelioConfig(rc remote.RemoteConfigGetter) *TelioConfig {
	return &TelioConfig{
		fetchers: []TelioConfigFetcher{
			&TelioLocalConfigFetcher{},
			&TelioRemoteConfigFetcher{rc},
		},
	}
}

func (tc *TelioConfig) GetConfig(version string) (string, error) {
	for _, c := range tc.fetchers {
		if c.IsAvailable() {
			return c.Fetch(version)
		}
	}
	return "", fmt.Errorf("telio config is not available")
}

type TelioLocalConfigFetcher struct{}

func (c *TelioLocalConfigFetcher) IsAvailable() bool {
	val, ok := os.LookupEnv(telioLocalConfigName)
	return ok && strings.TrimSpace(val) != ""
}

func (c *TelioLocalConfigFetcher) Fetch(string) (string, error) {
	val, ok := os.LookupEnv(telioLocalConfigName)
	val = strings.TrimSpace(val)
	if !ok || val == "" {
		return "", fmt.Errorf("telio local config is not available")
	}
	if strings.HasPrefix(val, "/") {
		// libtelio config is given as a path to json file
		log.Println(internal.InfoPrefix, "Fetch libtelio local config from file:", val)
		cfg, err := internal.FileRead(val)
		if err != nil {
			return "", fmt.Errorf("telio local config failed to read from file, err: %w", err)
		}
		return string(cfg), nil
	}
	log.Println(internal.InfoPrefix, "Fetch libtelio local config")
	return val, nil
}
