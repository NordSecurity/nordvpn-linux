package serverpicker

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

type ServerParameters struct {
	Country     string
	City        string
	Group       config.ServerGroup
	CountryCode string
	ServerName  string
}

// Undefined returns true if all fields of the ServerParameters struct are unset or empty.
func (sp ServerParameters) Undefined() bool {
	return sp.Country == "" &&
		sp.City == "" &&
		sp.Group == config.ServerGroup_UNDEFINED &&
		sp.CountryCode == "" &&
		sp.ServerName == ""
}

func GetServerParameters(serverTag string, groupTag string, countries core.Countries) ServerParameters {
	var parameters ServerParameters

	groupFromServerTag := groupConvert(serverTag)
	if groupFromServerTag != config.ServerGroup_UNDEFINED {
		parameters.Group = groupFromServerTag
	} else {
		parameters.Group = groupConvert(groupTag)
	}

	country, city := findCountryAndCity(serverTag, countries)

	if country == nil {
		if groupFromServerTag == config.ServerGroup_UNDEFINED {
			parameters.ServerName = serverTag
		}
		return parameters
	}

	parameters.Country = country.Name
	parameters.CountryCode = country.Code
	if city == nil {
		return parameters
	}

	parameters.City = city.Name
	return parameters
}
