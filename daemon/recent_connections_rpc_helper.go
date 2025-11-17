package daemon

import (
	"fmt"
	"slices"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// StorePendingRecentConnection stores current pending recent connection to memory
func StorePendingRecentConnection(
	store *recents.RecentConnectionsStore,
	eventPublisher func(events.DataRecentsChanged)) {
	exists, recentModel := store.PopPending()
	if !exists {
		return
	}

	if store.Add(recentModel) == nil {
		eventPublisher(events.DataRecentsChanged{})
	}
}

// isRecentConnectionSupported returns true if server connection can be used for reconnection,
// otherwise returns false
func isRecentConnectionSupported(rule config.ServerSelectionRule) bool {
	return rule != config.ServerSelectionRule_RECOMMENDED && rule != config.ServerSelectionRule_NONE
}

// applyObfuscationToRecentModel wraps the recent connection model with obfuscation handling.
func applyObfuscationToRecentModel(
	model recents.Model,
	parameters ServerParameters,
	event events.DataConnect,
	server *core.Server,
) recents.Model {
	if !event.IsObfuscated {
		return model
	}

	if parameters.Group != config.ServerGroup_UNDEFINED {
		return model
	}

	hasObfuscatedGroup := slices.ContainsFunc(server.Groups, func(g core.Group) bool {
		return g.ID == config.ServerGroup_OBFUSCATED
	})
	if !hasObfuscatedGroup {
		return model
	}

	switch model.ConnectionType {
	case config.ServerSelectionRule_CITY, config.ServerSelectionRule_COUNTRY:
		model.ConnectionType = config.ServerSelectionRule_COUNTRY_WITH_GROUP
		model.Group = config.ServerGroup_OBFUSCATED
	case config.ServerSelectionRule_SPECIFIC_SERVER:
		model.ConnectionType = config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP
		model.Group = config.ServerGroup_OBFUSCATED
	case config.ServerSelectionRule_NONE,
		config.ServerSelectionRule_RECOMMENDED,
		config.ServerSelectionRule_GROUP,
		config.ServerSelectionRule_COUNTRY_WITH_GROUP,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
	}

	return model
}

// extractSpecificServerName extracts the specific server name from the domain
func extractSpecificServerName(domain string) string {
	var name string
	if domain != "" {
		parts := strings.Split(domain, ".")
		if len(parts) > 0 && parts[0] != "" {
			name = parts[0]
		}
	}
	return name
}

// extractServerTechnologies extracts technology IDs from server technologies
func extractServerTechnologies(server *core.Server) []core.ServerTechnology {
	var serverTechs []core.ServerTechnology
	for _, v := range server.Technologies {
		serverTechs = append(serverTechs, v.ID)
	}
	return serverTechs
}

// isSingleCityCountry checks if this is a single-city country by looking up the country in the
// countries list
func isSingleCityCountry(country string, countries core.Countries) bool {
	if country == "" {
		return false
	}

	for _, c := range countries {
		if c.Name == country || c.Code == country {
			return len(c.Cities) == 1
		}
	}
	return false
}

// buildRecentConnectionModel creates a recent connection model from a successful VPN connection
// event.
func buildRecentConnectionModel(
	event events.DataConnect,
	parameters ServerParameters,
	server *core.Server,
	countries core.Countries,
) (recents.Model, error) {
	serverTechs := extractServerTechnologies(server)
	connectionType := event.TargetServerSelection
	cityForRecent := event.TargetServerCity

	// Normalize connection type for single-city countries to avoid duplicate recent entries.
	// Problem: CLI connects using COUNTRY type (no city specified), while GUI uses CITY type.
	// Solution: For single-city countries, always normalize to CITY type with the city populated.
	// This ensures CLI "nordvpn connect Greece" and GUI "connect to Athens" create the same recent entry.
	singleCityCountry := isSingleCityCountry(parameters.Country, countries)
	if singleCityCountry {
		// For single-city countries, use the city from the server event for consistency
		if connectionType == config.ServerSelectionRule_COUNTRY {
			connectionType = config.ServerSelectionRule_CITY
			cityForRecent = event.TargetServerCity
		}
		// For specialty group connections (e.g., "nordvpn connect Greece Athens --group p2p"),
		// also populate the city to match GUI behavior
		if connectionType == config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP {
			cityForRecent = event.TargetServerCity
		}

		if connectionType == config.ServerSelectionRule_COUNTRY_WITH_GROUP {
			connectionType = config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP
			cityForRecent = event.TargetServerCity
		}
	}

	recentModel := recents.Model{
		ConnectionType:     connectionType,
		ServerTechnologies: serverTechs,
		IsVirtual:          event.IsVirtualLocation,
	}

	// Populate model fields based on connection type
	switch recentModel.ConnectionType {
	case config.ServerSelectionRule_GROUP:
		recentModel.Group = parameters.Group

	case config.ServerSelectionRule_CITY:
		// City-level connection (e.g., "nordvpn connect New_York" or GUI city selection)
		recentModel.City = event.TargetServerCity
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_COUNTRY:
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		recentModel.Group = parameters.Group
		// recentModel.City = cityForRecent
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_SPECIFIC_SERVER:
		recentModel.SpecificServer = extractSpecificServerName(event.TargetServerDomain)
		recentModel.SpecificServerName = event.TargetServerName
		recentModel.City = event.TargetServerCity
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		// Note: We treat this as a city-level connection within the specialty group,
		// not as a specific server connection.
		// The specific server fields are intentionally left empty.
		recentModel.Group = parameters.Group
		recentModel.City = cityForRecent
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_NONE, config.ServerSelectionRule_RECOMMENDED:
		// These connection types should not create recent entries
		return recents.Model{}, fmt.Errorf("unexpected connection type in recent connections: %d", recentModel.ConnectionType)
	}

	// Apply obfuscation wrapping if needed (upgrades connection type when obfuscation is active)
	recentModel = applyObfuscationToRecentModel(recentModel, parameters, event, server)
	return recentModel, nil
}
