package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// SetFirewall controls whether firewall should be used by the app or not.
//
// This setting impacts the usage of these features:
// - Killswitch (impacts only next enabling)
// - Allowlist
// - Connect (impacts only connections, disconnect still works with the old setting)
func (r *RPC) SetFirewall(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	log.RPCSetFirewall.Tracef("enabled=%v", in.GetEnabled())
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.RPCSetFirewall.Error("loading config:", err)
	}

	if cfg.Firewall == in.GetEnabled() {
		log.RPCSetFirewall.Tracef("firewall already %v, nothing to do", cfg.Firewall)
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if cfg.KillSwitch && !in.GetEnabled() {
		log.RPCSetFirewall.Trace("kill switch is enabled, cannot disable firewall")
		return &pb.Payload{Type: internal.CodeDependencyError}, nil
	}

	log.RPCSetFirewall.Tracef("applying to networker: currentFirewall=%v killSwitch=%v", cfg.Firewall, cfg.KillSwitch)
	if in.GetEnabled() {
		if err := r.netw.EnableFirewall(); err != nil {
			log.RPCSetFirewall.Error("enabling firewall in networker:", err)
			return &pb.Payload{Type: internal.CodeFailure}, nil
		}
	} else {
		if err := r.netw.DisableFirewall(); err != nil {
			log.RPCSetFirewall.Error("disabling firewall in networker:", err)
			return &pb.Payload{Type: internal.CodeFailure}, nil
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.Firewall = in.GetEnabled()
		return c
	}); err != nil {
		log.RPCSetFirewall.Error("saving firewall config:", err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}
	r.events.Settings.Firewall.Publish(in.GetEnabled())

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}

func (r *RPC) SetFirewallMark(ctx context.Context, in *pb.SetUint32Request) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.RPCSetFirewall.Error("loading config:", err)
	}

	if cfg.FirewallMark == in.GetValue() {
		return &pb.Payload{Type: internal.CodeNothingToDo}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.FirewallMark = in.GetValue()
		return c
	}); err != nil {
		log.RPCSetFirewall.Error("saving firewall mark config:", err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}
	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
