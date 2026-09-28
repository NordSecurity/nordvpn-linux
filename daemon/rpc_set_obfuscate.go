package daemon

import (
	"context"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetObfuscate(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
	}

	if cfg.AutoConnectData.Obfuscate == in.GetEnabled() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if cfg.AutoConnect {
		// TODO: Fix this. Previously it was looking for a specific tag "de32", now the regenerated tag will be "country city"/"country"
		// How to deal with this?
		switch core.IsServerObfuscated(r.dm.GetServersData().Servers, config.ServerTagFromAutoconnectData(cfg.AutoConnectData)) {
		case core.ServerNotObfuscated:
			if in.GetEnabled() {
				return &pb.Payload{
					Type: internal.CodeAutoConnectServerNotObfuscated,
				}, nil
			}
		case core.ServerObfuscated:
			if !in.GetEnabled() {
				return &pb.Payload{
					Type: internal.CodeAutoConnectServerObfuscated,
				}, nil
			}
		case core.NotAServerName:
			// autoconnect is not set to a specific server
			// so obfuscation doesn't need to be validated
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Obfuscate = in.GetEnabled()
		return c
	}); err != nil {
		log.Error(err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	r.events.Settings.Obfuscate.Publish(in.GetEnabled())

	payload := &pb.Payload{}
	payload.Type = internal.CodeSuccess
	payload.Data = []string{strconv.FormatBool(r.netw.IsVPNActive())}
	return payload, nil
}
