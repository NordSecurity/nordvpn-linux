package network

import (
	"fmt"
	"net/netip"
	"sync"
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// Resolver is a DNSResolver implementation wrapping each DHCP request with
// allowing and blocking firewall rules
type Resolver struct {
	servers dns.Getter
	fwmark  uint32
	// store if VPN is connected or not. Based on this is decided if the firewall mark is used or not
	isVpnConnected atomic.Bool
	sync.Mutex
}

func NewResolver(
	servers dns.Getter,
	fwmark uint32,
	daemonEvents *daemonevents.ServiceEvents,
) DNSResolver {
	resolver := &Resolver{
		servers:        servers,
		fwmark:         fwmark,
		isVpnConnected: atomic.Bool{},
	}

	daemonEvents.Connect.Subscribe(func(dc events.DataConnect) error {
		resolver.updateVpnStatus(dc.EventStatus == events.StatusSuccess)
		return nil
	})

	daemonEvents.Disconnect.Subscribe(func(dd events.DataDisconnect) error {
		resolver.updateVpnStatus(false)
		return nil
	})

	return resolver
}

type DNSResolver interface {
	Resolve(domain string) ([]netip.Addr, error)
}

func (r *Resolver) Resolve(domain string) ([]netip.Addr, error) {
	nameservers := r.servers.Get(false)
	return r.resolveWithNameservers(domain, FilterInvalidIPs(nameservers), "udp")
}

func (r *Resolver) resolveWithNameservers(domain string, nameservers []string, protocol string) ([]netip.Addr, error) {
	r.Lock()
	defer r.Unlock()

	// get the addresses from DNS
	var ipAddrs []netip.Addr
	var err error
	for _, nameserver := range nameservers {
		var fwmark = r.fwmark
		if r.isVpnConnected.Load() {
			// While connected to VPN, send the DNS requests thru the tunnel so no fwmark
			fwmark = noFwMark
		}
		ipAddrs, err = lookupAddress(domain, nameserver, protocol, fwmark)

		if err == nil {
			return ipAddrs, nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("looking address up: %w", err)
	}
	return ipAddrs, nil
}

func (r *Resolver) updateVpnStatus(isConnected bool) {
	log.Println(internal.InfoPrefix, "resolver set VPN connected to", isConnected)
	r.isVpnConnected.Store(isConnected)
}
