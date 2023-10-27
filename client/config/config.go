// Package config provides an interface for managing persistent application configuration.
package config

// Config defines NordVPN application configuration
// TODO: fully remove this
type Config struct {
	Obfuscate bool `json:"obfuscate,omitempty"`
}

// NewConfig creates new config object
func NewConfig() Config {
	return Config{}
}
