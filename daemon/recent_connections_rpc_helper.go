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
	event events.DataConnect,
	server *core.Server,
	dm *DataManager,
	cfg config.Config,
) recents.Model {
	if !event.IsObfuscated {
		return model
	}

	hasObfuscatedGroup := slices.ContainsFunc(server.Groups, func(g core.Group) bool {
		return g.ID == config.ServerGroup_OBFUSCATED
	})
	if !hasObfuscatedGroup {
		return model
	}

	// Set group to OBFUSCATED for all obfuscated connections
	model.Group = config.ServerGroup_OBFUSCATED

	// Determine if city should be included based on connection type
	cityExplicitlySpecified := model.ConnectionType == config.ServerSelectionRule_CITY ||
		model.ConnectionType == config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP

	if cityExplicitlySpecified || isSingleCityCountry(model.CountryCode, dm, cfg) {
		model.ConnectionType = config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP
		model.City = event.TargetServerCity
	} else {
		model.ConnectionType = config.ServerSelectionRule_COUNTRY_WITH_GROUP
		model.City = ""
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

// isSingleCityCountry checks if this is a single-city country by checking how many cities
// are available for the given country code with current connection settings.
func isSingleCityCountry(countryCode string, dm *DataManager, cfg config.Config) bool {
	if countryCode == "" {
		return false
	}

	cities, err := dm.Cities(
		countryCode,
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		cfg.AutoConnectData.Obfuscate,
		cfg.VirtualLocation.Get(),
	)
	if err != nil {
		return false
	}

	return len(cities) == 1
}

// buildRecentConnectionModel creates a recent connection model from a successful VPN connection
// event.
func buildRecentConnectionModel(
	event events.DataConnect,
	parameters ServerParameters,
	server *core.Server,
	dm *DataManager,
	cfg config.Config,
) (recents.Model, error) {
	connectionType := event.TargetServerSelection

	// Normalize COUNTRY to CITY for single-city countries to avoid duplicate recent entries.
	if isSingleCityCountry(parameters.CountryCode, dm, cfg) {
		if connectionType == config.ServerSelectionRule_COUNTRY {
			connectionType = config.ServerSelectionRule_CITY
		}
		if connectionType == config.ServerSelectionRule_COUNTRY_WITH_GROUP {
			connectionType = config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP
		}
	}

	recentModel := recents.Model{
		ConnectionType:     connectionType,
		ServerTechnologies: extractServerTechnologies(server),
		IsVirtual:          event.IsVirtualLocation,
	}

	// Populate model fields based on connection type
	switch recentModel.ConnectionType {
	case config.ServerSelectionRule_GROUP:
		recentModel.Group = parameters.Group

	case config.ServerSelectionRule_CITY:
		recentModel.City = event.TargetServerCity
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_COUNTRY:
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		recentModel.Group = parameters.Group
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_SPECIFIC_SERVER:
		recentModel.SpecificServer = extractSpecificServerName(event.TargetServerDomain)
		recentModel.SpecificServerName = event.TargetServerName
		recentModel.City = event.TargetServerCity
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		recentModel.Group = parameters.Group
		recentModel.City = event.TargetServerCity // Always use the actual city for explicit city connections
		recentModel.CountryCode = event.TargetServerCountryCode
		recentModel.Country = event.TargetServerCountry

	case config.ServerSelectionRule_NONE, config.ServerSelectionRule_RECOMMENDED:
		// These connection types should not create recent entries
		return recents.Model{}, fmt.Errorf("unexpected connection type in recent connections: %d", recentModel.ConnectionType)
	}

	// Apply obfuscation wrapping if needed (upgrades connection type when obfuscation is active)
	recentModel = applyObfuscationToRecentModel(recentModel, event, server, dm, cfg)
	return recentModel, nil
}
