package recents

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

// Helper to create a basic city connection model
func cityModel(country, city, countryCode string, techs []core.ServerTechnology) Model {
	return Model{
		Country:            country,
		City:               city,
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        countryCode,
		SpecificServer:     "",
		SpecificServerName: "",
		ConnectionType:     config.ServerSelectionRule_CITY,
		ServerTechnologies: techs,
		IsVirtual:          false,
	}
}

// Helper to create a specific server connection model
func specificServerModel(country, city, countryCode, server, serverName string, techs []core.ServerTechnology) Model {
	return Model{
		Country:            country,
		City:               city,
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        countryCode,
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
		SpecificServer:     server,
		SpecificServerName: serverName,
		ServerTechnologies: techs,
		IsVirtual:          false,
	}
}

func TestFilter_Apply_WithTechnologies(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		target        Model
		candidates    []Model
		expectedCount int
		expectedCity  string
	}{
		{
			name:   "TechnologySuperset",
			target: cityModel("Australia", "Sydney", "AU", []core.ServerTechnology{1, 3, 5}),
			candidates: []Model{
				cityModel("Australia", "Sydney", "AU", []core.ServerTechnology{1, 3, 5, 21, 23}),
				cityModel("Australia", "Melbourne", "AU", []core.ServerTechnology{1, 3, 5, 21, 23}),
			},
			expectedCount: 1,
			expectedCity:  "Sydney",
		},
		{
			name:   "ExactMatch",
			target: cityModel("Germany", "Berlin", "DE", []core.ServerTechnology{1, 3, 5}),
			candidates: []Model{
				cityModel("Germany", "Berlin", "DE", []core.ServerTechnology{1, 3, 5}),
			},
			expectedCount: 1,
		},
		{
			name:   "NoMatch_MissingTechnology",
			target: cityModel("France", "Paris", "FR", []core.ServerTechnology{1, 3, 5, 21}),
			candidates: []Model{
				cityModel("France", "Paris", "FR", []core.ServerTechnology{1, 3, 5}),
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.target, tt.candidates)
			filter.WithSpecificServerOnlyFor([]config.ServerSelectionRule{
				config.ServerSelectionRule_SPECIFIC_SERVER,
				config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
			})
			filter.WithTechnologies(tt.target.ServerTechnologies)

			result := filter.Apply()

			assert.Len(t, result, tt.expectedCount)
			if tt.expectedCount > 0 && tt.expectedCity != "" {
				assert.Equal(t, tt.expectedCity, result[0].City)
			}
		})
	}
}

func TestFilter_Apply_SpecificServer(t *testing.T) {
	category.Set(t, category.Unit)

	target := specificServerModel("USA", "New York", "US", "us1234", "United States #1234", []core.ServerTechnology{1, 3})
	candidates := []Model{
		specificServerModel("USA", "New York", "US", "us1234", "United States #1234", []core.ServerTechnology{1, 3, 5}),
		specificServerModel("USA", "New York", "US", "us5678", "United States #5678", []core.ServerTechnology{1, 3, 5}),
	}

	filter := NewFilter(target, candidates)
	filter.WithSpecificServerOnlyFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithTechnologies(target.ServerTechnologies)

	result := filter.Apply()

	assert.Len(t, result, 1, "expected 1 match")
	assert.Equal(t, "us1234", result[0].SpecificServer)
}

func TestFilter_Apply_City_IgnoresSpecificServer(t *testing.T) {
	category.Set(t, category.Unit)

	target := cityModel("UK", "London", "GB", []core.ServerTechnology{1, 3})
	candidates := []Model{
		{
			Country:            "UK",
			City:               "London",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "GB",
			ConnectionType:     config.ServerSelectionRule_CITY,
			SpecificServer:     "uk123",
			SpecificServerName: "United Kingdom #123",
			ServerTechnologies: []core.ServerTechnology{1, 3, 5},
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
			ServerTechnologies: []core.ServerTechnology{1, 3},
			IsVirtual:          false,
		},
	}

	filter := NewFilter(target, candidates)
	filter.WithSpecificServerOnlyFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithTechnologies(target.ServerTechnologies)

	result := filter.Apply()

	assert.Len(t, result, 2, "expected 2 matches - specific server fields should be excluded for CITY connections")
}

func TestFilter_Apply_WithoutTechnologies(t *testing.T) {
	category.Set(t, category.Unit)

	target := cityModel("Canada", "Toronto", "CA", []core.ServerTechnology{1, 3, 5})
	candidates := []Model{
		cityModel("Canada", "Toronto", "CA", []core.ServerTechnology{1, 3, 5, 21, 23, 35}),
		cityModel("Canada", "Toronto", "CA", []core.ServerTechnology{7, 9}),
		cityModel("Canada", "Montreal", "CA", []core.ServerTechnology{1, 3, 5}),
	}

	filter := NewFilter(target, candidates)
	filter.WithSpecificServerOnlyFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithoutTechnologies()

	result := filter.Apply()

	assert.Len(t, result, 2, "expected 2 matches - technologies should be ignored")
	for _, r := range result {
		assert.Equal(t, "Toronto", r.City)
	}
}
