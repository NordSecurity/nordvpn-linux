/*
Package dns is responsible for configuring dns on various Linux distros.
*/
package dns

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// Setter is responsible for configuring DNS.
type Setter interface {
	Set(iface string, nameservers []string) error
	Unset(iface string) error
}

// Method is abstraction of DNS handling method
type Method interface {
	Set(iface string, nameservers []string) error
	Unset(iface string) error
	IsAvailable() bool
	Name() string
}

/*
DefaultSetter handleds DNS in this order:

1. If systemd-resolve command is available and systemd-resolved.service is
running, systemd-resolve DBUS API is used.

2. In case of systemd-resolve is not accessible, resolvectl (which is part of
systemd-resolve package) command line utility is used.

3. In absence of systemd-resolve, resolvconf command line utility is used, which
modifies /etc/resolv.conf by adding or removing lines.

4. In case the resolvconf command line utility fails, /etc/resolv.conf is
backed up and modified directly by NordVPN.
*/
type DefaultSetter struct {
	publisher events.Publisher[string]
	methods   []Method
}

func NewSetter(publisher events.Publisher[string]) *DefaultSetter {
	ds := DefaultSetter{
		publisher: publisher,
		methods:   []Method{},
	}
	ds.methods = append(ds.methods, &Resolved{})
	ds.methods = append(ds.methods, &Resolvectl{})
	ds.methods = append(ds.methods, &Resolvconf{})
	ds.methods = append(ds.methods, &ResolvConfFile{})
	return &ds
}

// Set DNS for a given iface if the system supports per interface DNS settings.
// Also, backup current DNS settings (only in case of direct resolv.conf edit).
// Backup is not overridden, so its safe to call this function multiple times in a row.
func (d *DefaultSetter) Set(iface string, nameservers []string) error {
	d.publisher.Publish(
		"setting dns to " + strings.Join(nameservers, " "),
	)

	if len(nameservers) == 0 {
		return errors.New("nameservers not provided")
	}

	for _, method := range d.methods {
		if method.IsAvailable() {
			d.publisher.Publish("set dns for interface [" + iface + "] using: " + method.Name())
			if err := method.Set(iface, nameservers); err != nil {
				return fmt.Errorf("setting dns with %s: %w", method.Name(), err)
			}
			return nil
		}
	}

	return fmt.Errorf("dns not set, no dns setting method is available")
}

// Unset DNS for network interface, restore DNS from a backup, if backup
// is available, and remove the backup on success.
func (d *DefaultSetter) Unset(iface string) error {
	d.publisher.Publish("unsetting DNS")

	for _, method := range d.methods {
		if method.IsAvailable() {
			d.publisher.Publish("unset dns for interface [" + iface + "] using: " + method.Name())
			if err := method.Unset(iface); err != nil {
				return fmt.Errorf("unsetting dns with %s: %w", method.Name(), err)
			}
			return nil
		}
	}

	return nil
}
