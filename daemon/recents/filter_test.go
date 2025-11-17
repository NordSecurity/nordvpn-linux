package recents

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/stretchr/testify/assert"
)

func TestFilter_Apply_TechnologySuperset(t *testing.T) {
	target := Model{
		Country:            "Australia",
		City:               "Sydney",
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        "AU",
		SpecificServer:     "",
		SpecificServerName: "",
		ConnectionType:     config.ServerSelectionRule_CITY,
		ServerTechnologies: []core.ServerTechnology{1, 3, 5},
	}

	candidates := []Model{
		{
			Country:            "Australia",
			City:               "Sydney",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "AU",
			SpecificServer:     "",
			SpecificServerName: "",
			ConnectionType:     config.ServerSelectionRule_CITY,
			ServerTechnologies: []core.ServerTechnology{1, 3, 5, 21, 23},
		},
		{
			Country:            "Australia",
			City:               "Melbourne",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "AU",
			SpecificServer:     "",
			SpecificServerName: "",
			ConnectionType:     config.ServerSelectionRule_CITY,
			ServerTechnologies: []core.ServerTechnology{1, 3, 5, 21, 23},
		},
	}

	filter := NewFilter(target, candidates)
	filter.WithoutSpecificServerFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithTechnologies(target.ServerTechnologies)

	result := filter.Apply()

	if assert.Len(t, result, 1, "expected 1 match") {
		assert.Equal(t, "Sydney", result[0].City, "expected Sydney match")
		assert.Len(t, result[0].ServerTechnologies, 5, "expected match with 5 technologies")
	}
}

func TestFilter_Apply_ExactMatch(t *testing.T) {
	target := Model{
		Country:            "Germany",
		City:               "Berlin",
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        "DE",
		SpecificServer:     "",
		SpecificServerName: "",
		ConnectionType:     config.ServerSelectionRule_CITY,
		ServerTechnologies: []core.ServerTechnology{1, 3, 5},
	}

	candidates := []Model{
		{
			Country:            "Germany",
			City:               "Berlin",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "DE",
			SpecificServer:     "",
			SpecificServerName: "",
			ConnectionType:     config.ServerSelectionRule_CITY,
			ServerTechnologies: []core.ServerTechnology{1, 3, 5},
		},
	}

	filter := NewFilter(target, candidates)
	filter.WithoutSpecificServerFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithTechnologies(target.ServerTechnologies)

	result := filter.Apply()

	assert.Len(t, result, 1, "expected 1 match")
}

func TestFilter_Apply_NoMatch_MissingTechnology(t *testing.T) {
	target := Model{
		Country:            "France",
		City:               "Paris",
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        "FR",
		SpecificServer:     "",
		SpecificServerName: "",
		ConnectionType:     config.ServerSelectionRule_CITY,
		ServerTechnologies: []core.ServerTechnology{1, 3, 5, 21},
	}

	candidates := []Model{
		{
			Country:            "France",
			City:               "Paris",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "FR",
			SpecificServer:     "",
			SpecificServerName: "",
			ConnectionType:     config.ServerSelectionRule_CITY,
			ServerTechnologies: []core.ServerTechnology{1, 3, 5},
		},
	}

	filter := NewFilter(target, candidates)
	filter.WithoutSpecificServerFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithTechnologies(target.ServerTechnologies)

	result := filter.Apply()

	assert.Empty(t, result, "expected 0 matches")
}

func TestFilter_Apply_SpecificServer_MatchesOnlyExactServer(t *testing.T) {
	target := Model{
		Country:            "USA",
		City:               "New York",
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        "US",
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
		SpecificServer:     "us1234",
		SpecificServerName: "United States #1234",
		ServerTechnologies: []core.ServerTechnology{1, 3},
	}

	candidates := []Model{
		{
			Country:            "USA",
			City:               "New York",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "US",
			ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
			SpecificServer:     "us1234",
			SpecificServerName: "United States #1234",
			ServerTechnologies: []core.ServerTechnology{1, 3, 5},
		},
		{
			Country:            "USA",
			City:               "New York",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "US",
			ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
			SpecificServer:     "us5678",
			SpecificServerName: "United States #5678",
			ServerTechnologies: []core.ServerTechnology{1, 3, 5},
		},
	}

	// For SPECIFIC_SERVER, we should NOT exclude specific server fields
	filter := NewFilter(target, candidates)
	filter.WithoutSpecificServerFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithTechnologies(target.ServerTechnologies)

	result := filter.Apply()

	assert.Len(t, result, 1, "expected 1 match")
	if len(result) > 0 {
		assert.Equal(t, "us1234", result[0].SpecificServer, "expected us1234")
	}
}

func TestFilter_Apply_City_IgnoresSpecificServer(t *testing.T) {
	target := Model{
		Country:            "UK",
		City:               "London",
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        "GB",
		ConnectionType:     config.ServerSelectionRule_CITY,
		SpecificServer:     "",
		SpecificServerName: "",
		ServerTechnologies: []core.ServerTechnology{1, 3},
	}

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
		},
	}

	filter := NewFilter(target, candidates)
	filter.WithoutSpecificServerFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithTechnologies(target.ServerTechnologies)

	result := filter.Apply()

	// Both should match because specific server fields are excluded for CITY connections
	assert.Len(t, result, 2, "expected 2 matches - specific server fields should be excluded for CITY connections")
}

func TestFilter_Apply_WithoutTechnologies_IgnoresTechnologies(t *testing.T) {
	target := Model{
		Country:            "Canada",
		City:               "Toronto",
		Group:              config.ServerGroup_UNDEFINED,
		CountryCode:        "CA",
		ConnectionType:     config.ServerSelectionRule_CITY,
		SpecificServer:     "",
		SpecificServerName: "",
		ServerTechnologies: []core.ServerTechnology{1, 3, 5},
	}

	candidates := []Model{
		{
			Country:            "Canada",
			City:               "Toronto",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "CA",
			ConnectionType:     config.ServerSelectionRule_CITY,
			SpecificServer:     "",
			SpecificServerName: "",
			ServerTechnologies: []core.ServerTechnology{1, 3, 5, 21, 23, 35},
		},
		{
			Country:            "Canada",
			City:               "Toronto",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "CA",
			ConnectionType:     config.ServerSelectionRule_CITY,
			SpecificServer:     "",
			SpecificServerName: "",
			ServerTechnologies: []core.ServerTechnology{7, 9},
		},
		{
			Country:            "Canada",
			City:               "Montreal",
			Group:              config.ServerGroup_UNDEFINED,
			CountryCode:        "CA",
			ConnectionType:     config.ServerSelectionRule_CITY,
			SpecificServer:     "",
			SpecificServerName: "",
			ServerTechnologies: []core.ServerTechnology{1, 3, 5},
		},
	}

	filter := NewFilter(target, candidates)
	filter.WithoutSpecificServerFor([]config.ServerSelectionRule{
		config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	})
	filter.WithoutTechnologies()

	result := filter.Apply()

	// Both Toronto entries should match regardless of technologies
	assert.Len(t, result, 2, "expected 2 matches - technologies should be ignored")
	for _, r := range result {
		assert.Equal(t, "Toronto", r.City, "all matches should be Toronto")
	}
}
