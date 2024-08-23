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

	// Try to find server tag type if it is unknown to maintain compatibility with older versions of the app that did
	// not save autoconnect server tag types.
	if cfg.AutoConnect && cfg.AutoConnectData.ServerTagType == config.ServerTagType_UNKNOWN {
		if cfg.AutoConnectData.ServerTag != "" {
			tagType := GetServerTagType(cfg.AutoConnectData.ServerTag, r.dm.countryData.Countries)
			if tagType == config.ServerTagType_UNKNOWN {
				log.Println(internal.ErrorPrefix,
					"failed to determine tag type when loading settings:",
					cfg.AutoConnectData.ServerTag)
			} else {
				// try converting country code to country name to maintain consistency
				if tagType == config.ServerTagType_COUNTRY {
					cfg.AutoConnectData.ServerTag = r.dm.CountryCodeToCountryName(cfg.AutoConnectData.ServerTag)
				}
				cfg.AutoConnectData.ServerTagType = tagType
			}
		} else {
			cfg.AutoConnectData.ServerTagType = config.ServerTagType_NONE
		}

		err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.AutoConnectData.ServerTag = cfg.AutoConnectData.ServerTag
			c.AutoConnectData.ServerTagType = cfg.AutoConnectData.ServerTagType
			return c
		})

		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to save new tag type:", err)
		}
	}

	return &pb.SettingsResponse{
		Type: internal.CodeSuccess,
		Data: &pb.UserSettings{
			Settings: &pb.Settings{
				Technology: cfg.Technology,
				Firewall:   cfg.Firewall,
				Fwmark:     cfg.FirewallMark,
				Routing:    cfg.Routing.Get(),
				Analytics:  cfg.Analytics.Get(),
				KillSwitch: cfg.KillSwitch,
				AutoConnectData: &pb.AutoconnectData{
					Enabled:       cfg.AutoConnect,
					ServerTag:     cfg.AutoConnectData.ServerTag,
					ServerTagType: cfg.AutoConnectData.ServerTagType,
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
				VirtualLocation: cfg.VirtualLocation.Get(),
			},
			UserSpecificSettings: &pb.UserSpecificSettings{
				Uid:    in.GetUid(),
				Notify: !cfg.UsersData.NotifyOff[in.GetUid()],
				Tray:   !cfg.UsersData.TrayOff[in.GetUid()],
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
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{
			config.Technology_OPENVPN.String(), config.Technology_NORDLYNX.String(),
		},
	}, nil
}
