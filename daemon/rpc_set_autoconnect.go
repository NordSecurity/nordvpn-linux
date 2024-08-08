package daemon

import (
	"context"
	"errors"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetAutoConnect(ctx context.Context, in *pb.SetAutoconnectRequest) (*pb.Payload, error) {
	if !r.ac.IsLoggedIn() {
		return nil, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.AutoConnect == in.GetAutoConnect() {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
		}, nil
	}

	if in.GetAutoConnect() {
		switch core.IsServerObfuscated(r.dm.GetServersData().Servers, in.GetServerTag()) {
		case core.ServerNotObfuscated:
			if cfg.AutoConnectData.Obfuscate {
				return &pb.Payload{
					Type: internal.CodeAutoConnectServerNotObfuscated,
				}, nil
			}
		case core.ServerObfuscated:
			if !cfg.AutoConnectData.Obfuscate {
				return &pb.Payload{
					Type: internal.CodeAutoConnectServerObfuscated,
				}, nil
			}
		case core.NotAServerName:
			// autoconnect is not set to a specific server
			// so obfuscation doesn't need to be validated
		}
	}

	if in.GetAutoConnect() {
		if in.GetServerTag() != "" {
			insights := r.dm.GetInsightsData().Insights

			server, _, err := selectServer(r, &insights, cfg, in.GetServerTag(), "")
			if err != nil {
				log.Println(internal.ErrorPrefix, "no server found for autoconnect", in.GetServerTag(), err)

				var errorCode *internal.ErrorWithCode
				if errors.As(err, &errorCode) {
					return &pb.Payload{
						Type: errorCode.Code,
					}, nil
				}

				return nil, err
			}
			log.Println(internal.InfoPrefix, "server for autoconnect found", server)
		}

		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.AutoConnect = in.GetAutoConnect()
			c.AutoConnectData = config.AutoConnectData{
				ID:                   cfg.AutoConnectData.ID,
				ServerTag:            in.GetServerTag(),
				Protocol:             cfg.AutoConnectData.Protocol,
				ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
				Obfuscate:            cfg.AutoConnectData.Obfuscate,
				DNS:                  cfg.AutoConnectData.DNS,
				Allowlist: config.NewAllowlist(
					in.GetAllowlist().GetPorts().GetTcp(),
					in.GetAllowlist().GetPorts().GetUdp(),
					in.GetAllowlist().GetSubnets(),
				),
			}
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.AutoConnect = in.GetAutoConnect()
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	}
	r.events.Settings.Autoconnect.Publish(in.GetAutoConnect())

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
