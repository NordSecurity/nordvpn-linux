package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetWhitelist(ctx context.Context, in *pb.SetWhitelistRequest) (*pb.Payload, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	whitelist := config.NewWhitelist(
		in.GetWhitelist().GetPorts().GetUdp(),
		in.GetWhitelist().GetPorts().GetTcp(),
		in.GetWhitelist().GetSubnets(),
	)

	if r.netw.IsVPNActive() || cfg.KillSwitch {
		if err := r.netw.UnsetWhitelist(); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeFailure,
			}, nil
		}
		if err := r.netw.SetWhitelist(whitelist); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeFailure,
			}, nil
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Whitelist = whitelist
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.events.Settings.Whitelist.Publish(events.DataWhitelist{
		TCPPorts: len(in.Whitelist.Ports.Tcp),
		UDPPorts: len(in.Whitelist.Ports.Udp),
		Subnets:  len(in.Whitelist.Subnets),
	})

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
