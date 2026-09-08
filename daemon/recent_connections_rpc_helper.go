package daemon

import (
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// storePendingRecentConnection stores current pending recent connection to memory
func storePendingRecentConnection(store *recents.RecentConnectionsStore) {
	exists, recentModel := store.PopPending()
	if !exists {
		return
	}

	_ = store.Add(recentModel)
}

// isRecentConnectionSupported returns true if server connection can be used for reconnection,
// otherwise false
func isRecentConnectionSupported(rule config.ServerSelectionRule) bool {
	return rule != config.ServerSelectionRule_RECOMMENDED && rule != config.ServerSelectionRule_NONE
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
	parameters serverpicker.ServerParameters,
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
		ConnectionType: connectionType,
		IsVirtual:      event.IsVirtualLocation,
		ConnectionTech: cfg.Technology,
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

	return recentModel, nil
}
