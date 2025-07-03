package mock

import (
	"fmt"
	"net"

	"github.com/NordSecurity/nordvpn-linux/config"
)

var TplNameserversV4 config.DNS = []string{
	"103.86.96.96",
	"103.86.99.99",
}

var DefaultNameserversV4 config.DNS = []string{
	"103.86.96.100",
	"103.86.99.100",
}

type RegisteredDomainsList map[string][]net.IP
type DNSGetter struct {
	RegisteredDomains RegisteredDomainsList
	Names             []string
}

func (md *DNSGetter) Get(isThreatProtectionLite bool) []string {
	if len(md.Names) != 0 {
		return md.Names
	}
	if isThreatProtectionLite {
		nameservers := TplNameserversV4
		return nameservers
	}

	nameservers := DefaultNameserversV4
	return nameservers
}

func (md *DNSGetter) LookupIP(host string) ([]net.IP, error) {
	if v, ok := md.RegisteredDomains[host]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("domain not found")
}
