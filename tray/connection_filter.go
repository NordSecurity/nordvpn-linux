package tray

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
)

// connectionSettingsChangeSensor monitors changes to the VPN protocol
type connectionSettingsChangeSensor struct {
	vpnProtocol config.VPNProtocol
	mu          sync.RWMutex
	changed     bool
}

// NewconnectionSettingsChangeSensor creates a new connection settings change sensor
// which tracks whether settings has changed since the last update
func newConnectionSettingsChangeSensor() *connectionSettingsChangeSensor {
	return &connectionSettingsChangeSensor{}
}

// Set sets the VPN protocol
func (s *connectionSettingsChangeSensor) Set(vpnProtocol config.VPNProtocol) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.changed = s.vpnProtocol != vpnProtocol
	if s.changed {
		s.vpnProtocol = vpnProtocol
	}
}

// Detected returns whether settings have changed since the last check
func (s *connectionSettingsChangeSensor) ChangeDetected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.changed
}
