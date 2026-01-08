/*
Package dns is responsible for configuring dns on various Linux distros.
*/
package dns

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	dnsPrefix = "[DNS]"
	// resolvconfFilePath defines path to resolv.conf file for DNS
	resolvconfFilePath = "/etc/resolv.conf"
	// resolvdComment is the comment used to mark resolv.conf managed by systemd-resolved.
	resolvdComment = "# This is /run/systemd/resolve/stub-resolv.conf managed by man:systemd-resolved(8)."

	resolvedLinkTarget = "../run/systemd/resolve/stub-resolv.conf"
)

var ErrDNSNotSet = fmt.Errorf("DNS not set")

// symlinkFilesystemHandle extends FilesystemHandle with symlink resolution capabilities
type symlinkFilesystemHandle interface {
	config.FilesystemHandle
	getLinkTarget(string) (string, error)
}

type stdSymlinkFilesystemHandle struct {
	config.StdFilesystemHandle
}

func (s *stdSymlinkFilesystemHandle) getLinkTarget(location string) (string, error) {
	return os.Readlink(location)
}

// Setter is responsible for configuring DNS.
type Setter interface {
	Set(iface string, nameservers []string) error
	Unset(iface string) error
}

// Method is abstraction of DNS handling method
type Method interface {
	Set(iface string, nameservers []string) error
	Unset(iface string) error
	Name() string
}

// DNSServiceSetter detects how OS is managing the DNS configuration and tries to set it using the appropriate method.
type DNSServiceSetter struct {
	// systemdResolvedSetter sets DNS using the most desired method:
	// 	1. systemd-resolved DBUS
	// 	2. resolvectl exec call
	systemdResolvedSetter Setter
	// resolvconfSetter sets DNS using the most desired method:
	//	1. resolvconf exec call
	//	2. direct write to /etc/resolv.conf
	resolvconfSetter Setter
	unsetter         Setter
	filesystemHandle symlinkFilesystemHandle
}

func NewDNSServiceSetter(publisher events.Publisher[string]) *DNSServiceSetter {
	return &DNSServiceSetter{
		systemdResolvedSetter: NewSetter(publisher, &Resolved{}, &Resolvectl{}),
		resolvconfSetter:      NewSetter(publisher, &Resolvconf{}, &ResolvConfFile{}),
		filesystemHandle:      &stdSymlinkFilesystemHandle{},
	}
}

// set sets DNS using the provided setter and sets a matching unsetter if the operation was successful
func (d *DNSServiceSetter) set(setter Setter, iface string, nameservers []string) error {
	err := setter.Set(iface, nameservers)
	if err != nil {
		return fmt.Errorf("failed to set DNS: %w", err)
	}

	d.unsetter = setter
	return nil
}

// setBasedOnComment sets DNS using the setter matching the comment specified in resolv.conf
func (d *DNSServiceSetter) setBasedOnComment(iface string, nameservers []string) error {
	resolvConfFileContents, err := d.filesystemHandle.ReadFile(resolvconfFilePath)
	if err != nil {
		return fmt.Errorf("reading resolv.conf file: %w", err)
	}

	resolvConfFileContentsStr := string(resolvConfFileContents)
	if strings.Contains(resolvConfFileContentsStr, resolvdComment) {
		log.Println(internal.InfoPrefix, dnsPrefix,
			"configuring DNS with systemd-resolved, inferred from resolv.conf comment")
		if err := d.set(d.systemdResolvedSetter, iface, nameservers); err != nil {
			return fmt.Errorf("setting DNS based on resolv.conf comment: %w", err)
		}
		log.Println(internal.InfoPrefix, dnsPrefix, "DNS configured with systemd-resolved")
		return nil
	}

	return fmt.Errorf("management service not recognized from resolv.conf comment")
}

// setBasedOnResolvConfLinkTarget sets DNS using the setter matching the link target of resolv.conf
func (d *DNSServiceSetter) setBasedOnResolvConfLinkTarget(iface string, nameservers []string) error {
	resolvConfLinkDestination, err := d.filesystemHandle.getLinkTarget(resolvconfFilePath)
	if err != nil {
		return fmt.Errorf("failed to obtain resolv.conf link target: %w", err)
	}

	if strings.Contains(resolvConfLinkDestination, resolvedLinkTarget) {
		log.Println(internal.InfoPrefix, dnsPrefix, "configuring DNS with systemd-resolved, inferred from link target")
		if err := d.set(d.systemdResolvedSetter, iface, nameservers); err != nil {
			return fmt.Errorf("setting DNS based on resolv.conf link target: %w", err)
		}
		log.Println(internal.InfoPrefix, dnsPrefix, "DNS configured with systemd-resolved")
		return nil
	}

	return fmt.Errorf("management service not recognized from resolv.conf link destination")
}

// setUsingBestAvailable sets DNS using the first setter in the bellow priority list:
//  1. systemd-resolvd DBUS
//  2. resolvctl utility
//  3. resolvconf utility
//  4. direct write to resovl.conf
func (d *DNSServiceSetter) setUsingBestAvailable(iface string, nameservers []string) error {
	if err := d.set(d.systemdResolvedSetter, iface, nameservers); err != nil {
		log.Println(internal.ErrorPrefix, dnsPrefix,
			"failed to configure DNS using systemd-resolvd, attempting with resolv.conf")
	} else {
		log.Println(internal.InfoPrefix, dnsPrefix, "DNS configured with systemd-resolvd")
		return nil
	}

	if err := d.set(d.resolvconfSetter, iface, nameservers); err != nil {
		return fmt.Errorf("failed to configure DNS with resolv.conf: %w", err)
	}

	log.Println(internal.InfoPrefix, dnsPrefix, "DNS configured with resolv.conf")

	return nil
}

// Set sets the DNS using the most appropriate method, it attempts to:
//  1. Infer which method to use from the comment in /etc/resolv.conf
//  2. If the above fails, infer which method to use based on /etc/resovl.conf link target
//  3. If the above fails, it attempts to use best available method(see setUsingBestAvailable)
//  4. If all of the above fail, it returns an error
func (d *DNSServiceSetter) Set(iface string, nameservers []string) error {
	if err := d.setBasedOnComment(iface, nameservers); err != nil {
		log.Println(internal.ErrorPrefix, dnsPrefix, "failed to configure DNS based on the resolv.conf comment:", err)
	} else {
		return nil
	}

	if err := d.setBasedOnResolvConfLinkTarget(iface, nameservers); err != nil {
		log.Println(internal.ErrorPrefix, dnsPrefix, "failed to configure DNS based on link destination:", err)
	} else {
		return nil
	}

	log.Println(internal.ErrorPrefix, dnsPrefix,
		"failed to detect DNS management service based on resolv.conf, attempting to use the best avaialable method")
	if err := d.setUsingBestAvailable(iface, nameservers); err != nil {
		return fmt.Errorf("failed to set DNS: %w", err)
	}

	return nil
}

// Uset unsets the DNS using the method that was used to set it
func (d *DNSServiceSetter) Unset(iface string) error {
	if d.unsetter == nil {
		return ErrDNSNotSet
	}

	log.Println(internal.DebugPrefix, dnsPrefix, "unsetting DNS")
	if err := d.unsetter.Unset(iface); err != nil {
		return fmt.Errorf("unsetting DNS: %w", err)
	}

	return nil
}

// DNSMethodSetter iterates over the list of DNS configuration methods and tries to apply the desired DNS config with each of
// them.
type DNSMethodSetter struct {
	publisher events.Publisher[string]
	methods   []Method
}

func NewSetter(publisher events.Publisher[string], methods ...Method) *DNSMethodSetter {
	ds := DNSMethodSetter{
		publisher: publisher,
		methods:   []Method{},
	}

	ds.methods = append(ds.methods, methods...)
	return &ds
}

// Set DNS for a given iface if the system supports per interface DNS settings.
// Also, backup current DNS settings (only in case of direct resolv.conf edit).
// Backup is not overridden, so its safe to call this function multiple times in a row.
func (d *DNSMethodSetter) Set(iface string, nameservers []string) error {
	d.publisher.Publish(
		"setting dns to " + strings.Join(nameservers, " "),
	)

	if len(nameservers) == 0 {
		return errors.New("nameservers not provided")
	}

	for _, method := range d.methods {
		d.publisher.Publish("set dns for interface [" + iface + "] using: " + method.Name())
		if err := method.Set(iface, nameservers); err != nil {
			log.Println(internal.ErrorPrefix, fmt.Errorf("setting dns with %s: %w", method.Name(), err))
			continue
		}
		return nil
	}

	return fmt.Errorf("dns not set, no dns setting method is available")
}

// Unset DNS for network interface, restore DNS from a backup, if backup
// is available, and remove the backup on success.
func (d *DNSMethodSetter) Unset(iface string) error {
	d.publisher.Publish("unsetting DNS")

	for _, method := range d.methods {
		d.publisher.Publish("unset dns for interface [" + iface + "] using: " + method.Name())
		if err := method.Unset(iface); err != nil {
			log.Println(internal.ErrorPrefix, fmt.Errorf("unsetting dns with %s: %w", method.Name(), err))
			continue
		}
		return nil
	}

	return nil
}

// RestoreResolvConfFile try to restore resolv.conf if target file contains Nordvpn changes
func RestoreResolvConfFile() {
	tryToRestoreDNS()
}
