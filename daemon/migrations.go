package daemon

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
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
