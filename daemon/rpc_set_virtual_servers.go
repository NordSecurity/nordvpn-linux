package daemon

import (
	"context"
	"log"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetVirtualLocation(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.VirtualLocation.Get() == in.Enabled {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.VirtualLocation.Set(in.Enabled)
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	r.events.Settings.VirtualLocation.Publish(in.Enabled)

	payload := &pb.Payload{}
	payload.Type = internal.CodeSuccess
	payload.Data = []string{strconv.FormatBool(r.netw.IsVPNActive())}
	return payload, nil
}
