package auth

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestIsTokenExpired(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "",
			expected: true,
		},
		{
			input:    "1990-01-01 09:18:53",
			expected: true,
		},
		{
			input:    "2990-01-01 09:18:53",
			expected: false,
		},
		{
			input:    "Wed Sep 18 09:27:12 UTC 2019",
			expected: true,
		},
	}

	for _, tt := range tests {
		got := IsTokenExpired(tt.input)
		assert.Equal(t, tt.expected, got)
	}
}
