/*
Package firewall provides firewall service to the caller
*/
package firewall

import (
	"errors"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
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

type VpnInfo struct{
	TunnelInterface *string
	Allowlist *config.Allowlist
	Killswitch bool
}

func NewVpnInfo(allowlist config.Allowlist, killswitch bool) *VpnInfo{
	// log.Printf("got values: tunint - %v, al - %v, ks - %v", tunnelInterface, allowlist, killswitch)
	return &VpnInfo{
		TunnelInterface: nil, 
		Allowlist: &allowlist, 
		Killswitch: killswitch}
}
func (fw *Firewall) Configure(vpnInfo *VpnInfo, meshMap *mesh.MachineMap) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "configure firewall, enabled:", fw.enabled)

	if !fw.enabled {
		return errors.New("firewall not enabled")
	}

	return fw.impl.Configure(vpnInfo, nil)
}

func (fw *Firewall) Remove() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "remove firewall, older status:", fw.enabled)

	if !fw.enabled {
		return errors.New("firewall not enabled")
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

	return fw.impl.Flush()
}
