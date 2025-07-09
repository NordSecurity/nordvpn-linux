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
	if ok, err := r.ac.IsLoggedIn(); ok {
		if errors.Is(err, core.ErrLoginTokenExpired) {
			return nil, err
		}
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
	serverGroup := in.GetServerGroup()
	if in.GetEnabled() {
		insights := r.dm.GetInsightsData().Insights

		server, _, err := selectServer(r, &insights, cfg, serverTag, serverGroup)
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
		parameters = GetServerParameters(serverTag, serverGroup, r.dm.GetCountryData().Countries)

		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.AutoConnect = in.GetEnabled()
			c.AutoConnectData = config.AutoConnectData{
				ID:                   cfg.AutoConnectData.ID,
				ServerTag:            serverTag,
				Country:              parameters.CountryCode,
				City:                 parameters.City,
				Group:                parameters.Group,
				Protocol:             cfg.AutoConnectData.Protocol,
				ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
				Obfuscate:            cfg.AutoConnectData.Obfuscate,
				DNS:                  cfg.AutoConnectData.DNS,
				Allowlist:            cfg.AutoConnectData.Allowlist,
				PostquantumVpn:       cfg.AutoConnectData.PostquantumVpn,
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
