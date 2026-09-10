package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestRecentConnDisplayLabel(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		conn          RecentConnection
		expectedLabel string
	}{
		{
			name: "ConnectionType city",
			conn: RecentConnection{
				Country:        "Lithuania",
				City:           "Vilnius",
				ConnectionType: config.ServerSelectionRule_CITY,
			},
			expectedLabel: "Lithuania, Vilnius",
		},
		{
			name: "ConnectionType country",
			conn: RecentConnection{
				Country:        "Lithuania",
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			},
			expectedLabel: "Lithuania",
		},
		{
			name: "ConnectionType specific server",
			conn: RecentConnection{
				SpecificServerName: "lt1234",
				ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
			},
			expectedLabel: "lt1234",
		},
		{
			name: "ConnectionType group",
			conn: RecentConnection{
				Group:          config.ServerGroup_DOUBLE_VPN,
				ConnectionType: config.ServerSelectionRule_GROUP,
			},
			expectedLabel: "Double VPN",
		},
		{
			name: "ConnectionType country with group",
			conn: RecentConnection{
				Country:        "Lithuania",
				Group:          config.ServerGroup_DOUBLE_VPN,
				ConnectionType: config.ServerSelectionRule_COUNTRY_WITH_GROUP,
			},
			expectedLabel: "Double VPN (Lithuania)",
		},
		{
			name: "ConnectionType specific server with group, country and city",
			conn: RecentConnection{
				Country:        "Lithuania",
				City:           "Vilnius",
				Group:          config.ServerGroup_DOUBLE_VPN,
				ConnectionType: config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
			},
			expectedLabel: "Double VPN (Lithuania, Vilnius)",
		},
		{
			name: "ConnectionType specific server with group, country without city",
			conn: RecentConnection{
				Country:            "Lithuania",
				SpecificServerName: "lt1234",
				Group:              config.ServerGroup_ONION_OVER_VPN,
				ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
			},
			expectedLabel: "Onion Over VPN (Lithuania)",
		},
		{
			name: "Obfuscated with Country + City for ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP",
			conn: RecentConnection{
				Country:        "Lithuania",
				City:           "Vilnius",
				Group:          config.ServerGroup_OBFUSCATED,
				ConnectionType: config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
			},
			expectedLabel: "Obfuscated Servers (Lithuania, Vilnius)",
		},
		{
			name: "Obfuscated with Country only for ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP",
			conn: RecentConnection{
				Country:        "Lithuania",
				ConnectionType: config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
				Group:          config.ServerGroup_OBFUSCATED,
			},
			expectedLabel: "Obfuscated Servers (Lithuania)",
		},
		{
			name: "Obfuscated with ServerSelectionRule_COUNTRY_WITH_GROUP",
			conn: RecentConnection{
				Country:        "Lithuania",
				ConnectionType: config.ServerSelectionRule_COUNTRY_WITH_GROUP,
				Group:          config.ServerGroup_OBFUSCATED,
			},
			expectedLabel: "Obfuscated Servers (Lithuania)",
		},
		{
			name: "NordWhisper with ConnectionType country and obfuscate group",
			conn: RecentConnection{
				Country:        "Lithuania",
				Group:          config.ServerGroup_OBFUSCATED,
				ConnectionType: config.ServerSelectionRule_COUNTRY_WITH_GROUP,
			},
			expectedLabel: "Obfuscated Servers (Lithuania)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedLabel, makeDisplayLabel(&test.conn))
		})
	}
}
