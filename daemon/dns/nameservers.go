package dns

import (
	"errors"
	"math/rand"
	"net"
	"slices"
	"sync/atomic"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const (
	primaryNameserver4                     = "103.86.96.100"
	secondaryNameserver4                   = "103.86.99.100"
	realTimeProtectionPrimaryNameserver4   = "103.86.96.108"
	realTimeProtectionSecondaryNameserver4 = "103.86.99.108"
)

var (
	defaultRtpServers = []string{
		realTimeProtectionPrimaryNameserver4, realTimeProtectionSecondaryNameserver4,
	}
	defaultServers = []string{primaryNameserver4, secondaryNameserver4}
)

type CalculateRetryDelayForAttempt func(attempt int) time.Duration
type ServersFetcher func() (*core.NameServers, error)

type Getter interface {
	Get(isRealTimeProtection bool) []string
	LookupIP(host string) ([]net.IP, error)
}

type NameServers struct {
	// Pointer to the List of RTP servers fetched from cloud
	rtpServers atomic.Pointer[[]string]
}

func NewNameServers() *NameServers {
	return &NameServers{}
}

// Get nameservers selected by the given criteria.
func (n *NameServers) Get(isRealTimeProtection bool) []string {
	if isRealTimeProtection {
		return n.getRtpServers()
	}

	return shuffleNameservers(slices.Clone(defaultServers))
}

func (n *NameServers) getRtpServers() []string {
	servers := n.rtpServers.Load()
	if servers != nil && len(*servers) != 0 {
		return shuffleNameservers(slices.Clone(*servers))
	}

	return shuffleNameservers(slices.Clone(defaultRtpServers))
}

func (n *NameServers) LookupIP(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

// FetchProtectionServers it is a blocking operation and fetches the protection servers until is successful.
// It uses exponential backoff between retries.
func (n *NameServers) FetchProtectionServers(fetcher ServersFetcher, timeoutFn CalculateRetryDelayForAttempt) error {
	if fetcher == nil || timeoutFn == nil {
		return errors.New("fetcher parameters cannot be nil")
	}

	for retry := 0; ; retry++ {
		servers, err := fetcher()
		if err == nil && len(servers.Servers) > 0 {
			// copy to ensure pointer is not later modified from outside
			log.Info("RTP servers updated to", servers.Servers)
			s := slices.Clone(servers.Servers)
			n.rtpServers.Store(&s)

			break
		}

		tryAfterDuration := timeoutFn(retry)
		log.Errorf("failed to fetch RTP servers. retry(%d) servers after %v: %v", retry, tryAfterDuration, err)
		<-time.After(tryAfterDuration)
	}

	return nil
}

func shuffleNameservers(nameservers []string) []string {
	// #nosec G404 - Using math/rand for nameserver shuffling is acceptable
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	r.Shuffle(len(nameservers), func(i, j int) {
		nameservers[i], nameservers[j] = nameservers[j], nameservers[i]
	})
	return nameservers
}
