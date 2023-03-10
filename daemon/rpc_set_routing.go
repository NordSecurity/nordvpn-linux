package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SetRouting controls whether routing should be used by the app or not.
//
// This setting impacts the usage of these features:
// - Whitelist
// - Connect
// - Meshnet
func (r *RPC) SetRouting(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.Routing.Get() == in.GetEnabled() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if cfg.Mesh && !in.GetEnabled() {
		return &pb.Payload{Type: internal.CodeDependencyError}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.Routing.Set(in.GetEnabled())
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	if in.GetEnabled() {
		r.netw.EnableRouting()
	} else {
		r.netw.DisableRouting()
	}
	r.events.Settings.Routing.Publish(in.GetEnabled())

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
