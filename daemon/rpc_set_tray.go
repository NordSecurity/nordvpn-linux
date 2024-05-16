package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetTray(ctx context.Context, in *pb.SetTrayRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	trayStatus := !cfg.UsersData.TrayOff[in.GetUid()]

	if in.GetTray() == trayStatus {
		getBool := func(label bool) string {
			if label {
				return "enabled"
			}
			return "disabled"
		}
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{getBool(trayStatus)},
		}, nil
	}

	if !in.GetTray() {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.UsersData.TrayOff[in.GetUid()] = true
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			delete(c.UsersData.TrayOff, in.GetUid())
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	}

	if err := r.norduser.Restart(uint32(in.GetUid())); err != nil {
		log.Println(internal.ErrorPrefix, "Cannot restart norduserd", err)
	}

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
