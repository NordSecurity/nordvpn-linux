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
		log.FW.Trace("fw configuration skipped, firewall is disabled")
		return nil
	}

	log.FW.Tracef("killSwitch=%v meshnet=%v tunnel=%q",
		config.KillSwitch, config.MeshnetInfo != nil, config.TunnelInterface)
	log.FW.Info("configuring firewall")
	if internal.IsDevEnv(fw.appEnvironment) {
		log.FW.Debug("configure fw from", internal.GetStack())
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

	log.FW.Info("enabling firewall")
	log.FW.Tracef("currentEnabled=%v", fw.enabled)

	if fw.enabled {
		return NewError(ErrFirewallAlreadyEnabled)
	}

	fw.enabled = true

	return nil
}

func (fw *Firewall) Disable() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.FW.Info("disabling firewall")
	log.FW.Tracef("currentEnabled=%v", fw.enabled)

	if !fw.enabled {
		return NewError(ErrFirewallAlreadyDisabled)
	}

	fw.enabled = false

	return fw.impl.Flush()
}

func (fw *Firewall) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.FW.Info("flush firewall rules")

	if !fw.enabled {
		log.FW.Info("ignoring flush because firewall is disabled")
		return nil
	}

	if internal.IsDevEnv(fw.appEnvironment) {
		log.FW.Debug("flush fw", internal.GetStack())
	}

	return fw.impl.Flush()
}
