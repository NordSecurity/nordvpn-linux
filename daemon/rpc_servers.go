package daemon

import (
	"context"
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func technologiesToProtobuf(technologies core.Technologies) []pb.Technology {
	technologiesProto := []pb.Technology{}
	for _, tech := range technologies {
		//nolint:exhaustive
		switch tech.ID {
		case core.OpenVPNUDP:
			technologiesProto = append(technologiesProto, pb.Technology_OPENVPN_UDP)
		case core.OpenVPNTCP:
			technologiesProto = append(technologiesProto, pb.Technology_OPENVPN_TCP)
		case core.WireguardTech:
			technologiesProto = append(technologiesProto, pb.Technology_NORDLYNX)
		}
	}

	return technologiesProto
}

// groupsToProtobuf converts core.Groups to a slice of config.ServerGroup. It also filters out the groups so that only
// ones returned are of interest to the GUI
func groupsToProtobuf(groups core.Groups, technology config.Technology) []config.ServerGroup {
	// the obfuscated group is never taken from the server tags. Only the OVPN XOR servers carry
	// that tag, and those are not connectable anymore
	filter := []config.ServerGroup{
		config.ServerGroup_P2P,
		config.ServerGroup_DOUBLE_VPN,
		config.ServerGroup_ONION_OVER_VPN,
		config.ServerGroup_DEDICATED_IP,
		config.ServerGroup_STANDARD_VPN_SERVERS,
	}

	desiredGroups := []config.ServerGroup{}
	for _, filterGroup := range filter {
		if slices.ContainsFunc(groups, func(g core.Group) bool {
			return g.ID == filterGroup
		}) {
			desiredGroups = append(desiredGroups, filterGroup)
		}
	}

	return withObfuscatedGroup(desiredGroups, technology)
}

// withObfuscatedGroup reports the standard servers as obfuscated ones as well, but only under a
// technology which is obfuscated. No server is tagged with that group by the API anymore.
func withObfuscatedGroup(
	groups []config.ServerGroup,
	technology config.Technology,
) []config.ServerGroup {
	if !serverpicker.IsObfuscatedTech(technology) {
		return groups
	}

	if !slices.Contains(groups, config.ServerGroup_STANDARD_VPN_SERVERS) {
		return groups
	}

	return append(groups, config.ServerGroup_OBFUSCATED)
}

func serversListToServersMap(
	internalServers core.Servers,
	allowVirtual bool,
	technology config.Technology,
) []*pb.ServerCountry {
	type serversMap map[string]map[string][]*pb.Server

	sMap := make(serversMap)
	// map country code to country name
	countryNames := make(map[string]string)

	for _, server := range internalServers {
		if !allowVirtual && server.IsVirtualLocation() {
			continue
		}

		s := &pb.Server{
			Id:           server.ID,
			HostName:     server.Hostname,
			Virtual:      server.IsVirtualLocation(),
			ServerGroups: groupsToProtobuf(server.Groups, technology),
			Technologies: technologiesToProtobuf(server.Technologies),
		}

		countryCode := server.Country().Code
		cityName := server.Country().City.Name

		if _, ok := sMap[countryCode]; !ok {
			sMap[countryCode] = make(map[string][]*pb.Server)
			countryNames[countryCode] = server.Country().Name
		}

		sMap[countryCode][cityName] = append(sMap[countryCode][cityName], s)
	}

	countries := []*pb.ServerCountry{}
	for countryCode, cityMap := range sMap {
		cities := []*pb.ServerCity{}
		for city, servers := range cityMap {
			cities = append(cities, &pb.ServerCity{
				CityName: city,
				Servers:  servers,
			})
		}
		countries = append(countries, &pb.ServerCountry{
			CountryCode: countryCode,
			CountryName: countryNames[countryCode],
			Cities:      cities,
		})
	}

	return countries
}

func (r *RPC) GetServers(ctx context.Context, in *pb.Empty) (*pb.ServersResponse, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Error("loading config:", err)
		return &pb.ServersResponse{Response: &pb.ServersResponse_Error{
			Error: pb.ServersError_GET_CONFIG_ERROR,
		}}, nil
	}

	servers := internal.Filter(r.dm.GetServersData().Servers, func(s core.Server) bool {
		return serverpicker.MatchesUserSettings(s, cfg)
	})

	if len(servers) == 0 {
		log.Error("filtering servers", err)
		return &pb.ServersResponse{Response: &pb.ServersResponse_Error{
			Error: pb.ServersError_FILTER_SERVERS_ERROR,
		}}, nil
	}

	return &pb.ServersResponse{Response: &pb.ServersResponse_Servers{
		Servers: &pb.ServersMap{
			ServersByCountry: serversListToServersMap(servers, cfg.VirtualLocation.Get(), cfg.Technology),
		},
	}}, nil
}
