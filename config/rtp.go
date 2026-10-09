package config

import (
	"encoding/json"

	"github.com/NordSecurity/nordvpn-linux/log"
)

func migrateTPL(c *Config, data []byte) {
	var legacy struct {
		AutoConnectData struct {
			ThreatProtectionLite *bool `json:"cybersec,omitempty"`
		} `json:"auto_connect_data"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		log.Warn("failed to unmarshal legacy threat protection lite value: ", err)
		return
	}

	if legacy.AutoConnectData.ThreatProtectionLite == nil {
		return
	}

	log.Debug("migrating threat protection lite to real time protection:",
		*legacy.AutoConnectData.ThreatProtectionLite)
	c.AutoConnectData.RealTimeProtection = *legacy.AutoConnectData.ThreatProtectionLite
}
