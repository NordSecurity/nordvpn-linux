package access

import (
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

type DisconnectInput struct {
	Netw          networker.Networker
	ConfigManager config.Manager
	Events        *daemonevents.Events
}

func Disconnect(input DisconnectInput) (bool, error) {
	startTime := time.Now()
	if !input.Netw.IsVPNActive() {
		if err := input.Netw.UnsetFirewall(); err != nil {
			log.Println(internal.WarningPrefix, "failed to force unset firewall on disconnect:", err)
		}
		return false, nil
	}

	var cfg config.Config
	var err error
	defer func(start time.Time) {
		status := events.StatusSuccess
		if err != nil {
			status = events.StatusFailure
		}
		input.Events.Service.Disconnect.Publish(events.DataDisconnect{
			Protocol:             cfg.AutoConnectData.Protocol,
			EventStatus:          status,
			Technology:           cfg.Technology,
			ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
			Duration:             time.Since(start),
			Error:                err,
		})
	}(startTime)

	if err = input.Netw.Stop(); err != nil {
		err = fmt.Errorf("stopping networker: %w", err)
		return true, err
	}
	if err = input.ConfigManager.Load(&cfg); err != nil {
		err = fmt.Errorf("loading config: %w", err)
		return true, err
	}

	return true, nil
}
