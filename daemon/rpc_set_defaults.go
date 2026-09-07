package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetDefaults(ctx context.Context, in *pb.SetDefaultsRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	if _, err := r.DoDisconnect(); err != nil {
		log.Error("error while disconnecting:", err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	if in.OffKillswitch && cfg.KillSwitch {
		if err := r.netw.UnsetKillSwitch(); err != nil {
			log.Error("error while disabling killswitch:", err)
			return &pb.Payload{Type: internal.CodeFailure}, nil
		}
	}

	// No error check in case mesh isn't even turned on
	if err := r.netw.UnSetMesh(); err != nil {
		log.Warn(err)
	}

	if !in.NoLogout {
		result := access.Logout(access.LogoutInput{
			AuthChecker:                  r.ac,
			CredentialsAPI:               r.credentialsAPI,
			Netw:                         r.netw,
			NcClient:                     r.ncClient,
			ConfigManager:                r.cm,
			UserLogoutEventPublisherFunc: r.events.User.Logout.Publish,
			DebugPublisherFunc:           r.publisher.Publish,
			DisconnectFunc:               r.DoDisconnect,
			DeviceKeyInvalidator:         r.dedicatedServerKeyManager,
		})

		if result.Err != nil {
			switch result.Err {
			case internal.ErrNotLoggedIn:
				log.Info("trying to log out with set defaults, user already logged out")
			default:
				log.Error("error while trying to logout:", result.Err)
				return &pb.Payload{
					Type: internal.CodeFailure,
				}, nil
			}
		}

		switch result.Status {
		case internal.CodeSuccess, internal.CodeTokenInvalidated, 0:
			log.Info("set defaults logout successful")
		default:
			log.Error("logout returned non success return code", result.Status)
			return &pb.Payload{
				Type: result.Status,
			}, nil
		}
	}

	if err := r.cm.Reset(in.NoLogout, in.OffKillswitch); err != nil {
		log.Error(err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	if err := r.recentVPNConnStore.Clean(); err != nil {
		return &pb.Payload{
			Type: internal.CodeCleanRecentConnectionError,
		}, nil
	}

	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
	}

	v, err := r.factory(cfg.Technology)
	if err != nil {
		log.Error(err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.netw.SetVPN(v)

	if err = r.netw.SetARPIgnore(cfg.ARPIgnore.Get()); err != nil {
		log.Warn("resetting arp ignore failed:", err)
	}
	r.netw.SetLanDiscovery(cfg.LanDiscovery)
	if err = r.netw.SetAllowlist(cfg.AutoConnectData.Allowlist); err != nil {
		log.Warn("resetting allowlist failed:", err)
	}

	r.events.Settings.Defaults.Publish(nil)
	r.events.Settings.Publish(cfg)

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
