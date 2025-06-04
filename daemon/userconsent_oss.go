//go:build !moose

package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

// AnalyticsConsentChecker here is a no-op stub struct used when there is no moose.
type AnalyticsConsentChecker struct{}

func NewConsentChecker(
	cm *config.FilesystemConfigManager,
	API *core.DefaultAPI,
	authChecker *auth.RenewingChecker,
) *AnalyticsConsentChecker {
	return &AnalyticsConsentChecker{}
}

// IsConsentFlowCompleted is a stub used when there is no moose.
func IsConsentFlowCompleted(_ config.Manager) bool {
	// consent is considered as always completed on non-moose app
	return true
}

// PrepareDaemonIfConsentNotCompleted is a no-op stub used when there is no moose.
func (acc *AnalyticsConsentChecker) PrepareDaemonIfConsentNotCompleted() {
	// nothing to do on non-moose app
	return
}
