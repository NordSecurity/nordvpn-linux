package daemon

import (
	"context"
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func techToProto(tech core.ServerTechnology) pb.Technology {
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
	}

	return pb.Technology_UNKNWON
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
	internalServers := r.dm.GetServersData().Servers

	servers := []*pb.Server{}
	for _, server := range internalServers {
		technologies := []pb.Technology{}
		for _, technology := range server.Technologies {
			protoTech := techToProto(technology.ID)
			if protoTech == pb.Technology_UNKNWON {
				continue
			}
			technologies = append(technologies, protoTech)
		}

		ips := []string{}
		for _, ip := range server.IPs() {
			ips = append(ips, ip.String())
		}

		s := pb.Server{
			Ips:          ips,
			CountryCode:  server.Country().Code,
			CityName:     server.Country().City.Name,
			HostName:     server.Hostname,
			Virtual:      server.IsVirtualLocation(),
			ServerGroup:  groupFilter(server.Groups),
			Technologies: technologies,
		}

		servers = append(servers, &s)
	}

	return &pb.ServersResponse{Servers: servers}, nil
}
