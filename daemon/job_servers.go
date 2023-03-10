package daemon

import (
	"errors"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"

	mapset "github.com/deckarep/golang-set"
)

var alphanumeric = regexp.MustCompile(`[^0-9a-zA-Z ]+`)

// JobServers is responsible for population of local server cache which is needed
// to avoid excees requests to the backend API.
func JobServers(dm *DataManager, cm config.Manager, api core.ServersAPI, validate bool) func() error {
	return func() error {
		var cfg config.Config
		err := cm.Load(&cfg)
		if err != nil {
			log.Println(internal.ErrorPrefix, err)
		}
		if validate && dm.ServerDataExists() {
			// always fill app data even if db file is outdated
			SetAppData(dm, cfg.Technology, dm.GetServersData().Servers)

			// if db is still valid, make sure it's locked and do nothing
			if dm.IsServersDataValid() {
				return nil
			}
		}

		// save execution start time
		currentTime := time.Now()
		servers, headers, err := api.Servers()
		if err != nil {
			return err
		}

		if len(servers) == 0 {
			return errors.New("empty servers list")
		}

		geoInfoData := dm.GetInsightsData()
		randomComponent := randFloat(time.Now().UnixNano(), RandomComponentMin, RandomComponentMax)

		// format first server beforehand to create initial values
		// TODO: change server date format to equivalent from time.RFCXXXX
		parsedTime, err := time.Parse(internal.ServerDateFormat, servers[0].CreatedAt)
		if err != nil {
			return err
		}
		timestamp := parsedTime.Unix()
		dist := distance(
			geoInfoData.Insights.Latitude,
			geoInfoData.Insights.Longitude,
			servers[0].Locations[0].Country.City.Latitude,
			servers[0].Locations[0].Country.City.Longitude,
		)
		servers[0].Timestamp = timestamp
		servers[0].Distance = dist

		// set initial minmax values
		timestampMin := timestamp
		timestampMax := timestamp
		distanceMin := dist
		distanceMax := dist

		var filteredServers core.Servers

		// first iteration to filter "bad" servers and find minmax values
		for idx, server := range servers {
			// store keys to find server easier
			loweredHostnameID := strings.ToLower(strings.Split(server.Hostname, ".")[0])
			loweredCountryName := strings.ToLower(strings.Join(strings.Split(server.Locations[0].Country.Name, " "), "_"))
			loweredCountryCode := strings.ToLower(strings.Join(strings.Split(server.Locations[0].Country.Code, " "), "_"))
			loweredCityName := strings.ToLower(strings.Join(strings.Split(server.Locations[0].Country.City.Name, " "), "_"))
			loweredGroupTitles := make([]string, len(server.Groups))
			for idx, group := range server.Groups {
				loweredGroupTitles[idx] = strings.ToLower(strings.Join(strings.Split(group.Title, " "), "_"))
			}

			if loweredCountryCode == "gb" {
				loweredCountryCode = "uk"
			}

			servers[idx].Keys = append([]string{
				loweredCountryName,
				loweredCountryCode,
				loweredCountryName + loweredCityName,
				loweredCountryCode + loweredCityName,
				loweredCityName,
				loweredHostnameID,
			}, loweredGroupTitles...)

			// calculate minmax distance and timestamp
			parsedTime, err := time.Parse(internal.ServerDateFormat, server.CreatedAt)
			if err != nil {
				return err
			}
			timestamp := parsedTime.Unix()
			dist := distance(
				geoInfoData.Insights.Latitude,
				geoInfoData.Insights.Longitude,
				server.Locations[0].Country.City.Latitude,
				server.Locations[0].Country.City.Longitude,
			)
			servers[idx].Timestamp = timestamp
			servers[idx].Distance = dist

			if dist < distanceMin {
				distanceMin = dist
			}
			if dist > distanceMax {
				distanceMax = dist
			}
			if timestamp < timestampMin {
				timestampMin = timestamp
			}
			if timestamp > timestampMax {
				timestampMax = timestamp
			}

			filteredServers = append(filteredServers, servers[idx])
		}
		servers = filteredServers

		// second iteration to calculate penalty scores
		for idx, server := range servers {
			penal, partialPenalty := penalty(
				core.IsObfuscated()(server),
				server.Distance, distanceMin, distanceMax,
				server.Timestamp, timestampMin, timestampMax,
				server.Load,
				geoInfoData.Insights.CountryCode, server.Locations[0].Country.Code,
				server.Locations[0].Country.City.HubScore,
				randomComponent,
			)
			servers[idx].Penalty = penal
			servers[idx].PartialPenalty = partialPenalty
		}

		// sort by penalty score
		sort.SliceStable(servers, func(i, j int) bool {
			return servers[i].Penalty < servers[j].Penalty
		})

		SetAppData(dm, cfg.Technology, servers)
		err = dm.SetServersData(currentTime, servers, headers.Get(core.HeaderDigest))
		if err != nil {
			return err
		}
		return nil
	}
}

func SetAppData(dm *DataManager, tech config.Technology, servers core.Servers) {
	countryNames := map[bool]map[config.Protocol]mapset.Set{
		false: {
			config.Protocol_UDP: mapset.NewSet(),
			config.Protocol_TCP: mapset.NewSet(),
		},
		true: {
			config.Protocol_UDP: mapset.NewSet(),
			config.Protocol_TCP: mapset.NewSet(),
		},
	}
	cityNames := map[bool]map[config.Protocol]map[string]mapset.Set{
		false: {
			config.Protocol_UDP: make(map[string]mapset.Set, 0),
			config.Protocol_TCP: make(map[string]mapset.Set, 0),
		},
		true: {
			config.Protocol_UDP: make(map[string]mapset.Set, 0),
			config.Protocol_TCP: make(map[string]mapset.Set, 0),
		},
	}
	groupNames := map[bool]map[config.Protocol]mapset.Set{
		false: {
			config.Protocol_UDP: mapset.NewSet(),
			config.Protocol_TCP: mapset.NewSet(),
		},
		true: {
			config.Protocol_UDP: mapset.NewSet(),
			config.Protocol_TCP: mapset.NewSet(),
		},
	}

	for _, server := range servers {
		var (
			hasUDP bool
			hasTCP bool
		)

		switch tech {
		case config.Technology_OPENVPN:
			if core.IsConnectableVia(core.OpenVPNUDP)(server) ||
				core.IsConnectableVia(core.OpenVPNUDPObfuscated)(server) {
				hasUDP = true
			}
			if core.IsConnectableVia(core.OpenVPNTCP)(server) ||
				core.IsConnectableVia(core.OpenVPNTCPObfuscated)(server) {
				hasTCP = true
			}
		case config.Technology_NORDLYNX:
			if core.IsConnectableVia(core.WireguardTech)(server) {
				hasUDP = true
			}
		case config.Technology_UNKNOWN_TECHNOLOGY:
			fallthrough
		default:
			continue
		}

		countryTitle := internal.Title(server.Locations[0].Country.Name)
		loweredCountryTitle := strings.ToLower(countryTitle)
		cityTitle := internal.Title(server.Locations[0].Country.City.Name)
		groupTitles := make([]string, len(server.Groups))
		for idx, group := range server.Groups {
			groupTitles[idx] = internal.Title(alphanumeric.ReplaceAllString(group.Title, ""))
		}

		var protos []config.Protocol
		if hasUDP {
			protos = append(protos, config.Protocol_UDP)
		}
		if hasTCP {
			protos = append(protos, config.Protocol_TCP)
		}

		for _, proto := range protos {
			obfuscated := core.IsObfuscated()(server)
			countryNames[obfuscated][proto].Add(countryTitle)
			if _, ok := cityNames[obfuscated][proto][loweredCountryTitle]; ok {
				cityNames[obfuscated][proto][loweredCountryTitle].Add(cityTitle)
			} else {
				cityNames[obfuscated][proto][loweredCountryTitle] = mapset.NewSetWith(cityTitle)
			}
			for _, group := range groupTitles {
				groupNames[obfuscated][proto].Add(group)
			}
		}
	}

	dm.SetAppData(countryNames, cityNames, groupNames)
}
