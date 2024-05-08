package logger

import (
	"fmt"
	"log"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
)

type DaemonSettingsSubscriber struct{}

func NewSubscriber() *DaemonSettingsSubscriber {
	return &DaemonSettingsSubscriber{}
}

func (l *DaemonSettingsSubscriber) NotifyTechnology(data config.Technology) error {
	printSettingsChange("Technology", data.String())
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyMeshnet(data bool) error {
	printSettingsChange("Meshnet", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyProtocol(data config.Protocol) error {
	printSettingsChange("Protocol", data.String())
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyFirewall(data bool) error {
	printSettingsChange("Firewall", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyRouting(data bool) error {
	printSettingsChange("Routing", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyKillswitch(data bool) error {
	printSettingsChange("Kill Switch", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyThreatProtectionLite(data bool) error {
	printSettingsChange("ThreatProtectionLite", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyObfuscate(data bool) error {
	printSettingsChange("Obfuscate", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyNotify(data bool) error {
	printSettingsChange("Notify", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyAutoconnect(data bool) error {
	printSettingsChange("Auto-connect", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyDNS(data events.DataDNS) error {
	printSettingsChange("DNS", strings.Join(data.Ips, " "))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyAllowlist(data events.DataAllowlist) error {
	printSettingsChange("Allowlist", fmt.Sprintf("%+v", data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyIpv6(data bool) error {
	printSettingsChange("IPv6", boolToString(data))
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyDefaults(any) error {
	log.Printf("Settings have been restored to their default values")
	return nil
}

func (l *DaemonSettingsSubscriber) NotifyLANDiscovery(data bool) error {
	printSettingsChange("LAN Discovery", boolToString(data))
	return nil
}

func printSettingsChange(settingName string, val string) {
	log.Printf("%s set to: %s", settingName, val)
}

func boolToString(val bool) string {
	if val {
		return "enabled"
	}
	return "disabled"
}
