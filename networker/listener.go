package networker

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Reconnect refresh connectivity on network state change
func (c *Combined) Reconnect(stateIsUp bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isVpnSet {
		// disable IPv6 as soon as possible to prevent leaks when a new Ethernet is inserted
		if err := c.disableIPv6IfNeeded(); err != nil {
			log.Println(internal.ErrorPrefix, "refreshing network", err)
		}
	}

	if stateIsUp {
		if err := c.refreshVPN(); err != nil {
			log.Println(internal.ErrorPrefix, "refreshing vpn", err)
		}
	}
}
