/*
Package dns is responsible for configuring dns on various Linux distros.
*/
package dns

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Daemons and Services
const (
	// serviceSystemdResolved defines resolved service managed by systemd
	serviceSystemdResolved = "systemd-resolved"
)

// Executables
const (
	// execBusctl defines busctl executable
	execBusctl = "busctl"

	// execResolvconf defines resolvconf executable
	execResolvconf = "resolvconf"
)

// Files
const (
	// resolvconfFilePath defines path to resolv.conf file for DNS
	resolvconfFilePath = "/etc/resolv.conf"

	// resolvconfBackupPath defines where resolv.conf backup file is stored
	resolvconfBackupPath = internal.BakFilesPath + "resolv.conf"

	// resolvconfInterfaceFilePath defines path of the file used to control
	// the order in which resolvconf nameserver information records are processed.
	resolconfInterfaceFilePath = "/etc/resolvconf/interface-order"
)

// Setter is responsible for configuring DNS.
type Setter interface {
	Set(iface string, nameservers []string) error
	Unset(iface string) error
}

/*
DefaultSetter handleds DNS in this order:

1. If systemd-resolve command is available and systemd-resolved.service is
runnning, systemd-resolve is used. It replaces /etc/resolv.conf file with
a generated one.

2. In absence of systemd-resolve, resolvconf command line utility is used, which
modifies /etc/resolv.conf by adding or removing lines.

3. In case the resolvconf command line utility fails, /etc/resolv.conf is
backed up and modified directly by NordVPN.
*/
type DefaultSetter struct {
	publisher events.Publisher[string]
}

func NewSetter(publisher events.Publisher[string]) *DefaultSetter {
	return &DefaultSetter{
		publisher: publisher,
	}
}

// Set DNS for a given iface if the system supports per interface DNS settings.
// Also, backup current DNS settings. Backup is not overriden, so its safe to
// call this function multiple times in a row.
func (d *DefaultSetter) Set(iface string, nameservers []string) error {
	d.publisher.Publish(
		"setting dns to " + strings.Join(nameservers, " "),
	)
	if err := internal.FileUnlock(resolvconfFilePath); err != nil {
		log.Println(internal.WarningPrefix, err)
	}
	defer internal.FileLock(resolvconfFilePath)

	if len(nameservers) == 0 {
		return errors.New("nameservers not provided")
	}

	if internal.IsServiceActive(serviceSystemdResolved) {
		d.publisher.Publish("using systemd-resolved")
		err := setDNSWithSystemdResolve(iface, nameservers)
		if err != nil {
			return fmt.Errorf("setting dns with systemd resolved: %w", err)
		}
		return nil
	}

	if internal.IsCommandAvailable(execResolvconf) {
		d.publisher.Publish("using resolvconf")
		err := setDNSWithResolvconf(iface, nameservers)
		if err != nil {
			log.Println(internal.WarningPrefix, err)
			return setDNSDefault(nameservers)
		}
	}
	return setDNSDefault(nameservers)
}

// Unset DNS from a backup and remove the backup on success.
func (d *DefaultSetter) Unset(iface string) error {
	d.publisher.Publish("unsetting DNS")
	if err := internal.FileUnlock(resolvconfFilePath); err != nil {
		log.Println(internal.WarningPrefix, err)
	}
	if internal.IsServiceActive(serviceSystemdResolved) {
		d.publisher.Publish("using systemd-resolved")
		err := unsetDNSWithSystemdResolve(iface)
		if err != nil {
			return fmt.Errorf("unsetting dns with systemd resolved: %w", err)
		}
		return nil
	}

	if internal.FileExists(resolvconfBackupPath) {
		err := unsetDNSDefault()
		if err != nil {
			return fmt.Errorf("unsetting dns from backup: %w", err)
		}
		return nil
	}
	if internal.IsCommandAvailable(execResolvconf) {
		d.publisher.Publish("using resolvconf")
		err := unsetDNSWithResolvconf(iface)
		if err != nil {
			return fmt.Errorf("unsetting dns with resolvconf: %w", err)
		}
		return nil
	}
	return nil
}

// setDNSWithSystemdResolve uses systemd-resolve dbus API to manage DNS
// https://www.freedesktop.org/wiki/Software/systemd/resolved/
func setDNSWithSystemdResolve(ifname string, addresses []string) error {
	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		return err
	}
	// Set dns
	args := []string{
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"SetLinkDNS", "ia(iay)", fmt.Sprintf("%d", iface.Index), fmt.Sprintf("%d", len(addresses)),
	}
	// prepare addresses for busctl
	for _, address := range addresses {
		ip := net.ParseIP(address)
		if ip4 := ip.To4(); ip4 != nil {
			ip = ip4
			args = append(args, "2", "4")
		} else {
			args = append(args, "10", "16")
		}
		for _, octet := range ip {
			args = append(args, fmt.Sprintf("%d", octet))
		}
	}
	// #nosec G204 -- input is properly validated
	out, err := exec.Command(execBusctl, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting link dns for %s via dbus: %s: %w", iface.Name, strings.TrimSpace(string(out)), err)
	}

	// Set domains
	// #nosec G204 -- input is properly validated
	out, err = exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"SetLinkDomains", "ia(sb)", fmt.Sprintf("%d", iface.Index), "1", ".", "true",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting link domains for %s via dbus: %s: %w", iface.Name, strings.TrimSpace(string(out)), err)
	}

	// Use secure DNS extension, but allow to downgrade if it's unsupported
	// #nosec G204 -- input is properly validated
	out, err = exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"SetLinkDNSSEC", "is", fmt.Sprintf("%d", iface.Index), "allow-downgrade",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting link dns sec for %s via dbus: %s: %w", iface.Name, strings.TrimSpace(string(out)), err)
	}

	links, err := internal.NetworkLinks()
	if err != nil {
		return fmt.Errorf("listing network links: %w", err)
	}
	// Setup other links
	for _, link := range links {
		// lo is managed by systemd-networkd
		// vpn and managed links should be ignored
		if link.Name == "lo" || link.Name == iface.Name || !internal.IsNetworkLinkUnmanaged(link.Name) {
			continue
		}

		// Remove domains
		// #nosec G204 -- input is properly validated
		out, err = exec.Command(execBusctl,
			"call",
			"org.freedesktop.resolve1",
			"/org/freedesktop/resolve1",
			"org.freedesktop.resolve1.Manager",
			"SetLinkDomains", "ia(sb)", fmt.Sprintf("%d", link.Index), "0",
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf("setting link domains for %s via dbus: %s: %w", link.Name, strings.TrimSpace(string(out)), err)
		}
	}

	out, err = exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"FlushCaches",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("flushing local dns caches via dbus: %s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func unsetDNSWithSystemdResolve(ifname string) error {
	if ifname == "" {
		return nil
	}

	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		return err
	}

	// #nosec G204 -- input is properly validated
	out, err := exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"RevertLink", "i", fmt.Sprintf("%d", iface.Index),
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("reverting link %s via dbus: %s: %w", iface.Name, strings.TrimSpace(string(out)), err)
	}

	out, err = exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"FlushCaches",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("flushing local dns caches via dbus: %s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func resolvconfIfacePrefix() (string, error) {
	if internal.FileExists(resolconfInterfaceFilePath) {
		file, err := os.Open(resolconfInterfaceFilePath)
		if err != nil {
			return "", fmt.Errorf("opening %s: %w", resolconfInterfaceFilePath, err)
		}
		// #nosec G307 -- no writes are made
		defer file.Close()
		fscanner := bufio.NewScanner(file)
		re := regexp.MustCompile("^([A-Za-z0-9-]+)\\*$")
		for fscanner.Scan() {
			match := re.FindStringSubmatch(fscanner.Text())
			if len(match) > 0 {
				return fmt.Sprintf("%s.", match[1]), nil
			}
		}
		return "", nil
	}
	return "", nil
}

func setDNSWithResolvconf(iface string, addresses []string) error {
	var addrs = make([]string, len(addresses))
	for idx, address := range addresses {
		addrs[idx] = "nameserver " + address
	}
	content := strings.Join(addrs, "\n")
	prefix, err := resolvconfIfacePrefix()
	if err != nil {
		return fmt.Errorf("determining interface prefix: %w", err)
	}

	// #nosec G204 -- the code would have failed already if iface did not belong
	// to an actual network interface on the system
	cmd := exec.Command(execResolvconf, "-a", prefix+iface, "-m", "0", "-x")
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting dns with resolvconf: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func unsetDNSWithResolvconf(iface string) error {
	prefix, err := resolvconfIfacePrefix()
	if err != nil {
		return fmt.Errorf("determining interface prefix: %w", err)
	}

	// #nosec G204 -- the code would have failed already if iface did not belong
	// to an actual network interface on the system
	cmd := exec.Command(execResolvconf, "-d", prefix+iface, "-f")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("calling %s: %s", cmd.String(), strings.Trim(string(out), "\n"))
	}
	return nil
}

func setDNSDefault(addresses []string) error {
	err := backupDNS()
	if err != nil {
		return fmt.Errorf("backing up dns: %w", err)
	}
	return resetDNSDefault(addresses)
}

func resetDNSDefault(addresses []string) error {
	var addrs = make([]string, len(addresses))
	for idx, address := range addresses {
		addrs[idx] = "nameserver " + address
	}

	// set DNS
	content := "# Generated by NordVPN\n" + strings.Join(addrs, "\n")
	return internal.FileWrite(resolvconfFilePath, []byte(content), internal.PermUserRWGroupROthersR)
}

func unsetDNSDefault() error {
	out, err := internal.FileRead(resolvconfFilePath)
	if err != nil {
		return fmt.Errorf("reading resolv.conf: %w", err)
	}
	if strings.Contains(string(out), "Generated by NordVPN") {
		return restoreDNS()
	}
	return nil
}

func backupDNS() error {
	if internal.FileExists(resolvconfBackupPath) {
		return nil
	}
	out, err := internal.FileRead(resolvconfFilePath)
	if err != nil {
		return fmt.Errorf("reading resolv.conf: %w", err)
	}
	return internal.FileWrite(resolvconfBackupPath, out, internal.PermUserRWGroupROthersR)
}

func restoreDNS() error {
	defer internal.FileDelete(resolvconfBackupPath)
	out, err := internal.FileRead(resolvconfBackupPath)
	if err != nil {
		return fmt.Errorf("reading resolv.conf backup: %w", err)
	}
	return internal.FileWrite(resolvconfFilePath, out, internal.PermUserRWGroupROthersR)
}
