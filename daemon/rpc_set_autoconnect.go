package daemon

import (
	"context"
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetAutoConnect(ctx context.Context, in *pb.SetAutoconnectRequest) (*pb.Payload, error) {
	if serverpicker.IsServerTag(in.GetServerTag()) {
		return &pb.Payload{
			Type: internal.CodeAutoconnectToSpecificServer,
		}, nil
	}

	if ok, err := r.ac.IsLoggedIn(); !ok {
		if errors.Is(err, core.ErrUnauthorized) {
			return nil, err
		}
		return nil, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error("configuration loading for auto-connect", err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	if !cfg.AutoConnect && !in.GetEnabled() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	var parameters serverpicker.ServerParameters
	if in.GetEnabled() {
		serverSelection, err := r.validateServerInput(in.ServerTag, in.ServerGroup, cfg)
		if err != nil {
			var errorCode *internal.ErrorWithCode
			if errors.As(err, &errorCode) {
				return &pb.Payload{
					Type: errorCode.Code,
				}, nil
			}

			return nil, err
		}

		log.Info("server for autoconnect found", serverSelection.Server)
		parameters = serverpicker.GetServerParameters(in.ServerTag, in.ServerGroup, r.dm.GetCountryData().Countries)
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnect = in.GetEnabled()
		c.AutoConnectData.ServerTag = in.ServerTag
		c.AutoConnectData.Country = parameters.Country
		c.AutoConnectData.CountryCode = parameters.CountryCode
		c.AutoConnectData.City = parameters.City
		c.AutoConnectData.Group = parameters.Group

		return c
	}); err != nil {
		log.Error("saving config for auto-connect", err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}
	r.events.Settings.Autoconnect.Publish(in.GetEnabled())

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
