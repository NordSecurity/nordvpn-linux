package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestExtractZone(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    string
		expected string
	}{
		{
			input: `Timezone=Europe/Vilnius
LocalRTC=no
CanNTP=yes
NTP=yes
NTPSynchronized=yes
TimeUSec=Wed 2020-05-06 17:06:01 EEST
RTCTimeUSec=Wed 2020-05-06 17:06:01 EEST
`,
			expected: "Europe/Vilnius",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		zone := extractZone([]byte(tt.input))
		assert.Equal(t, tt.expected, zone)
	}
}
