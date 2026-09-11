package recents

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestRecentConnectionsEncode(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		connections    []Model
		expectedOutput string
		expectedError  error
	}{
		{
			name:           "no connections",
			connections:    []Model{},
			expectedOutput: `{"version":1,"connections":[]}`,
		},
		{
			name: "encode connection",
			connections: []Model{
				{
					Country:            "Germany",
					CountryCode:        "DE",
					City:               "Berlin",
					Group:              config.ServerGroup_STANDARD_VPN_SERVERS,
					SpecificServerName: "de123",
					ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
					ConnectionTech:     config.Technology_NORDLYNX,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
			expectedOutput: `{"version":1,"connections":[{"country":"Germany","city":"Berlin","group":11,"country-code":"DE","specific-server-name":"de123","connection-type":6,"connection-tech":2},{"country":"Lithuania","country-code":"LT","connection-type":3,"connection-tech":3}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := encode(test.connections)
			assert.Equal(t, test.expectedError, err)
			assert.JSONEq(t, test.expectedOutput, string(data))
			assert.Equal(t, test.expectedOutput, string(data))
		})
	}
}

func TestRecentConnectionsDecode(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name               string
		data               string
		expectedOutput     []Model
		expectedFileFormat int
		expectedError      error
	}{
		{
			name:               "data input is empty",
			data:               ``,
			expectedFileFormat: 1,
			expectedOutput:     []Model{},
		},
		{
			name:               "no connections",
			expectedFileFormat: 1,
			data:               `{"version":1,"connections":[]}`,
			expectedOutput:     []Model{},
		},
		{
			name:               "decode 2 valid connections",
			expectedFileFormat: 1,
			data:               `{"version":1,"connections":[{"country":"Germany","city":"Berlin","group":11,"country-code":"DE","specific-server-name":"de123","connection-type":6,"connection-tech":2},{"country":"Lithuania","country-code":"LT","connection-type":3,"connection-tech":3}]}`,
			expectedOutput: []Model{
				{
					Country:            "Germany",
					CountryCode:        "DE",
					City:               "Berlin",
					Group:              config.ServerGroup_STANDARD_VPN_SERVERS,
					SpecificServerName: "de123",
					ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
					ConnectionTech:     config.Technology_NORDLYNX,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
		},
		{
			name:               "decode older format",
			expectedFileFormat: 0,
			data:               `[{"country":"Germany","city":"Berlin","group":11,"country-code":"DE","specific-server-name":"de123","connection-type":6,"connection-tech":2},{"country":"Lithuania","country-code":"LT","connection-type":3,"connection-tech":3}]`,
			expectedOutput: []Model{
				{
					Country:            "Germany",
					CountryCode:        "DE",
					City:               "Berlin",
					Group:              config.ServerGroup_STANDARD_VPN_SERVERS,
					SpecificServerName: "de123",
					ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
					ConnectionTech:     config.Technology_NORDLYNX,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := decode([]byte(test.data))
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedFileFormat, f.Version)
			assert.Equal(t, test.expectedOutput, f.Connections)
		})
	}
}

func TestRecentConnectionsFileMigration(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		connections    []Model
		expectedOutput []Model
	}{
		{
			name:           "nil input",
			connections:    nil,
			expectedOutput: []Model{},
		},
		{
			name: "set connection type to nordlynx for Technology_UNKNOWN_TECHNOLOGY",
			connections: []Model{
				{
					Country:            "Germany",
					CountryCode:        "DE",
					City:               "Berlin",
					Group:              config.ServerGroup_STANDARD_VPN_SERVERS,
					SpecificServerName: "de123",
					ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
					ConnectionTech:     config.Technology_UNKNOWN_TECHNOLOGY,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
			expectedOutput: []Model{
				{
					Country:            "Germany",
					CountryCode:        "DE",
					City:               "Berlin",
					Group:              config.ServerGroup_STANDARD_VPN_SERVERS,
					SpecificServerName: "de123",
					ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
					ConnectionTech:     config.Technology_NORDLYNX,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
		},
		{
			name: "obfuscated group is deleted",
			connections: []Model{
				{
					Country:            "Germany",
					CountryCode:        "DE",
					City:               "Berlin",
					Group:              config.ServerGroup_STANDARD_VPN_SERVERS,
					SpecificServerName: "de123",
					ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
					ConnectionTech:     config.Technology_NORDLYNX,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					Group:          config.ServerGroup_OBFUSCATED,
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
			expectedOutput: []Model{
				{
					Country:            "Germany",
					CountryCode:        "DE",
					City:               "Berlin",
					Group:              config.ServerGroup_STANDARD_VPN_SERVERS,
					SpecificServerName: "de123",
					ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
					ConnectionTech:     config.Technology_NORDLYNX,
				},
				{
					Country:        "Lithuania",
					CountryCode:    "LT",
					ConnectionType: config.ServerSelectionRule_COUNTRY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := migrateToVersion1(test.connections)
			assert.Equal(t, currentFile, f.Version)
			assert.Equal(t, test.expectedOutput, f.Connections)
		})
	}
}
