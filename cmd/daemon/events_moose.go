//go:build moose

package main

import (
	"os"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events/moose"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	EventsDomain    = ""
	EventsSubdomain = ""
)

func newAnalytics(eventsDbPath string, fs *config.FilesystemConfigManager,
	ver, env, id string) *moose.Subscriber {
	_ = os.Setenv("MOOSE_LOG_FILE", "Stdout")
	logLevel := "error"
	if !internal.IsProdEnv(env) {
		logLevel = "debug"
	}
	_ = os.Setenv("MOOSE_LOG", logLevel)
	return &moose.Subscriber{
		EventsDbPath: eventsDbPath,
		Config:       fs,
		Version:      ver,
		Environment:  env,
		Domain:       EventsDomain,
		Subdomain:    EventsSubdomain,
		DeviceID:     id,
	}
}
