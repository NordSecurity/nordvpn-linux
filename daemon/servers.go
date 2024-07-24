package daemon

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/exp/slices"
)

var tag = regexp.MustCompile(`^[a-z]{2}[0-9]{2,4}$`)
var ErrDedicatedIPServer = fmt.Errorf("selected dedicated IP servers group")

// PickServer by the specified criteria.
func PickServer(
	api core.ServersAPI,
	countries core.Countries,
	servers core.Servers,
	longitude float64,
	latitude float64,
	tech config.Technology,
	protocol config.Protocol,
	obfuscated bool,
	tag string,
	groupFlag string,
	allowVirtualServer bool,
) (core.Server, bool, error) {
	result, remote, err := getServers(
		api,
		countries,
		servers,
		longitude,
		latitude,
		tech,
		protocol,
		obfuscated,
		tag,
		groupFlag,
		1,
		allowVirtualServer,
	)
	if err != nil {
		return core.Server{}, remote, err
	}

	// #nosec G404 -- not used for cryptographic purposes
	return result[rand.Intn(len(result))], remote, nil
}

func getServers(
	api core.ServersAPI,
	countries core.Countries,
	servers core.Servers,
	longitude,
	latitude float64,
	tech config.Technology,
	protocol config.Protocol,
	obfuscated bool,
	tag string,
	groupFlag string,
	count int,
	allowVirtualServer bool,
) ([]core.Server, bool, error) {
	var remote = true
	var err error
	ret := []core.Server{}

	serverGroup, err := resolveServerGroup(groupFlag, tag)
	if err != nil {
		return ret, false, err
	}

	if serverGroup == config.DedicatedIP {
		// DIP servers are taken from the user subscription services
		return nil, false, ErrDedicatedIPServer
	}

	isGroupFlagSet := groupFlag != ""
	serverTag, err := serverTagFromString(countries, api, tag, serverGroup, servers, isGroupFlagSet)
	if errors.Is(err, internal.ErrTagDoesNotExist) {
		return ret, false, err
	}
	if err == nil {
		if serverTag.Action == core.ServerByName {
			ret, err = getSpecificServerRemote(
				api,
				tech,
				protocol,
				obfuscated,
				serverTag,
				serverGroup,
				tag,
			)
		} else {
			ret, err = getServersRemote(
				api,
				longitude,
				latitude,
				tech,
				protocol,
				obfuscated,
				serverTag,
				serverGroup,
				count,
			)
		}
	}

	if err != nil {
		// if server cannot be selected from the API, try from locally cached servers
		remote = false
		log.Println(internal.ErrorPrefix, "failed to select server from remote", err)
		ret, err = filterServers(
			servers,
			tech,
			protocol,
			tag,
			serverGroup,
			obfuscated,
		)
	}

	if err != nil {
		return ret, false, err
	}

	if !allowVirtualServer && len(ret) > 0 {
		ret = slices.DeleteFunc(ret, func(s core.Server) bool { return s.IsVirtualLocation() })
		if len(ret) == 0 {
			// if the selected servers are only virtual, but user has this disabled return an error
			return ret, false, internal.ErrVirtualServerSelected
		}
	}

	if count == 1 && len(ret) > 0 {
		// #nosec G404 -- not used for cryptographic purposes
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		ret = []core.Server{ret[rng.Int63n(int64(len(ret)))]}
	}

	return ret, remote, nil
}

func resolveServerGroup(flag, tag string) (config.ServerGroup, error) {
	tagServerGroup := groupConvert(tag)
	flagServerGroup := groupConvert(flag)

	if tagServerGroup != config.UndefinedGroup && flagServerGroup != config.UndefinedGroup {
		return config.UndefinedGroup, internal.ErrDoubleGroup
	}
	if flag != "" {
		if flagServerGroup == config.UndefinedGroup {
			return config.UndefinedGroup, internal.ErrGroupDoesNotExist
		}

		return flagServerGroup, nil
	}

	return tagServerGroup, nil
}

func getSpecificServerRemote(
	api core.ServersAPI,
	tech config.Technology,
	protocol config.Protocol,
	obfuscated bool,
	serverTag core.ServerTag,
	group config.ServerGroup,
	tag string,
) ([]core.Server, error) {
	server, err := api.Server(serverTag.ID)
	if err != nil {
		return nil, err
	}

	filteredServers := internal.Filter(core.Servers{*server}, func(s core.Server) bool {
		return core.IsConnectableWithProtocol(tech, protocol)(s) &&
			(core.IsObfuscated()(s) == obfuscated)
	})

	if len(filteredServers) == 0 {
		log.Println(internal.DebugPrefix, "server", tag, "not available for:", tech, protocol, group, obfuscated)
		return nil, internal.ErrServerIsUnavailable
	}
	return filteredServers, nil
}

func getServersRemote(
	api core.ServersAPI,
	longitude,
	latitude float64,
	tech config.Technology,
	protocol config.Protocol,
	obfuscated bool,
	tag core.ServerTag,
	group config.ServerGroup,
	count int,
) ([]core.Server, error) {
	serverTech := techToServerTech(tech, protocol, obfuscated)
	if serverTech == core.Unknown {
		return nil, errors.New("unknown technology")
	}
	limit := 20
	if count != 1 {
		limit = count
	}

	if group == config.UndefinedGroup && obfuscated {
		group = config.Obfuscated
	}

	filter := core.ServersFilter{
		Group: group,
		Tech:  serverTech,
		Tag:   tag,
		Limit: limit,
	}

	servers, _, err := api.RecommendedServers(filter, longitude, latitude)
	if err != nil {
		return nil, err
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("recommended: empty list")
	}

	return servers, nil
}

func filterServers(
	servers core.Servers,
	tech config.Technology,
	protocol config.Protocol,
	serverTag string,
	group config.ServerGroup,
	obfuscated bool,
) ([]core.Server, error) {
	ret := internal.Filter(servers, canConnect(tech, protocol, serverTag, group, obfuscated))
	if len(ret) == 0 {
		log.Println(internal.DebugPrefix, "no servers found for:", tech, protocol, serverTag, group, obfuscated)
		return nil, internal.ErrServerIsUnavailable
	}
	return ret, nil
}

func serverTagToServerBy(serverTag string, srv core.Server) core.ServerBy {
	countryName := strings.ReplaceAll(srv.Locations[0].Country.Name, " ", "_")
	countryCode := strings.ReplaceAll(srv.Locations[0].Country.Code, " ", "_")
	cityName := strings.ReplaceAll(srv.Locations[0].Country.City.Name, " ", "_")
	if strings.EqualFold(countryCode, "gb") {
		countryCode = "uk"
	}
	var by core.ServerBy

	switch {
	case len(serverTag) == 0:
		by = core.ServerBySpeed
	case slices.ContainsFunc(srv.Groups, core.ByTag(serverTag)):
		by = core.ServerBySpeed
	case strings.EqualFold(serverTag, strings.Split(srv.Hostname, ".")[0]):
		by = core.ServerByName
	case strings.EqualFold(serverTag, countryName),
		strings.EqualFold(serverTag, countryCode):
		by = core.ServerByCountry
	case strings.EqualFold(serverTag, cityName),
		strings.EqualFold(serverTag, countryName+cityName),
		strings.EqualFold(serverTag, countryCode+cityName):
		by = core.ServerByCity
	}
	return by
}

func serverTagFromString(
	countries core.Countries,
	api core.ServersAPI,
	serverTag string,
	group config.ServerGroup,
	servers core.Servers,
	isGroupFlagSet bool,
) (core.ServerTag, error) {
	if len(serverTag) == 0 {
		return core.ServerTag{Action: core.ServerByUnknown, ID: 0}, nil
	}

	if group != config.UndefinedGroup && !isGroupFlagSet {
		return core.ServerTag{Action: core.ServerBySpeed, ID: int64(group)}, nil
	}

	if strings.EqualFold(serverTag, "uk") {
		serverTag = "gb"
	}

	var err error
	if len(countries) == 0 {
		countries, _, err = api.ServersCountries()
		if err != nil {
			return core.ServerTag{Action: core.ServerByUnknown, ID: 0}, err
		}
	}
	for _, country := range countries {
		countryName := internal.SnakeCase(country.Name)
		countryCode := internal.SnakeCase(country.Code)

		if strings.EqualFold(serverTag, countryName) || strings.EqualFold(serverTag, countryCode) {
			return core.ServerTag{Action: core.ServerByCountry, ID: country.ID}, nil
		}
		for _, city := range country.Cities {
			cityName := internal.SnakeCase(city.Name)
			if strings.EqualFold(serverTag, cityName) ||
				strings.EqualFold(serverTag, countryName+" "+cityName) ||
				strings.EqualFold(serverTag, countryCode+" "+cityName) {
				return core.ServerTag{Action: core.ServerByCity, ID: city.ID}, nil
			}
		}
	}
	for _, server := range servers {
		if strings.EqualFold(serverTag, strings.Split(server.Hostname, ".")[0]) {
			return core.ServerTag{Action: core.ServerByName, ID: server.ID}, nil
		}
	}
	if !tag.MatchString(serverTag) {
		return core.ServerTag{}, internal.ErrTagDoesNotExist
	}
	return core.ServerTag{}, fmt.Errorf("could not determine server tag from %q", serverTag)
}

func groupConvert(group string) config.ServerGroup {
	key := internal.SnakeCase(group)
	if _, ok := config.GroupMap[key]; ok {
		return config.GroupMap[key]
	}
	return config.UndefinedGroup
}

func techToServerTech(tech config.Technology, protocol config.Protocol, obfuscated bool) core.ServerTechnology {
	switch tech {
	case config.Technology_NORDLYNX:
		return core.WireguardTech
	case config.Technology_OPENVPN:
		switch protocol {
		case config.Protocol_TCP:
			if obfuscated {
				return core.OpenVPNTCPObfuscated
			}
			return core.OpenVPNTCP
		case config.Protocol_UDP:
			if obfuscated {
				return core.OpenVPNUDPObfuscated
			}
			return core.OpenVPNUDP
		case config.Protocol_UNKNOWN_PROTOCOL:
			break
		}
	case config.Technology_UNKNOWN_TECHNOLOGY:
		break
	}
	return core.Unknown
}

func canConnect(
	tech config.Technology,
	protocol config.Protocol,
	serverTag string,
	group config.ServerGroup,
	obfuscated bool,
) core.Predicate {
	return func(s core.Server) bool {
		return core.IsConnectableWithProtocol(tech, protocol)(s) &&
			(core.IsObfuscated()(s) == obfuscated) &&
			selectFilter(serverTag, group, obfuscated)(s)
	}
}

func selectFilter(tag string, group config.ServerGroup, obfuscated bool) core.Predicate {
	if tag != "" && group != config.UndefinedGroup {
		return func(s core.Server) bool {
			return slices.ContainsFunc(s.Groups, core.ByGroup(group)) && slices.Contains(s.Keys, tag)
		}
	}

	if group != config.UndefinedGroup {
		return func(s core.Server) bool {
			return slices.ContainsFunc(s.Groups, core.ByGroup(group))
		}
	}

	if tag != "" {
		return func(s core.Server) bool {
			return slices.Contains(s.Keys, tag)
		}
	}

	return func(s core.Server) bool {
		getGroup := func() config.ServerGroup {
			if obfuscated {
				return config.Obfuscated
			}
			return config.StandardVPNServers
		}
		return slices.ContainsFunc(s.Groups, core.ByGroup(getGroup()))
	}
}

// isAnyDIPServersAvailable returns true if dedicated IP server is selected for any of the provided services
func isAnyDIPServersAvailable(dedicatedIPServices []auth.DedicatedIPService) bool {
	index := slices.IndexFunc(dedicatedIPServices, func(dipService auth.DedicatedIPService) bool {
		return dipService.ServerID != -1
	})

	return index != -1
}

func selectServer(r *RPC, insights *core.Insights, cfg config.Config, tag string, groupFlag string) (*core.Server, bool, error) {
	serversList := r.dm.GetServersData().Servers
	server, remote, err := PickServer(
		r.serversAPI,
		r.dm.GetCountryData().Countries,
		serversList,
		insights.Longitude,
		insights.Latitude,
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		cfg.AutoConnectData.Obfuscate,
		tag,
		groupFlag,
		cfg.VirtualLocation.Get(),
	)

	if err != nil {
		log.Println(internal.ErrorPrefix, "picking servers:", err)
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			if err := r.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				return nil, false, err
			}
			return nil, false, internal.ErrNotLoggedIn
		case errors.Is(err, internal.ErrTagDoesNotExist),
			errors.Is(err, internal.ErrGroupDoesNotExist),
			errors.Is(err, internal.ErrServerIsUnavailable),
			errors.Is(err, internal.ErrDoubleGroup),
			errors.Is(err, internal.ErrVirtualServerSelected):
			return nil, false, err

		case errors.Is(err, ErrDedicatedIPServer):
			dedicatedIPServer, err := selectDedicatedIPServer(r.ac, serversList)
			if err != nil {
				return nil, false, err
			}
			server = *dedicatedIPServer

		default:
			return nil, false, internal.ErrUnhandled
		}
	}

	log.Println(internal.InfoPrefix, "server", server.Hostname, "remote", remote)

	if isDedicatedIP(server) {
		dedicatedIPServices, err := r.ac.GetDedicatedIPServices()
		if err != nil {
			log.Println(internal.ErrorPrefix, "getting dedicated IP service data:", err)
			return nil, false, internal.ErrUnhandled
		}

		if len(dedicatedIPServices) == 0 {
			return nil, false, internal.NewErrorWithCode(internal.CodeDedicatedIPRenewError)
		}

		if !isAnyDIPServersAvailable(dedicatedIPServices) {
			return nil, false, internal.NewErrorWithCode(internal.CodeDedicatedIPServiceButNoServers)
		}

		index := slices.IndexFunc(dedicatedIPServices, func(s auth.DedicatedIPService) bool {
			return s.ServerID == server.ID
		})
		if index == -1 {
			log.Println(internal.ErrorPrefix, "server is not in the DIP servers list")
			return nil, false, internal.NewErrorWithCode(internal.CodeDedicatedIPNoServer)
		}
	}

	return &server, remote, nil
}

func getServerByID(servers core.Servers, serverID int64) (*core.Server, error) {
	index := slices.IndexFunc(servers, func(server core.Server) bool {
		return server.ID == serverID
	})

	if index == -1 {
		return nil, fmt.Errorf("server not found")
	}

	return &servers[index], nil
}

func selectDedicatedIPServer(authChecker auth.Checker, servers core.Servers) (*core.Server, error) {
	dedicatedIPServices, err := authChecker.GetDedicatedIPServices()
	if err != nil {
		log.Println(internal.ErrorPrefix, "getting dedicated IP service data:", err)
		return nil, internal.ErrUnhandled
	}

	if len(dedicatedIPServices) == 0 {
		return nil, internal.NewErrorWithCode(internal.CodeDedicatedIPRenewError)
	}

	if !isAnyDIPServersAvailable(dedicatedIPServices) {
		return nil, internal.NewErrorWithCode(internal.CodeDedicatedIPServiceButNoServers)
	}

	service := dedicatedIPServices[rand.Intn(len(dedicatedIPServices))]
	server, err := getServerByID(servers, service.ServerID)
	if err != nil {
		log.Println(internal.ErrorPrefix, "DIP server not found:", err)
		return nil, internal.ErrServerIsUnavailable
	}

	return server, nil
}
