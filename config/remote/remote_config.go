package remote

import (
	"time"
)

const (
	UpdatePeriod              = time.Hour * 24
	RcNatTraversalMinVerKey   = "nat_traversal_min_version"
	RcTelioAnalyticsMinVerKey = "telio_analytics_min_version"
	RcFileSharingMinVerKey    = "fileshare_min_version"

	// Telio remote config has field with app version e.g. telio_config_3_16_2
	// (remote config field naming does not allow dots and dashes, only letters
	// and digits, underscores) app will try to find field corresponding to
	// app's version, but if exact match is not found then first older version
	// is chosen, if that one is also not available, then use local defaults.
	RcTelioConfigFieldPrefix  = "telio_config_"
	RcQuenchConfigFieldPrefix = "quench_enabled_"
)

// RemoteConfigGetter get values from remote config
type RemoteConfigGetter interface {
	GetTelioConfig(version string) (string, error)
	GetQuenchEnabled(version string) (bool, error)
}
