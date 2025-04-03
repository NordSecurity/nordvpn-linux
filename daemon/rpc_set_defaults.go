package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetDefaults(ctx context.Context, in *pb.Empty) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	if _, err := r.DoDisconnect(); err != nil {
		log.Println(internal.ErrorPrefix, "Error while disconnecting:", err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	// No error check in case mesh isn't even turned on
	if err := r.netw.UnSetMesh(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	if err := r.ncClient.Stop(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	if !r.ncClient.Revoke() {
		log.Println(internal.WarningPrefix, "error revoking token")
	}

	if err := r.cm.Reset(); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	v, err := r.factory(cfg.Technology)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.netw.SetVPN(v)

	r.events.Settings.Defaults.Publish(nil)
	r.events.Settings.Publish(cfg)

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
