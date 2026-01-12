package dns

import (
	"log"
	"math/rand"
	"net"
	"slices"
	"sync/atomic"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	primaryNameserver4                       = "103.86.96.100"
	secondaryNameserver4                     = "103.86.99.100"
	threatProtectionLitePrimaryNameserver4   = "103.86.96.96"
	threatProtectionLiteSecondaryNameserver4 = "103.86.99.99"
)

var (
	defaultTpServers = []string{
		threatProtectionLitePrimaryNameserver4, threatProtectionLiteSecondaryNameserver4,
	}
	defaultServers = []string{primaryNameserver4, secondaryNameserver4}
)

type ServersFetcher func() (*core.NameServers, error)

type Getter interface {
	Get(isThreatProtectionLite bool) []string
	LookupIP(host string) ([]net.IP, error)
}

type NameServers struct {
	// Pointer to the List of TP servers fetched from cloud
	tpServers atomic.Pointer[[]string]
}

func NewNameServers(fetcher ServersFetcher, timeoutFn internal.CalculateRetryDelayForAttempt) *NameServers {
	n := &NameServers{}
	// start async to fetch the TP server names
	go n.fetchTpServers(fetcher, timeoutFn)
	return n
}

// Get nameservers selected by the given criteria.
func (n *NameServers) Get(isThreatProtectionLite bool) []string {
	if isThreatProtectionLite {
		return n.getTpServers()
	}

	return shuffleNameservers(slices.Clone(defaultServers))
}

func (n *NameServers) getTpServers() []string {
	servers := n.tpServers.Load()
	if servers != nil && len(*servers) != 0 {
		return shuffleNameservers(slices.Clone(*servers))
	}

	return shuffleNameservers(slices.Clone(defaultTpServers))
}

func (n *NameServers) LookupIP(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

func (n *NameServers) fetchTpServers(fetcher ServersFetcher, timeoutFn internal.CalculateRetryDelayForAttempt) {
	if fetcher == nil {
		return
	}

	for {
		servers, err := fetcher()
		if err == nil && len(servers.Servers) > 0 {
			// copy to ensure pointer is not later modified from outside
			s := slices.Clone(servers.Servers)
			n.tpServers.Store(&s)

			return
		}

		var i = 1
		tryAfterDuration := timeoutFn(i)
		log.Printf("%s failed to fetch TP servers: %v, retry(%d) servers after %v\n", internal.WarningPrefix, err, i, tryAfterDuration)
		i += 1
		<-time.After(tryAfterDuration)
	}
}

func shuffleNameservers(nameservers []string) []string {
	// #nosec G404 - Using math/rand for nameserver shuffling is acceptable
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	r.Shuffle(len(nameservers), func(i, j int) {
		nameservers[i], nameservers[j] = nameservers[j], nameservers[i]
	})
	return nameservers
}
