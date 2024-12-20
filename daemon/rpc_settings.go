package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/peer"
)

// Settings returns system daemon settings
func (r *RPC) Settings(ctx context.Context, in *pb.Empty) (*pb.SettingsResponse, error) {
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

	// Storing autoconnect parameters was introduced later on so they might not be save in a config yet. We need to
	// perform an update in such cases to maintain compatibility.
	autoconnectParamsNotSet := cfg.AutoConnectData.Country == "" &&
		cfg.AutoConnectData.City == "" &&
		cfg.AutoConnectData.Group == config.ServerGroup_UNDEFINED
	if cfg.AutoConnect && cfg.AutoConnectData.ServerTag != "" && autoconnectParamsNotSet {
		// use group tag as a second prameter once it is implemented
		parameters := GetServerParameters(cfg.AutoConnectData.ServerTag,
			cfg.AutoConnectData.ServerTag,
			r.dm.GetCountryData().Countries)
		cfg.AutoConnectData.Country = parameters.Country
		cfg.AutoConnectData.City = parameters.City
		cfg.AutoConnectData.Group = parameters.Group

		err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.AutoConnectData.Country = cfg.AutoConnectData.Country
			c.AutoConnectData.City = cfg.AutoConnectData.City
			c.AutoConnectData.Group = cfg.AutoConnectData.Group

			return c
		})

		if err != nil {
			log.Println(internal.WarningPrefix, "failed to set autoconnect parameters during the settings RPC:", err)
		}
	}

	peer, ok := peer.FromContext(ctx)
	var uid int64
	if ok {
		cred, ok := peer.AuthInfo.(internal.UcredAuth)
		if !ok {
			return &pb.SettingsResponse{
				Type: internal.CodeFailure,
			}, nil
		}
		uid = int64(cred.Uid)
	}

	return &pb.SettingsResponse{
		Type: internal.CodeSuccess,
		Data: &pb.Settings{
			Technology: cfg.Technology,
			Firewall:   cfg.Firewall,
			Fwmark:     cfg.FirewallMark,
			Routing:    cfg.Routing.Get(),
			Analytics:  cfg.Analytics.Get(),
			KillSwitch: cfg.KillSwitch,
			AutoConnectData: &pb.AutoconnectData{
				Enabled:     cfg.AutoConnect,
				Country:     cfg.AutoConnectData.Country,
				City:        cfg.AutoConnectData.City,
				ServerGroup: cfg.AutoConnectData.Group,
			},
			Ipv6:                 cfg.IPv6,
			Meshnet:              cfg.Mesh,
			Dns:                  cfg.AutoConnectData.DNS,
			ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
			Protocol:             cfg.AutoConnectData.Protocol,
			LanDiscovery:         cfg.LanDiscovery,
			Allowlist: &pb.Allowlist{
				Ports:   &ports,
				Subnets: subnets,
			},
			Obfuscate:       cfg.AutoConnectData.Obfuscate,
			PostquantumVpn:  cfg.AutoConnectData.PostquantumVpn,
			VirtualLocation: cfg.VirtualLocation.Get(),
			UserSettings: &pb.UserSpecificSettings{
				Uid:    uid,
				Notify: !cfg.UsersData.NotifyOff[uid],
				Tray:   !cfg.UsersData.TrayOff[uid],
			},
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
	technologies := []string{
		config.Technology_OPENVPN.String(), config.Technology_NORDLYNX.String(),
	}

	if r.isNordWhisperEnabled() {
		technologies = append(technologies, config.Technology_NORDWHISPER.String())
	}

	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: technologies,
	}, nil
}
