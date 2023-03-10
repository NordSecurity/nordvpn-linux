package logger

import (
	"log"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
)

type DaemonSettingsSubscriber struct {
	enabled bool
	cm      config.Manager
}

func NewSubscriber(enabled bool, cm config.Manager) *DaemonSettingsSubscriber {
	return &DaemonSettingsSubscriber{
		enabled: enabled,
		cm:      cm,
	}
}

func (l *DaemonSettingsSubscriber) NotifyTechnology(data config.Technology) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Technology", data.String(), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyMeshnet(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Meshnet", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyProtocol(data config.Protocol) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Protocol", data.String(), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyFirewall(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Firewall", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyRouting(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Routing", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyKillswitch(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Kill Switch", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyThreatProtectionLite(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("ThreatProtectionLite", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyObfuscate(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Obfuscate", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyNotify(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Notify", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyAutoconnect(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("Auto-connect", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyDNS(data events.DataDNS) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("DNS", boolToString(data.Enabled), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyWhitelist(data events.DataWhitelist) error {
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyIpv6(data bool) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	printSettingsChange("IPv6", boolToString(data), cfg)
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyDefaults(any) error {
	var cfg config.Config
	err := l.cm.Load(&cfg)
	if err != nil {
		return err
	}
	log.Printf("Settings have been restored to their default value\n%s", readSettings(cfg))
	return nil
}

func readSettings(cfg config.Config) string {
	var settings strings.Builder
	settings.WriteString("NordVPN App Connection Settings:\n")
	settings.WriteString("Technology: " + cfg.Technology.String() + "\n")
	settings.WriteString("Meshnet: " + boolToString(cfg.Mesh) + "\n")
	settings.WriteString("Protocol: " + cfg.AutoConnectData.Protocol.String() + "\n")
	settings.WriteString("Firewall: " + boolToString(cfg.Firewall) + "\n")
	settings.WriteString("KillSwitch: " + boolToString(cfg.KillSwitch) + "\n")
	settings.WriteString("Obfuscate: " + boolToString(cfg.AutoConnectData.Obfuscate) + "\n")
	settings.WriteString("ThreatProtectionLite: " + boolToString(cfg.AutoConnectData.ThreatProtectionLite) + "\n")
	settings.WriteString("DNS: " + strings.Join(cfg.AutoConnectData.DNS, " ") + "\n")
	settings.WriteString("IPv6: " + boolToString(cfg.IPv6) + "\n")
	if cfg.UsersData.Notify != nil && len(cfg.UsersData.Notify) > 0 {
		settings.WriteString("Notify: enabled\n")
	} else {
		settings.WriteString("Notify: disabled\n")
	}
	settings.WriteString("Auto-connect: " + boolToString(cfg.AutoConnect) + "\n")
	settings.WriteString("\n")

	return settings.String()
}

func printSettingsChange(settingName string, val string, c config.Config) {
	log.Printf("%s set to: %s\n%s", settingName, val, readSettings(c))
}

func boolToString(val bool) string {
	if val {
		return "enabled"
	}
	return "disabled"
}
