package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestMakeDisplayLabel_SpecificServerWithGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		conn     RecentConnection
		expected string
	}{
		{
			name:     "country and city",
			conn:     RecentConnection{Group: config.ServerGroup_DOUBLE_VPN, Country: "Lithuania", City: "Vilnius"},
			expected: "Double VPN (Lithuania, Vilnius)",
		},
		{
			name:     "country only",
			conn:     RecentConnection{Group: config.ServerGroup_OBFUSCATED, Country: "Lithuania"},
			expected: "Obfuscated Servers (Lithuania)",
		},
		{
			name:     "no country",
			conn:     RecentConnection{Group: config.ServerGroup_DOUBLE_VPN, City: "Vilnius"},
			expected: "",
		},
		{
			name:     "undefined group",
			conn:     RecentConnection{Group: config.ServerGroup_UNDEFINED, Country: "Lithuania", City: "Vilnius"},
			expected: "",
		},
		{
			name:     "group without label",
			conn:     RecentConnection{Group: 19, Country: "Lithuania", City: "Vilnius"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.conn.ConnectionType = config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP
			assert.Equal(t, tt.expected, makeDisplayLabel(&tt.conn))
		})
	}
}
