package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SetFileshareStatusNotifications enables pop-up notifications for fileshare transfer status
func (r *RPC) SetFileshareStatusNotifications(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.FileshareStatusNotifications == in.Enabled {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.FileshareStatusNotifications = in.Enabled
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
