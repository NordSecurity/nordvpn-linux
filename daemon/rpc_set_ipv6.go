package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SetIpv6 controls whether ipv6 usage should be allowed
func (r *RPC) SetIpv6(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.IPv6 == in.Enabled {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.IPv6 = in.Enabled
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	r.events.Settings.Ipv6.Publish(in.GetEnabled())

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
