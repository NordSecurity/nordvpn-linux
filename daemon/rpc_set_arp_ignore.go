package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetARPIgnore(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, "failed to load config:", err)

		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	if cfg.ARPIgnore.Get() == in.Enabled {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
		}, nil
	}

	if err := r.netw.SetARPIgnore(in.Enabled); err != nil {
		log.Println(internal.ErrorPrefix, "failed to set ARP ignore:", err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.ARPIgnore.Set(in.Enabled)
		return c
	})
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to save config:", err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
