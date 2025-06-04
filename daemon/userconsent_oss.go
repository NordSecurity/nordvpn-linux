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

func IsConsentFlowCompleted(cm config.Manager) (bool, error) {
	return true, nil
}

func (acc *AnalyticsConsentChecker) PrepareDaemonIfConsentNotCompleted() {
	// nothing to do on non-moose app
	return
}
