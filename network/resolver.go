package network

import (
	"errors"
	"fmt"
	"net/netip"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
)

// Resolver is a DNSResolver implementation wrapping each DHCP request with
// allowing and blocking firewall rules
type Resolver struct {
	servers dns.Getter
	fwmark  uint32
	sync.Mutex
}

func NewResolver(servers dns.Getter, fwmark uint32) *Resolver {
	return &Resolver{
		servers: servers,
		fwmark:  fwmark,
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

	// get the addresses from DNS
	var ipAddrs []netip.Addr
	var err error
	for _, nameserver := range nameservers {
		ipAddrs, err = LookupAddressWithCustomDNS(domain, nameserver.String(), protocol, r.fwmark)
		if err == nil {
			return ipAddrs, nil
		}
	}

	return nil, fmt.Errorf("looking address up: %w", err)
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
	return nil, errors.New("Not implemeted")
}
