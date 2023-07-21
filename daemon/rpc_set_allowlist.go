package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetAllowlist(ctx context.Context, in *pb.SetAllowlistRequest) (*pb.Payload, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	allowlist := config.NewAllowlist(
		in.GetAllowlist().GetPorts().GetUdp(),
		in.GetAllowlist().GetPorts().GetTcp(),
		in.GetAllowlist().GetSubnets(),
	)

	if r.netw.IsVPNActive() || cfg.KillSwitch {
		if err := r.netw.UnsetAllowlist(); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeFailure,
			}, nil
		}
		if err := r.netw.SetAllowlist(allowlist); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeFailure,
			}, nil
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Allowlist = allowlist
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.events.Settings.Allowlist.Publish(events.DataAllowlist{
		TCPPorts: len(in.Allowlist.Ports.Tcp),
		UDPPorts: len(in.Allowlist.Ports.Udp),
		Subnets:  len(in.Allowlist.Subnets),
	})

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
