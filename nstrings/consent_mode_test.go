package nstrings

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config/consent"
	"gotest.tools/v3/assert"
)

func TestUserConsent_ReturnsProperLabel(t *testing.T) {
	tests := []struct {
		name     string
		mode     consent.ConsentMode
		expected string
	}{
		{
			name:     "ConsentMode_DENIED returns disabled",
			mode:     consent.ConsentMode_DENIED,
			expected: disabled,
		},
		{
			name:     "ConsentMode_GRANTED returns enabled",
			mode:     consent.ConsentMode_GRANTED,
			expected: enabled,
		},
		{
			name:     "ConsentMode_UNDEFINED returns undefined",
			mode:     consent.ConsentMode_UNDEFINED,
			expected: undefined,
		},
		{
			name:     "Unknown mode returns undefined",
			mode:     consent.ConsentMode(999),
			expected: undefined,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UserConsent(tt.mode)
			assert.Equal(t, result, tt.expected, "UserConsent(%v) = %q; want %q", tt.mode, result, tt.expected)
		})
	}
}
