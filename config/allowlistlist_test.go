package config

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSubnet(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		existing        []string
		newSubnet       string
		expectError     bool
		expectedSubnets []string
	}{
		{
			name:            "add to empty list",
			existing:        []string{},
			newSubnet:       "10.0.0.0/8",
			expectedSubnets: []string{"10.0.0.0/8"},
		},
		{
			name:            "add non-overlapping subnet",
			existing:        []string{"10.0.0.0/8"},
			newSubnet:       "192.168.0.0/16",
			expectedSubnets: []string{"10.0.0.0/8", "192.168.0.0/16"},
		},
		{
			name:            "wider subnet replaces existing narrower one",
			existing:        []string{"10.1.0.0/16"},
			newSubnet:       "10.0.0.0/8",
			expectedSubnets: []string{"10.0.0.0/8"},
		},
		{
			name:            "wider subnet replaces multiple narrower ones",
			existing:        []string{"10.1.0.0/16", "10.2.0.0/16", "192.168.0.0/24"},
			newSubnet:       "10.0.0.0/8",
			expectedSubnets: []string{"192.168.0.0/24", "10.0.0.0/8"},
		},
		{
			name:            "ipv6 wider replaces narrower",
			existing:        []string{"fd00:1::/32"},
			newSubnet:       "fd00::/16",
			expectedSubnets: []string{"fd00::/16"},
		},
		{
			name:            "ipv4 and ipv6 coexist",
			existing:        []string{"10.0.0.0/8"},
			newSubnet:       "fd00::/16",
			expectedSubnets: []string{"10.0.0.0/8", "fd00::/16"},
		},
		{
			name:            "subnet replaces single host",
			existing:        []string{"10.1.2.3/32"},
			newSubnet:       "10.0.0.0/8",
			expectedSubnets: []string{"10.0.0.0/8"},
		},
		{
			name:            "narrower subnet rejected when existing is wider",
			existing:        []string{"1.1.0.0/16"},
			newSubnet:       "1.1.1.2/24",
			expectError:     true,
			expectedSubnets: []string{"1.1.0.0/16"},
		},
		{
			name:            "smaller subnet rejected when existing is wider",
			existing:        []string{"10.0.0.0/8"},
			newSubnet:       "10.1.0.0/16",
			expectError:     true,
			expectedSubnets: []string{"10.0.0.0/8"},
		},
		{
			name:            "duplicate subnet rejected",
			existing:        []string{"10.0.0.0/8"},
			newSubnet:       "10.0.0.0/8",
			expectError:     true,
			expectedSubnets: []string{"10.0.0.0/8"},
		},
		{
			name:            "narrower subnet rejected among multiple existing",
			existing:        []string{"10.0.0.0/8", "172.16.0.0/12"},
			newSubnet:       "172.16.1.0/24",
			expectError:     true,
			expectedSubnets: []string{"10.0.0.0/8", "172.16.0.0/12"},
		},
		{
			name:            "ipv6 narrower rejected",
			existing:        []string{"fd00::/16"},
			newSubnet:       "fd00:1::/32",
			expectError:     true,
			expectedSubnets: []string{"fd00::/16"},
		},
		{
			name:            "single host rejected when subnet exists",
			existing:        []string{"10.0.0.0/8"},
			newSubnet:       "10.1.2.3/32",
			expectError:     true,
			expectedSubnets: []string{"10.0.0.0/8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAllowlist(nil, nil, tt.existing)
			err := a.addSubnet(tt.newSubnet, nil)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedSubnets, a.Subnets)
		})
	}
}

func TestNormalizeSubnets(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		subnets   []string
		removeCnt int
		expected  []string
	}{
		{
			name:     "empty list",
			subnets:  []string{},
			expected: nil,
		},
		{
			name:     "single subnet unchanged",
			subnets:  []string{"10.0.0.0/8"},
			expected: []string{"10.0.0.0/8"},
		},
		{
			name:     "non-overlapping subnets unchanged",
			subnets:  []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"},
			expected: []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"},
		},
		{
			name:      "narrower subnet removed when wider exists",
			subnets:   []string{"10.0.0.0/8", "10.1.0.0/16"},
			removeCnt: 1,
			expected:  []string{"10.0.0.0/8"},
		},
		{
			name:      "narrower listed first is removed",
			subnets:   []string{"10.1.0.0/16", "10.0.0.0/8"},
			removeCnt: 1,
			expected:  []string{"10.0.0.0/8"},
		},
		{
			name:      "multiple narrower subnets removed",
			subnets:   []string{"10.1.0.0/16", "10.2.0.0/16", "10.0.0.0/8"},
			removeCnt: 2,
			expected:  []string{"10.0.0.0/8"},
		},
		{
			name:      "mixed overlapping and non-overlapping",
			subnets:   []string{"10.1.0.0/16", "192.168.0.0/24", "10.0.0.0/8", "172.16.0.0/12"},
			removeCnt: 1,
			expected:  []string{"192.168.0.0/24", "10.0.0.0/8", "172.16.0.0/12"},
		},
		{
			name:      "duplicate subnets reduced to one",
			subnets:   []string{"10.0.0.0/8", "10.0.0.0/8"},
			removeCnt: 1,
			expected:  []string{"10.0.0.0/8"},
		},
		{
			name:      "three duplicates reduced to one",
			subnets:   []string{"10.0.0.0/8", "10.0.0.0/8", "10.0.0.0/8"},
			removeCnt: 2,
			expected:  []string{"10.0.0.0/8"},
		},
		{
			name:      "host covered by subnet removed",
			subnets:   []string{"10.1.2.3/32", "10.0.0.0/8"},
			removeCnt: 1,
			expected:  []string{"10.0.0.0/8"},
		},
		{
			name:      "deeply nested subnets all removed",
			subnets:   []string{"10.0.0.0/8", "10.1.0.0/16", "10.1.1.0/24", "10.1.1.1/32"},
			removeCnt: 3,
			expected:  []string{"10.0.0.0/8"},
		},
		{
			name:      "two independent overlapping groups",
			subnets:   []string{"10.1.0.0/16", "10.0.0.0/8", "192.168.1.0/24", "192.168.0.0/16"},
			removeCnt: 2,
			expected:  []string{"10.0.0.0/8", "192.168.0.0/16"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeCnt := 0
			a := NewAllowlist(nil, nil, tt.subnets)
			a.NormalizeSubnets(func(_, _ string) { removeCnt++ })
			assert.Equal(t, tt.removeCnt, removeCnt)
			assert.Equal(t, tt.expected, a.Subnets)
		})
	}
}

func TestNormalizeSubnetsInvalidInput(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		subnets   []string
		removeCnt int
	}{
		{
			name:      "single invalid entry",
			subnets:   []string{"not-a-subnet"},
			removeCnt: 1,
		},
		{
			name:      "invalid entry among valid ones",
			subnets:   []string{"10.0.0.0/8", "garbage", "192.168.0.0/16"},
			removeCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeCnt := 0
			a := NewAllowlist(nil, nil, tt.subnets)
			a.NormalizeSubnets(func(_, _ string) { removeCnt++ })
			assert.Equal(t, tt.removeCnt, removeCnt)
		})
	}
}

func TestSubnetsCoveredBy(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name               string
		existing           []string
		newSubnet          string
		expectedEliminated []string
	}{
		{
			name:               "no existing subnets",
			existing:           []string{},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{},
		},
		{
			name:               "no overlap",
			existing:           []string{"192.168.0.0/16"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{},
		},
		{
			name:               "new subnet is narrower than existing",
			existing:           []string{"10.0.0.0/8"},
			newSubnet:          "10.1.0.0/16",
			expectedEliminated: []string{},
		},
		{
			name:               "new subnet equals existing",
			existing:           []string{"10.0.0.0/8"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{},
		},
		{
			name:               "new subnet is wider and eliminates existing",
			existing:           []string{"10.1.0.0/16"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{"10.1.0.0/16"},
		},
		{
			name:               "new subnet eliminates multiple existing",
			existing:           []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/24"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/24"},
		},
		{
			name:               "new subnet eliminates some but not others",
			existing:           []string{"10.1.0.0/16", "192.168.0.0/24"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{"10.1.0.0/16"},
		},
		{
			name:               "new subnet eliminates single host",
			existing:           []string{"10.1.2.3/32"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{"10.1.2.3/32"},
		},
		{
			name:               "new subnet eliminates deeply nested",
			existing:           []string{"10.1.0.0/16", "10.1.1.0/24", "10.1.1.1/32"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{"10.1.0.0/16", "10.1.1.0/24", "10.1.1.1/32"},
		},
		{
			name:               "does not modify existing subnets",
			existing:           []string{"10.1.0.0/16", "192.168.0.0/24"},
			newSubnet:          "10.0.0.0/8",
			expectedEliminated: []string{"10.1.0.0/16"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAllowlist(nil, nil, tt.existing)
			existingCopy := make([]string, len(tt.existing))
			copy(existingCopy, tt.existing)

			eliminated, err := a.SubnetsCoveredBy(tt.newSubnet)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedEliminated, eliminated)
			assert.Equal(t, existingCopy, a.Subnets, "existing subnets must not be modified")
		})
	}
}

func TestWouldEliminateSubnetsInvalidInput(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		existing  []string
		newSubnet string
	}{
		{
			name:      "invalid new subnet",
			existing:  []string{"10.0.0.0/8"},
			newSubnet: "not-a-subnet",
		},
		{
			name:      "invalid existing subnet with overlapping new subnet",
			existing:  []string{"garbage"},
			newSubnet: "10.0.0.0/8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAllowlist(nil, nil, tt.existing)
			eliminated, err := a.SubnetsCoveredBy(tt.newSubnet)
			assert.Error(t, err)
			assert.Nil(t, eliminated)
		})
	}
}

func TestAddSubnetInvalidInput(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		existing  []string
		newSubnet string
	}{
		{
			name:      "invalid new subnet",
			existing:  []string{"10.0.0.0/8"},
			newSubnet: "not-a-subnet",
		},
		{
			name:      "invalid existing subnet",
			existing:  []string{"garbage"},
			newSubnet: "10.0.0.0/8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAllowlist(nil, nil, tt.existing)
			err := a.addSubnet(tt.newSubnet, nil)
			assert.Error(t, err)
		})
	}
}
