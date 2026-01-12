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

type CalculateRetryDelayForAttempt func(attempt int) time.Duration
type ServersFetcher func() (*core.NameServers, error)

type Getter interface {
	Get(isThreatProtectionLite bool) []string
	LookupIP(host string) ([]net.IP, error)
}

type NameServers struct {
	// Pointer to the List of TP servers fetched from cloud
	tpServers atomic.Pointer[[]string]
}

func NewNameServers(fetcher ServersFetcher, timeoutFn CalculateRetryDelayForAttempt) *NameServers {
	n := &NameServers{}
	if fetcher != nil && timeoutFn != nil {
		// start async to fetch the TP server names
		go n.fetchTPServers(fetcher, timeoutFn)
	} else {
		log.Println(internal.ErrorPrefix, "no fetcher set for TP servers")
	}
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

// fetches the TP servers until is successful. It uses exponential backoff between retries
func (n *NameServers) fetchTPServers(fetcher ServersFetcher, timeoutFn CalculateRetryDelayForAttempt) {
	if fetcher == nil {
		return
	}

	for i := 0; ; i++ {
		servers, err := fetcher()
		if err == nil && len(servers.Servers) > 0 {
			// copy to ensure pointer is not later modified from outside
			log.Println(internal.InfoPrefix, "TP servers updated to", servers.Servers)
			s := slices.Clone(servers.Servers)
			n.tpServers.Store(&s)

			return
		}

		tryAfterDuration := timeoutFn(i)
		log.Printf("%s failed to fetch TP servers. retry(%d) servers after %v: %v\n", internal.ErrorPrefix, i, tryAfterDuration, err)
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
