package firewallmock

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
)

// --- Firewall mock (firewall.Service) ---

type Firewall struct {
	enabled bool
	config  firewall.Config
	Err     error
}

func NewFirewall() *Firewall {
	return &Firewall{}
}

// Enable firewall
func (mf *Firewall) Enable() error {
	if mf.Err != nil {
		return mf.Err
	}
	mf.enabled = true
	return nil
}

// Disable firewall
func (mf *Firewall) Disable() error {
	if mf.Err != nil {
		return mf.Err
	}
	mf.enabled = false
	return nil
}

// Flush firewall
func (mf *Firewall) Flush() error {
	if mf.Err != nil {
		return mf.Err
	}
	mf.config = firewall.Config{}
	return nil
}

// Configure firewall
func (mf *Firewall) Configure(config firewall.Config) error {
	if mf.Err != nil {
		return mf.Err
	}
	mf.config = config
	return nil
}

// IsEnabled returns the current enable status
func (mf *Firewall) IsEnabled() bool {
	return mf.enabled
}

// Config returns the currently stored config
func (mf *Firewall) Config() firewall.Config {
	return mf.config
}

// --- end Firewall mock ---

// --- FirewallBackend mock (firewall.FirewallBackend) ---

type FirewallBackend struct {
	config firewall.Config
	Err    error
}

func (m *FirewallBackend) Configure(config firewall.Config) error {
	if m.Err != nil {
		return m.Err
	}
	m.config = config
	return nil
}

func (m *FirewallBackend) Flush() error {
	if m.Err != nil {
		return m.Err
	}
	m.config = firewall.Config{}
	return nil
}

// Config returns the currently stored config
func (m *FirewallBackend) Config() firewall.Config {
	return m.config
}

// --- end FirewallBackend mock ---
