package network

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
)

// Resolver is a DNSResolver implementation wrapping each DHCP request with
// allowing and blocking firewall rules
type Resolver struct {
	fw      firewall.Service
	servers dns.Getter
	sync.Mutex
}

func NewResolver(fw firewall.Service, servers dns.Getter) *Resolver {
	return &Resolver{
		fw:      fw,
		servers: servers,
	}
}

type DNSResolver interface {
	Resolve(domain string) ([]netip.Addr, error)
}

func (r *Resolver) Resolve(domain string) ([]netip.Addr, error) {
	nameservers := r.servers.Get(false)
	return r.ResolveWithNameservers(domain, StringsToIPs(nameservers), "udp")
}

func (r *Resolver) ResolveWithNameservers(domain string, nameservers []netip.Addr, protocol string) ([]netip.Addr, error) {
	r.Lock()
	defer r.Unlock()

	err := allowlistIP(r.fw, "allow_dns", nameservers...)
	if err != nil {
		return nil, fmt.Errorf("allowlisting DNS IP addresses %+v: %w", nameservers, err)
	}
	defer r.fw.Delete([]string{"allow_dns"}) // ignore error here
	// get the addresses from DNS
	var ipAddrs []netip.Addr
	for _, nameserver := range nameservers {
		ipAddrs, err = LookupAddressWithCustomDNS(domain, nameserver.String(), protocol)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("looking address up: %w", err)
	}
	return ipAddrs, nil
}

func allowlistIP(fw firewall.Service, name string, ips ...netip.Addr) error {
	ifaces, err := device.ListPhysical()
	if err != nil {
		return fmt.Errorf("listing physical interfaces: %w", err)
	}

	var networks []netip.Prefix
	for _, ip := range ips {
		networks = append(networks, netip.PrefixFrom(ip, ip.BitLen()))
	}
	if err := fw.Add([]firewall.Rule{
		{
			Name:           name,
			RemoteNetworks: networks,
			Interfaces:     ifaces,
			Direction:      firewall.TwoWay,
			Allow:          true,
			Physical:       true,
		},
	}); err != nil {
		return fmt.Errorf("adding firewall rule %s for %+v: %w", name, ips, err)
	}
	return nil
}

// ResolverChain tries each resolver until the first successful one.
type ResolverChain struct {
	resolvers []EndpointResolver
}

func NewDefaultResolverChain(fw firewall.Service) ResolverChain {
	return ResolverChain{
		resolvers: []EndpointResolver{
			NewPingConnectionChecker(fw),
		},
	}
}

func (c ResolverChain) Resolve(endpointIP netip.Addr) ([]netip.Addr, error) {
	for _, resolver := range c.resolvers {
		ip, err := resolver.Resolve(endpointIP)
		if err == nil {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("unable to resolve ip %s", endpointIP)
}

// PingConnectionChecker is the only resolver used by ResolverChain
type PingConnectionChecker struct {
	fw firewall.Service
}

func NewPingConnectionChecker(fw firewall.Service) PingConnectionChecker {
	return PingConnectionChecker{fw}
}

func (c PingConnectionChecker) Resolve(endpointIP netip.Addr) ([]netip.Addr, error) {
	if err := allowlistIP(c.fw, "allow_ping", endpointIP); err != nil {
		return nil, err
	}
	defer c.fw.Delete([]string{"allow_ping"})

	var err error
	for i := 0; i < 3; i++ {
		err = Ping(endpointIP.String(), 1)
		if err == nil {
			return []netip.Addr{endpointIP}, nil
		}
	}
	return nil, err
}
