//go:build moose

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events/moose"
)

var (
	EventsDomain    = ""
	EventsSubdomain = ""
)

func newAnalytics(eventsDbPath string, fs *config.Filesystem,
	ver, env, id string) *moose.Subscriber {
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
