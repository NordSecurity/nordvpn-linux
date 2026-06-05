package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// HandleVPNConnectionError is the placeholder (for now) listener for generic VPN connection
// error events published on the internal VPN events bus
func (r *RPC) HandleVPNConnectionError(e events.VPNConnectionErrorEvent) error {
	log.Debug("received VPN connection error event, code:", e.Code.String())
	return nil
}
