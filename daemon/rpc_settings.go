package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Settings returns system daemon settings
func (r *RPC) Settings(ctx context.Context, in *pb.SettingsRequest) (*pb.SettingsResponse, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	ports := pb.Ports{}
	for port := range cfg.AutoConnectData.Allowlist.Ports.TCP {
		ports.Tcp = append(ports.Tcp, port)
	}
	for port := range cfg.AutoConnectData.Allowlist.Ports.UDP {
		ports.Udp = append(ports.Udp, port)
	}

	subnets := []string{}
	for subnet := range cfg.AutoConnectData.Allowlist.Subnets {
		subnets = append(subnets, subnet)
	}

	return &pb.SettingsResponse{
		Type: internal.CodeSuccess,
		Data: &pb.Settings{
			Technology:           cfg.Technology,
			Firewall:             cfg.Firewall,
			Fwmark:               cfg.FirewallMark,
			Routing:              cfg.Routing.Get(),
			Analytics:            cfg.Analytics.Get(),
			KillSwitch:           cfg.KillSwitch,
			AutoConnect:          cfg.AutoConnect,
			Ipv6:                 cfg.IPv6,
			Notify:               cfg.UsersData.Notify[in.GetUid()],
			Tray:                 !cfg.UsersData.TrayOff[in.GetUid()],
			Meshnet:              cfg.Mesh,
			Dns:                  cfg.AutoConnectData.DNS,
			ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
			Protocol:             cfg.AutoConnectData.Protocol,
			LanDiscovery:         cfg.LanDiscovery,
			Allowlist: &pb.Allowlist{
				Ports:   &ports,
				Subnets: subnets,
			},
			Obfuscate: cfg.AutoConnectData.Obfuscate,
		},
	}, nil
}

func (r *RPC) SettingsProtocols(ctx context.Context, _ *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{config.Protocol_UDP.String(), config.Protocol_TCP.String()},
	}, nil
}

func (r *RPC) SettingsTechnologies(ctx context.Context, _ *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{
			config.Technology_OPENVPN.String(), config.Technology_NORDLYNX.String(),
		},
	}, nil
}
