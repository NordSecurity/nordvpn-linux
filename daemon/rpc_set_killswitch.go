package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetKillSwitch(ctx context.Context, in *pb.SetKillSwitchRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if !cfg.Firewall {
		return &pb.Payload{Type: internal.CodeDependencyError}, nil
	}

	if cfg.KillSwitch == in.GetKillSwitch() {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
		}, nil
	}

	if in.KillSwitch {
		allowlist := config.NewAllowlist(
			in.GetAllowlist().GetPorts().GetUdp(),
			in.GetAllowlist().GetPorts().GetTcp(),
			in.GetAllowlist().GetSubnets(),
		)

		if err := r.netw.SetKillSwitch(allowlist); err != nil {
			log.Println(internal.ErrorPrefix, "enabling killswitch:", err)
			return &pb.Payload{
				Type: internal.CodeKillSwitchError,
			}, nil
		}
	} else {
		if err := r.netw.UnsetKillSwitch(); err != nil {
			log.Println(internal.ErrorPrefix, "disabling killswitch:", err)
			return &pb.Payload{
				Type: internal.CodeKillSwitchError,
			}, nil
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.KillSwitch = in.GetKillSwitch()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.events.Settings.Killswitch.Publish(in.GetKillSwitch())

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
