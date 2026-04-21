package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetNotify(ctx context.Context, in *pb.SetNotifyRequest) (*pb.Payload, error) {
	cred, err := getCallerCred(ctx)
	if err != nil {
		log.Println(internal.ErrorPrefix, "SetNotify:", err)
		return &pb.Payload{Type: internal.CodeInternalError}, nil
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	notifyStatus := !cfg.UsersData.NotifyOff[int64(cred.Uid)]

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
			c.UsersData.NotifyOff[int64(cred.Uid)] = true
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			delete(c.UsersData.NotifyOff, int64(cred.Uid))
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
