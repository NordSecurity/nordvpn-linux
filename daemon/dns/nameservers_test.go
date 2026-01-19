package dns

import (
	"errors"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

// helper function to return the needed format for the fetcher used in NewNameServers
func wrapServersList(servers []string, wg *sync.WaitGroup) func() (*core.NameServers, error) {
	return func() (*core.NameServers, error) {
		if wg != nil {
			wg.Done()
		}
		return &core.NameServers{
			Servers: servers,
		}, nil
	}
}

func TestDiscoverNameserverIp(t *testing.T) {
	category.Set(t, category.Unit)

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
			name:                 "default DNS servers, TP=false",
			threatProtectionLite: false,
			initial:              defaultTpServers,
			expected:             defaultServers,
		},
		{
			name:                 "fetch TP list and return it",
			threatProtectionLite: true,
			initial:              defaultTpServers,
			expected:             defaultTpServers,
		},
		{
			name:                 "fetched servers are returned for TP servers",
			threatProtectionLite: true,
			initial:              []string{"1.2.3.4"},
			expected:             []string{"1.2.3.4"},
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
			servers := NewNameServers()

			// before fetching the servers from the API check the default values
			assert.ElementsMatch(t, defaultTpServers, servers.Get(true))
			assert.ElementsMatch(t, defaultServers, servers.Get(false))

			var wg sync.WaitGroup
			wg.Add(1)
			go servers.FetchTPServers(wrapServersList(test.initial, &wg), func(attempt int) time.Duration {
				assert.True(t, len(test.initial) == 0, "this must be called only when test.initial is empty")
				return time.Minute
			})

			wg.Wait()

			// retry several times to fetch the servers, because at first attempt internal members might not be stored
			for retry := 0; retry < 2; retry++ {
				if slices.Contains(servers.Get(test.threatProtectionLite), test.expected[0]) {
					break
				}
				time.Sleep(time.Millisecond * 2)
			}
			assert.ElementsMatch(t, test.expected, servers.Get(test.threatProtectionLite))
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
			servers := NewNameServers()

			// fetch in blocking mode
			servers.FetchTPServers(wrapServersList(test.initial, nil), func(attempt int) time.Duration { return time.Minute })

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
	nameservers := NewNameServers()
	assert.Error(t, nameservers.FetchTPServers(nil, nil))
	// check that the default servers are returned
	assert.ElementsMatch(t, defaultTpServers, nameservers.Get(true))
}

func TestNameserversRetriesToFetchTPOnError(t *testing.T) {
	category.Set(t, category.Unit)

	servers := []string{"1.2.3.4"}

	var retries atomic.Int32
	retries.Store(3)

	nameservers := NewNameServers()

	// run as blocking the fetch because there is no need to fetch in parallel and is successful after number of "retries"
	nameservers.FetchTPServers(
		func() (*core.NameServers, error) {
			// return error for `retries` times, before returning servers list
			if retries.Load() == 0 {
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

	assert.ElementsMatch(t, servers, nameservers.Get(true))
	assert.Equal(t, int32(0), retries.Load())
}
