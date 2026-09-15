package serverpicker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

var tagRegExp = regexp.MustCompile(`^[a-z]{2}[0-9]{2,4}$`)

// resolveServerGroup returns the detected group and clears the tag from the request when the group
// was provided via the tag instead of --group.
func resolveServerGroup(input SearchParams) (SearchParams, config.ServerGroup, error) {
	groupFromTag := groupConvert(input.Tag)
	groupFromFlag := groupConvert(input.Group)

	tagIsGroup := groupFromTag != config.ServerGroup_UNDEFINED
	flagValid := groupFromFlag != config.ServerGroup_UNDEFINED

	if tagIsGroup && flagValid {
		return input, config.ServerGroup_UNDEFINED, internal.ErrDoubleGroup
	}

	if input.Group != "" {
		if !flagValid {
			return input, config.ServerGroup_UNDEFINED, internal.ErrGroupDoesNotExist
		}

		return input, groupFromFlag, nil
	}

	// if server group was set as a tag and not as a --group flag, then we need to reset it
	if tagIsGroup {
		log.ServerSel.Debug("reset tag because group is specified in it")
		input.Tag = ""
	}

	return input, groupFromTag, nil
}

func serverTagFromString(
	tag string,
	group config.ServerGroup,
	countries core.Countries,
	servers core.Servers,
) (core.ServerTag, error) {
	if len(tag) == 0 {
		if group != config.ServerGroup_UNDEFINED {
			return core.ServerTag{Action: core.ServerBySpeed, ID: int64(group)}, nil
		}
		return core.ServerTag{Action: core.ServerByUnknown, ID: 0}, nil
	}

	if strings.EqualFold(tag, "uk") {
		tag = "gb"
	}

	if country, city := findCountryAndCity(tag, countries); country != nil {
		if city != nil {
			return core.ServerTag{Action: core.ServerByCity, ID: city.ID}, nil
		}
		return core.ServerTag{Action: core.ServerByCountry, ID: country.ID}, nil
	}

	for _, server := range servers {
		if strings.EqualFold(tag, strings.Split(server.Hostname, ".")[0]) {
			return core.ServerTag{Action: core.ServerByName, ID: server.ID}, nil
		}
		if serverTag := matchInputToServerLocation(server, tag); serverTag != nil {
			return *serverTag, nil
		}
	}

	if !tagRegExp.MatchString(tag) {
		return core.ServerTag{}, internal.ErrTagDoesNotExist
	}

	return core.ServerTag{}, fmt.Errorf("could not determine server tag from %q", tag)
}

// findCountryAndCity returns:
//
// * the matched country and a nil city if name is a country name or country code
// * the matched country and city if name is a city
// * nil, nil if name is neither a country nor a city
func findCountryAndCity(name string, countries core.Countries) (*core.Country, *core.City) {
	for _, country := range countries {
		countryName := internal.SnakeCase(country.Name)
		countryCode := internal.SnakeCase(country.Code)

		if strings.EqualFold(name, countryName) || strings.EqualFold(name, countryCode) {
			return &country, nil
		}
		for _, city := range country.Cities {
			cityName := internal.SnakeCase(city.Name)
			if strings.EqualFold(name, cityName) ||
				strings.EqualFold(name, countryName+" "+cityName) ||
				strings.EqualFold(name, countryCode+" "+cityName) {
				return &country, &city
			}
		}
	}
	return nil, nil
}

func matchInputToServerLocation(server core.Server, tag string) *core.ServerTag {
	for _, country := range server.Locations {
		countryName := internal.SnakeCase(country.Name)
		countryCode := internal.SnakeCase(country.Code)

		if strings.EqualFold(tag, countryName) || strings.EqualFold(tag, countryCode) {
			if country.ID == 0 {
				// return only when it contains valid data
				return nil
			}
			return &core.ServerTag{Action: core.ServerByCountry, ID: country.ID}
		}
		for _, city := range append(country.Cities, country.City) {
			if city.ID == 0 {
				// return only when it contains valid data
				return nil
			}
			cityName := internal.SnakeCase(city.Name)
			if strings.EqualFold(tag, cityName) ||
				strings.EqualFold(tag, countryName+" "+cityName) ||
				strings.EqualFold(tag, countryCode+" "+cityName) {
				return &core.ServerTag{Action: core.ServerByCity, ID: city.ID}
			}
		}
	}

	return nil
}

func groupConvert(group string) config.ServerGroup {
	key := internal.SnakeCase(group)
	if _, ok := config.GroupMap[key]; ok {
		return config.GroupMap[key]
	}
	return config.ServerGroup_UNDEFINED
}
