package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
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
		if err := r.ncClient.Stop(); err != nil {
			log.Warn("error stoping notification center client:", err)
		}

		if !r.ncClient.Revoke() {
			log.Warn("error revoking notification center token")
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
	_ = r.netw.SetARPIgnore(cfg.ARPIgnore.Get())

	r.events.Settings.Defaults.Publish(nil)
	r.events.Settings.Publish(cfg)

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
