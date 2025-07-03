package dns

import (
	"math/rand"
	"net"
	"time"
)

const (
	primaryNameserver4                       = "103.86.96.100"
	secondaryNameserver4                     = "103.86.99.100"
	threatProtectionLitePrimaryNameserver4   = "103.86.96.96"
	threatProtectionLiteSecondaryNameserver4 = "103.86.99.99"
)

type Getter interface {
	Get(isThreatProtectionLite bool) []string
	LookupIP(host string) ([]net.IP, error)
}

type NameServers struct {
	// List of TP servers fetched from cloud
	servers []string
}

func NewNameServers(servers []string) *NameServers {
	return &NameServers{servers}
}

// Get nameservers selected by the given criteria.
func (n *NameServers) Get(isThreatProtectionLite bool) []string {
	if isThreatProtectionLite {
		var nameservers []string
		v4Nameservers := n.servers
		if len(v4Nameservers) == 0 {
			v4Nameservers = []string{threatProtectionLitePrimaryNameserver4, threatProtectionLiteSecondaryNameserver4}
		}

		nameservers = append(nameservers, v4Nameservers...)
		return shuffleNameservers(nameservers)
	}

	var nameservers []string
	nameservers = append(nameservers, primaryNameserver4, secondaryNameserver4)
	return shuffleNameservers(nameservers)
}

func (n *NameServers) LookupIP(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

func shuffleNameservers(nameservers []string) []string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	r.Shuffle(len(nameservers), func(i, j int) {
		nameservers[i], nameservers[j] = nameservers[j], nameservers[i]
	})
	return nameservers
}
