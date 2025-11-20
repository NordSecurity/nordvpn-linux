package daemon

import (
	"context"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var callInitOnceGuard = sync.Once{}

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

	previousAnalyticsConsentState := cfg.AnalyticsConsent
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

	var err error = nil
	callInitOnceGuard.Do(func() {
		err = r.analytics.Init()
	})

	if err != nil {
		log.Println(internal.ErrorPrefix, "moose initialization failure:", err)
		return &pb.Payload{Type: internal.CodeInternalError}, nil
	}

	if in.GetEnabled() {
		if err := r.analytics.Enable(previousAnalyticsConsentState); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	} else {
		if err := r.analytics.Disable(previousAnalyticsConsentState); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return &pb.Payload{
				Type: internal.CodeConfigError,
			}, nil
		}
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
