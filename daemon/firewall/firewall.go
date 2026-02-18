/*
Package firewall provides firewall service to the caller
*/
package firewall

import (
	"errors"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
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
	impl    FwImpl
	enabled bool
}

// NewFirewall produces an instance of Firewall.
func NewFirewall(impl FwImpl, enabled bool) *Firewall {
	return &Firewall{
		impl:    impl,
		enabled: enabled,
	}
}

func (fw *Firewall) Configure(vpnInfo *VpnInfo, meshnetMap *mesh.MachineMap) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "configure firewall, enabled:", fw.enabled)

	if !fw.enabled {
		log.Println("firewall not enabled")
		return nil
	}

	return fw.impl.Configure(vpnInfo, meshnetMap)
}

func (fw *Firewall) Remove() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "remove firewall, older status:", fw.enabled)

	if !fw.enabled {
		log.Println("firewall not enabled")
		return nil
	}

	return fw.impl.Flush()

}

func (fw *Firewall) Enable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "enabling firewall:", fw.enabled)

	if fw.enabled {
		return errors.New("already enabled")
	}

	fw.enabled = true

	return nil
}

func (fw *Firewall) Disable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "disable firewall, older status:", fw.enabled)

	if !fw.enabled {
		return errors.New("already disabled")
	}

	fw.enabled = false

	return fw.impl.Flush()
}

func (fw *Firewall) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.enabled {
		log.Println("firewall not enabled")
		return nil
	}

	return fw.impl.Flush()
}
