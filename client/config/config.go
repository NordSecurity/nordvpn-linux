// Package config provides an interface for managing persistent application configuration.
package config

import (
	"github.com/NordSecurity/nordvpn-linux/config"
)

// Config defines NordVPN application configuration
type Config struct {
	Technology config.Technology `json:"technology,omitempty"`
	Protocol   config.Protocol   `json:"protocol,omitempty"`
	// TODO: rename json key when v4 comes out.
	ThreatProtectionLite bool      `json:"cybersec,omitempty"`
	Obfuscate            bool      `json:"obfuscate,omitempty"`
	DNS                  []string  `json:"dns,omitempty"`
	Whitelist            Whitelist `json:"whitelist,omitempty"`
}

// NewConfig creates new config object
func NewConfig() Config {
	c := Config{}
	c.setDefaultsIfEmpty()
	return c
}
