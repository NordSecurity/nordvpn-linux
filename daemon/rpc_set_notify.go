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

	notifyStatus := !cfg.UsersData.NotifyOff[in.GetUid()]

	if in.GetNotify() == notifyStatus {
		getBool := func(label bool) string {
			if label {
				return "enabled"
			}
			return "disabled"
		}
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{getBool(notifyStatus)},
		}, nil
	}

	if !in.GetNotify() {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.UsersData.NotifyOff[in.GetUid()] = true
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			delete(c.UsersData.NotifyOff, in.GetUid())
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
