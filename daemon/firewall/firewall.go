/*
Package firewall provides firewall service to the caller
*/
package firewall

import (
	"errors"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	logPrefix = "[fw]"
)

// Firewall is responsible for correctly changing one firewall agent over another.
//
// Thread-safe.
type Firewall struct {
	mu      sync.Mutex
	impl    FirewallImpl
	enabled bool
}

// NewFirewall produces an instance of Firewall.
func NewFirewall(impl FirewallImpl, enabled bool) *Firewall {
	return &Firewall{
		impl:    impl,
		enabled: enabled,
	}
}

func (fw *Firewall) Configure(config Config) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "configure firewall")

	if !fw.enabled {
		log.Println(internal.InfoPrefix, logPrefix, "ignoring configure because firewall is disabled")
		return nil
	}

	return fw.impl.Configure(config)
}

func (fw *Firewall) Enable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "enabling firewall")

	if fw.enabled {
		return errors.New("firewall already enabled")
	}

	fw.enabled = true

	return nil
}

func (fw *Firewall) Disable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "disable firewall, older status:", fw.enabled)

	if !fw.enabled {
		return errors.New("firewall already disabled")
	}

	fw.enabled = false

	return fw.impl.Flush()
}

func (fw *Firewall) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "flush firewall rules")

	if !fw.enabled {
		log.Println(internal.InfoPrefix, logPrefix, "ignoring flush because firewall is disabled")
		return nil
	}

	return fw.impl.Flush()
}
