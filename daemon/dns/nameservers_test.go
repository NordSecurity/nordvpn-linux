package dns

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

// helper function to return the needed format for the fetcher used in NewNameServers
func wrapServersList(servers []string) func() (*core.NameServers, error) {
	return func() (*core.NameServers, error) {
		return &core.NameServers{
			Servers: servers,
		}, nil
	}
}

func TestDiscoverNameserverIp(t *testing.T) {
	ip, err := discoverNameserverIp()
	assert.NoError(t, err)
	assert.NotNil(t, ip)
}

func TestNameservers(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		threatProtectionLite bool
		initial              []string
		expected             []string
	}{
		{
			name:                 "ipv4",
			threatProtectionLite: false,
			initial:              defaultTpServers,
			expected:             defaultServers,
		},
		{
			name:                 "ipv4 threat protection lite",
			threatProtectionLite: true,
			initial:              defaultTpServers,
			expected:             defaultTpServers,
		},
		{
			name:                 "empty initial list",
			threatProtectionLite: true,
			initial:              nil,
			expected:             defaultTpServers,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			servers := NewNameServers(wrapServersList(test.initial), internal.ExponentialBackoff)
			nameservers := servers.Get(test.threatProtectionLite)
			assert.ElementsMatch(t, test.expected, nameservers)
		})
	}
}

func TestNameserversRandomness(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		threatProtectionLite bool
		initial              []string
		expected             []string
	}{
		{
			name:                 "randomness",
			threatProtectionLite: true,
			initial: []string{
				"1.1.1.1", "1.0.0.1", "8.8.8.8", "8.8.4.4",
				threatProtectionLitePrimaryNameserver4, threatProtectionLitePrimaryNameserver4,
			},
			expected: []string{
				"1.1.1.1", "1.0.0.1", "8.8.8.8", "8.8.4.4",
				threatProtectionLitePrimaryNameserver4, threatProtectionLitePrimaryNameserver4,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			servers := NewNameServers(wrapServersList(test.initial), internal.ExponentialBackoff)

			// give time to fetch the data, before checking it
			time.Sleep(time.Millisecond * 5)

			nameservers1 := servers.Get(test.threatProtectionLite)
			nameservers2 := servers.Get(test.threatProtectionLite)

			// Make sure they contain the expected elements
			assert.ElementsMatch(t, test.expected, nameservers1)
			assert.ElementsMatch(t, test.expected, nameservers2)

			// If by any chance the lists have the same order
			// Generate a third one and if that has the same order
			// with the first two, then we have a problem with shuffle
			if reflect.DeepEqual(nameservers1, nameservers2) {
				nameservers3 := servers.Get(test.threatProtectionLite)
				assert.ElementsMatch(t, test.expected, nameservers3)
				assert.NotEqual(t, nameservers1, nameservers3)
			}
		})
	}
}

func TestNameserversNotCrashingWithNilServersFetcher(t *testing.T) {
	category.Set(t, category.Unit)
	nameservers := NewNameServers(nil, nil)

	// check that the default servers are returned
	assert.ElementsMatch(t, defaultTpServers, nameservers.Get(true))
}

func TestNameserversRetriesToFetchTPOnError(t *testing.T) {
	category.Set(t, category.Unit)

	servers := []string{"1.2.3.4"}

	var retries atomic.Int32
	retries.Store(3)

	var wg sync.WaitGroup
	wg.Add(1)

	nameservers := NewNameServers(
		func() (*core.NameServers, error) {
			// return error for `retries` times, before returning servers list
			if retries.Load() == 0 {
				defer wg.Done()
				return &core.NameServers{Servers: servers}, nil
			}

			return nil, errors.New("fail to fetch")
		},
		func(attempt int) time.Duration {
			retries.Add(-1)
			assert.True(t, retries.Load() >= 0, "called too many times")
			return time.Millisecond
		},
	)

	wg.Wait()
	time.Sleep(time.Millisecond * 5)
	assert.ElementsMatch(t, servers, nameservers.Get(true))
	assert.Equal(t, int32(0), retries.Load())
}
