package recents

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

// Helper to create a basic city connection model
func cityModel(country, city, countryCode string) Model {
	return Model{
		Country:            country,
		City:               city,
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        countryCode,
		SpecificServer:     "",
		SpecificServerName: "",
		ConnectionType:     config.ServerSelectionRule_CITY,
		IsVirtual:          false,
		ConnectionTech:     config.Technology_NORDLYNX,
	}
}

// Helper to create a specific server connection model
func specificServerModel(country, city, countryCode, server, serverName string) Model {
	return Model{
		Country:            country,
		City:               city,
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        countryCode,
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
		SpecificServer:     server,
		SpecificServerName: serverName,
		IsVirtual:          false,
		ConnectionTech:     config.Technology_NORDLYNX,
	}
}

func TestFilter_Apply_SpecificServer(t *testing.T) {
	category.Set(t, category.Unit)

	target := specificServerModel("USA", "New York", "US", "us1234", "United States #1234")
	candidates := []Model{
		specificServerModel("USA", "New York", "US", "us1234", "United States #1234"),
		specificServerModel("USA", "New York", "US", "us5678", "United States #5678"),
	}

	filter := newFilter(target, candidates)
	filter.withSpecificServerOnlyFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})

	result := filter.apply()

	assert.Len(t, result, 1, "expected 1 match")
	assert.Equal(t, "us1234", result[0].SpecificServer)
}

func TestFilter_Apply_City_IgnoresSpecificServer(t *testing.T) {
	category.Set(t, category.Unit)

	target := cityModel("UK", "London", "GB")
	candidates := []Model{
		{
			Country:            "UK",
			City:               "London",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "GB",
			ConnectionType:     config.ServerSelectionRule_CITY,
			SpecificServer:     "uk123",
			SpecificServerName: "United Kingdom #123",
			IsVirtual:          false,
		},
		{
			Country:            "UK",
			City:               "London",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "GB",
			ConnectionType:     config.ServerSelectionRule_CITY,
			SpecificServer:     "uk456",
			SpecificServerName: "United Kingdom #456",
			IsVirtual:          false,
		},
	}

	filter := newFilter(target, candidates)
	filter.withSpecificServerOnlyFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})

	result := filter.apply()

	assert.Len(t, result, 2, "expected 2 matches - specific server fields should be excluded for CITY connections")
}

func TestMatches(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		target         Model
		candidates     []Model
		expectedResult []Model
	}{
		{
			name:           "country match",
			target:         Model{Country: "Germany"},
			candidates:     []Model{{Country: "Germany"}, {Country: "France"}},
			expectedResult: []Model{{Country: "Germany"}},
		},
		{
			name:           "city match",
			target:         Model{Country: "Germany", City: "Berlin"},
			candidates:     []Model{{Country: "Germany", City: "Berlin"}, {Country: "Germany", City: "Munich"}},
			expectedResult: []Model{{Country: "Germany", City: "Berlin"}},
		},
		{
			name:   "group match",
			target: Model{Country: "Germany", City: "Berlin", Group: config.ServerGroup_STANDARD_VPN_SERVERS},
			candidates: []Model{
				{Country: "Germany", City: "Berlin", Group: config.ServerGroup_STANDARD_VPN_SERVERS},
				{Country: "Germany", City: "Munich", Group: config.ServerGroup_STANDARD_VPN_SERVERS},
			},
			expectedResult: []Model{{Country: "Germany", City: "Berlin", Group: config.ServerGroup_STANDARD_VPN_SERVERS}},
		},
		{
			name:   "server selection match",
			target: Model{Country: "Germany", City: "Berlin", Group: config.ServerGroup_STANDARD_VPN_SERVERS, ConnectionType: config.ServerSelectionRule_CITY},
			candidates: []Model{
				{Country: "Germany", City: "Berlin", Group: config.ServerGroup_STANDARD_VPN_SERVERS, ConnectionType: config.ServerSelectionRule_CITY},
				{Country: "Germany", City: "Berlin", Group: config.ServerGroup_STANDARD_VPN_SERVERS, ConnectionType: config.ServerSelectionRule_COUNTRY},
			},
			expectedResult: []Model{
				{Country: "Germany", City: "Berlin", Group: config.ServerGroup_STANDARD_VPN_SERVERS, ConnectionType: config.ServerSelectionRule_CITY},
			},
		},
		{
			name: "connection technology match, not NordWisper",
			target: Model{
				Country:        "Germany",
				City:           "Berlin",
				Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
				ConnectionType: config.ServerSelectionRule_CITY,
				ConnectionTech: config.Technology_NORDLYNX,
			},
			candidates: []Model{
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_NORDLYNX,
				},
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_OPENVPN,
				},
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
			expectedResult: []Model{
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_NORDLYNX,
				},
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_OPENVPN,
				},
			},
		},
		{
			name: "NordWisper connection technology match",
			target: Model{
				Country:        "Germany",
				City:           "Berlin",
				Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
				ConnectionType: config.ServerSelectionRule_CITY,
				ConnectionTech: config.Technology_NORDWHISPER,
			},
			candidates: []Model{
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_NORDLYNX,
				},
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_OPENVPN,
				},
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
			expectedResult: []Model{
				{
					Country:        "Germany",
					City:           "Berlin",
					Group:          config.ServerGroup_STANDARD_VPN_SERVERS,
					ConnectionType: config.ServerSelectionRule_CITY,
					ConnectionTech: config.Technology_NORDWHISPER,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := newFilter(tt.target, tt.candidates)
			result := filter.apply()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
