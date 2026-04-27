//go:build !moose

package main

import (
	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	devicekey "github.com/NordSecurity/nordvpn-linux/device_key"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// AnalyticsConsentChecker here is a no-op stub struct used when there is no moose.
type NoOpConsentChecker struct{}

// PrepareDaemonIfConsentNotCompleted is a no-op stub used when there is no moose.
func (acc *NoOpConsentChecker) PrepareDaemonIfConsentNotCompleted() {
	// nothing to do on non-moose app
}

// IsConsentFlowCompleted is a stub used when there is no moose.
func (acc *NoOpConsentChecker) IsConsentFlowCompleted() bool {
	// consent is considered as always completed on non-moose app
	return true
}

func newConsentChecker(
	_ bool,
	_ config.Manager,
	_ core.InsightsAPI,
	_ auth.Checker,
	_ events.Analytics,
	_ devicekey.DeviceKeyInvalidator,
) daemon.ConsentChecker {
	return &NoOpConsentChecker{}
}
