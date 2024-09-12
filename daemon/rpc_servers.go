package daemon

import (
	"context"
	"log"
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func techToProto(tech core.ServerTechnology) pb.Technology {
	//nolint:exhaustive
	switch tech {
	case core.OpenVPNUDP:
		return pb.Technology_OPENVPN_UDP
	case core.OpenVPNTCP:
		return pb.Technology_OPENVPN_TCP
	case core.OpenVPNUDPObfuscated:
		return pb.Technology_OBFUSCATED_OPENVPN_UDP
	case core.OpenVPNTCPObfuscated:
		return pb.Technology_OBFUSCATED_OPENVPN_TCP
	case core.WireguardTech:
		return pb.Technology_NORDLYNX
	default:
		return pb.Technology_UNKNOWN_TECHNLOGY
	}
}

func groupFilter(groups core.Groups) config.ServerGroup {
	filter := []config.ServerGroup{
		// P2P(the servers part of this group have from API also standard group which needs to be ignored)
		config.ServerGroup_P2P,
		// double VPN
		config.ServerGroup_DoubleVPN,
		// Onion over VPN
		config.ServerGroup_ONION_OVER_VPN,
		// dedicated IP
		config.ServerGroup_DEDICATED_IP,
		// obfuscated openVPN
		config.ServerGroup_OBFUSCATED,
		// standard VPN
		config.ServerGroup_STANDARD_VPN_SERVERS,
	}

	for _, filterGroup := range filter {
		if slices.ContainsFunc(groups, func(g core.Group) bool {
			return g.ID == filterGroup
		}) {
			return filterGroup
		}
	}

	return config.ServerGroup_UNDEFINED
}

func (r *RPC) GetServers(ctx context.Context, in *pb.Empty) (*pb.ServersResponse, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, "loading config:", err)
		return &pb.ServersResponse{Response: &pb.ServersResponse_Error{
			Error: pb.ServersError_GET_CONFIG_ERROR,
		}}, nil
	}

	internalServers := r.dm.GetServersData().Servers
	internalServers, err = filterServers(internalServers,
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		"",
		config.ServerGroup_UNDEFINED,
		cfg.AutoConnectData.Obfuscate)
	if err != nil {
		log.Println(internal.ErrorPrefix, "filtering servers", err)
		return &pb.ServersResponse{Response: &pb.ServersResponse_Error{
			Error: pb.ServersError_FILTER_SERVERS_ERROR,
		}}, nil
	}

	servers := []*pb.Server{}
	for _, server := range internalServers {
		if !cfg.VirtualLocation.Get() && server.IsVirtualLocation() {
			continue
		}
		technologies := []pb.Technology{}
		for _, technology := range server.Technologies {
			protoTech := techToProto(technology.ID)
			if protoTech == pb.Technology_UNKNOWN_TECHNLOGY {
				continue
			}
			technologies = append(technologies, protoTech)
		}

		s := pb.Server{
			Id:           server.ID,
			CountryCode:  server.Country().Code,
			CityName:     server.Country().City.Name,
			HostName:     server.Hostname,
			Virtual:      server.IsVirtualLocation(),
			ServerGroup:  groupFilter(server.Groups),
			Technologies: technologies,
		}

		servers = append(servers, &s)
	}

	return &pb.ServersResponse{Response: &pb.ServersResponse_Servers{
		Servers: &pb.Servers{
			Servers: servers,
		},
	}}, nil
}
