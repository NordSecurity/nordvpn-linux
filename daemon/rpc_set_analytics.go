package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// SetAnalytics
func (r *RPC) SetAnalytics(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	if cfg.AnalyticsConsent != config.ConsentUndefined {
		enabled := cfg.AnalyticsConsent == config.ConsentGranted
		if enabled == in.GetEnabled() {
			return &pb.Payload{Type: internal.CodeNothingToDo}, nil
		}
	}

	// moose requires consent status updated in the config before triggering initialization
	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		if in.GetEnabled() {
			c.AnalyticsConsent = config.ConsentGranted
		} else {
			c.AnalyticsConsent = config.ConsentDenied
		}
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	if err := r.analytics.Init(); err != nil {
		log.Println(internal.ErrorPrefix, "moose failed to initialize with error:", err)
	}
	if in.GetEnabled() {
		if err := r.analytics.Enable(); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.analytics.Disable(); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
