package dns

import (
	"errors"
	"fmt"
	"net"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/godbus/dbus/v5"
)

const (
	dbusDestination = "org.freedesktop.resolve1"
	dbusPath        = "/org/freedesktop/resolve1"
)

// linkDomain represents a domain for SetLinkDomains
type linkDomain struct {
	Domain      string
	RoutingOnly bool
}

// dnsServer represents a DNS server address for SetLinkDNS
type dnsServer struct {
	Family  int32
	Address []byte
}

func newDNSServer(ip netip.Addr) dnsServer {
	// family: 2=IPv4, 10=IPv6
	var family int32
	var raw []byte
	if ip.Is4() {
		b := ip.As4()
		raw = b[:]
		family = 2
	} else {
		b := ip.As16()
		raw = b[:]
		family = 10
	}
	return dnsServer{
		Family:  family,
		Address: raw,
	}
}

// Systemd-resolved DBUS API based DNS handling method
type Resolved struct{}

func (m *Resolved) Set(iface string, nameservers []string) error {
	return m.setDNSWithSystemdResolve(iface, nameservers)
}

func (m *Resolved) Unset(iface string) error {
	return m.unsetDNSWithSystemdResolve(iface)
}

func (m *Resolved) Name() string {
	return "resolved"
}

// setDNSWithSystemdResolve uses systemd-resolve dbus API to manage DNS
// https://www.freedesktop.org/wiki/Software/systemd/resolved/
// Following the logic from https://systemd.io/RESOLVED-VPNS/
// resolvectl domain privacy0 '~.'
// resolvectl default-route privacy0 true
// resolvectl dns privacy0 <IP>
func (m *Resolved) setDNSWithSystemdResolve(ifname string, addresses []string) error {
	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		return err
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		log.DNS.Error("cannot create dbus connection", ifname, err)
		return err
	}

	obj := conn.Object(dbusDestination, dbusPath)

	return m.configure(obj, iface, addresses)
}

func (m *Resolved) configure(obj dbus.BusObject, iface *net.Interface, addresses []string) error {
	if err := setLinkDNS(obj, iface.Index, addresses); err != nil {
		return err
	}

	if call := obj.Call(methodName("SetLinkDomains"), 0, iface.Index, []linkDomain{{Domain: ".", RoutingOnly: true}}); call.Err != nil {
		return fmt.Errorf("setting DNS domain for %s v: %w", iface.Name, call.Err)
	}

	// empty string makes the link use the global DNSSEC setting
	dnsSECOption := ""
	if areOnlyNordDNSServers(addresses) {
		dnsSECOption = "no"
	}

	if call := obj.Call(methodName("SetLinkDNSSEC"), 0, iface.Index, dnsSECOption); call.Err != nil {
		return fmt.Errorf("setting link DNS sec for %s v: %w", iface.Name, call.Err)
	}

	if call := obj.Call(methodName("FlushCaches"), 0); call.Err != nil {
		return fmt.Errorf("configure DNS flushing local dns caches via dbuss: %w", call.Err)
	}

	return nil
}

func (m *Resolved) unsetDNSWithSystemdResolve(ifname string) error {
	if ifname == "" {
		log.DNS.Info("ignore unset DNS because interface name is empty")
		return nil
	}

	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		log.DNS.Error("cannot find interface with name:", ifname, err)
		return err
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		log.DNS.Error("cannot create dbus connection for unset", ifname, err)
		return err
	}

	obj := conn.Object(dbusDestination, dbusPath)

	if err := m.unset(obj, iface); err != nil {
		return err
	}

	return nil
}

func (*Resolved) unset(obj dbus.BusObject, iface *net.Interface) error {
	if call := obj.Call(methodName("RevertLink"), 0, iface.Index); call.Err != nil {
		return fmt.Errorf("reverting link %s via dbus: %w", iface.Name, call.Err)
	}

	if call := obj.Call(methodName("FlushCaches"), 0); call.Err != nil {
		return fmt.Errorf("unset DNS flushing local dns caches via dbuss: %w", call.Err)
	}
	return nil
}

func methodName(name string) string {
	return fmt.Sprintf("org.freedesktop.resolve1.Manager.%s", name)
}

func setLinkDNS(obj dbus.BusObject, ifIndex int, addresses []string) error {
	servers := make([]dnsServer, 0, len(addresses))
	for _, address := range addresses {
		ip, err := netip.ParseAddr(address)
		if err != nil {
			log.DNS.Error("cannot convert DNS address, ignoring it", address)
			continue
		}
		servers = append(servers, newDNSServer(ip))
	}
	if len(servers) == 0 {
		return errors.New("DNS addresses list is empty")
	}

	if call := obj.Call(methodName("SetLinkDNS"), 0, ifIndex, servers); call.Err != nil {
		return fmt.Errorf("setting link dns for %d v: %w", ifIndex, call.Err)
	}

	return nil
}
