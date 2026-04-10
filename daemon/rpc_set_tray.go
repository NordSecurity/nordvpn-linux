package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"google.golang.org/grpc/peer"
)

func (r *RPC) SetTray(ctx context.Context, in *pb.SetTrayRequest) (*pb.Payload, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		log.Println(internal.ErrorPrefix, "failed to retrieve gRPC peer information from the context")
		return &pb.Payload{
			Type: internal.CodeInternalError,
		}, nil
	}

	cred, ok := peer.AuthInfo.(internal.UcredAuth)
	if !ok {
		log.Println(internal.ErrorPrefix, "failed to extract ucred out of gRPC peer info")
		return &pb.Payload{
			Type: internal.CodeInternalError,
		}, nil
	}

	uid := int64(cred.Uid)

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	trayStatus := !cfg.UsersData.TrayOff[uid]

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
			c.UsersData.TrayOff[uid] = true
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.cm.SaveWith(func(c config.Config) config.Config {
			delete(c.UsersData.TrayOff, uid)
			return c
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	}

	if err := r.norduser.Restart(cred.Uid); err != nil {
		log.Println(internal.ErrorPrefix, "Cannot restart norduserd", err)
	}

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
