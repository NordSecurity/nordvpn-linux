package tray

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
)

// connectionSettings represents a part of VPN connection configuration parameters
type connectionSettings struct {
	Obfuscated      bool
	Protocol        config.Protocol
	Technology      config.Technology
	VirtualLocation bool
}

// connectionSettingsChangeSensor monitors changes to connection settings
type connectionSettingsChangeSensor struct {
	settings connectionSettings
	mu       sync.RWMutex
	changed  bool
}

// NewconnectionSettingsChangeSensor creates a new connection settings change sensor
// which tracks whether settings has changed since the last update
func newConnectionSettingsChangeSensor() *connectionSettingsChangeSensor {
	return &connectionSettingsChangeSensor{
		settings: connectionSettings{},
		changed:  false,
	}
}

// Set sets connection related settings
func (s *connectionSettingsChangeSensor) Set(settings connectionSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.changed = s.settings != settings
	if s.changed {
		s.settings = settings
	}
}

// Detected returns whether settings have changed since the last check
func (s *connectionSettingsChangeSensor) ChangeDetected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.changed
}
