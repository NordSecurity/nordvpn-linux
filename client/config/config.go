// Package config provides an interface for managing persistent application configuration.
package config

import (
	"github.com/NordSecurity/nordvpn-linux/config"
)

// Config defines NordVPN application configuration
type Config struct {
	Technology config.Technology `json:"technology,omitempty"`
	Obfuscate  bool              `json:"obfuscate,omitempty"`
	Allowlist  Allowlist         `json:"whitelist,omitempty"`
}

// NewConfig creates new config object
func NewConfig() Config {
	c := Config{}
	c.setDefaultsIfEmpty()
	return c
}
