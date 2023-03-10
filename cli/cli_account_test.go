package cli

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestOrdinal(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		day  int
		name string
	}{
		{
			day:  1,
			name: "1st",
		},
		{
			day:  21,
			name: "21st",
		},
		{
			day:  31,
			name: "31st",
		},
		{
			day:  2,
			name: "2nd",
		},
		{
			day:  22,
			name: "22nd",
		},
		{
			day:  3,
			name: "3rd",
		},
		{
			day:  23,
			name: "23rd",
		},
		{
			day:  15,
			name: "15th",
		},
		{
			day:  17,
			name: "17th",
		},
		{
			day:  24,
			name: "24th",
		},
	}

	for _, test := range tests {
		got := ordinal(test.day)
		assert.Equal(t, got, test.name)
	}
}

func TestActiveBoolToString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		isActive bool
		result   string
	}{
		{
			isActive: true,
			result:   "Active",
		},
		{
			isActive: false,
			result:   "Inactive",
		},
	}

	for _, test := range tests {
		got := activeBoolToString(test.isActive)
		assert.Equal(t, got, test.result)
	}
}
