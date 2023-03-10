package networker

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Reconnect refresh connectivity on network state change
func (c *Combined) Reconnect(stateIsUp bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	isStarted := c.isVpnSet
	ipv6Enabled := c.ipv6Enabled

	if isStarted {
		if !ipv6Enabled {
			if err := c.denyIPv6(); err != nil {
				log.Println(internal.ErrorPrefix, "refreshing network", err)
			}
		}
	}

	if stateIsUp {
		if err := c.refreshVPN(); err != nil {
			log.Println(internal.ErrorPrefix, "refreshing vpn", err)
		}
	}
}
