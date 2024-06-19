package dns

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverNameserverIp(t *testing.T) {
	ip, err := discoverNameserverIp()
	assert.NoError(t, err)
	assert.NotNil(t, ip)
}

func TestNameservers(t *testing.T) {
	category.Set(t, category.Unit)
	tpNameservers := []string{threatProtectionLitePrimaryNameserver4, threatProtectionLiteSecondaryNameserver4}

	servers := NewNameServers(tpNameservers)
	tests := []struct {
		name                 string
		threatProtectionLite bool
		ipv6                 bool
		initial              []string
		expected             []string
	}{
		{
			name:                 "ipv4",
			threatProtectionLite: false,
			ipv6:                 false,
			initial:              tpNameservers,
			expected:             []string{primaryNameserver4, secondaryNameserver4},
		},
		{
			name:                 "ipv6",
			threatProtectionLite: false,
			ipv6:                 true,
			expected: []string{
				primaryNameserver6, secondaryNameserver6,
				primaryNameserver4, secondaryNameserver4,
			},
		},
		{
			name:                 "ipv4 threat protection lite",
			threatProtectionLite: true,
			ipv6:                 false,
			initial:              tpNameservers,
			expected: []string{
				threatProtectionLitePrimaryNameserver4, threatProtectionLiteSecondaryNameserver4,
			},
		},
		{
			name:                 "ipv6 threat protection lite",
			threatProtectionLite: true,
			ipv6:                 true,
			initial:              tpNameservers,
			expected: []string{
				threatProtectionLitePrimaryNameserver6, threatProtectionLiteSecondaryNameserver6,
				threatProtectionLitePrimaryNameserver4, threatProtectionLiteSecondaryNameserver4,
			},
		},
		{
			name:                 "empty initial list",
			threatProtectionLite: true,
			ipv6:                 false,
			initial:              nil,
			expected: []string{
				threatProtectionLitePrimaryNameserver4, threatProtectionLiteSecondaryNameserver4,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nameservers := servers.Get(test.threatProtectionLite, test.ipv6)
			assert.ElementsMatch(t, test.expected, nameservers)
		})
	}
}
