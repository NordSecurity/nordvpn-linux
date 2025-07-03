package networker

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Reconnect refresh connectivity on network state change
func (c *Combined) Reconnect(stateIsUp bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Disable IPv6 as soon as possible to prevent leaks when a new adapter is inserted
	if c.isVpnSet {
		c.blockIPv6()
	}

	if stateIsUp {
		if err := c.refreshVPN(context.Background()); err != nil {
			log.Println(internal.ErrorPrefix, "refreshing vpn", err)
		}
	}
}
