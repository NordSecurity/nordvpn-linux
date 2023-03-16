package remote

import (
	"time"

	"github.com/coreos/go-semver/semver"
)

const (
	UpdatePeriod              = time.Hour * 24
	RcNatTraversalMinVerKey   = "nat_traversal_min_version"
	RcTelioAnalyticsMinVerKey = "telio_analytics_min_version"
	RcFileSharingMinVerKey    = "fileshare_min_version"
)

type SupportedVersionGetter interface {
	GetMinFeatureVersion(featureKey string) (*semver.Version, error)
}
