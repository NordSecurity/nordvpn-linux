package access

import (
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

type DisconnectInput struct {
	Networker                  networker.Networker
	ConfigManager              config.Manager
	PublishDisconnectEventFunc func(events.DataDisconnect)
}

// Disconnect disconnects the user from the current VPN server. Returning boolean indicates
// whether the user was connection or not before this call.
func Disconnect(input DisconnectInput) (bool, error) {
	startTime := time.Now()
	if !input.Networker.IsVPNActive() {
		if err := input.Networker.UnsetFirewall(); err != nil {
			log.Println(internal.WarningPrefix, "failed to force unset firewall on disconnect:", err)
		}
		return false, nil
	}

	var cfg config.Config
	var err error
	if err = input.Networker.Stop(); err != nil {
		err = fmt.Errorf("stopping networker: %w", err)
		return true, err
	}
	if err = input.ConfigManager.Load(&cfg); err != nil {
		err = fmt.Errorf("loading config: %w", err)
		return true, err
	}

	defer func() {
		status := events.StatusSuccess
		if err != nil {
			status = events.StatusFailure
		}
		input.PublishDisconnectEventFunc(events.DataDisconnect{
			Protocol:             cfg.AutoConnectData.Protocol,
			EventStatus:          status,
			Technology:           cfg.Technology,
			ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
			Duration:             time.Since(startTime),
			Error:                err,
		})
	}()

	return true, nil
}
