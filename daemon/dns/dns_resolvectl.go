package dns

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
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

func (m *Resolvectl) IsAvailable() bool {
	return internal.IsCommandAvailable(execResolvectl)
}

func (m *Resolvectl) Name() string {
	return "resolvectl"
}

func setDNSWithResolvectl(iface string, addresses []string) error {
	cmdStr := []string{"dns", iface}
	cmdStr = append(cmdStr, addresses...)
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, cmdStr...).CombinedOutput(); err != nil {
		return fmt.Errorf("setting dns with resolvectl: %s: %w", strings.TrimSpace(string(out)), err)
	}
	// "Catch-all" domain routing for interface, more here: https://github.com/poettering/systemd/commit/8cedb0aef94da880e61b4c8cfeb7f450f8760ec6
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command("resolvectl", "domain", iface, "~.").CombinedOutput(); err != nil {
		log.Println("dns domain routing with resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command("resolvectl", "default-route", iface, "true").CombinedOutput(); err != nil {
		log.Println("dns domain default-route with resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command("resolvectl", "flush-caches").CombinedOutput(); err != nil {
		log.Println("flushing dns caches resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	return nil
}

func unsetDNSWithResolvectl(iface string) error {
	// Just set empty/no DNS server for interface
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command(execResolvectl, "dns", iface, "").CombinedOutput(); err != nil {
		return fmt.Errorf("unsetting dns with resolvectl: %s: %w", strings.TrimSpace(string(out)), err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command("resolvectl", "domain", iface, "").CombinedOutput(); err != nil {
		log.Println("dns domain routing with resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command("resolvectl", "default-route", iface, "false").CombinedOutput(); err != nil {
		log.Println("dns domain default-route with resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	// #nosec G204 -- input is properly validated
	if out, err := exec.Command("resolvectl", "flush-caches").CombinedOutput(); err != nil {
		log.Println("flushing dns caches resolvectl:", strings.TrimSpace(string(out)), "err:", err)
	}
	return nil
}
