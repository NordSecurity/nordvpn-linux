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

	if !cfg.AutoConnect && !in.GetEnabled() {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
		}, nil
	}

	if in.GetEnabled() {
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

	var parameters ServerParameters
	serverTag := in.GetServerTag()
	if in.GetEnabled() {
		if serverTag != "" {
			insights := r.dm.GetInsightsData().Insights

			server, _, err := selectServer(r, &insights, cfg, serverTag, "")
			if err != nil {
				log.Println(internal.ErrorPrefix, "no server found for autoconnect", serverTag, err)

				var errorCode *internal.ErrorWithCode
				if errors.As(err, &errorCode) {
					return &pb.Payload{
						Type: errorCode.Code,
					}, nil
				}

				return nil, err
			}
			log.Println(internal.InfoPrefix, "server for autoconnect found", server)
			// On the cli side, using the --group flag overrides any other arguments and group name will replace the
			// server tag. Once this is fixed and this RPC accepts both server tag and a group flag, group flag should
			// be used as a second argument in this call.s
			parameters = GetServerParameters(serverTag, serverTag, r.dm.GetCountryData().Countries)
		}

		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.AutoConnect = in.GetEnabled()
			c.AutoConnectData = config.AutoConnectData{
				ID:                   cfg.AutoConnectData.ID,
				ServerTag:            serverTag,
				Country:              parameters.Country,
				City:                 parameters.City,
				Group:                parameters.Group,
				Protocol:             cfg.AutoConnectData.Protocol,
				ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
				Obfuscate:            cfg.AutoConnectData.Obfuscate,
				DNS:                  cfg.AutoConnectData.DNS,
				Allowlist:            cfg.AutoConnectData.Allowlist,
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
			c.AutoConnect = in.GetEnabled()
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	}
	r.events.Settings.Autoconnect.Publish(in.GetEnabled())

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
