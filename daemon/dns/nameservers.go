package dns

import (
	"math/rand"
	"net"
	"slices"
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
		if len(n.servers) == 0 {
			return shuffleNameservers([]string{threatProtectionLitePrimaryNameserver4, threatProtectionLiteSecondaryNameserver4})
		} else {
			return shuffleNameservers(slices.Clone(n.servers))
		}
	}

	return shuffleNameservers([]string{primaryNameserver4, secondaryNameserver4})
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
