package libtelio

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
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

func NewTelioConfig(cm config.Manager) *TelioConfig {
	return &TelioConfig{
		fetchers: []TelioConfigFetcher{
			&TelioLocalConfigFetcher{},
			&TelioRemoteConfigFetcher{cm: cm},
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
	log.Println(internal.InfoPrefix, "Fetch libtelio local config")
	return val, nil
}
