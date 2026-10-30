package dns

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/log"
)

// Executables
const (
	// execResolvectl defines resolvectl executable
	execResolvectl = "resolvectl"
)

// Systemd-resolved and resolvectl based DNS handling method
type Resolvectl struct{}

func (m *Resolvectl) Set(iface string, nameservers []string) error {
	return setDNSWithResolvectl(iface, nameservers)
}

func (m *Resolvectl) Unset(iface string) error {
	return unsetDNSWithResolvectl(iface)
}

func (m *Resolvectl) Name() string {
	return "resolvectl"
}

// Following the logic from https://systemd.io/RESOLVED-VPNS/
// resolvectl domain privacy0 '~.'
// resolvectl default-route privacy0 true
// resolvectl dns privacy0 <IP>
func setDNSWithResolvectl(iface string, addresses []string) error {
	cmdStr := []string{"dns", iface}
	cmdStr = append(cmdStr, addresses...)
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, cmdStr...).CombinedOutput(); err != nil {
		return fmt.Errorf("setting dns with resolvectl: %s: %w", strings.TrimSpace(string(out)), err)
	}
	// "Catch-all" domain routing for interface
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, "domain", iface, "~.").CombinedOutput(); err != nil {
		log.DNS.Warn("dns domain routing with resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	// empty string makes the link use the global DNSSEC setting
	dnsSECOption := ""
	if areOnlyNordDNSServers(addresses) {
		dnsSECOption = "off"
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, "dnssec", iface, dnsSECOption).CombinedOutput(); err != nil {
		log.DNS.Warn("dnssec with resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, "default-route", iface, "true").CombinedOutput(); err != nil {
		log.DNS.Warn("dns domain default-route with resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, "flush-caches").CombinedOutput(); err != nil {
		log.DNS.Warn("flushing dns caches resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	return nil
}

func unsetDNSWithResolvectl(iface string) error {
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, "revert", iface).CombinedOutput(); err != nil {
		return fmt.Errorf("reverting dns with resolvectl: %s: %w", strings.TrimSpace(string(out)), err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, "flush-caches").CombinedOutput(); err != nil {
		log.DNS.Warn("flushing dns caches resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	return nil
}
