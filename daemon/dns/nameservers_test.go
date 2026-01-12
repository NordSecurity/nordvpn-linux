package dns

import (
	"reflect"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

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

			// give time to fetch the data
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

func TestNameserversNotCrashingForNilServersFetcher(t *testing.T) {
	category.Set(t, category.Unit)
	nameservers := NewNameServers(nil, internal.ExponentialBackoff)

	// check that the default servers are returned
	assert.ElementsMatch(t, defaultTpServers, nameservers.Get(true))
}
