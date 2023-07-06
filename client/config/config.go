// Package config provides an interface for managing persistent application configuration.
package config

import (
	"github.com/NordSecurity/nordvpn-linux/config"
)

// Config defines NordVPN application configuration
type Config struct {
	Technology config.Technology `json:"technology,omitempty"`
	Protocol   config.Protocol   `json:"protocol,omitempty"`
	Obfuscate  bool              `json:"obfuscate,omitempty"`
	Whitelist  Whitelist         `json:"whitelist,omitempty"`
}

// NewConfig creates new config object
func NewConfig() Config {
	c := Config{}
	c.setDefaultsIfEmpty()
	return c
}
