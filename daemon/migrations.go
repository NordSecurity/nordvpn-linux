package daemon

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/features"
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

func MigrateConfig(c config.Config) config.Config {
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

	// switch from OpenVPN obfuscated to NordWhisper or simple OpenVPN
	return migrateObfuscatedSettingsToNordWhisper(c, features.NordWhisperEnabled)
}

func migrateObfuscatedSettingsToNordWhisper(c config.Config, isNordWhisperEnabled bool) config.Config {
	if !c.AutoConnectData.Obfuscate {
		return c
	}

	log.Info("migrating OpenVPN obfuscated")

	c.AutoConnectData.Obfuscate = false

	if isNordWhisperEnabled {
		c.Technology = config.Technology_NORDWHISPER
		c.AutoConnectData.Protocol = config.Protocol_Webtunnel
	} else if c.AutoConnectData.Group == config.ServerGroup_OVPN_OBFUSCATED {
		// for open source builds change the group to non obfuscated
		c.AutoConnectData.Group = config.ServerGroup_UNDEFINED
	}

	return c
}

// strToServerTag converts a regular string to a server tag style string, i.e it changes all spaces to underscores and
// all upper case letters to lowercase.
func strToServerTag(tag string) string {
	return strings.ReplaceAll(strings.ToLower(tag), " ", "_")
}

// MigrateDeprecatedAutoconnectToSpecificServer checks if autoconnect target is a specific server. If it is, it sets the
// target to its country/city.
//
// Autoconnect to a specific server was deprecated in version 6.0.0.
func MigrateDeprecatedAutoconnectToSpecificServer(configManager config.Manager, dataManager *DataManager) error {
	var cfg config.Config
	if err := configManager.Load(&cfg); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.AutoConnectData.ServerTag == "" {
		return nil
	}

	serverTag := serverpicker.UnitedKingdomServerTagFix(cfg.AutoConnectData.ServerTag)

	if !serverpicker.IsServerTag(serverTag) {
		return nil
	}

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

	if err := configManager.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData = cfg.AutoConnectData
		return c
	}); err != nil {
		return fmt.Errorf("saving migrated config: %w", err)
	}

	return nil
}
