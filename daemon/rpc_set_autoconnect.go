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

			server, _, err := PickServer(
				r.serversAPI,
				r.dm.GetCountryData().Countries,
				r.dm.GetServersData().Servers,
				insights.Longitude,
				insights.Latitude,
				cfg.Technology,
				cfg.AutoConnectData.Protocol,
				cfg.AutoConnectData.Obfuscate,
				in.GetServerTag(),
				"",
				cfg.VirtualLocation.Get(),
			)
			if err != nil {
				log.Println(internal.ErrorPrefix, "no server found for autoconnect", in.GetServerTag(), err)
				switch {
				case errors.Is(err, core.ErrUnauthorized):
					return nil, internal.ErrNotLoggedIn

				case errors.Is(err, internal.ErrTagDoesNotExist),
					errors.Is(err, internal.ErrGroupDoesNotExist),
					errors.Is(err, internal.ErrServerIsUnavailable),
					errors.Is(err, internal.ErrDoubleGroup),
					errors.Is(err, internal.ErrVirtualServerSelected):
					return nil, err

				default:
					return nil, internal.ErrUnhandled
				}
			}

			log.Println(internal.InfoPrefix, "server for autoconnect found", server)

			if isDedicatedIP(server) {
				expired, err := r.ac.IsDedicatedIPExpired()
				if err != nil {
					log.Println(internal.ErrorPrefix, "checking dedicated IP expiration: ", err)
					return &pb.Payload{
						Type: internal.CodeDedicatedIPRenewError,
					}, nil
				}

				if expired {
					return &pb.Payload{
						Type: internal.CodeDedicatedIPRenewError,
					}, nil
				}
			}
		}

		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.AutoConnect = in.GetAutoConnect()
			c.AutoConnectData = config.AutoConnectData{
				ID:                   cfg.AutoConnectData.ID,
				ServerTag:            in.GetServerTag(),
				Protocol:             cfg.AutoConnectData.Protocol,
				ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
				Obfuscate:            in.GetObfuscate(),
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
