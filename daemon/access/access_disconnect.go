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
// whether the user was connected or not before this call.
func Disconnect(input DisconnectInput) (bool, error) {
	startTime := time.Now()
	if !input.Networker.IsVPNActive() {
		if err := input.Networker.UnsetFirewall(); err != nil {
			log.Println(internal.WarningPrefix, "failed to force unset firewall on disconnect:", err)
		}
		return false, nil
	}

	var err error
	defer func() {
		var cfg config.Config
		if err := input.ConfigManager.Load(&cfg); err != nil {
			log.Printf("%s loading config during disconnect: %v", internal.WarningPrefix, err)
			return
		}

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

	if err = input.Networker.Stop(); err != nil {
		err = fmt.Errorf("stopping networker: %w", err)
		return true, err
	}

	return true, nil
}
