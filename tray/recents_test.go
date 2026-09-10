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
		// obfuscated tech
		{
			name: "NordWhisper with ConnectionType city",
			conn: RecentConnection{
				Country:        "Lithuania",
				City:           "Vilnius",
				ConnectionTech: config.Technology_NORDWHISPER,
				ConnectionType: config.ServerSelectionRule_CITY,
			},
			expectedLabel: "Obfuscated Servers (Lithuania, Vilnius)",
		},
		{
			name: "NordWhisper with ConnectionType country",
			conn: RecentConnection{
				Country:        "Lithuania",
				ConnectionType: config.ServerSelectionRule_COUNTRY,
				ConnectionTech: config.Technology_NORDWHISPER,
			},
			expectedLabel: "Obfuscated Servers (Lithuania)",
		},
		{
			name: "NordWhisper with ConnectionType specific server",
			conn: RecentConnection{
				SpecificServerName: "lt1234",
				ConnectionTech:     config.Technology_NORDWHISPER,
				ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
			},
			expectedLabel: "Obfuscated Servers (lt1234)",
		},
		{
			name: "NordWhisper with ConnectionType group, not obfuscate group",
			conn: RecentConnection{
				Group:          config.ServerGroup_DOUBLE_VPN,
				ConnectionType: config.ServerSelectionRule_GROUP,
				ConnectionTech: config.Technology_NORDWHISPER,
			},
			expectedLabel: "",
		},
		{
			name: "NordWhisper with ConnectionType country and obfuscate group",
			conn: RecentConnection{
				Country:        "Lithuania",
				Group:          config.ServerGroup_OBFUSCATED,
				ConnectionTech: config.Technology_NORDWHISPER,
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
