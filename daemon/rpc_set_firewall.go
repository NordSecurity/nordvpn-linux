package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SetFirewall controls whether firewall should be used by the app or not.
//
// This setting impacts the usage of these features:
// - Killswitch (impacts only next enabling)
// - Allowlist
// - Connect (impacts only connections, disconnect still works with the old setting)
func (r *RPC) SetFirewall(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.Firewall == in.GetEnabled() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if cfg.KillSwitch && !in.GetEnabled() {
		return &pb.Payload{Type: internal.CodeDependencyError}, nil
	}

	if in.GetEnabled() {
		if err := r.netw.EnableFirewall(); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{Type: internal.CodeFailure}, nil
		}
	} else {
		if err := r.netw.DisableFirewall(); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{Type: internal.CodeFailure}, nil
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.Firewall = in.GetEnabled()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}
	r.events.Settings.Firewall.Publish(in.GetEnabled())

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}

func (r *RPC) SetFirewallMark(ctx context.Context, in *pb.SetUint32Request) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.FirewallMark == in.GetValue() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.FirewallMark = in.GetValue()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}
	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
