package daemon

import (
	"errors"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func selectServer(
	r *RPC,
	insights *core.Insights,
	cfg config.Config,
	tag string,
	groupFlag string,
) (serverpicker.ServerSelection, error) {
	serversList := r.dm.GetServersData().Servers
	searchParams := serverpicker.NewSearchParams(tag, groupFlag)
	selection, err := serverpicker.PickServer(
		r.serversAPI,
		serversList,
		r.dm.GetCountryData().Countries,
		*insights,
		cfg,
		searchParams,
	)

	if err != nil {
		if !errors.Is(err, serverpicker.ErrDedicatedIPServer) && !errors.Is(err, serverpicker.ErrDedicatedServer) {
			log.Error("picking servers error", err)
		}
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			if err := r.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, r.events.User.Logout, events.ReasonUnauthorized)); err != nil {
				return serverpicker.ServerSelection{}, err
			}
			return serverpicker.ServerSelection{}, internal.ErrNotLoggedIn
		case errors.Is(err, internal.ErrTagDoesNotExist),
			errors.Is(err, internal.ErrGroupDoesNotExist),
			errors.Is(err, internal.ErrServerIsUnavailable),
			errors.Is(err, internal.ErrDoubleGroup),
			errors.Is(err, internal.ErrVirtualServerSelected):
			return serverpicker.ServerSelection{}, err

		case errors.Is(err, serverpicker.ErrDedicatedIPServer):
			return serverpicker.SelectDedicatedIPServer(r.ac, serversList, cfg)

		case errors.Is(err, serverpicker.ErrDedicatedServer):
			return serverpicker.SelectDedicatedServer(
				r.ac, r.dedicatedServersAPI,
				r.dedicatedServerKeyManager,
			)

		default:
			return serverpicker.ServerSelection{}, internal.ErrUnhandled
		}
	}

	log.Info("server", selection.Server.Hostname, "remote", selection.Remote)

	if core.IsDedicatedIP(*selection.Server) {
		if err := serverpicker.CheckDIPServerInSubscription(r.ac, *selection.Server, cfg); err != nil {
			return serverpicker.ServerSelection{}, err
		}
	}

	return selection, nil
}
