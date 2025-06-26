//go:build moose

package main

import (
	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/events"
)

func newConsentChecker(
	isDevEnv bool,
	cm config.Manager,
	insightsAPI core.InsightsAPI,
	authChecker auth.Checker,
	analytics events.Analytics,
) daemon.ConsentChecker {
	return daemon.NewConsentChecker(isDevEnv, cm, insightsAPI, authChecker, analytics)
}
