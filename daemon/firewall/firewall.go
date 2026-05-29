/*
Package firewall provides firewall service to the caller
*/
package firewall

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	logPrefix = "[fw]"
)

// Firewall is responsible for correctly changing one firewall agent over another.
//
// Thread-safe.
type Firewall struct {
	mu             sync.Mutex
	impl           FirewallBackend
	enabled        bool
	debuggerEvents events.Publisher[events.DebuggerEvent]
	appEnvironment string
}

// NewFirewall produces an instance of Firewall.
func NewFirewall(
	impl FirewallBackend,
	enabled bool,
	appEnvironment string,
	debuggerEvents events.Publisher[events.DebuggerEvent],
) *Firewall {
	return &Firewall{
		impl:           impl,
		enabled:        enabled,
		appEnvironment: appEnvironment,
		debuggerEvents: debuggerEvents,
	}
}

func (fw *Firewall) Configure(config Config) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.enabled {
		return nil
	}

	log.Println(internal.InfoPrefix, logPrefix, "configuring firewall")
	if internal.IsDevEnv(fw.appEnvironment) {
		log.Println(internal.DebugPrefix, logPrefix, "configure fw from", internal.GetStack())
	}

	err := fw.impl.Configure(config)
	fw.emitConfigureEvent(config, err)
	return err
}

// emitConfigureEvent publishes a firewall configuration analytics event.
func (fw *Firewall) emitConfigureEvent(config Config, err error) {
	if fw.debuggerEvents == nil {
		return
	}
	event := newConfigureEvent(config, err)
	fw.debuggerEvents.Publish(*event.ToDebuggerEvent())
}

func (fw *Firewall) Enable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "enabling firewall")

	if fw.enabled {
		return NewError(ErrFirewallAlreadyEnabled)
	}

	fw.enabled = true

	return nil
}

func (fw *Firewall) Disable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Println(internal.InfoPrefix, logPrefix, "disabling firewall")

	if !fw.enabled {
		return NewError(ErrFirewallAlreadyDisabled)
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

	if internal.IsDevEnv(fw.appEnvironment) {
		log.Println(internal.DebugPrefix, logPrefix, "flush fw", internal.GetStack())
	}

	return fw.impl.Flush()
}
