package libtelio

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	// telioLocalConfigName environment variable name
	telioLocalConfigName = "TELIO_LOCAL_CFG"
)

type FetchFn func() (string, error)

// TelioConfigFetcher is abstract interface to cover local and remote config
type TelioConfigFetcher interface {
	IsAvailable() bool
	Fetch() (string, error)
}

type TelioConfig struct {
	fetchers []TelioConfigFetcher
}

func NewTelioConfig(rcFn FetchFn) *TelioConfig {
	return &TelioConfig{
		fetchers: []TelioConfigFetcher{
			&TelioLocalConfigFetcher{},
			&TelioRemoteConfigFetcher{rcFn},
		},
	}
}

func (tc *TelioConfig) GetConfig() (string, error) {
	for _, c := range tc.fetchers {
		if c.IsAvailable() {
			return c.Fetch()
		}
	}
	return "", fmt.Errorf("telio config is not available")
}

type TelioLocalConfigFetcher struct{}

func (c *TelioLocalConfigFetcher) IsAvailable() bool {
	val, ok := os.LookupEnv(telioLocalConfigName)
	return ok && strings.TrimSpace(val) != ""
}

func (c *TelioLocalConfigFetcher) Fetch() (string, error) {
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

type TelioRemoteConfigFetcher struct {
	rcFn FetchFn
}

func (c *TelioRemoteConfigFetcher) IsAvailable() bool {
	return true
}

func (c *TelioRemoteConfigFetcher) Fetch() (string, error) {
	return c.rcFn()
}
