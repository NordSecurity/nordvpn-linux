package dns

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Executables
const (
	// execBusctl defines busctl executable
	execBusctl = "busctl"
)

// Systemd-resolved DBUS API based DNS handling method
type Resolved struct{}

func (m *Resolved) Set(iface string, nameservers []string) error {
	return setDNSWithSystemdResolve(iface, nameservers)
}

func (m *Resolved) Unset(iface string) error {
	return unsetDNSWithSystemdResolve(iface)
}

func (m *Resolved) Name() string {
	return "resolved"
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

	// Set routing domains (more info: https://github.com/poettering/systemd/commit/8cedb0aef94da880e61b4c8cfeb7f450f8760ec6)
	// #nosec G204 -- input is properly validated
	out, err = exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"SetLinkDomains", "ia(sb)", fmt.Sprintf("%d", iface.Index), "1", ".", "true",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting link routing domains for %s via dbus: %s: %w", iface.Name, strings.TrimSpace(string(out)), err)
	}

	// Set Default route to tunnel interface
	// #nosec G204 -- input is properly validated
	out, err = exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"SetLinkDefaultRoute", "ib", fmt.Sprintf("%d", iface.Index), "true",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting link default route for %s via dbus: %s: %w", iface.Name, strings.TrimSpace(string(out)), err)
	}

	// Use secure DNS extension, but allow to downgrade if it's unsupported
	// #nosec G204 -- input is properly validated
	out, err = exec.Command(execBusctl,
		"call",
		"org.freedesktop.resolve1",
		"/org/freedesktop/resolve1",
		"org.freedesktop.resolve1.Manager",
		"SetLinkDNSSEC", "is", fmt.Sprintf("%d", iface.Index), "no",
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
