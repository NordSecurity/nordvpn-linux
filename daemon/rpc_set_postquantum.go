package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetPostQuantum(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.AutoConnectData.PostquantumVpn == in.GetEnabled() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if cfg.Mesh && in.GetEnabled() {
		return &pb.Payload{Type: internal.CodePqAndMeshnetSimultaneously}, nil
	}

	if cfg.Technology != config.Technology_NORDLYNX {
		return &pb.Payload{Type: internal.CodePqWitoughNordlynx}, nil
	}

	if cfg.AutoConnect && in.GetEnabled() {
		//TODO(msz): validate that server supports PQ
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.PostquantumVpn = in.GetEnabled()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	r.events.Settings.PostquantumVPN.Publish(in.GetEnabled())

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
