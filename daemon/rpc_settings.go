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
			Meshnet:              cfg.Mesh,
			Dns:                  cfg.AutoConnectData.DNS,
			ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
			Protocol:             cfg.AutoConnectData.Protocol,
			LanDiscovery:         cfg.LanDiscovery,
		},
	}, nil
}

func (r RPC) SettingsProtocols(ctx context.Context, _ *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{config.Protocol_UDP.String(), config.Protocol_TCP.String()},
	}, nil
}

func (r RPC) SettingsTechnologies(ctx context.Context, _ *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{
			config.Technology_OPENVPN.String(), config.Technology_NORDLYNX.String(),
		},
	}, nil
}
