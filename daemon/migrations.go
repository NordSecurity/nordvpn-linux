package daemon

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// MigrateDeprecatedGroupsAutoconnect removes the deprecated regional and P2P groups from autoconnect.
// If it was the only target, ServerTag is cleared so autoconnect falls back to quick connect.
// The migration is idempotent.
func MigrateDeprecatedGroupsAutoconnect(cm config.Manager) error {
	var cfg config.Config
	if err := cm.Load(&cfg); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if !config.IsRegionalGroup(cfg.AutoConnectData.Group) && !config.IsDeprecatedP2PGroup(cfg.AutoConnectData.Group) {
		return nil
	}
	return cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Group = config.ServerGroup_UNDEFINED
		if c.AutoConnectData.Country == "" && c.AutoConnectData.City == "" {
			c.AutoConnectData.ServerTag = ""
			c.AutoConnectData.CountryCode = ""
		}
		return c
	})
}

// MigrateLegacyAllowlist moves the allowlist from the pre-6.0.0 "whitelist" key into new one.
func MigrateLegacyAllowlist(c config.Config) config.Config {
	if c.AutoConnectData.LegacyAllowlist != nil {
		c.AutoConnectData.Allowlist = *c.AutoConnectData.LegacyAllowlist
		c.AutoConnectData.LegacyAllowlist = nil
	}
	return c
}

// ConfigCleanup - validate/cleanup DNS addresses, allowlist subnets
func ConfigCleanup(c config.Config) config.Config {
	// Remove all nameservers with IPv6 addresses
	var dnsList []string
	for _, addr := range c.AutoConnectData.DNS {
		if internal.IsAddressValidAsDNSServer(addr) {
			dnsList = append(dnsList, addr)
		} else {
			log.Warn("remove invalid DNS address from the list", addr)
		}
	}
	c.AutoConnectData.DNS = dnsList

	// Remove overlapping, invalid and IPv6 subnets, if any
	c.AutoConnectData.Allowlist.NormalizeSubnets(func(removed, reason string) {
		log.Warn("On start, allowlist remove subnet:", removed, "; reason:", reason)
	})

	return c
}

// strToServerTag converts a regular string to a server tag style string, i.e it changes all spaces to underscores and
// all upper case letters to lowercase.
func strToServerTag(tag string) string {
	return strings.ReplaceAll(strings.ToLower(tag), " ", "_")
}

// MigrateAutoconnectTarget normalizes the deprecated autoconnect ServerTag into a
// country or city target and clears the ServerTag.
//
// It runs in two stages:
//  1. If the target is a specific server (deprecated in 6.0.0), it falls back to
//     that server's country/city, or to the country code embedded in the tag when
//     the server is no longer known
//  2. Any remaining country/city tag is resolved into the Country/City fields
func MigrateAutoconnectTarget(configManager config.Manager, dataManager *DataManager) error {
	var cfg config.Config
	if err := configManager.Load(&cfg); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.AutoConnectData.ServerTag == "" {
		return nil
	}

	// Stage 1: resolve a specific-server target to its country/city
	serverTag := serverpicker.UnitedKingdomServerTagFix(cfg.AutoConnectData.ServerTag)
	if serverpicker.IsServerTag(serverTag) {
		serversData := dataManager.GetServersData()
		serverIndex := slices.IndexFunc(serversData.Servers, func(server core.Server) bool {
			return serverpicker.MatchTagToHostname(serverTag, server)
		})

		if serverIndex != -1 {
			log.Info("autoconnection target set to a specific server tag, falling back to server's country/city")

			server := serversData.Servers[serverIndex]
			serversCountry := server.Country()

			cfg.AutoConnectData.City = serversCountry.City.Name
			cfg.AutoConnectData.Country = serversCountry.Name
			cfg.AutoConnectData.CountryCode = serversCountry.Code
			cfg.AutoConnectData.ServerTag = strToServerTag(serversCountry.City.Name)
		} else {
			serverNameRegEx := regexp.MustCompile(`^([a-zA-Z]{2}(?:-[a-zA-Z]{2})?)(\d+)$`)
			match := serverNameRegEx.FindStringSubmatch(serverTag)
			// it has 3 elements, [0] - full, [1] - country code, [2] - server number
			if len(match) == 3 {
				countryCode := match[1]
				log.Warn(
					"autoconnection target set to a specific server tag, server not found, falling back to server's country code:",
					countryCode,
				)

				cfg.AutoConnectData.ServerTag = countryCode
				cfg.AutoConnectData.CountryCode = strings.ToUpper(countryCode)

				if dataManager.CountryDataExists() {
					countries := dataManager.GetCountryData().Countries
					countryIndex := slices.IndexFunc(countries, func(c core.Country) bool {
						return c.Code == countryCode
					})
					if countryIndex != -1 {
						cfg.AutoConnectData.Country = strToServerTag(countries[countryIndex].Name)
					}
				}
			} else {
				cfg.AutoConnectData.ServerTag = ""
				log.Warn("failed to extract country code out of the server name, will fall back to fastest server")
			}
		}
	}

	// Stage 2: resolve a country/city tag into the Country/City fields.
	//          Anything that's left unresolved is discarded and connect to fastest as fallback.
	tag := cfg.AutoConnectData.ServerTag
	if tag != "" && dataManager.CountryDataExists() {
		countries := dataManager.GetCountryData().Countries
		country, city := serverpicker.FindCountryAndCity(tag, countries)
		if country != nil {
			cfg.AutoConnectData.Country = strToServerTag(country.Name)
		}
		if city != nil {
			cfg.AutoConnectData.City = strToServerTag(city.Name)
		}
		if country == nil && city == nil {
			log.Warn("Failed to migrate Country/City property for ServerTag:", tag, " - migrating to empty")
		}
	}
	cfg.AutoConnectData.ServerTag = ""

	if err := configManager.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData = cfg.AutoConnectData
		return c
	}); err != nil {
		return fmt.Errorf("saving migrated config: %w", err)
	}

	return nil
}
