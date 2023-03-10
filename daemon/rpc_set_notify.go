package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetNotify(ctx context.Context, in *pb.SetNotifyRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if in.GetNotify() && cfg.UsersData.Notify[in.GetUid()] || !in.GetNotify() && !cfg.UsersData.Notify[in.GetUid()] {
		getBool := func(label bool) string {
			if label {
				return "enabled"
			}
			return "disabled"
		}
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{getBool(in.GetNotify())},
		}, nil
	}

	if in.GetNotify() {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.UsersData.Notify[in.GetUid()] = true
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			delete(c.UsersData.Notify, in.GetUid())
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	}

	r.events.Settings.Notify.Publish(in.GetNotify())

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
