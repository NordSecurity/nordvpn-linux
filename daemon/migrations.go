package daemon

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// MigrateDeprecatedRegionalAutoconnect removes the deprecated regional group from autoconnect.
// If it was the only target, ServerTag is cleared so autoconnect falls back to quick connect.
// The migration is idempotent.
func MigrateDeprecatedRegionalAutoconnect(cm config.Manager) error {
	var cfg config.Config
	if err := cm.Load(&cfg); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if !config.IsRegionalGroup(cfg.AutoConnectData.Group) {
		return nil
	}
	return cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Group = config.ServerGroup_UNDEFINED
		if c.AutoConnectData.Country == "" && c.AutoConnectData.City == "" {
			c.AutoConnectData.ServerTag = ""
		}
		return c
	})
}

// ConfigCleanup - validate/cleanup DNS addresses, allowlist subnets
func ConfigCleanup(c config.Config) config.Config {
	// Remove all nameservers with IPv6 addresses
	var dnsList []string
	for _, addr := range c.AutoConnectData.DNS {
		if internal.IsAddressValidAsDNSServer(addr) {
			dnsList = append(dnsList, addr)
		} else {
			log.Warn("remove invalid DNS address from the list", addr)
		}
	}
	c.AutoConnectData.DNS = dnsList

	// Remove overlapping, invalid and IPv6 subnets, if any
	c.AutoConnectData.Allowlist.NormalizeSubnets(func(removed, reason string) {
		log.Warn("On start, allowlist remove subnet:", removed, "; reason:", reason)
	})

	return c
}
